package engine

import (
	"math"
	"strings"
	"testing"

	"github.com/Seb0sti0n/Ohmline/backend/internal/config"
)

func runReal(t *testing.T, mutate func([]Event) []Event) map[string]AnomalyResult {
	t.Helper()
	by, events := loadFixture(t)
	if mutate != nil {
		events = mutate(events)
	}
	out := map[string]AnomalyResult{}
	for _, r := range Run(flatten(by), events, config.DefaultEngine()) {
		out[r.MeterID] = r
	}
	return out
}

func TestRunOnRealDataset(t *testing.T) {
	by, events := loadFixture(t)
	results := Run(flatten(by), events, config.DefaultEngine())

	want := []struct {
		meter string
		typ   AnomalyType
		sev   Severity
	}{
		{"M-109", RealAnomaly, High},
		{"M-112", DataQuality, High},
		{"M-104", ExplainableAnomaly, Medium},
		{"M-106", FalsePositive, Low},
	}
	if len(results) != len(want) {
		t.Fatalf("got %d anomalies, want %d (healthy meters must not produce any): %+v", len(results), len(want), ids(results))
	}
	// Order is the priority order: M-109 > M-112 > M-104 > M-106.
	for i, w := range want {
		r := results[i]
		if r.MeterID != w.meter || r.Type != w.typ || r.Severity != w.sev {
			t.Errorf("position %d: got %s %s %s, want %s %s %s", i+1, r.MeterID, r.Type, r.Severity, w.meter, w.typ, w.sev)
		}
		if !r.Anomaly || r.Confidence <= 0 || r.Confidence > 0.99 {
			t.Errorf("%s: anomaly=%v confidence=%v", r.MeterID, r.Anomaly, r.Confidence)
		}
		if r.Reason == "" || r.Explanation == "" || r.RecommendedAction == "" || len(r.Evidence.RulesFired) == 0 {
			t.Errorf("%s: missing explanation or evidence", r.MeterID)
		}
		if len(r.Evidence.PriorityBreakdown) == 0 || len(r.Evidence.ConfidenceBreakdown) == 0 {
			t.Errorf("%s: missing score breakdowns", r.MeterID)
		}
	}
	for i := 1; i < len(results); i++ {
		if results[i].PriorityScore > results[i-1].PriorityScore {
			t.Errorf("results not sorted by priority: %v", ids(results))
		}
	}
	if fp := results[3]; fp.PriorityScore >= 20 {
		t.Errorf("FALSE_POSITIVE priority %.1f must be below 20", fp.PriorityScore)
	}
}

func TestRealDatasetNumbers(t *testing.T) {
	r := runReal(t, nil)
	m109 := r["M-109"]
	if v := m109.Evidence.Metrics.LastDayKWh; v < 2200 || v > 2215 {
		t.Errorf("M-109 last day = %.1f kWh, want ≈ 2.208", v)
	}
	if v := m109.Evidence.Metrics.LastDayVariation; v < 105 || v > 115 {
		t.Errorf("M-109 last day variation = %.1f%%, want ≈ +110", v)
	}
	if got := m109.WindowStart.Format("2006-01-02 15:04"); got != "2026-09-12 14:00" {
		t.Errorf("M-109 window starts %s, want 2026-09-12 14:00", got)
	}
	pf, _ := changeOf(m109.Evidence.ChangedVariables, VarPowerFactor)
	if !pf.Significant || pf.Delta > -0.15 {
		t.Errorf("M-109 power factor change not detected: %+v", pf)
	}
	if h := r["M-106"].Evidence.Window.Hours; h != 12 {
		t.Errorf("M-106 outage window = %d h, want 12", h)
	}
}

// The event type must never be used as a label: the engine has to reach the same conclusion
// from the readings, and only use events as corroborating or explaining evidence.
func TestDataQualityIsIndependentOfEventLabel(t *testing.T) {
	relabel := func(typ string, drop bool) func([]Event) []Event {
		return func(events []Event) []Event {
			var out []Event
			for _, e := range events {
				if e.MeterID == "M-112" {
					if drop {
						continue
					}
					e.Type = typ
				}
				out = append(out, e)
			}
			return out
		}
	}
	tests := []struct {
		name   string
		mutate func([]Event) []Event
	}{
		{"event relabeled UNKNOWN", relabel(EventUnknown, false)},
		{"event relabeled OPERATIONAL_CHANGE", relabel(EventOperationalChange, false)},
		{"event removed", relabel("", true)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, ok := runReal(t, tt.mutate)["M-112"]
			if !ok || r.Type != DataQuality || r.Severity != High {
				t.Fatalf("M-112 = %+v, want DATA_QUALITY/HIGH", r.Type)
			}
			if tt.name == "event removed" && len(r.Evidence.Events) != 0 {
				t.Errorf("no event should be cited, got %v", r.Evidence.Events)
			}
		})
	}
}

