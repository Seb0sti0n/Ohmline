package engine

import (
	"math"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
)

// Window is a persistent consumption change: a run of consecutive hours whose robust z-score
// exceeds the threshold in the same direction.
type Window struct {
	StartIdx, EndIdx  int // indexes into the meter's readings, inclusive
	Hours             int
	Direction         int  // +1 above baseline, -1 below
	Ongoing           bool // still active at the last reading
	ReturnsToBaseline bool // for finished windows: consumption right after is back within tolerance
	MeanVariationPct  float64
	MedianAbsZ        float64
	ExtraKWh          float64
}

const madFloor = 1e-6

func robustZ(x float64, s HourStat) float64 { return (x - s.Median) / math.Max(s.MAD, madFloor) }

// DetectConsumption returns the persistent windows and the number of isolated spikes (runs shorter
// than MinPersistHours), which are minor findings and never become anomalies.
func DetectConsumption(rs []Reading, b Baseline, cfg config.Engine) (windows []Window, spikes int) {
	z := make([]float64, len(rs))
	for i, r := range rs {
		z[i] = robustZ(r.ConsumptionKWh, b.Consumption[r.Timestamp.Hour()])
	}
	for i := 0; i < len(rs); {
		if math.Abs(z[i]) <= cfg.ZThreshold {
			i++
			continue
		}
		dir := 1
		if z[i] < 0 {
			dir = -1
		}
		j := i
		for j+1 < len(rs) && math.Abs(z[j+1]) > cfg.ZThreshold && (z[j+1] > 0) == (z[i] > 0) {
			j++
		}
		if j-i+1 >= cfg.MinPersistHours {
			windows = append(windows, buildWindow(rs, b, z, i, j, dir, cfg))
		} else {
			spikes++
		}
		i = j + 1
	}
	return windows, spikes
}

func buildWindow(rs []Reading, b Baseline, z []float64, i, j, dir int, cfg config.Engine) Window {
	w := Window{StartIdx: i, EndIdx: j, Hours: j - i + 1, Direction: dir, Ongoing: j == len(rs)-1}
	var actual, expected, absZ []float64
	for k := i; k <= j; k++ {
		actual = append(actual, rs[k].ConsumptionKWh)
		expected = append(expected, b.Consumption[rs[k].Timestamp.Hour()].Median)
		absZ = append(absZ, math.Abs(z[k]))
	}
	w.MeanVariationPct = (mean(actual)/mean(expected) - 1) * 100
	w.ExtraKWh = sum(actual) - sum(expected)
	w.MedianAbsZ = median(absZ)

	if !w.Ongoing {
		// Look at up to 6 hours after the window: has consumption gone back to normal?
		var after, exp []float64
		for k := j + 1; k < len(rs) && k <= j+6; k++ {
			after = append(after, rs[k].ConsumptionKWh)
			exp = append(exp, b.Consumption[rs[k].Timestamp.Hour()].Median)
		}
		if len(after) > 0 {
			w.ReturnsToBaseline = math.Abs(mean(after)/mean(exp)-1)*100 <= cfg.RecoveryTolPct
		}
	}
	return w
}

func sum(x []float64) float64 {
	s := 0.0
	for _, v := range x {
		s += v
	}
	return s
}
