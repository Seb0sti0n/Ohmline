// Package analysis orchestrates an analysis run: it loads the data, runs the engine, tracks the
// progress of each stage and persists the results.
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
	"github.com/Seb0sti0n/astrophage/backend/internal/db"
	"github.com/Seb0sti0n/astrophage/backend/internal/engine"
)

// StepNames are the visible pipeline stages, in order.
var StepNames = []string{"Lecturas", "Baseline", "Detección", "Correlación", "Eventos", "Explicación", "Recomendación"}

type Step struct {
	Name       string `json:"name"`
	Status     string `json:"status"` // PENDING, RUNNING, DONE
	DurationMS int64  `json:"duration_ms"`
}

type Runner struct {
	store  *db.Store
	engine config.Engine
	delay  time.Duration // pause after each stage so the progress is visible
	mu     sync.Mutex    // serializes Start so two requests cannot both create a run
}

func NewRunner(store *db.Store, cfg config.Config) *Runner {
	return &Runner{store: store, engine: cfg.Engine, delay: cfg.StepDelay}
}

func newSteps() []Step {
	steps := make([]Step, len(StepNames))
	for i, n := range StepNames {
		steps[i] = Step{Name: n, Status: "PENDING"}
	}
	return steps
}

// Start creates a run and executes it in the background. If a run is already in progress it
// returns that run's id and inProgress = true instead of starting another.
func (r *Runner) Start(ctx context.Context) (id int, inProgress bool, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id, active, err := r.store.ActiveRunID(ctx); err != nil {
		return 0, false, err
	} else if active {
		return id, true, nil
	}
	steps, _ := json.Marshal(newSteps())
	id, err = r.store.CreateRun(ctx, steps)
	if err != nil {
		return 0, false, err
	}
	go r.execute(id)
	return id, false, nil
}

// execute runs in its own goroutine, so it uses a background context and never panics the server.
//
// The engine is a single pure computation: it runs during the "Detección" stage and covers
// baseline, correlation, event matching and the template explanations for all meters. The
// stepper mirrors the pipeline so the operator can follow it; results are persisted at the end.
func (r *Runner) execute(runID int) {
	ctx := context.Background()
	started := time.Now()
	steps := newSteps()

	fail := func(err error) {
		log.Printf("analysis %d failed: %v", runID, err)
		if ferr := r.store.FailRun(ctx, runID, err.Error()); ferr != nil {
			log.Printf("analysis %d: could not record failure: %v", runID, ferr)
		}
	}
	defer func() {
		if p := recover(); p != nil {
			fail(fmt.Errorf("panic: %v", p))
		}
	}()

	persist := func(current string) error {
		b, _ := json.Marshal(steps)
		return r.store.UpdateProgress(ctx, runID, current, b)
	}
	// stage marks a stage running, does its work, waits the visible delay and marks it done.
	stage := func(i int, work func() error) error {
		steps[i].Status = "RUNNING"
		if err := persist(steps[i].Name); err != nil {
			return err
		}
		t := time.Now()
		if work != nil {
			if err := work(); err != nil {
				return err
			}
		}
		time.Sleep(r.delay)
		steps[i].Status, steps[i].DurationMS = "DONE", time.Since(t).Milliseconds()
		return persist(steps[i].Name)
	}

	var readings []engine.Reading
	var events []engine.Event
	var results []engine.AnomalyResult

	stages := []func() error{
		func() (err error) { // Lecturas
			if readings, err = r.store.AllReadings(ctx); err != nil {
				return err
			}
			events, err = r.store.AllEvents(ctx)
			return err
		},
		nil, // Baseline
		func() error { results = engine.Run(readings, events, r.engine); return nil }, // Detección
		nil, // Correlación
		nil, // Eventos
		nil, // Explicación (template texts come from the engine)
		nil, // Recomendación
	}
	for i, work := range stages {
		if err := stage(i, work); err != nil {
			fail(err)
			return
		}
	}

	stepsJSON, _ := json.Marshal(steps)
	summaryJSON, _ := json.Marshal(summarize(results, readings, started))
	if err := r.store.SaveResults(ctx, runID, results, stepsJSON, summaryJSON); err != nil {
		fail(err)
	}
}

type Summary struct {
	MetersAnalyzed int     `json:"meters_analyzed"`
	Anomalies      int     `json:"anomalies"`
	HighPriority   int     `json:"high_priority"`
	AvgConfidence  float64 `json:"avg_confidence"`
	DurationMS     int64   `json:"duration_ms"`
}

// summarize counts what the dashboard shows: HIGH severity anomalies are the ones needing attention first.
func summarize(results []engine.AnomalyResult, readings []engine.Reading, started time.Time) Summary {
	meters := map[string]bool{}
	for _, r := range readings {
		meters[r.MeterID] = true
	}
	s := Summary{MetersAnalyzed: len(meters), Anomalies: len(results), DurationMS: time.Since(started).Milliseconds()}
	total := 0.0
	for _, r := range results {
		if r.Severity == engine.High {
			s.HighPriority++
		}
		total += r.Confidence
	}
	if len(results) > 0 {
		s.AvgConfidence = float64(int(total/float64(len(results))*100+0.5)) / 100
	}
	return s
}
