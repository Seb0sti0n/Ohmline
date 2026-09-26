// Package engine is the analysis core: pure functions that turn meter readings and events
// into classified, prioritized and explained anomalies. It has no HTTP or database dependencies.
package engine

import "time"

type Reading struct {
	MeterID        string
	Timestamp      time.Time
	ConsumptionKWh float64
	VoltageV       float64
	CurrentA       float64
	PowerFactor    float64
}

type Event struct {
	MeterID     string
	Timestamp   time.Time
	Type        string
	Description string
}

type AnomalyType string

const (
	RealAnomaly        AnomalyType = "REAL_ANOMALY"
	ExplainableAnomaly AnomalyType = "EXPLAINABLE_ANOMALY"
	FalsePositive      AnomalyType = "FALSE_POSITIVE"
	DataQuality        AnomalyType = "DATA_QUALITY"
)

type Severity string

const (
	High   Severity = "HIGH"
	Medium Severity = "MEDIUM"
	Low    Severity = "LOW"
)

// AnomalyResult is the engine output for one meter. Evidence is what backs the conclusion.
type AnomalyResult struct {
	MeterID           string      `json:"meter_id"`
	Anomaly           bool        `json:"anomaly"`
	Type              AnomalyType `json:"type"`
	Severity          Severity    `json:"severity"`
	Confidence        float64     `json:"confidence"`
	PriorityScore     float64     `json:"priority_score"`
	Reason            string      `json:"reason"`
	Explanation       string      `json:"explanation"`
	RecommendedAction string      `json:"recommended_action"`
	WindowStart       time.Time   `json:"window_start"`
	WindowEnd         time.Time   `json:"window_end"`
	Evidence          Evidence    `json:"evidence"`
	// ExplanationSource says who wrote reason/explanation/action: the engine templates or an LLM.
	ExplanationSource string `json:"explanation_source"`
}

const (
	SourceTemplate = "TEMPLATE"
	SourceLLM      = "LLM"
)

type Evidence struct {
	Kind                string             `json:"kind"` // "consumption_shift" or "data_quality"
	Window              WindowInfo         `json:"window"`
	Metrics             Metrics            `json:"metrics"`
	ChangedVariables    []Change           `json:"changed_variables"`
	Events              []EventMatch       `json:"events"`
	RulesFired          []string           `json:"rules_fired"`
	Quality             *QualityInfo       `json:"quality,omitempty"`
	PriorityBreakdown   map[string]float64 `json:"priority_breakdown"`
	ConfidenceBreakdown map[string]float64 `json:"confidence_breakdown"`
	// Maximum contribution of each component, to show the breakdowns as fractions.
	PriorityWeights   map[string]float64 `json:"priority_weights"`
	ConfidenceWeights map[string]float64 `json:"confidence_weights"`
}

type WindowInfo struct {
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Hours   int       `json:"hours"`
	Ongoing bool      `json:"ongoing"`
}

type Metrics struct {
	BaselineDailyKWh float64 `json:"baseline_daily_kwh"`
	LastDayKWh       float64 `json:"last_day_kwh"`
	LastDayVariation float64 `json:"last_day_variation_pct"`
	WindowVariation  float64 `json:"window_variation_pct"` // mean over the window vs baseline
	ExtraKWh         float64 `json:"extra_kwh"`            // window consumption minus baseline for the same hours
	MedianAbsZ       float64 `json:"median_abs_z"`
}
