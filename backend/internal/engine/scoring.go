package engine

import "math"

// Priority weights (sum = 100) and the saturation points of each component.
const (
	wMagnitude    = 25.0
	wPersistence  = 15.0
	wElectrical   = 20.0
	wUnexplained  = 20.0
	wExtraKWh     = 20.0
	magnitudeCap  = 100.0  // % variation at which magnitude saturates
	persistenceH  = 48.0   // hours at which persistence saturates
	extraKWhCap   = 1500.0 // kWh at which the extra-energy component saturates
	falsePosLimit = 19.0   // FALSE_POSITIVE is always forced below 20
)

// priorityWeights and confidenceWeights are also published in the evidence, so the UI can show each
// component of a score against its maximum without duplicating these numbers.
var (
	priorityWeights = map[string]float64{
		"magnitude": wMagnitude, "persistence": wPersistence, "electrical": wElectrical,
		"unexplained": wUnexplained, "extra_energy": wExtraKWh,
	}
	confidenceWeights = map[string]float64{
		"base": 0.40, "signal_strength": 0.25, "independent_signals": 0.20, "explanation_clarity": 0.10,
	}
)

func copyWeights(m map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

type scoreInput struct {
	Type        AnomalyType
	Magnitude   float64 // |variation %| for consumption shifts, flagged-hours fraction*100 for data quality
	Hours       int
	Ongoing     bool
	Electrical  bool    // electrical deterioration or unreliable electrical readings
	Unexplained bool    // no compatible operational event
	ExtraKWh    float64 // absolute
}

// priorityScore is a weighted sum of magnitude, persistence (and whether it is still going),
// electrical deterioration, lack of explanation and extra energy. Returns the score and its parts.
func priorityScore(in scoreInput) (float64, map[string]float64) {
	parts := map[string]float64{
		"magnitude":    wMagnitude * clamp01(in.Magnitude/magnitudeCap),
		"persistence":  wPersistence * 2 / 3 * clamp01(float64(in.Hours)/persistenceH),
		"electrical":   0,
		"unexplained":  0,
		"extra_energy": wExtraKWh * clamp01(math.Abs(in.ExtraKWh)/extraKWhCap),
	}
	if in.Ongoing {
		parts["persistence"] += wPersistence / 3
	}
	if in.Electrical {
		parts["electrical"] = wElectrical
	}
	if in.Unexplained {
		parts["unexplained"] = wUnexplained
	}
	total := 0.0
	for k, v := range parts {
		parts[k] = round(v, 1)
		total += v
	}
	if in.Type == FalsePositive && total > falsePosLimit {
		total = falsePosLimit
		parts["capped_false_positive"] = falsePosLimit
	}
	return round(math.Min(total, 100), 1), parts
}

type confidenceInput struct {
	Strength float64 // 0..1, how strong the main signal is
	Signals  int     // independent signals that agree
	Clarity  float64 // 0..1, how clear the explanation (or lack of it) is
}

// confidence combines signal strength, the number of independent signals and the clarity of the
// event match. It is capped at 0.99: the engine never claims certainty.
func confidence(in confidenceInput) (float64, map[string]float64) {
	parts := map[string]float64{
		"base":                confidenceWeights["base"],
		"signal_strength":     confidenceWeights["signal_strength"] * clamp01(in.Strength),
		"independent_signals": confidenceWeights["independent_signals"] * clamp01(float64(in.Signals)/4),
		"explanation_clarity": confidenceWeights["explanation_clarity"] * clamp01(in.Clarity),
	}
	total := 0.0
	for k, v := range parts {
		parts[k] = round(v, 3)
		total += v
	}
	return round(math.Min(total, 0.99), 2), parts
}
