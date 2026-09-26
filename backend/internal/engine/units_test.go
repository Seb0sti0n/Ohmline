package engine

import (
	"math"
	"testing"
	"time"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
)

var t0 = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

// series builds hourly readings for `days` days. f returns kWh, voltage, current, pf for the index.
func series(days int, f func(i int, hour int) (kwh, v, a, pf float64)) []Reading {
	var rs []Reading
	for i := 0; i < days*24; i++ {
		ts := t0.Add(time.Duration(i) * time.Hour)
		k, v, a, pf := f(i, ts.Hour())
		rs = append(rs, Reading{"M-T", ts, k, v, a, pf})
	}
	return rs
}

// steady has a small deterministic day-to-day wobble so the MAD is not zero.
func steady(i, hour int) (float64, float64, float64, float64) {
	day := i / 24
	kwh := 10 + 0.1*float64(day%3)
	return kwh, 220, kwh * 4.5, 0.95
}

func TestBaselineByHour(t *testing.T) {
	cfg := config.DefaultEngine()
	rs := series(10, func(i, hour int) (float64, float64, float64, float64) {
		k := 5.0
		if hour >= 8 && hour < 18 {
			k = 20
		}
		return k + 0.1*float64((i/24)%3), 220, 100, 0.95
	})
	b := BuildBaseline(rs, cfg)
	if got := b.Consumption[3].Median; math.Abs(got-5.1) > 1e-9 {
		t.Errorf("night median = %v, want 5.1", got)
	}
	if got := b.Consumption[12].Median; math.Abs(got-20.1) > 1e-9 {
		t.Errorf("day median = %v, want 20.1", got)
	}
	if b.Consumption[3].MAD <= 0 {
		t.Errorf("MAD should be positive")
	}
	wantDaily := 14*5.1 + 10*20.1
	if math.Abs(b.DailyKWh-wantDaily) > 1e-6 {
		t.Errorf("daily baseline = %v, want %v", b.DailyKWh, wantDaily)
	}
}

func TestBaselineUsesOnlyReferenceWindow(t *testing.T) {
	cfg := config.DefaultEngine()
	rs := series(14, func(i, hour int) (float64, float64, float64, float64) {
		k, v, a, pf := steady(i, hour)
		if i >= 8*24 { // the second week is far higher and must not leak into the baseline
			k *= 10
		}
		return k, v, a, pf
	})
	if got := BuildBaseline(rs, cfg).Consumption[0].Median; got > 11 {
		t.Errorf("baseline leaked later data: %v", got)
	}
}

func TestPersistence(t *testing.T) {
	cfg := config.DefaultEngine()
	// jump lifts consumption by +5 kWh for the given hours of day 10 (after the baseline week).
	withJump := func(from, to int, dir float64) []Reading {
		return series(14, func(i, hour int) (float64, float64, float64, float64) {
			k, v, a, pf := steady(i, hour)
			if i >= 10*24+from && i < 10*24+to {
				k += 5 * dir
			}
			return k, v, a, pf
		})
	}
	tests := []struct {
		name        string
		rs          []Reading
		wantWindows int
		wantSpikes  int
		wantHours   int
		wantDir     int
	}{
		{"5 h is a spike, not a change", withJump(2, 7, 1), 0, 1, 0, 0},
		{"6 h is a persistent change", withJump(2, 8, 1), 1, 0, 6, 1},
		{"drop is detected with negative direction", withJump(2, 12, -1), 1, 0, 10, -1},
		{"nothing happens", withJump(0, 0, 1), 0, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BuildBaseline(tt.rs, cfg)
			ws, spikes := DetectConsumption(tt.rs, b, cfg)
			if len(ws) != tt.wantWindows || spikes != tt.wantSpikes {
				t.Fatalf("windows=%d spikes=%d, want %d and %d", len(ws), spikes, tt.wantWindows, tt.wantSpikes)
			}
			if tt.wantWindows == 1 && (ws[0].Hours != tt.wantHours || ws[0].Direction != tt.wantDir) {
				t.Errorf("window = %+v", ws[0])
			}
		})
	}
}

