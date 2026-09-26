package http

import (
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/Seb0sti0n/astrophage/backend/internal/db"
	"github.com/Seb0sti0n/astrophage/backend/internal/engine"
)

const topPriorities = 4

type dailyPoint struct {
	Date           string  `json:"date"`
	ConsumptionKWh float64 `json:"consumption_kwh"`
	BaselineKWh    float64 `json:"baseline_kwh"` // sum of the meters' expected daily consumption
}

// dashboardSummary feeds the whole home screen. The analysis-derived fields are empty until the
// first analysis has completed (last_analysis is null before any run).
func (s *Server) dashboardSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	meters, err := s.allSummaries(r)
	if err != nil {
		serverError(w, r, err)
		return
	}
	readings, err := s.store.AllReadings(ctx)
	if err != nil {
		serverError(w, r, err)
		return
	}
	anomalies, err := s.store.Anomalies(ctx, db.AnomalyFilter{})
	if err != nil {
		serverError(w, r, err)
		return
	}
	run, err := s.store.LatestRun(ctx)
	if err != nil {
		serverError(w, r, err)
		return
	}

	total := 0.0
	perDay := map[string]float64{}
	for _, rd := range readings {
		total += rd.ConsumptionKWh
		perDay[rd.Timestamp.Format("2006-01-02")] += rd.ConsumptionKWh
	}
	baseline := 0.0
	for _, m := range meters {
		baseline += m.BaselineKWh
	}
	days := make([]string, 0, len(perDay))
	for d := range perDay {
		days = append(days, d)
	}
	sort.Strings(days)
	daily := make([]dailyPoint, 0, len(days))
	for _, d := range days {
		daily = append(daily, dailyPoint{d, r1(perDay[d]), r1(baseline)})
	}

	high, confSum := 0, 0.0
	for _, a := range anomalies {
		if a.Severity == string(engine.High) {
			high++
		}
		confSum += a.Confidence
	}
	var avgConf *float64
	if len(anomalies) > 0 {
		v := math.Round(confSum/float64(len(anomalies))*100) / 100
		avgConf = &v
	}
	top := anomalies
	if len(top) > topPriorities {
		top = top[:topPriorities]
	}
	if top == nil {
		top = []db.Anomaly{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"meters_count":          len(meters),
		"total_consumption_kwh": r1(total),
		"anomalies_count":       len(anomalies),
		"high_priority_count":   high,
		"avg_confidence":        avgConf,
		"last_analysis":         run,
		"daily_consumption":     daily,
		"top_priorities":        top,
		"meters":                meters,
		"generated_at":          time.Now().UTC(),
	})
}
