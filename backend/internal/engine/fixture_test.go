package engine

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/Seb0sti0n/astrophage/backend/internal/seed"
)

// loadFixture reads the real dataset from /data (the same CSVs the app is seeded with).
func loadFixture(t testing.TB) (map[string][]Reading, []Event) {
	t.Helper()
	dir := filepath.Join("..", "..", "..", "data")
	rr, _, err := seed.ParseReadings(filepath.Join(dir, "readings.csv"))
	if err != nil {
		t.Fatal(err)
	}
	ee, err := seed.ParseEvents(filepath.Join(dir, "events.csv"))
	if err != nil {
		t.Fatal(err)
	}
	by := map[string][]Reading{}
	for _, r := range rr {
		by[r.MeterID] = append(by[r.MeterID], Reading{r.MeterID, r.Timestamp, r.ConsumptionKWh, r.VoltageV, r.CurrentA, r.PowerFactor})
	}
	for _, rs := range by {
		sort.Slice(rs, func(i, j int) bool { return rs[i].Timestamp.Before(rs[j].Timestamp) })
	}
	var events []Event
	for _, e := range ee {
		events = append(events, Event{e.MeterID, e.Timestamp, e.Type, e.Description})
	}
	return by, events
}

func flatten(by map[string][]Reading) []Reading {
	var all []Reading
	for _, rs := range by {
		all = append(all, rs...)
	}
	return all
}