func TestOpposingDirectionsDoNotMerge(t *testing.T) {
	cfg := config.DefaultEngine()
	rs := series(14, func(i, hour int) (float64, float64, float64, float64) {
		k, v, a, pf := steady(i, hour)
		if i >= 10*24 && i < 10*24+4 {
			k += 5
		} else if i >= 10*24+4 && i < 10*24+8 {
			k -= 5
		}
		return k, v, a, pf
	})
	b := BuildBaseline(rs, cfg)
	if ws, spikes := DetectConsumption(rs, b, cfg); len(ws) != 0 || spikes != 2 {
		t.Errorf("windows=%d spikes=%d, want 0 and 2 (two 4 h runs in opposite directions)", len(ws), spikes)
	}
}

func TestQualityFlags(t *testing.T) {
	cfg := config.DefaultEngine()
	// bad(i) decides which hours have erratic voltage (alternating 240 / 200 V).
	build := func(bad func(i int) bool) []Reading {
		return series(14, func(i, hour int) (float64, float64, float64, float64) {
			k, v, a, pf := steady(i, hour)
			if bad(i) {
				v = 240
				if i%2 == 0 {
					v = 200
				}
			}
			return k, v, a, pf
		})
	}
	tests := []struct {
		name          string
		bad           func(i int) bool
		wantRecurrent bool
		wantFlagged   bool
	}{
		{"healthy meter has no flags", func(i int) bool { return false }, false, false},
		{"one isolated bad reading is not recurrent", func(i int) bool { return i == 200 }, false, true},
		{"3 bad readings (4 flagged hours with the recovery jump) are not recurrent", func(i int) bool { return i >= 200 && i < 203 }, false, true},
		{"4 bad readings (5 flagged hours with the recovery jump) are recurrent", func(i int) bool { return i >= 200 && i < 204 }, true, true},
		{"8 bad readings in a day are recurrent", func(i int) bool { return i >= 200 && i < 208 }, true, true},
		{"bad readings spread over many days are not recurrent", func(i int) bool { return i%30 == 0 && i > 100 }, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := build(tt.bad)
			q := DetectQuality(rs, BuildBaseline(rs, cfg), cfg)
			if q.Info.Recurrent != tt.wantRecurrent {
				t.Errorf("recurrent = %v, want %v (flags=%d)", q.Info.Recurrent, tt.wantRecurrent, q.Info.TotalFlags)
			}
			if (len(q.Flagged) > 0) != tt.wantFlagged {
				t.Errorf("flagged = %d hours", len(q.Flagged))
			}
		})
	}
}

func TestQualityFlagKinds(t *testing.T) {
	cfg := config.DefaultEngine()
	rs := series(10, steady)
	rs[200].PowerFactor = 0.60    // jump of 0.35 from 0.95, and next hour jumps back
	rs[220].VoltageV = 232        // outside ±5% of 220 (and a jump of 12 V)
	rs[240-1].ConsumptionKWh *= 2 // kWh no longer matches V·I·PF
	q := DetectQuality(rs, BuildBaseline(rs, cfg), cfg)
	for _, kind := range []string{FlagPFJump, FlagVoltageRange, FlagVoltageJump, FlagCoherence} {
		if q.Info.ByKind[kind] == 0 {
			t.Errorf("expected a %s flag, got %v", kind, q.Info.ByKind)
		}
	}
}

func TestOutageHours(t *testing.T) {
	tests := []struct {
		desc string
		want int
		ok   bool
	}{
		{"Scheduled maintenance outage for 12 hours", 12, true},
		{"planned stop of 8 hrs", 8, true},
		{"outage for 6h", 6, true},
		{"Scheduled maintenance outage", 0, false},
	}
	for _, tt := range tests {
		got, ok := outageHours(tt.desc)
		if got != tt.want || ok != tt.ok {
			t.Errorf("outageHours(%q) = %d,%v want %d,%v", tt.desc, got, ok, tt.want, tt.ok)
		}
	}
}

