package engine

import (
	"regexp"
	"strconv"
	"time"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
)

const (
	EventScheduledOutage   = "SCHEDULED_OUTAGE"
	EventOperationalChange = "OPERATIONAL_CHANGE"
	EventDataQuality       = "DATA_QUALITY"
	EventUnknown           = "UNKNOWN"
)

// EventMatch says whether a nearby event actually explains a window (compatibility, not just timing).
type EventMatch struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	Compatible  bool      `json:"compatible"`
	Quality     float64   `json:"quality"` // 0..1, how well it explains the window
	Reason      string    `json:"reason"`
}

var hoursRe = regexp.MustCompile(`(?i)(\d+)\s*(?:hours?|hrs?|h)\b`)

// outageHours extracts the outage length from a description such as "outage for 12 hours".
func outageHours(desc string) (int, bool) {
	m := hoursRe.FindStringSubmatch(desc)
	if m == nil {
		return 0, false
	}
	h, err := strconv.Atoi(m[1])
	return h, err == nil
}

// nearWindow reports whether an event falls in [start - lookback, start + lookahead].
func nearWindow(e Event, start time.Time, cfg config.Engine) bool {
	from := start.Add(-time.Duration(cfg.EventLookbackHours) * time.Hour)
	to := start.Add(time.Duration(cfg.EventLookaheadHours) * time.Hour)
	return !e.Timestamp.Before(from) && !e.Timestamp.After(to)
}

// MatchConsumptionEvents evaluates the events near a consumption window. Only the event type and
// description are used to judge compatibility with the observed shape of the window.
func MatchConsumptionEvents(events []Event, w Window, start time.Time, deteriorated bool, cfg config.Engine) []EventMatch {
	var out []EventMatch
	for _, e := range events {
		if !nearWindow(e, start, cfg) {
			continue
		}
		m := EventMatch{Type: e.Type, Timestamp: e.Timestamp, Description: e.Description}
		switch e.Type {
		case EventScheduledOutage:
			m.Compatible, m.Quality, m.Reason = matchOutage(e, w, cfg)
		case EventOperationalChange:
			switch {
			case w.Direction < 0:
				m.Reason = "un cambio operativo no explica una caída de consumo"
			case deteriorated:
				m.Reason = "hay deterioro eléctrico (FP o voltaje) que un cambio operativo no explica"
			default:
				m.Compatible, m.Quality = true, 0.9
				m.Reason = "escalón persistente sin deterioro eléctrico, coherente con un cambio operativo"
			}
		case EventDataQuality:
			m.Reason = "un evento de calidad de datos no explica un cambio de consumo"
		default:
			m.Reason = "evento sin información operativa que explique el cambio"
		}
		out = append(out, m)
	}
	return out
}

func matchOutage(e Event, w Window, cfg config.Engine) (bool, float64, string) {
	if w.Direction > 0 {
		return false, 0, "una parada no explica un aumento de consumo"
	}
	if w.Ongoing || !w.ReturnsToBaseline {
		return false, 0, "el consumo no volvió al baseline tras la parada"
	}
	hours, ok := outageHours(e.Description)
	if !ok {
		return false, 0, "la descripción del evento no indica la duración de la parada"
	}
	diff := w.Hours - hours
	if diff < 0 {
		diff = -diff
	}
	if diff > cfg.OutageToleranceHours {
		return false, 0, "la duración observada no coincide con la parada descrita"
	}
	quality := 1 - 0.15*float64(diff)
	return true, quality, "caída de consumo con la duración de la parada programada y retorno al baseline"
}

// CorroborateQuality returns the DATA_QUALITY events near a quality window. They only corroborate
// a finding already made from the readings; they never create it.
func CorroborateQuality(events []Event, start time.Time, cfg config.Engine) []EventMatch {
	var out []EventMatch
	for _, e := range events {
		if e.Type == EventDataQuality && nearWindow(e, start, cfg) {
			out = append(out, EventMatch{
				Type: e.Type, Timestamp: e.Timestamp, Description: e.Description,
				Compatible: true, Quality: 0.8,
				Reason: "el evento reportado corrobora el problema detectado en las lecturas",
			})
		}
	}
	return out
}
