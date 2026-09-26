package engine

import (
	"math"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
)

// Change describes how one electrical variable moved inside an anomalous window versus baseline.
type Change struct {
	Variable    string  `json:"variable"`
	Baseline    float64 `json:"baseline"`
	Observed    float64 `json:"observed"`
	Delta       float64 `json:"delta"`
	DeltaPct    float64 `json:"delta_pct"`
	Significant bool    `json:"significant"`
	Direction   string  `json:"direction"` // "up", "down" or "stable"
}

const (
	VarConsumption = "consumption_kwh"
	VarCurrent     = "current_a"
	VarPowerFactor = "power_factor"
	VarVoltage     = "voltage_v"
	VarVoltageStd  = "voltage_std"
)

// Correlate compares consumption, current, power factor and voltage inside the window with the
// baseline for the same hours. deteriorated is true when power factor dropped or voltage shifted
// or became unstable, i.e. the electrical health of the installation got worse.
func Correlate(rs []Reading, b Baseline, startIdx, endIdx int, cfg config.Engine) (changes []Change, deteriorated bool) {
	var kwh, cur, pf, volt, ekwh, ecur, epf, evolt []float64
	for k := startIdx; k <= endIdx; k++ {
		h := rs[k].Timestamp.Hour()
		kwh = append(kwh, rs[k].ConsumptionKWh)
		cur = append(cur, rs[k].CurrentA)
		pf = append(pf, rs[k].PowerFactor)
		volt = append(volt, rs[k].VoltageV)
		ekwh = append(ekwh, b.Consumption[h].Median)
		ecur = append(ecur, b.Current[h].Median)
		epf = append(epf, b.PowerFactor[h].Median)
		evolt = append(evolt, b.Voltage[h].Median)
	}

	pct := func(obs, base float64) float64 {
		if base == 0 {
			return 0
		}
		return (obs/base - 1) * 100
	}
	mk := func(name string, base, obs float64, significant bool) Change {
		dir := "stable"
		if significant {
			dir = "up"
			if obs < base {
				dir = "down"
			}
		}
		return Change{name, round(base, 3), round(obs, 3), round(obs-base, 3), round(pct(obs, base), 1), significant, dir}
	}

	curPct := pct(mean(cur), mean(ecur))
	pfDelta := mean(pf) - mean(epf)
	voltDelta := mean(volt) - mean(evolt)
	stdRatio := 1.0
	if b.VoltageStd > 0 {
		stdRatio = std(volt) / b.VoltageStd
	}

	changes = []Change{
		mk(VarConsumption, mean(ekwh), mean(kwh), math.Abs(pct(mean(kwh), mean(ekwh))) > 0),
		mk(VarCurrent, mean(ecur), mean(cur), math.Abs(curPct) > cfg.CurrentChangePct),
		mk(VarPowerFactor, mean(epf), mean(pf), math.Abs(pfDelta) > cfg.PFDrop),
		mk(VarVoltage, mean(evolt), mean(volt), math.Abs(voltDelta) > cfg.VoltageShiftV),
		mk(VarVoltageStd, b.VoltageStd, std(volt), stdRatio > cfg.VoltageStdRatio),
	}

	pfDropped := pfDelta < -cfg.PFDrop
	voltShifted := math.Abs(voltDelta) > cfg.VoltageShiftV
	voltUnstable := stdRatio > cfg.VoltageStdRatio
	return changes, pfDropped || voltShifted || voltUnstable
}