func TestEventMatcher(t *testing.T) {
	cfg := config.DefaultEngine()
	start := t0.Add(10 * 24 * time.Hour)
	ev := func(typ, desc string, offset time.Duration) []Event {
		return []Event{{"M-T", start.Add(offset), typ, desc}}
	}
	drop12 := Window{Hours: 12, Direction: -1, ReturnsToBaseline: true}
	step := Window{Hours: 60, Direction: 1, Ongoing: true}

	tests := []struct {
		name         string
		events       []Event
		w            Window
		deteriorated bool
		want         int  // number of matches (events near the window)
		compatible   bool // first match compatible
	}{
		{"outage of the same length that recovers", ev(EventScheduledOutage, "outage for 12 hours", 0), drop12, false, 1, true},
		{"outage 2 h off is still compatible", ev(EventScheduledOutage, "outage for 10 hours", 0), drop12, false, 1, true},
		{"outage of a very different length", ev(EventScheduledOutage, "outage for 4 hours", 0), drop12, false, 1, false},
		{"outage without duration in the description", ev(EventScheduledOutage, "maintenance", 0), drop12, false, 1, false},
		{"outage but consumption never recovers", ev(EventScheduledOutage, "outage for 12 hours", 0), Window{Hours: 12, Direction: -1, Ongoing: true}, false, 1, false},
		{"outage cannot explain an increase", ev(EventScheduledOutage, "outage for 12 hours", 0), step, false, 1, false},
		{"operational change explains a clean step", ev(EventOperationalChange, "new line", 0), step, false, 1, true},
		{"operational change with electrical deterioration", ev(EventOperationalChange, "new line", 0), step, true, 1, false},
		{"UNKNOWN never explains", ev(EventUnknown, "nothing reported", 0), step, false, 1, false},
		{"event 6 h before the window counts", ev(EventOperationalChange, "new line", -6*time.Hour), step, false, 1, true},
		{"event 7 h before is too early", ev(EventOperationalChange, "new line", -7*time.Hour), step, false, 0, false},
		{"event 3 h after the start is too late", ev(EventOperationalChange, "new line", 3*time.Hour), step, false, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchConsumptionEvents(tt.events, tt.w, start, tt.deteriorated, cfg)
			if len(got) != tt.want {
				t.Fatalf("got %d matches, want %d", len(got), tt.want)
			}
			if tt.want > 0 && got[0].Compatible != tt.compatible {
				t.Errorf("compatible = %v (%s), want %v", got[0].Compatible, got[0].Reason, tt.compatible)
			}
		})
	}
}

func TestPriorityScore(t *testing.T) {
	strong := scoreInput{Type: RealAnomaly, Magnitude: 110, Hours: 58, Ongoing: true, Electrical: true, Unexplained: true, ExtraKWh: 2800}
	explained := scoreInput{Type: ExplainableAnomaly, Magnitude: 47, Hours: 96, Ongoing: true, ExtraKWh: 2100}
	outage := scoreInput{Type: FalsePositive, Magnitude: 80, Hours: 12, ExtraKWh: -500}

	s, _ := priorityScore(strong)
	e, _ := priorityScore(explained)
	o, parts := priorityScore(outage)
	if !(s > e && e > o) {
		t.Errorf("want strong > explained > outage, got %v %v %v", s, e, o)
	}
	if s > 100 || o >= 20 {
		t.Errorf("bounds violated: strong=%v outage=%v", s, o)
	}
	// Even a huge, unexplained, deteriorated outage-typed finding stays below 20.
	huge := strong
	huge.Type = FalsePositive
	if v, _ := priorityScore(huge); v >= 20 {
		t.Errorf("FALSE_POSITIVE must be forced below 20, got %v", v)
	}
	_ = parts
}

func TestConfidenceIsCapped(t *testing.T) {
	c, _ := confidence(confidenceInput{Strength: 10, Signals: 10, Clarity: 10})
	if c > 0.99 {
		t.Errorf("confidence %v above cap", c)
	}
	low, _ := confidence(confidenceInput{})
	if low < 0.4 || low >= c {
		t.Errorf("confidence range wrong: low=%v high=%v", low, c)
	}
}

func TestNumberFormatting(t *testing.T) {
	tests := []struct {
		got, want string
	}{
		{num(2825, 0), "2.825"},
		{num(1047.66, 1), "1.047,7"},
		{num(110.54, 1), "110,5"},
		{num(-2.9, 1), "-2,9"},
		{num(0.94, 2), "0,94"},
		{num(1234567.8, 1), "1.234.567,8"},
		{signedPct(110.5), "+110,5%"},
		{signedPct(-1.4), "-1,4%"},
		{when(time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC)), "12 sep, 14:00"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("got %q, want %q", tt.got, tt.want)
		}
	}
}
