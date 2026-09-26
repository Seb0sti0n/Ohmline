package engine

import (
	"time"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
)

// HourStat is the median and scaled MAD of a variable at one hour of the day.
type HourStat struct{ Median, MAD float64 }

// Baseline is a meter's expected behavior, by hour of the day (UTC), from its first BaselineDays days.
type Baseline struct {
	Consumption, Voltage, Current, PowerFactor [24]HourStat
	DailyKWh                                   float64 // sum of the 24 hourly consumption medians
	CoherenceRatio                             float64 // median of kWh / (V·I·PF/1000)
	VoltageStd                                 float64 // std of voltage over the whole baseline window
}

// electricalKWh is the energy implied by V, I and PF for one hour.
func electricalKWh(r Reading) float64 { return r.VoltageV * r.CurrentA * r.PowerFactor / 1000 }

// BuildBaseline expects readings of a single meter sorted by time.
func BuildBaseline(rs []Reading, cfg config.Engine) Baseline {
	var b Baseline
	if len(rs) == 0 {
		return b
	}
	end := rs[0].Timestamp.Add(time.Duration(cfg.BaselineDays) * 24 * time.Hour)

	var kwh, volt, cur, pf [24][]float64
	var ratios, allVolt []float64
	for _, r := range rs {
		if !r.Timestamp.Before(end) {
			break
		}
		h := r.Timestamp.Hour()
		kwh[h] = append(kwh[h], r.ConsumptionKWh)
		volt[h] = append(volt[h], r.VoltageV)
		cur[h] = append(cur[h], r.CurrentA)
		pf[h] = append(pf[h], r.PowerFactor)
		allVolt = append(allVolt, r.VoltageV)
		if e := electricalKWh(r); e > 0 {
			ratios = append(ratios, r.ConsumptionKWh/e)
		}
	}
	for h := 0; h < 24; h++ {
		b.Consumption[h] = HourStat{median(kwh[h]), scaledMAD(kwh[h])}
		b.Voltage[h] = HourStat{median(volt[h]), scaledMAD(volt[h])}
		b.Current[h] = HourStat{median(cur[h]), scaledMAD(cur[h])}
		b.PowerFactor[h] = HourStat{median(pf[h]), scaledMAD(pf[h])}
		b.DailyKWh += b.Consumption[h].Median
	}
	b.CoherenceRatio = median(ratios)
	b.VoltageStd = std(allVolt)
	return b
}
