package http

import (
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Seb0sti0n/astrophage/backend/internal/engine"
)

// readingPoint is one point of a meter series with the expected (baseline) values for the same
// hour, so the charts can draw the band without recomputing anything.
type readingPoint struct {
	Timestamp           time.Time `json:"timestamp"`
	ConsumptionKWh      float64   `json:"consumption_kwh"`
	VoltageV            float64   `json:"voltage_v"`
	CurrentA            float64   `json:"current_a"`
	PowerFactor         float64   `json:"power_factor"`
	BaselineKWh         float64   `json:"baseline_kwh"`
	BandLowKWh          float64   `json:"band_low_kwh"`  // baseline − 3·MAD
	BandHighKWh         float64   `json:"band_high_kwh"` // baseline + 3·MAD
	BaselineVoltageV    float64   `json:"baseline_voltage_v"`
	BaselineCurrentA    float64   `json:"baseline_current_a"`
	BaselinePowerFactor float64   `json:"baseline_power_factor"`
}

const bandMADs = 3

func parseTime(v string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.UTC(), nil
	}
	return time.ParseInLocation("2006-01-02", v, time.UTC)
}

func (s *Server) getReadings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "meterId")
	q := r.URL.Query()

	gran := q.Get("granularity")
	if gran == "" {
		gran = "hour"
	}
	if gran != "hour" && gran != "day" {
		writeError(w, http.StatusBadRequest, "granularity must be hour or day")
		return
	}
	var from, to time.Time
	var err error
	if v := q.Get("from"); v != "" {
		if from, err = parseTime(v); err != nil {
			writeError(w, http.StatusBadRequest, "from must be RFC3339 or YYYY-MM-DD")
			return
		}
	}
	if v := q.Get("to"); v != "" {
		if to, err = parseTime(v); err != nil {
			writeError(w, http.StatusBadRequest, "to must be RFC3339 or YYYY-MM-DD")
			return
		}
	}
	if !s.meterExists(w, r, id) {
		return
	}

	// The baseline always comes from the full series, whatever range is requested.
	rs, err := s.store.ReadingsByMeter(r.Context(), id)
	if err != nil {
		serverError(w, r, err)
		return
	}
	b := engine.BuildBaseline(rs, s.cfg.Engine)

	var hours []readingPoint
	for _, rd := range rs {
		if (!from.IsZero() && rd.Timestamp.Before(from)) || (!to.IsZero() && rd.Timestamp.After(to)) {
			continue
		}
		h := rd.Timestamp.Hour()
		c := b.Consumption[h]
		hours = append(hours, readingPoint{
			Timestamp: rd.Timestamp, ConsumptionKWh: rd.ConsumptionKWh, VoltageV: rd.VoltageV,
			CurrentA: rd.CurrentA, PowerFactor: rd.PowerFactor,
			BaselineKWh: c.Median, BandLowKWh: math.Max(0, c.Median-bandMADs*c.MAD), BandHighKWh: c.Median + bandMADs*c.MAD,
			BaselineVoltageV: b.Voltage[h].Median, BaselineCurrentA: b.Current[h].Median, BaselinePowerFactor: b.PowerFactor[h].Median,
		})
	}
	if gran == "day" {
		hours = aggregateDays(hours)
	}
	if hours == nil {
		hours = []readingPoint{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"meter_id": id, "granularity": gran, "points": roundPoints(hours)})
}

// aggregateDays groups hourly points by UTC day: consumption and baselines are summed, electrical
// variables averaged.
func aggregateDays(hours []readingPoint) []readingPoint {
	var out []readingPoint
	var n float64
	flush := func() {
		if n == 0 {
			return
		}
		p := &out[len(out)-1]
		p.VoltageV, p.CurrentA, p.PowerFactor = p.VoltageV/n, p.CurrentA/n, p.PowerFactor/n
		p.BaselineVoltageV, p.BaselineCurrentA, p.BaselinePowerFactor = p.BaselineVoltageV/n, p.BaselineCurrentA/n, p.BaselinePowerFactor/n
	}
	for _, h := range hours {
		day := h.Timestamp.Truncate(24 * time.Hour)
		if len(out) == 0 || !out[len(out)-1].Timestamp.Equal(day) {
			flush()
			n = 0
			out = append(out, readingPoint{Timestamp: day})
		}
		p := &out[len(out)-1]
		p.ConsumptionKWh += h.ConsumptionKWh
		p.BaselineKWh += h.BaselineKWh
		p.BandLowKWh += h.BandLowKWh
		p.BandHighKWh += h.BandHighKWh
		p.VoltageV += h.VoltageV
		p.CurrentA += h.CurrentA
		p.PowerFactor += h.PowerFactor
		p.BaselineVoltageV += h.BaselineVoltageV
		p.BaselineCurrentA += h.BaselineCurrentA
		p.BaselinePowerFactor += h.BaselinePowerFactor
		n++
	}
	flush()
	return out
}

func roundPoints(ps []readingPoint) []readingPoint {
	r := func(v float64) float64 { return math.Round(v*1000) / 1000 }
	for i := range ps {
		p := &ps[i]
		p.ConsumptionKWh, p.VoltageV, p.CurrentA, p.PowerFactor = r(p.ConsumptionKWh), r(p.VoltageV), r(p.CurrentA), r(p.PowerFactor)
		p.BaselineKWh, p.BandLowKWh, p.BandHighKWh = r(p.BaselineKWh), r(p.BandLowKWh), r(p.BandHighKWh)
		p.BaselineVoltageV, p.BaselineCurrentA, p.BaselinePowerFactor = r(p.BaselineVoltageV), r(p.BaselineCurrentA), r(p.BaselinePowerFactor)
	}
	return ps
}
