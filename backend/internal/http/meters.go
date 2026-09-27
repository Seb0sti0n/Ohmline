package http

import (
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
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
	DailyKWh       []float64   `json:"daily_kwh"`  // consumption per UTC day, for the sparkline
	DailyFrom      string      `json:"daily_from"` // YYYY-MM-DD of the first daily value
	Anomaly        *db.Anomaly `json:"anomaly"`    // from the latest completed analysis, or null
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
	s.DailyKWh = dailyTotals(rs)
	s.DailyFrom = rs[0].Timestamp.Format("2006-01-02")
	return s
}

// dailyTotals sums consumption per UTC day, in order.
func dailyTotals(rs []engine.Reading) []float64 {
	out := []float64{}
	var day time.Time
	for _, r := range rs {
		d := r.Timestamp.Truncate(24 * time.Hour)
		if len(out) == 0 || !d.Equal(day) {
			out, day = append(out, 0), d
		}
		out[len(out)-1] += r.ConsumptionKWh
	}
	for i := range out {
		out[i] = r1(out[i])
	}
	return out
}

var severityRank = map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}

// allSummaries builds the summary of every meter, in meter_id order.
// statusUnevaluated is what a meter's status is until the first analysis completes: the stored
// status only means something once an analysis has produced it, so we do not claim "OK" before that.
const statusUnevaluated = "UNEVALUATED"

func (s *Server) allSummaries(r *http.Request) ([]meterSummary, error) {
	ctx := r.Context()
	meters, err := s.store.Meters(ctx)
	if err != nil {
		return nil, err
	}
	analysed, err := s.store.HasCompletedRun(ctx)
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
		if !analysed {
			m.Status = statusUnevaluated
		}
		out = append(out, summarizeMeter(m, byMeter[m.MeterID], s.cfg.Engine, anomalyOf[m.MeterID]))
	}
	return out, nil
}

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// meterPage is one page of the meters table. Total is the number of meters that match the filters
// (for the pager); Counts is the number of meters per status over *all* meters, ignoring filters
// and search (for the filter chips), so the UI never needs the full list.
type meterPage struct {
	Items    []meterSummary `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Counts   map[string]int `json:"counts"`
}

// intParam reads an optional integer query parameter within [min, max].
func intParam(q url.Values, name string, fallback, min, max int) (int, bool) {
	v := q.Get(name)
	if v == "" {
		return fallback, true
	}
	n, err := strconv.Atoi(v)
	return n, err == nil && n >= min && n <= max
}

func (s *Server) listMeters(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := strings.ToUpper(q.Get("status"))
	if status != "" && status != "ALL" && status != "OK" && status != "ALERT" && status != "CRITICAL" && status != statusUnevaluated {
		writeError(w, http.StatusBadRequest, "status must be OK, ALERT, CRITICAL or UNEVALUATED")
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
	page, ok := intParam(q, "page", 1, 1, math.MaxInt32)
	if !ok {
		writeError(w, http.StatusBadRequest, "page must be a positive integer")
		return
	}
	pageSize, ok := intParam(q, "page_size", defaultPageSize, 1, maxPageSize)
	if !ok {
		writeError(w, http.StatusBadRequest, "page_size must be between 1 and 100")
		return
	}

	all, err := s.allSummaries(r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	counts := map[string]int{"ALL": len(all), "OK": 0, "ALERT": 0, "CRITICAL": 0, statusUnevaluated: 0}
	for _, m := range all {
		counts[m.Status]++
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
				return math.Abs(m.VariationPct) // the biggest change first, up or down
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

	// Without an explicit sort the order is by meter_id (allSummaries), which keeps pages stable.
	total := len(list)
	start := (page - 1) * pageSize
	items := []meterSummary{}
	if start < total {
		items = list[start:min(start+pageSize, total)]
	}
	writeJSON(w, http.StatusOK, meterPage{Items: items, Total: total, Page: page, PageSize: pageSize, Counts: counts})
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
