package engine

import (
	"fmt"
	"sort"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
)

// qualityOngoingHours: a quality problem is "ongoing" if its last flag is this close to the end of the data.
const qualityOngoingHours = 3

// Run analyzes every meter and returns its anomalies ordered by priority (highest first).
// Meters without findings produce no result. The event type is never used as a label: a data
// quality problem must be visible in the readings themselves.
func Run(readings []Reading, events []Event, cfg config.Engine) []AnomalyResult {
	byMeter := map[string][]Reading{}
	for _, r := range readings {
		byMeter[r.MeterID] = append(byMeter[r.MeterID], r)
	}
	eventsBy := map[string][]Event{}
	for _, e := range events {
		eventsBy[e.MeterID] = append(eventsBy[e.MeterID], e)
	}

	var results []AnomalyResult
	for id, rs := range byMeter {
		sort.Slice(rs, func(i, j int) bool { return rs[i].Timestamp.Before(rs[j].Timestamp) })
		if res, ok := analyzeMeter(id, rs, eventsBy[id], cfg); ok {
			results = append(results, res)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].PriorityScore != results[j].PriorityScore {
			return results[i].PriorityScore > results[j].PriorityScore
		}
		return results[i].MeterID < results[j].MeterID
	})
	return results
}

func analyzeMeter(id string, rs []Reading, events []Event, cfg config.Engine) (AnomalyResult, bool) {
	if len(rs) == 0 {
		return AnomalyResult{}, false
	}
	b := BuildBaseline(rs, cfg)
	windows, _ := DetectConsumption(rs, b, cfg) // isolated spikes are minor findings, not anomalies
	quality := DetectQuality(rs, b, cfg)

	// Data quality only when flags are recurrent AND consumption has no persistent change.
	if len(windows) == 0 {
		if !quality.Info.Recurrent {
			return AnomalyResult{}, false
		}
		return qualityResult(id, rs, b, events, quality, cfg), true
	}

	var best AnomalyResult
	for i, w := range windows {
		res := consumptionResult(id, rs, b, events, w, cfg)
		if i == 0 || res.PriorityScore > best.PriorityScore {
			best = res
		}
	}
	return best, true
}

func baseMetrics(rs []Reading, b Baseline) Metrics {
	m := Metrics{BaselineDailyKWh: round(b.DailyKWh, 1)}
	n := len(rs)
	last := 24
	if n < last {
		last = n
	}
	for _, r := range rs[n-last:] {
		m.LastDayKWh += r.ConsumptionKWh
	}
	m.LastDayKWh = round(m.LastDayKWh, 1)
	if b.DailyKWh > 0 {
		m.LastDayVariation = round((m.LastDayKWh/b.DailyKWh-1)*100, 1)
	}
	return m
}

func consumptionResult(id string, rs []Reading, b Baseline, events []Event, w Window, cfg config.Engine) AnomalyResult {
	changes, deteriorated := Correlate(rs, b, w.StartIdx, w.EndIdx, cfg)
	start, end := rs[w.StartIdx].Timestamp, rs[w.EndIdx].Timestamp
	matches := MatchConsumptionEvents(events, w, start, deteriorated, cfg)
	typ, sev, match := classifyConsumption(matches)

	m := baseMetrics(rs, b)
	m.WindowVariation = round(w.MeanVariationPct, 1)
	m.ExtraKWh = round(w.ExtraKWh, 1)
	m.MedianAbsZ = round(w.MedianAbsZ, 1)

	rules := []string{
		fmt.Sprintf("Cambio persistente: %d h consecutivas con |z| > %.0f (z mediano %s)", w.Hours, cfg.ZThreshold, num(w.MedianAbsZ, 1)),
	}
	if deteriorated {
		rules = append(rules, "Deterioro eléctrico: cambió el factor de potencia o el voltaje")
	}
	if match != nil {
		rules = append(rules, fmt.Sprintf("Evento compatible: %s (%s)", match.Type, match.Reason))
	} else {
		rules = append(rules, "Sin evento compatible que explique el cambio")
	}
	if w.Ongoing {
		rules = append(rules, "El cambio sigue en curso al final de los datos")
	}

	// Independent signals: consumption, plus each electrical variable that changed, plus the event.
	signals := 1
	for _, c := range changes {
		if c.Significant && c.Variable != VarConsumption && c.Variable != VarVoltageStd {
			signals++
		}
	}
	clarity := 0.9 // checked the events and found nothing that explains it
	if match != nil {
		signals++
		clarity = match.Quality
	}
	conf, confParts := confidence(confidenceInput{Strength: w.MedianAbsZ / 10, Signals: signals, Clarity: clarity})
	score, scoreParts := priorityScore(scoreInput{
		Type: typ, Magnitude: abs(w.MeanVariationPct), Hours: w.Hours, Ongoing: w.Ongoing,
		Electrical: deteriorated, Unexplained: match == nil, ExtraKWh: w.ExtraKWh,
	})

	res := AnomalyResult{
		MeterID: id, Anomaly: true, Type: typ, Severity: sev, Confidence: conf, PriorityScore: score,
		WindowStart: start, WindowEnd: end,
		Evidence: Evidence{
			Kind:                "consumption_shift",
			Window:              WindowInfo{Start: start, End: end, Hours: w.Hours, Ongoing: w.Ongoing},
			Metrics:             m,
			ChangedVariables:    changes,
			Events:              matches,
			RulesFired:          rules,
			PriorityBreakdown:   scoreParts,
			ConfidenceBreakdown: confParts,
		},
	}
	explainConsumption(&res, match)
	return res
}

