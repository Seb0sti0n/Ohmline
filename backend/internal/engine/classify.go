package engine

// classifyConsumption decides the type and severity of a persistent consumption window from the
// event matches. An event only explains the window if it is compatible with its shape.
func classifyConsumption(matches []EventMatch) (AnomalyType, Severity, *EventMatch) {
	for i := range matches {
		m := &matches[i]
		if !m.Compatible {
			continue
		}
		switch m.Type {
		case EventScheduledOutage:
			return FalsePositive, Low, m
		case EventOperationalChange:
			return ExplainableAnomaly, Medium, m
		}
	}
	return RealAnomaly, High, nil
}
