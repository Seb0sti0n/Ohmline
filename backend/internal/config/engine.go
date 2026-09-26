package config

// Engine holds every threshold of the analysis engine. Values are tuned to the demo dataset;
// the README explains why each one was chosen.
type Engine struct {
	// Baseline
	BaselineDays int // reference window at the start of the series

	// Consumption detection
	ZThreshold      float64 // robust z-score needed to flag a reading
	MinPersistHours int     // consecutive flagged hours needed to call it a persistent change

	// Data quality flags
	VoltageNominal      float64 // volts
	VoltageTolerancePct float64 // allowed deviation from nominal
	VoltageJumpV        float64 // max change between consecutive hours
	PFJump              float64 // max power factor change between consecutive hours
	CoherenceTolPct     float64 // allowed deviation of kWh/(V·I·PF) from its baseline ratio
	QualityFlagsPer24h  int     // flags within 24 h that make the problem recurrent

	// Electrical correlation (what counts as a relevant change inside an anomalous window)
	CurrentChangePct float64 // |Δ current| in %
	PFDrop           float64 // power factor drop
	VoltageShiftV    float64 // |Δ mean voltage|
	VoltageStdRatio  float64 // window voltage std / baseline voltage std

	// Event association
	EventLookbackHours   int     // events up to this long before the window start
	EventLookaheadHours  int     // events up to this long after the window start
	OutageToleranceHours int     // allowed difference between described and observed outage length
	RecoveryTolPct       float64 // consumption after an outage must be within this of baseline
}

func DefaultEngine() Engine {
	return Engine{
		BaselineDays: 7,

		ZThreshold:      4,
		MinPersistHours: 6,

		VoltageNominal:      220,
		VoltageTolerancePct: 5,
		VoltageJumpV:        10,
		PFJump:              0.2,
		CoherenceTolPct:     30,
		QualityFlagsPer24h:  5,

		CurrentChangePct: 15,
		PFDrop:           0.05,
		VoltageShiftV:    1.5,
		VoltageStdRatio:  1.5,

		EventLookbackHours:   6,
		EventLookaheadHours:  2,
		OutageToleranceHours: 2,
		RecoveryTolPct:       15,
	}
}
