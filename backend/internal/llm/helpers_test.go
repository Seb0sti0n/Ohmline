package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Seb0sti0n/Ohmline/backend/internal/config"
	"github.com/Seb0sti0n/Ohmline/backend/internal/engine"
)

var start = time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC)

// sampleResult is an M-109-like result: numbers here are the ones a model may cite.
func sampleResult() engine.AnomalyResult {
	end := time.Date(2026, 9, 14, 23, 0, 0, 0, time.UTC)
	return engine.AnomalyResult{
		MeterID: "M-109", Anomaly: true, Type: engine.RealAnomaly, Severity: engine.High,
		Confidence: 0.94, PriorityScore: 100,
		Reason: "template reason", Explanation: "template explanation", RecommendedAction: "Investigar medidor e instalación.",
		WindowStart: start, WindowEnd: end, ExplanationSource: engine.SourceTemplate,
		Evidence: engine.Evidence{
			Kind:   "consumption_shift",
			Window: engine.WindowInfo{Start: start, End: end, Hours: 58, Ongoing: true},
			Metrics: engine.Metrics{BaselineDailyKWh: 1047.7, LastDayKWh: 2207.6, LastDayVariation: 110.7,
				WindowVariation: 110.5, ExtraKWh: 2825, MedianAbsZ: 35},
			ChangedVariables: []engine.Change{
				{Variable: "current_a", Baseline: 202.058, Observed: 424.697, Delta: 222.639, DeltaPct: 110.2, Significant: true, Direction: "up"},
				{Variable: "power_factor", Baseline: 0.94, Observed: 0.74, Delta: -0.2, DeltaPct: -21.3, Significant: true, Direction: "down"},
			},
			Events:              []engine.EventMatch{{Type: "UNKNOWN", Timestamp: start, Description: "No operational event reported", Reason: "sin información"}},
			RulesFired:          []string{"Cambio persistente: 58 h consecutivas con |z| > 4 (z mediano 35,0)"},
			PriorityBreakdown:   map[string]float64{"magnitude": 25},
			ConfidenceBreakdown: map[string]float64{"base": 0.4},
		},
	}
}

// goodContent is a valid, grounded answer for sampleResult.
const goodContent = `{"reason":"El consumo subió 110,5% desde el 12 sep, 14:00 sin evento que lo explique.",` +
	`"explanation":"El consumo diario pasó de 1.047,7 kWh a 2.207,6 kWh. La corriente subió 110,2% y el factor de potencia cayó de 0,94 a 0,74. No hay ningún evento operativo que lo explique.",` +
	`"recommended_action":"Investigar medidor e instalación y revisar la compensación reactiva."}`

func chatBody(content string) string {
	b, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": content}}}})
	return string(b)
}

// fakeAPI is an OpenAI-compatible server. handler decides each response; calls counts requests.
type fakeAPI struct {
	*httptest.Server
	calls atomic.Int32
	last  atomic.Value // last request body (string)
}

func newFakeAPI(t *testing.T, handler func(call int, w http.ResponseWriter, r *http.Request)) *fakeAPI {
	t.Helper()
	f := &fakeAPI{}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(f.calls.Add(1))
		raw, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(raw)) // restore it: handler may want to read it too
		f.last.Store(string(raw))
		if r.URL.Path != "/chat/completions" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		handler(n, w, r)
	}))
	t.Cleanup(f.Close)
	return f
}

func reply(content string) func(int, http.ResponseWriter, *http.Request) {
	return func(_ int, w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(chatBody(content)))
	}
}

func (f *fakeAPI) client(mutate ...func(*config.LLM)) *Client {
	cfg := config.LLM{BaseURL: f.URL, APIKey: "test-key", Model: "test-model", Timeout: 2 * time.Second}
	for _, m := range mutate {
		m(&cfg)
	}
	return New(cfg)
}