func qualityResult(id string, rs []Reading, b Baseline, events []Event, q QualityResult, cfg config.Engine) AnomalyResult {
	start, end := rs[q.StartIdx].Timestamp, rs[q.EndIdx].Timestamp
	info := q.Info
	matches := CorroborateQuality(events, start, cfg)

	// Consumption stays within its baseline while the electrical readings misbehave.
	var kwh, expected []float64
	for k := q.StartIdx; k <= q.EndIdx; k++ {
		kwh = append(kwh, rs[k].ConsumptionKWh)
		expected = append(expected, b.Consumption[rs[k].Timestamp.Hour()].Median)
	}
	m := baseMetrics(rs, b)
	m.WindowVariation = round((mean(kwh)/mean(expected)-1)*100, 1)

	ongoing := q.EndIdx >= len(rs)-qualityOngoingHours
	info.FlagHours = 0 // count only the flagged hours inside the window
	for _, k := range q.Flagged {
		if k >= q.StartIdx && k <= q.EndIdx {
			info.FlagHours++
		}
	}
	flagFraction := float64(info.FlagHours) / float64(info.WindowHours)

	rules := []string{
		fmt.Sprintf("Flags de calidad recurrentes: %d horas con flags en %d h (máx. %d en 24 h)", info.FlagHours, info.WindowHours, info.FlagsIn24h),
		fmt.Sprintf("Consumo sin cambio persistente (variación media %s)", signedPct(m.WindowVariation)),
	}
	if len(matches) > 0 {
		rules = append(rules, "Un evento DATA_QUALITY corrobora el hallazgo (no lo origina)")
	}

	signals := len(info.ByKind)
	clarity := 0.5
	if len(matches) > 0 {
		signals++
		clarity = matches[0].Quality
	}
	conf, confParts := confidence(confidenceInput{Strength: float64(info.FlagsIn24h) / 20, Signals: signals, Clarity: clarity})
	score, scoreParts := priorityScore(scoreInput{
		Type: DataQuality, Magnitude: flagFraction * 100, Hours: info.WindowHours, Ongoing: ongoing,
		Electrical: true, Unexplained: true,
	})

	res := AnomalyResult{
		MeterID: id, Anomaly: true, Type: DataQuality, Severity: High, Confidence: conf, PriorityScore: score,
		WindowStart: start, WindowEnd: end,
		Evidence: Evidence{
			Kind:                "data_quality",
			Window:              WindowInfo{Start: start, End: end, Hours: info.WindowHours, Ongoing: ongoing},
			Metrics:             m,
			ChangedVariables:    []Change{},
			Events:              matches,
			RulesFired:          rules,
			Quality:             &info,
			PriorityBreakdown:   scoreParts,
			ConfidenceBreakdown: confParts,
		},
	}
	explainQuality(&res)
	return res
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