// An event only explains a change if it is compatible with it, not merely because it is close in time.
func TestEventsMustBeCompatibleToExplain(t *testing.T) {
	retype := func(meter, typ string) func([]Event) []Event {
		return func(events []Event) []Event {
			for i := range events {
				if events[i].MeterID == meter {
					events[i].Type = typ
				}
			}
			return events
		}
	}
	tests := []struct {
		name   string
		meter  string
		typ    string
		wantTy AnomalyType
	}{
		{"M-106 without a scheduled outage is a real anomaly", "M-106", EventUnknown, RealAnomaly},
		{"M-104 without an operational change is a real anomaly", "M-104", EventUnknown, RealAnomaly},
		{"an operational change does not explain M-109's electrical deterioration", "M-109", EventOperationalChange, RealAnomaly},
		{"an outage does not explain an increase", "M-109", EventScheduledOutage, RealAnomaly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := runReal(t, retype(tt.meter, tt.typ))[tt.meter]
			if r.Type != tt.wantTy {
				t.Errorf("%s = %s, want %s", tt.meter, r.Type, tt.wantTy)
			}
		})
	}
}

func TestRunWithNoData(t *testing.T) {
	if got := Run(nil, nil, config.DefaultEngine()); len(got) != 0 {
		t.Errorf("got %d results for empty input", len(got))
	}
}

func ids(rs []AnomalyResult) []string {
	var out []string
	for _, r := range rs {
		out = append(out, r.MeterID)
	}
	return out
}

// The evidence publishes each score component's maximum, so the UI can draw breakdowns as fractions.
func TestBreakdownsComeWithTheirWeights(t *testing.T) {
	for meter, r := range runReal(t, nil) {
		e := r.Evidence
		if len(e.ConfidenceWeights) == 0 || len(e.PriorityWeights) == 0 {
			t.Fatalf("%s: weights missing from the evidence", meter)
		}
		for name, weights := range map[string]struct{ parts, max map[string]float64 }{
			"confidence": {e.ConfidenceBreakdown, e.ConfidenceWeights},
			"priority":   {e.PriorityBreakdown, e.PriorityWeights},
		} {
			for k, v := range weights.parts {
				if k == "capped_false_positive" {
					continue
				}
				max, ok := weights.max[k]
				if !ok {
					t.Errorf("%s %s: component %q has no weight", meter, name, k)
				} else if v < 0 || v > max+0.001 {
					t.Errorf("%s %s: %q = %v is outside 0..%v", meter, name, k, v, max)
				}
			}
		}
		// The confidence is exactly the sum of its parts, capped.
		sum := 0.0
		for _, v := range e.ConfidenceBreakdown {
			sum += v
		}
		if want := math.Min(sum, 0.99); math.Abs(r.Confidence-round(want, 2)) > 0.011 {
			t.Errorf("%s: confidence %v does not match its breakdown (%v)", meter, r.Confidence, sum)
		}
	}
}

// The evidence is shown to operators as is: it must not leak internal identifiers.
func TestEvidenceTextsAreForOperators(t *testing.T) {
	for meter, r := range runReal(t, nil) {
		for _, rule := range r.Evidence.RulesFired {
			for _, raw := range []string{"OPERATIONAL_CHANGE", "SCHEDULED_OUTAGE", "DATA_QUALITY", "UNKNOWN", "flags"} {
				if strings.Contains(rule, raw) {
					t.Errorf("%s: rule %q contains %q", meter, rule, raw)
				}
			}
		}
	}
}

// Isolated spikes (short runs outside the band) are detected but are noise: the eight healthy meters have
// plenty of them (M-107 has 23) and must still produce no anomaly.
func TestIsolatedSpikesAreNotAnomalies(t *testing.T) {
	by, events := loadFixture(t)
	cfg := config.DefaultEngine()
	results := map[string]bool{}
	for _, r := range Run(flatten(by), events, cfg) {
		results[r.MeterID] = true
	}
	for _, id := range []string{"M-101", "M-102", "M-103", "M-105", "M-107", "M-108", "M-110", "M-111"} {
		rs := by[id]
		_, spikes := DetectConsumption(rs, BuildBaseline(rs, cfg), cfg)
		if spikes == 0 {
			t.Errorf("%s: expected some isolated spikes in the real data, found none", id)
		}
		if results[id] {
			t.Errorf("%s has %d isolated spikes and must not be reported as an anomaly", id, spikes)
		}
	}
}
