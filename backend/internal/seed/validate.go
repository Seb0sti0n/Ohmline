package seed

import (
	"sort"
	"time"
)

// Validate checks the readings for duplicates, negative values and gaps in the hourly grid.
// It returns the readings to load (duplicates removed) and the issues found.
func Validate(readings []Reading) ([]Reading, []Issue) {
	var issues []Issue
	type key struct {
		meter string
		ts    time.Time
	}
	seen := map[key]bool{}
	byMeter := map[string][]time.Time{}
	var clean []Reading

	for _, r := range readings {
		k := key{r.MeterID, r.Timestamp}
		if seen[k] {
			issues = append(issues, Issue{r.MeterID, 0, "duplicate reading at " + r.Timestamp.Format("2006-01-02 15:04")})
			continue
		}
		seen[k] = true
		if r.ConsumptionKWh < 0 || r.VoltageV < 0 || r.CurrentA < 0 || r.PowerFactor < 0 {
			issues = append(issues, Issue{r.MeterID, 0, "negative value at " + r.Timestamp.Format("2006-01-02 15:04")})
		}
		byMeter[r.MeterID] = append(byMeter[r.MeterID], r.Timestamp)
		clean = append(clean, r)
	}

	for meter, times := range byMeter {
		sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
		for i := 1; i < len(times); i++ {
			if gap := times[i].Sub(times[i-1]); gap != time.Hour {
				issues = append(issues, Issue{meter, 0, "gap of " + gap.String() + " after " + times[i-1].Format("2006-01-02 15:04")})
			}
		}
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].String() < issues[j].String() })
	return clean, issues
}
