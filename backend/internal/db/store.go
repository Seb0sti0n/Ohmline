package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Seb0sti0n/astrophage/backend/internal/engine"
)

// Store groups the SQL queries used by the API and the analysis runner.
type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// ---- meters ----

type Meter struct {
	MeterID  string `json:"meter_id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Status   string `json:"status"`
}

func (s *Store) Meters(ctx context.Context) ([]Meter, error) {
	rows, err := s.pool.Query(ctx, `SELECT meter_id, name, location, status FROM meters ORDER BY meter_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Meter
	for rows.Next() {
		var m Meter
		if err := rows.Scan(&m.MeterID, &m.Name, &m.Location, &m.Status); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Meter returns nil (and no error) when the meter does not exist.
func (s *Store) Meter(ctx context.Context, meterID string) (*Meter, error) {
	var m Meter
	err := s.pool.QueryRow(ctx, `SELECT meter_id, name, location, status FROM meters WHERE meter_id = $1`, meterID).
		Scan(&m.MeterID, &m.Name, &m.Location, &m.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

// ---- readings and events (returned as engine types; timestamps are always UTC) ----

func (s *Store) readings(ctx context.Context, where string, args ...any) ([]engine.Reading, error) {
	rows, err := s.pool.Query(ctx, `SELECT meter_id, timestamp, consumption_kwh, voltage_v, current_a, power_factor
		FROM readings `+where+` ORDER BY meter_id, timestamp`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []engine.Reading
	for rows.Next() {
		var r engine.Reading
		if err := rows.Scan(&r.MeterID, &r.Timestamp, &r.ConsumptionKWh, &r.VoltageV, &r.CurrentA, &r.PowerFactor); err != nil {
			return nil, err
		}
		r.Timestamp = r.Timestamp.UTC()
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) AllReadings(ctx context.Context) ([]engine.Reading, error) {
	return s.readings(ctx, "")
}

func (s *Store) ReadingsByMeter(ctx context.Context, meterID string) ([]engine.Reading, error) {
	return s.readings(ctx, "WHERE meter_id = $1", meterID)
}

func (s *Store) events(ctx context.Context, where string, args ...any) ([]engine.Event, error) {
	rows, err := s.pool.Query(ctx, `SELECT meter_id, timestamp, type, description FROM events `+where+` ORDER BY timestamp, id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []engine.Event
	for rows.Next() {
		var e engine.Event
		if err := rows.Scan(&e.MeterID, &e.Timestamp, &e.Type, &e.Description); err != nil {
			return nil, err
		}
		e.Timestamp = e.Timestamp.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) AllEvents(ctx context.Context) ([]engine.Event, error) { return s.events(ctx, "") }

func (s *Store) EventsByMeter(ctx context.Context, meterID string) ([]engine.Event, error) {
	return s.events(ctx, "WHERE meter_id = $1", meterID)
}

// ---- users ----

// PasswordHash returns the hash for an email, or "" if the user does not exist.
func (s *Store) PasswordHash(ctx context.Context, email string) (string, error) {
	var h string
	err := s.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE email = $1`, email).Scan(&h)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return h, err
}
