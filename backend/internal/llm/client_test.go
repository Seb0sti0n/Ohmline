package llm

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Seb0sti0n/Ohmline/backend/internal/config"
	"github.com/Seb0sti0n/Ohmline/backend/internal/engine"
)

func TestExplainSendsAWellFormedRequest(t *testing.T) {
	var auth, ctype string
	f := newFakeAPI(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		auth, ctype = r.Header.Get("Authorization"), r.Header.Get("Content-Type")
		reply(goodContent)(0, w, r)
	})
	rw, err := f.client(func(c *config.LLM) { c.ReasoningEffort = "low" }).Explain(context.Background(), sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	if rw.Reason == "" || rw.Explanation == "" || rw.RecommendedAction == "" {
		t.Errorf("rewrite incomplete: %+v", rw)
	}
	if auth != "Bearer test-key" || ctype != "application/json" {
		t.Errorf("headers: auth=%q content-type=%q", auth, ctype)
	}
	body := f.last.Load().(string)
	for _, want := range []string{`"model":"test-model"`, `"response_format":{"type":"json_object"}`, `"reasoning_effort":"low"`,
		`"role":"system"`, `\"meter_id\":\"M-109\"`, `\"type\":\"REAL_ANOMALY\"`, "110.5", "0.74"} {
		if !strings.Contains(body, want) {
			t.Errorf("request body is missing %s", want)
		}
	}
	if strings.Contains(body, "priority_breakdown\":{") {
		t.Errorf("score breakdowns should not be sent to the model")
	}
	// The API key must never travel in the prompt.
	if strings.Contains(body, "test-key") {
		t.Errorf("API key leaked into the request body")
	}
}

func TestReasoningEffortIsOptional(t *testing.T) {
	f := newFakeAPI(t, reply(goodContent))
	if _, err := f.client().Explain(context.Background(), sampleResult()); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(f.last.Load().(string), "reasoning_effort") {
		t.Errorf("reasoning_effort must not be sent when not configured")
	}
}

func TestAnswersThatAreRejected(t *testing.T) {
	long := strings.Repeat("Texto largo con cifras 12 y 14. ", 40)
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{"not JSON", "Claro, aquí está mi análisis.", "not valid JSON"},
		{"missing field", `{"reason":"El consumo subió 110,5% el 12 sep.","explanation":"Algo."}`, "missing"},
		{"empty field", `{"reason":"","explanation":"a 12","recommended_action":"b 12"}`, "missing"},
		{"reason with two sentences", `{"reason":"Subió 110,5%. Es grave.","explanation":"Cayó el FP a 0,74.","recommended_action":"Revisar."}`, "single short sentence"},
		{"explanation with five sentences", `{"reason":"Subió 110,5%.","explanation":"Uno 12. Dos 14. Tres 12. Cuatro 14. Cinco 12.","recommended_action":"Revisar."}`, "longer than"},
		{"explanation too long", `{"reason":"Subió 110,5%.","explanation":"` + long + `","recommended_action":"Revisar."}`, "longer than"},
		{"cites no numbers", `{"reason":"El consumo subió mucho.","explanation":"Hay un cambio grande.","recommended_action":"Revisar el medidor."}`, "no numbers"},
		{"invented integer", `{"reason":"El consumo subió 110,5% en 37 horas.","explanation":"Ok 12.","recommended_action":"Revisar."}`, `"37"`},
		{"invented decimal", `{"reason":"El consumo subió 173,2%.","explanation":"Ok 12.","recommended_action":"Revisar."}`, "173,2"},
		{"invented thousands", `{"reason":"Acumula 987.654 kWh extra.","explanation":"Ok 12.","recommended_action":"Revisar."}`, "987.654"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeAPI(t, reply(tt.content))
			_, err := f.client().Explain(context.Background(), sampleResult())
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
			}
			if got := f.calls.Load(); got != maxAttempts {
				t.Errorf("made %d calls, want %d (one retry)", got, maxAttempts)
			}
		})
	}
}

