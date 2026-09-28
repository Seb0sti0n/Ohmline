package seed

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/Seb0sti0n/Ohmline/backend/migrations"
)

const (
	DemoEmail    = "demo@energy.io"
	DemoPassword = "Ohmline#2026"
)

// profiles are the demo names and locations of the meters (from the approved design). A meter that
// is not listed gets a generic name so any CSV can still be loaded.
var profiles = map[string]struct{ Name, Location string }{
	"M-101": {"Compresores A", "Planta Norte"},
	"M-102": {"Hornos línea 1", "Planta Norte"},
	"M-103": {"Oficinas", "Sede administrativa"},
	"M-104": {"Ensamble", "Planta Sur"},
	"M-105": {"Bombeo", "Planta Norte"},
	"M-106": {"Extrusión", "Planta Sur"},
	"M-107": {"Iluminación patio", "Centro logístico"},
	"M-108": {"Refrigeración", "Centro logístico"},
	"M-109": {"Tablero principal B", "Planta Sur"},
	"M-110": {"Talleres", "Planta Norte"},
	"M-111": {"Climatización", "Sede administrativa"},
	"M-112": {"Empaque", "Centro logístico"},
}

func profile(id string) (name, location string) {
	if p, ok := profiles[id]; ok {
		return p.Name, p.Location
	}
	return "Medidor " + id, "Sin ubicación"
}

type Summary struct {
	Meters, Readings, Events int
	Issues                   []Issue
}

// Migrate applies all pending migrations.
func Migrate(databaseURL string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, strings.Replace(databaseURL, "postgres://", "pgx5://", 1))
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// Run migrates, truncates and reloads all data. It is idempotent.
func Run(ctx context.Context, pool *pgxpool.Pool, dataDir string) (*Summary, error) {
	rawReadings, parseIssues, err := ParseReadings(filepath.Join(dataDir, "readings.csv"))
	if err != nil {
		return nil, err
	}
	events, err := ParseEvents(filepath.Join(dataDir, "events.csv"))
	if err != nil {
		return nil, err
	}
	readings, issues := Validate(rawReadings)
	issues = append(parseIssues, issues...)

	meterIDs := uniqueMeters(readings)
	known := map[string]bool{}
	for _, id := range meterIDs {
		known[id] = true
	}
	for _, e := range events {
		if !known[e.MeterID] {
			return nil, fmt.Errorf("event references unknown meter %s", e.MeterID)
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(DemoPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `TRUNCATE anomalies, analysis_runs, readings, events, meters, users RESTART IDENTITY CASCADE`); err != nil {
		return nil, err
	}

	meterRows := make([][]any, len(meterIDs))
	for i, id := range meterIDs {
		name, location := profile(id)
		meterRows[i] = []any{id, name, location}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"meters"}, []string{"meter_id", "name", "location"}, pgx.CopyFromRows(meterRows)); err != nil {
		return nil, fmt.Errorf("meters: %w", err)
	}

	readingRows := make([][]any, len(readings))
	for i, r := range readings {
		readingRows[i] = []any{r.MeterID, r.Timestamp, r.ConsumptionKWh, r.VoltageV, r.CurrentA, r.PowerFactor, r.Status}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"readings"},
		[]string{"meter_id", "timestamp", "consumption_kwh", "voltage_v", "current_a", "power_factor", "status"},
		pgx.CopyFromRows(readingRows)); err != nil {
		return nil, fmt.Errorf("readings: %w", err)
	}

	eventRows := make([][]any, len(events))
	for i, e := range events {
		eventRows[i] = []any{e.MeterID, e.Timestamp, e.Type, e.Description}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"events"}, []string{"meter_id", "timestamp", "type", "description"}, pgx.CopyFromRows(eventRows)); err != nil {
		return nil, fmt.Errorf("events: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO users (email, password_hash) VALUES ($1, $2)`, DemoEmail, string(hash)); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &Summary{Meters: len(meterIDs), Readings: len(readings), Events: len(events), Issues: issues}, nil
}

func uniqueMeters(readings []Reading) []string {
	set := map[string]bool{}
	for _, r := range readings {
		set[r.MeterID] = true
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
