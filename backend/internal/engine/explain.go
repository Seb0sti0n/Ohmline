package engine

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Deterministic Spanish templates. They are good enough to stand alone; an LLM may rewrite the
// text later but never changes the classification.

var monthsES = [...]string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}

// num formats with es-CO conventions: thousands with ".", decimals with ",".
func num(v float64, decimals int) string {
	s := fmt.Sprintf("%.*f", decimals, math.Abs(v))
	intPart, frac, _ := strings.Cut(s, ".")
	var b strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	out := b.String()
	if frac != "" {
		out += "," + frac
	}
	if v < 0 && strings.Trim(s, "0.") != "" {
		out = "-" + out
	}
	return out
}

func signedPct(v float64) string {
	s := num(v, 1)
	if v > 0 {
		s = "+" + s
	}
	return s + "%"
}

func when(t time.Time) string {
	return fmt.Sprintf("%d %s, %02d:%02d", t.Day(), monthsES[t.Month()-1], t.Hour(), t.Minute())
}

func changeOf(cs []Change, variable string) (Change, bool) {
	for _, c := range cs {
		if c.Variable == variable {
			return c, true
		}
	}
	return Change{}, false
}

// electricalSentence summarizes the significant electrical changes, or says they stayed stable.
func electricalSentence(cs []Change) string {
	var parts []string
	if c, ok := changeOf(cs, VarCurrent); ok && c.Significant {
		parts = append(parts, fmt.Sprintf("la corriente cambió %s", signedPct(c.DeltaPct)))
	}
	if c, ok := changeOf(cs, VarPowerFactor); ok && c.Significant {
		verb := "subió"
		if c.Delta < 0 {
			verb = "cayó"
		}
		parts = append(parts, fmt.Sprintf("el factor de potencia %s de %s a %s", verb, num(c.Baseline, 2), num(c.Observed, 2)))
	}
	if c, ok := changeOf(cs, VarVoltage); ok && c.Significant {
		verb := "subió"
		if c.Delta < 0 {
			verb = "bajó"
		}
		parts = append(parts, fmt.Sprintf("el voltaje %s %s V", verb, num(math.Abs(c.Delta), 1)))
	} else if c, ok := changeOf(cs, VarVoltageStd); ok && c.Significant {
		parts = append(parts, "el voltaje se volvió más inestable")
	}
	if len(parts) == 0 {
		return "La corriente, el factor de potencia y el voltaje se mantienen estables."
	}
	s := strings.Join(parts, "; ")
	return strings.ToUpper(s[:1]) + s[1:] + "."
}

// explainConsumption writes reason, explanation and action for a persistent consumption window.
func explainConsumption(r *AnomalyResult, match *EventMatch) {
	e := &r.Evidence
	m := e.Metrics
	dir := "por encima"
	if m.WindowVariation < 0 {
		dir = "por debajo"
	}
	span := fmt.Sprintf("desde el %s", when(r.WindowStart))
	if !e.Window.Ongoing {
		span = fmt.Sprintf("durante %d h desde el %s", e.Window.Hours, when(r.WindowStart))
	}
	cons, _ := changeOf(e.ChangedVariables, VarConsumption)

	s1 := fmt.Sprintf("El consumo por hora pasó de %s a %s kWh (%s) %s", num(cons.Baseline, 1), num(cons.Observed, 1), signedPct(cons.DeltaPct), span)
	if e.Window.Ongoing {
		s1 += " y sigue fuera del baseline."
	} else {
		s1 += " y luego volvió al baseline."
	}
	sentences := []string{s1, electricalSentence(e.ChangedVariables)}

	switch r.Type {
	case RealAnomaly:
		r.Reason = fmt.Sprintf("Consumo %s %s del baseline %s, sin evento conocido que lo explique.", num(math.Abs(m.WindowVariation), 1)+"%", dir, span)
		if len(e.Events) > 0 {
			sentences = append(sentences, fmt.Sprintf("El evento cercano del %s dice «%s» y no explica el cambio.", when(e.Events[0].Timestamp), e.Events[0].Description))
		} else {
			sentences = append(sentences, "No hay ningún evento operativo registrado que lo explique.")
		}
		sentences = append(sentences, fmt.Sprintf("Acumula %s kWh por encima de lo esperado.", num(m.ExtraKWh, 0)))
		r.RecommendedAction = "Investigar medidor e instalación; revisar cargas conectadas"
		if pf, ok := changeOf(e.ChangedVariables, VarPowerFactor); ok && pf.Significant && pf.Delta < 0 {
			r.RecommendedAction += " y compensación reactiva (FP bajo)"
		}
		r.RecommendedAction += "."
	case ExplainableAnomaly:
		r.Reason = fmt.Sprintf("Consumo %s %s del baseline %s, coherente con el evento «%s».", num(math.Abs(m.WindowVariation), 1)+"%", dir, span, match.Description)
		sentences = append(sentences, fmt.Sprintf("Coincide con el evento del %s: %s.", when(match.Timestamp), match.Reason))
		sentences = append(sentences, fmt.Sprintf("Acumula %s kWh sobre el baseline anterior.", num(m.ExtraKWh, 0)))
		r.RecommendedAction = "Validar con operación y actualizar el baseline tras confirmar la nueva línea base."
	case FalsePositive:
		r.Reason = fmt.Sprintf("Consumo %s %s del baseline %s, explicado por la parada programada.", num(math.Abs(m.WindowVariation), 1)+"%", dir, span)
		sentences = append(sentences, fmt.Sprintf("Coincide con el evento del %s: %s.", when(match.Timestamp), match.Reason))
		r.RecommendedAction = "No escalar; registrar como explicado por la parada programada."
	}
	r.Explanation = strings.Join(sentences, " ")
}

// explainQuality writes the texts for a data quality finding.
func explainQuality(r *AnomalyResult) {
	q := r.Evidence.Quality
	r.Reason = fmt.Sprintf("Consumo estable, pero %d de %d horas con lecturas eléctricas inconsistentes desde el %s.", q.FlagHours, q.WindowHours, when(r.WindowStart))

	var kinds []string
	if n := q.ByKind[FlagVoltageJump]; n > 0 {
		kinds = append(kinds, fmt.Sprintf("%d saltos de voltaje de más de 10 V", n))
	}
	if n := q.ByKind[FlagVoltageRange]; n > 0 {
		kinds = append(kinds, fmt.Sprintf("%d lecturas de voltaje fuera de ±5%% del nominal", n))
	}
	if n := q.ByKind[FlagPFJump]; n > 0 {
		kinds = append(kinds, fmt.Sprintf("%d saltos bruscos del factor de potencia", n))
	}
	if n := q.ByKind[FlagCoherence]; n > 0 {
		kinds = append(kinds, fmt.Sprintf("%d lecturas donde el kWh no cuadra con V·I·FP", n))
	}
	sentences := []string{
		"El consumo se mantiene dentro de su baseline, por lo que no hay un cambio de carga real.",
		"Las lecturas eléctricas muestran " + strings.Join(kinds, ", ") + ".",
	}
	if len(r.Evidence.Events) > 0 {
		ev := r.Evidence.Events[0]
		sentences = append(sentences, fmt.Sprintf("El evento reportado el %s («%s») corrobora el hallazgo.", when(ev.Timestamp), ev.Description))
	}
	r.Explanation = strings.Join(sentences, " ")
	r.RecommendedAction = "Validar medidor y comunicaciones; no usar sus lecturas eléctricas para decisiones hasta verificar."
}