func TestAnswersThatAreAccepted(t *testing.T) {
	tests := map[string]string{
		"plain":          goodContent,
		"json fence":     "```json\n" + goodContent + "\n```",
		"preamble":       "Aquí tienes el JSON:\n" + goodContent,
		"en-US numbers":  `{"reason":"El consumo subió 110.5% el 12 sep.","explanation":"Pasó de 1,047.7 a 2,207.6 kWh.","recommended_action":"Investigar el medidor."}`,
		"percent of pf":  `{"reason":"El FP cayó de 0,94 a 0,74, un 94% a 74%.","explanation":"Ok 12.","recommended_action":"Revisar."}`,
		"rounded values": `{"reason":"El consumo subió 111% desde el 12 sep, 14:00.","explanation":"La corriente pasó de 202 A a 425 A.","recommended_action":"Revisar."}`,
		"year and month": `{"reason":"Anomalía desde el 12 de septiembre de 2026.","explanation":"Dura 58 h.","recommended_action":"Revisar."}`,
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFakeAPI(t, reply(content))
			if _, err := f.client().Explain(context.Background(), sampleResult()); err != nil {
				t.Errorf("rejected a valid answer: %v", err)
			}
		})
	}
}

func TestRetryRecoversFromABadFirstAnswer(t *testing.T) {
	f := newFakeAPI(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			reply("no es JSON")(n, w, r)
			return
		}
		reply(goodContent)(n, w, r)
	})
	if _, err := f.client().Explain(context.Background(), sampleResult()); err != nil {
		t.Fatalf("second attempt should succeed: %v", err)
	}
	if f.calls.Load() != 2 {
		t.Errorf("calls = %d, want 2", f.calls.Load())
	}
}

