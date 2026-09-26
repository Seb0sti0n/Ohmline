package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Seb0sti0n/astrophage/backend/internal/engine"
)

// Run is an analysis run as exposed by the API.
type Run struct {
	ID          int             `json:"id"`
	Status      string          `json:"status"`
	CurrentStep *string         `json:"current_step"`
	Steps       json.RawMessage `json:"steps"`
	Summary     json.RawMessage `json:"summary"`
	StartedAt   time.Time       `json:"started_at"`
	FinishedAt  *time.Time      `json:"finished_at"`
}

const runCols = `id, status, current_step, steps, summary, started_at, finished_at`

func scanRun(row pgx.Row) (*Run, error) {
	var r Run
	var steps, summary []byte
	err := row.Scan(&r.ID, &r.Status, &r.CurrentStep, &steps, &summary, &r.StartedAt, &r.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.Steps, r.Summary = steps, summary
	return &r, nil
}

func (s *Store) CreateRun(ctx context.Context, steps []byte) (int, error) {
	var id int
	err := s.pool.QueryRow(ctx, `INSERT INTO analysis_runs (status, steps) VALUES ('RUNNING', $1) RETURNING id`, steps).Scan(&id)
	return id, err
}

func (s *Store) Run(ctx context.Context, id int) (*Run, error) {
	return scanRun(s.pool.QueryRow(ctx, `SELECT `+runCols+` FROM analysis_runs WHERE id = $1`, id))
}

func (s *Store) LatestRun(ctx context.Context) (*Run, error) {
	return scanRun(s.pool.QueryRow(ctx, `SELECT `+runCols+` FROM analysis_runs ORDER BY id DESC LIMIT 1`))
}

// ActiveRunID returns the id of a run that is still in progress, if any.
func (s *Store) ActiveRunID(ctx context.Context) (int, bool, error) {
	var id int
	err := s.pool.QueryRow(ctx, `SELECT id FROM analysis_runs WHERE status IN ('PENDING','RUNNING') ORDER BY id DESC LIMIT 1`).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, err
}

// UpdateProgress records the current step and the per-step state of a running analysis.
func (s *Store) UpdateProgress(ctx context.Context, id int, currentStep string, steps []byte) error {
	_, err := s.pool.Exec(ctx, `UPDATE analysis_runs SET current_step = $2, steps = $3 WHERE id = $1`, id, currentStep, steps)
	return err
}

func (s *Store) FailRun(ctx context.Context, id int, reason string) error {
	summary, _ := json.Marshal(map[string]string{"error": reason})
	_, err := s.pool.Exec(ctx, `UPDATE analysis_runs SET status = 'FAILED', finished_at = now(), summary = $2 WHERE id = $1`, id, summary)
	return err
}

// FailStaleRuns marks runs left in progress by a previous process as failed.
func (s *Store) FailStaleRuns(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE analysis_runs SET status = 'FAILED', finished_at = now(),
		summary = '{"error":"interrupted by a server restart"}' WHERE status IN ('PENDING','RUNNING')`)
	return err
}

// meterStatus maps an anomaly type to the status of its meter.
func meterStatus(t engine.AnomalyType) string {
	switch t {
	case engine.RealAnomaly:
		return "CRITICAL"
	case engine.DataQuality, engine.ExplainableAnomaly:
		return "ALERT"
	}
	return "OK" // false positives are explained: the meter is fine
}

// SaveResults stores the anomalies of a run, updates each meter's status from them and completes
// the run, all in one transaction.
func (s *Store) SaveResults(ctx context.Context, runID int, results []engine.AnomalyResult, steps, summary []byte) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE meters SET status = 'OK'`); err != nil {
		return err
	}
	for _, r := range results {
		evidence, err := json.Marshal(r.Evidence)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO anomalies
			(meter_id, analysis_run_id, type, severity, confidence, priority_score, reason, explanation,
			 recommended_action, window_start, window_end, evidence, explanation_source)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			r.MeterID, runID, r.Type, r.Severity, r.Confidence, r.PriorityScore, r.Reason, r.Explanation,
			r.RecommendedAction, r.WindowStart, r.WindowEnd, evidence, r.ExplanationSource)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE meters SET status = $2 WHERE meter_id = $1`, r.MeterID, meterStatus(r.Type)); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE analysis_runs SET status = 'COMPLETED', finished_at = now(),
		current_step = NULL, steps = $2, summary = $3 WHERE id = $1`, runID, steps, summary); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
