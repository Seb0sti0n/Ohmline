package llm

import (
	"encoding/json"
	"fmt"

	"github.com/Seb0sti0n/Ohmline/backend/internal/engine"
)

const systemPrompt = `Eres un analista energético. Recibes el resultado de un motor de detección de anomalías sobre un medidor eléctrico y debes redactar la explicación para el operador, en español.

Reglas:
- El motor ya decidió el tipo, la severidad, la prioridad y la confianza. NO los cambies ni los cuestiones; solo redacta.
- Usa únicamente datos del JSON. Cita cifras concretas tomadas tal cual del JSON, sin calcular otras nuevas ni mezclar cifras de ventanas distintas.
- No inventes causas, equipos ni eventos que no estén en el JSON. Si no hay un evento que lo explique, dilo.
- Escribe para un operador, no para un analista de datos: no uses nombres de campos del JSON (consumption_kwh, current_a...), ni z-score, MAD ni desviaciones estándar.
- Cifras: coma decimal y punto de miles (2.825 kWh, +110,5%). Fechas como "12 sep, 14:00", nunca en formato ISO.
- Cada frase termina en punto, incluida "reason".
- "recommended_action" debe conservar el sentido de "default_recommended_action"; puedes precisarla con los datos.

Cómo leer los datos:
- evidence.metrics: baseline_daily_kwh es el consumo diario esperado; last_day_kwh y last_day_variation_pct son el último día; window_variation_pct y extra_kwh son la variación media y la energía extra dentro de la ventana anómala.
- evidence.changed_variables: cómo cambiaron corriente, factor de potencia y voltaje en la ventana frente a su baseline (solo las "significant": true importan).
- evidence.quality (problemas de calidad de datos): flagged_hours son las horas con algún indicador de calidad dentro de window_hours; by_kind cuenta cada indicador por tipo (saltos de voltaje, saltos de factor de potencia, voltaje fuera de rango, kWh que no cuadra con V·I·FP).

Responde SOLO con un objeto JSON con estas claves:
- "reason": una sola frase que resume el hallazgo.
- "explanation": como máximo 4 frases: qué cambió, qué variables lo respaldan y si hay un evento que lo explique.
- "recommended_action": una frase con la acción para el operador.`

// typeMeaning tells the model what the engine's classification means, so it does not have to guess.
var typeMeaning = map[engine.AnomalyType]string{
	engine.RealAnomaly:        "anomalía real: cambio persistente de consumo sin evento operativo que lo explique",
	engine.ExplainableAnomaly: "anomalía explicable: cambio persistente coherente con un evento operativo",
	engine.FalsePositive:      "falso positivo: desviación temporal explicada por un evento programado y ya normalizada",
	engine.DataQuality:        "problema de calidad de datos: lecturas eléctricas inconsistentes con el consumo estable",
}

type promptInput struct {
	MeterID                  string             `json:"meter_id"`
	Type                     engine.AnomalyType `json:"type"`
	TypeMeaning              string             `json:"type_meaning"`
	Severity                 engine.Severity    `json:"severity"`
	Confidence               float64            `json:"confidence"`
	PriorityScore            float64            `json:"priority_score"`
	DefaultRecommendedAction string             `json:"default_recommended_action"`
	Evidence                 engine.Evidence    `json:"evidence"`
}

// buildPrompt returns the messages and the set of numbers the model is allowed to cite.
func buildPrompt(r engine.AnomalyResult) (system, user string, allowed *numberSet, err error) {
	in := promptInput{
		MeterID: r.MeterID, Type: r.Type, TypeMeaning: typeMeaning[r.Type], Severity: r.Severity,
		Confidence: r.Confidence, PriorityScore: r.PriorityScore,
		DefaultRecommendedAction: r.RecommendedAction, Evidence: r.Evidence,
	}
	allowed, err = numbersIn(in) // computed from the full evidence, before trimming the prompt
	if err != nil {
		return "", "", nil, err
	}
	// The score breakdowns explain how the engine ranked things; the model does not need them, and
	// on a free tier every token counts (requests per minute are limited by tokens).
	in.Evidence.PriorityBreakdown, in.Evidence.ConfidenceBreakdown = nil, nil
	in.Evidence.PriorityWeights, in.Evidence.ConfidenceWeights = nil, nil
	b, err := json.Marshal(in)
	if err != nil {
		return "", "", nil, err
	}
	return systemPrompt, fmt.Sprintf("Datos del motor:\n%s", b), allowed, nil
}