func TestHTTPFailures(t *testing.T) {
	status := func(code int, header map[string]string) func(int, http.ResponseWriter, *http.Request) {
		return func(_ int, w http.ResponseWriter, _ *http.Request) {
			for k, v := range header {
				w.Header().Set(k, v)
			}
			w.WriteHeader(code)
			_, _ = w.Write([]byte(`{"error":{"message":"nope"}}`))
		}
	}
	tests := []struct {
		name      string
		handler   func(int, http.ResponseWriter, *http.Request)
		wantCalls int32
		wantErr   string
	}{
		{"server error is retried once", status(500, nil), 2, "HTTP 500"},
		{"bad request is retried once", status(400, nil), 2, "HTTP 400"},
		{"invalid key is not retried", status(401, nil), 1, "HTTP 401"},
		{"unknown model is not retried", status(404, nil), 1, "HTTP 404"},
		{"long rate limit is not waited for", status(429, map[string]string{"Retry-After": "60"}), 1, "429"},
		{"short rate limit is retried", status(429, map[string]string{"Retry-After": "0"}), 2, "429"},
		{"empty choices", func(_ int, w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"choices":[]}`)) }, 2, "unexpected response"},
		{"garbage body", func(_ int, w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`<html>`)) }, 2, "unexpected response"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeAPI(t, tt.handler)
			t0 := time.Now()
			_, err := f.client().Explain(context.Background(), sampleResult())
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if f.calls.Load() != tt.wantCalls {
				t.Errorf("calls = %d, want %d", f.calls.Load(), tt.wantCalls)
			}
			if time.Since(t0) > time.Second {
				t.Errorf("took %v: a long Retry-After must not be waited for", time.Since(t0))
			}
		})
	}
}

func TestShortRateLimitThenSuccess(t *testing.T) {
	f := newFakeAPI(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		reply(goodContent)(n, w, r)
	})
	if _, err := f.client().Explain(context.Background(), sampleResult()); err != nil {
		t.Fatalf("should recover after a short 429: %v", err)
	}
}

func TestTimeout(t *testing.T) {
	f := newFakeAPI(t, func(_ int, w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	})
	t0 := time.Now()
	_, err := f.client(func(c *config.LLM) { c.Timeout = 80 * time.Millisecond }).Explain(context.Background(), sampleResult())
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if time.Since(t0) > time.Second {
		t.Errorf("timeout not enforced: took %v", time.Since(t0))
	}
}

func TestCancelledContextStopsRetrying(t *testing.T) {
	f := newFakeAPI(t, reply("nope"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.client().Explain(ctx, sampleResult()); err == nil {
		t.Fatal("expected an error")
	}
	if f.calls.Load() > 1 {
		t.Errorf("calls = %d after cancellation", f.calls.Load())
	}
}

// ---- Enhance ----

func TestEnhanceDisabledWithoutKeyMakesNoCalls(t *testing.T) {
	f := newFakeAPI(t, reply(goodContent))
	c := f.client(func(c *config.LLM) { c.APIKey = "" })
	if c.Enabled() {
		t.Fatal("client without key must be disabled")
	}
	rs := []engine.AnomalyResult{sampleResult()}
	if n := c.Enhance(context.Background(), rs); n != 0 || f.calls.Load() != 0 {
		t.Errorf("enhanced %d, calls %d: must do nothing without a key", n, f.calls.Load())
	}
	if rs[0].ExplanationSource != engine.SourceTemplate || rs[0].Reason != "template reason" {
		t.Errorf("result changed without a key: %+v", rs[0])
	}
	var nilClient *Client
	if nilClient.Enabled() {
		t.Error("nil client must be disabled")
	}
}

// The model only writes text: whatever it returns, the classification stays exactly as decided.
func TestEnhanceNeverChangesTheClassification(t *testing.T) {
	// The model "argues" with the engine in its answer, and only M-109 gets a valid one.
	var f *fakeAPI
	f = newFakeAPI(t, func(n int, w http.ResponseWriter, r *http.Request) {
		if strings.Contains(f.last.Load().(string), `M-109`) {
			reply(`{"reason":"Esto es un falso positivo de severidad LOW, 110,5%.","explanation":"Confianza 0,94 pero lo considero un falso positivo.","recommended_action":"No escalar."}`)(n, w, r)
			return
		}
		reply("not json")(n, w, r)
	})
	other := sampleResult()
	other.MeterID, other.Type, other.Severity = "M-104", engine.ExplainableAnomaly, engine.Medium
	rs := []engine.AnomalyResult{sampleResult(), other}
	before := []engine.AnomalyResult{sampleResult(), other}

	n := f.client().Enhance(context.Background(), rs)
	if n != 1 {
		t.Fatalf("enhanced %d, want 1", n)
	}
	if rs[0].ExplanationSource != engine.SourceLLM || !strings.Contains(rs[0].Reason, "falso positivo") {
		t.Errorf("M-109 text was not applied: %+v", rs[0])
	}
	if rs[1].ExplanationSource != engine.SourceTemplate || rs[1].Reason != "template reason" {
		t.Errorf("M-104 must keep its template: %+v", rs[1])
	}
	for i := range rs {
		a, b := rs[i], before[i]
		if a.MeterID != b.MeterID || a.Type != b.Type || a.Severity != b.Severity || a.Confidence != b.Confidence ||
			a.PriorityScore != b.PriorityScore || !a.WindowStart.Equal(b.WindowStart) || !reflect.DeepEqual(a.Evidence, b.Evidence) {
			t.Errorf("%s: classification, scores or evidence changed", a.MeterID)
		}
	}
}

func TestEnhanceRunsInParallel(t *testing.T) {
	f := newFakeAPI(t, func(n int, w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		reply(goodContent)(n, w, r)
	})
	rs := []engine.AnomalyResult{sampleResult(), sampleResult(), sampleResult(), sampleResult()}
	t0 := time.Now()
	if n := f.client().Enhance(context.Background(), rs); n != 4 {
		t.Fatalf("enhanced %d, want 4", n)
	}
	if d := time.Since(t0); d > 600*time.Millisecond {
		t.Errorf("4 calls of 200 ms took %v: they should run in parallel", d)
	}
}
