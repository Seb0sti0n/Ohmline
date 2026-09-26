package http

import (
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
	"github.com/Seb0sti0n/astrophage/backend/internal/db"
	"github.com/Seb0sti0n/astrophage/backend/internal/engine"
)

type meterSummary struct {
	db.Meter
	ConsumptionKWh float64     `json:"consumption_kwh"` // last 24 h
	BaselineKWh    float64     `json:"baseline_kwh"`    // expected daily consumption
	VariationPct   float64     `json:"variation_pct"`
	Anomaly        *db.Anomaly `json:"anomaly"` // from the latest completed analysis, or null
}

func r1(v float64) float64 { return math.Round(v*10) / 10 }

// summarize computes last-day consumption against the baseline. The baseline does not depend on
// the analysis, so the table works before the first run.
func summarizeMeter(m db.Meter, rs []engine.Reading, cfg config.Engine, a *db.Anomaly) meterSummary {
	s := meterSummary{Meter: m, Anomaly: a}
	if len(rs) == 0 {
		return s
	}
	b := engine.BuildBaseline(rs, cfg)
	last := rs
	if len(last) > 24 {
		last = last[len(last)-24:]
	}
	for _, r := range last {
		s.ConsumptionKWh += r.ConsumptionKWh
	}
	s.BaselineKWh = b.DailyKWh
	if b.DailyKWh > 0 {
		s.VariationPct = (s.ConsumptionKWh/b.DailyKWh - 1) * 100
	}
	s.ConsumptionKWh, s.BaselineKWh, s.VariationPct = r1(s.ConsumptionKWh), r1(s.BaselineKWh), r1(s.VariationPct)
	return s
}

var severityRank = map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}

// allSummaries builds the summary of every meter, in meter_id order.
func (s *Server) allSummaries(r *http.Request) ([]meterSummary, error) {
	ctx := r.Context()
	meters, err := s.store.Meters(ctx)
	if err != nil {
		return nil, err
	}
	readings, err := s.store.AllReadings(ctx)
	if err != nil {
		return nil, err
	}
	anomalies, err := s.store.Anomalies(ctx, db.AnomalyFilter{})
	if err != nil {
		return nil, err
	}
	byMeter := map[string][]engine.Reading{}
	for _, rd := range readings {
		byMeter[rd.MeterID] = append(byMeter[rd.MeterID], rd)
	}
	anomalyOf := map[string]*db.Anomaly{}
	for i := range anomalies {
		anomalyOf[anomalies[i].MeterID] = &anomalies[i]
	}
	out := make([]meterSummary, 0, len(meters))
	for _, m := range meters {
		out = append(out, summarizeMeter(m, byMeter[m.MeterID], s.cfg.Engine, anomalyOf[m.MeterID]))
	}
	return out, nil
}

func (s *Server) listMeters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := strings.ToUpper(q.Get("status"))
	if status != "" && status != "ALL" && status != "OK" && status != "ALERT" && status != "CRITICAL" {
		writeError(w, http.StatusBadRequest, "status must be OK, ALERT or CRITICAL")
		return
	}
	sortBy := q.Get("sort")
	if sortBy != "" && sortBy != "consumption" && sortBy != "variation" && sortBy != "severity" {
		writeError(w, http.StatusBadRequest, "sort must be consumption, variation or severity")
		return
	}
	if o := q.Get("order"); o != "" && o != "asc" && o != "desc" {
		writeError(w, http.StatusBadRequest, "order must be asc or desc")
		return
	}

	all, err := s.allSummaries(r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	search := strings.ToLower(strings.TrimSpace(q.Get("search")))
	list := []meterSummary{}
	for _, m := range all {
		if status != "" && status != "ALL" && m.Status != status {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(m.MeterID), search) {
			continue
		}
		list = append(list, m)
	}

	if sortBy != "" {
		key := func(m meterSummary) float64 {
			switch sortBy {
			case "consumption":
				return m.ConsumptionKWh
			case "variation":
				return m.VariationPct
			}
			if m.Anomaly == nil {
				return 0
			}
			return float64(severityRank[m.Anomaly.Severity])*1000 + m.Anomaly.PriorityScore
		}
		asc := q.Get("order") == "asc"
		sort.SliceStable(list, func(i, j int) bool {
			if a, b := key(list[i]), key(list[j]); a != b {
				return (a < b) == asc
			}
			return list[i].MeterID < list[j].MeterID
		})
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) getMeter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "meterId")
	all, err := s.allSummaries(r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	for _, m := range all {
		if m.MeterID == id {
			writeJSON(w, http.StatusOK, m)
			return
		}
	}
	writeError(w, http.StatusNotFound, "meter not found")
}

func (s *Server) getMeterEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "meterId")
	if !s.meterExists(w, r, id) {
		return
	}
	events, err := s.store.EventsByMeter(r.Context(), id)
	if err != nil {
		serverError(w, r, err)
		return
	}
	type ev struct {
		Timestamp   time.Time `json:"timestamp"`
		Type        string    `json:"type"`
		Description string    `json:"description"`
	}
	out := []ev{}
	for _, e := range events {
		out = append(out, ev{e.Timestamp, e.Type, e.Description})
	}
	writeJSON(w, http.StatusOK, out)
}

// meterExists writes a 404 and returns false when the meter is unknown.
func (s *Server) meterExists(w http.ResponseWriter, r *http.Request, id string) bool {
	m, err := s.store.Meter(r.Context(), id)
	if err != nil {
		serverError(w, r, err)
		return false
	}
	if m == nil {
		writeError(w, http.StatusNotFound, "meter not found")
		return false
	}
	return true
}
