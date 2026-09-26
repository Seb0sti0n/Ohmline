package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Anomaly is an anomaly row joined with its meter. Evidence is only filled by Anomaly (detail).
type Anomaly struct {
	ID                int             `json:"id"`
	MeterID           string          `json:"meter_id"`
	MeterName         string          `json:"meter_name"`
	AnalysisRunID     int             `json:"analysis_run_id"`
	DetectedAt        time.Time       `json:"detected_at"`
	Type              string          `json:"type"`
	Severity          string          `json:"severity"`
	Confidence        float64         `json:"confidence"`
	PriorityScore     float64         `json:"priority_score"`
	Reason            string          `json:"reason"`
	Explanation       string          `json:"explanation"`
	RecommendedAction string          `json:"recommended_action"`
	Status            string          `json:"status"`
	WindowStart       *time.Time      `json:"window_start"`
	WindowEnd         *time.Time      `json:"window_end"`
	ExplanationSource string          `json:"explanation_source"`
	Evidence          json.RawMessage `json:"evidence,omitempty"`
}

type AnomalyFilter struct{ Type, Severity, Status string }

const anomalyCols = `a.id, a.meter_id, m.name, a.analysis_run_id, a.detected_at, a.type, a.severity, a.confidence,
	a.priority_score, a.reason, a.explanation, a.recommended_action, a.status, a.window_start, a.window_end, a.explanation_source`

func scanAnomaly(row pgx.Row, withEvidence bool) (*Anomaly, error) {
	var a Anomaly
	dest := []any{&a.ID, &a.MeterID, &a.MeterName, &a.AnalysisRunID, &a.DetectedAt, &a.Type, &a.Severity, &a.Confidence,
		&a.PriorityScore, &a.Reason, &a.Explanation, &a.RecommendedAction, &a.Status, &a.WindowStart, &a.WindowEnd, &a.ExplanationSource}
	var evidence []byte
	if withEvidence {
		dest = append(dest, &evidence)
	}
	if err := row.Scan(dest...); err != nil {
		return nil, err
	}
	a.Evidence = evidence
	return &a, nil
}

// Anomalies returns the anomalies of the latest completed analysis, highest priority first.
func (s *Store) Anomalies(ctx context.Context, f AnomalyFilter) ([]Anomaly, error) {
	q := `SELECT ` + anomalyCols + ` FROM anomalies a JOIN meters m ON m.meter_id = a.meter_id
		WHERE a.analysis_run_id = (SELECT id FROM analysis_runs WHERE status = 'COMPLETED' ORDER BY id DESC LIMIT 1)`
	var args []any
	add := func(col, val string) {
		if val != "" {
			args = append(args, val)
			q += fmt.Sprintf(" AND a.%s = $%d", col, len(args))
		}
	}
	add("type", f.Type)
	add("severity", f.Severity)
	add("status", f.Status)
	q += ` ORDER BY a.priority_score DESC, a.id`

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Anomaly{}
	for rows.Next() {
		a, err := scanAnomaly(rows, false)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

// Anomaly returns the detail with evidence, or nil if the id does not exist.
func (s *Store) Anomaly(ctx context.Context, id int) (*Anomaly, error) {
	a, err := scanAnomaly(s.pool.QueryRow(ctx, `SELECT `+anomalyCols+`, a.evidence FROM anomalies a
		JOIN meters m ON m.meter_id = a.meter_id WHERE a.id = $1`, id), true)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

// UpdateAnomalyStatus returns false if the anomaly does not exist.
func (s *Store) UpdateAnomalyStatus(ctx context.Context, id int, status string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE anomalies SET status = $2 WHERE id = $1`, id, status)
	return tag.RowsAffected() > 0, err
}
