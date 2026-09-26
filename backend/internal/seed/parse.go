// Package seed loads the CSV data into PostgreSQL and validates it.
package seed

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Reading struct {
	MeterID        string
	Timestamp      time.Time
	ConsumptionKWh float64
	VoltageV       float64
	CurrentA       float64
	PowerFactor    float64
	Status         string
}

type Event struct {
	MeterID     string
	Timestamp   time.Time
	Type        string
	Description string
}

// Issue is a data problem found while loading. Row is the 1-based CSV line (0 if not applicable).
type Issue struct {
	MeterID string
	Row     int
	Message string
}

func (i Issue) String() string {
	if i.Row > 0 {
		return fmt.Sprintf("%s (line %d): %s", i.MeterID, i.Row, i.Message)
	}
	return fmt.Sprintf("%s: %s", i.MeterID, i.Message)
}

func readCSV(path string, wantCols int) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(rows) < 1 || len(rows[0]) != wantCols {
		return nil, fmt.Errorf("%s: unexpected header", path)
	}
	return rows[1:], nil
}

// ParseReadings reads readings.csv. Rows with empty or non-numeric fields are skipped and reported.
func ParseReadings(path string) ([]Reading, []Issue, error) {
	rows, err := readCSV(path, 7)
	if err != nil {
		return nil, nil, err
	}
	var out []Reading
	var issues []Issue
	for i, r := range rows {
		line := i + 2 // header is line 1
		ts, err := time.ParseInLocation("2006-01-02 15:04:05", r[1], time.UTC)
		if err != nil {
			issues = append(issues, Issue{r[0], line, "invalid timestamp " + strconv.Quote(r[1])})
			continue
		}
		nums := make([]float64, 4)
		bad := false
		for j := range nums {
			if strings.TrimSpace(r[2+j]) == "" {
				issues = append(issues, Issue{r[0], line, "empty value in column " + strconv.Itoa(3+j)})
				bad = true
				break
			}
			if nums[j], err = strconv.ParseFloat(r[2+j], 64); err != nil {
				issues = append(issues, Issue{r[0], line, "non-numeric value " + strconv.Quote(r[2+j])})
				bad = true
				break
			}
		}
		if bad {
			continue
		}
		out = append(out, Reading{r[0], ts, nums[0], nums[1], nums[2], nums[3], r[6]})
	}
	return out, issues, nil
}

func ParseEvents(path string) ([]Event, error) {
	rows, err := readCSV(path, 4)
	if err != nil {
		return nil, err
	}
	var out []Event
	for i, r := range rows {
		ts, err := time.ParseInLocation("2006-01-02 15:04", r[1], time.UTC)
		if err != nil {
			return nil, fmt.Errorf("events.csv line %d: invalid timestamp %q", i+2, r[1])
		}
		out = append(out, Event{r[0], ts, r[2], r[3]})
	}
	return out, nil
}
