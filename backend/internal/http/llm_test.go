package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Seb0sti0n/Ohmline/backend/internal/config"
)

const fakeAnswer = `{"reason":"Frase del modelo sobre el 12 sep.","explanation":"Explicación del modelo. Ventana de 14 h.","recommended_action":"Acción del modelo."}`

// fakeLLM serves an OpenAI-compatible endpoint. status != 200 makes it fail.
func fakeLLM(t *testing.T, status int) (config.LLM, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		b, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": fakeAnswer}}}})
		_, _ = w.Write(b)
	}))
	t.Cleanup(srv.Close)
	return config.LLM{BaseURL: srv.URL, APIKey: "k", Model: "m", Timeout: 2 * time.Second}, &calls
}

type llmAnomaly struct {
	MeterID       string  `json:"meter_id"`
	Type          string  `json:"type"`
	Severity      string  `json:"severity"`
	Confidence    float64 `json:"confidence"`
	PriorityScore float64 `json:"priority_score"`
	Source        string  `json:"explanation_source"`
	Reason        string  `json:"reason"`
	Explanation   string  `json:"explanation"`
	Action        string  `json:"recommended_action"`
}

type llmSummary struct {
	Status  string
	Summary struct {
		Anomalies    int
		LLMEnabled   bool `json:"llm_enabled"`
		LLMExplained int  `json:"llm_explanations"`
	}
}

// The classification is the engine's regardless of who writes the text.
var expected = []struct{ meter, typ, sev string }{
	{"M-109", "REAL_ANOMALY", "HIGH"}, {"M-112", "DATA_QUALITY", "HIGH"},
	{"M-104", "EXPLAINABLE_ANOMALY", "MEDIUM"}, {"M-106", "FALSE_POSITIVE", "LOW"},
}

func checkClassification(t *testing.T, list []llmAnomaly) {
	t.Helper()
	if len(list) != len(expected) {
		t.Fatalf("got %d anomalies", len(list))
	}
	for i, w := range expected {
		if a := list[i]; a.MeterID != w.meter || a.Type != w.typ || a.Severity != w.sev {
			t.Errorf("position %d = %s %s %s, want %s %s %s", i+1, a.MeterID, a.Type, a.Severity, w.meter, w.typ, w.sev)
		}
	}
}

func TestAnalysisUsesTheLLMForTheTexts(t *testing.T) {
	cfg, calls := fakeLLM(t, http.StatusOK)
	e := newEnvLLM(t, 0, cfg)
	e.analyze()

	list := decode[[]llmAnomaly](t, e.mustGet("/api/anomalies", 200))
	checkClassification(t, list)
	for _, a := range list {
		if a.Source != "LLM" || a.Reason != "Frase del modelo sobre el 12 sep." || a.Action != "Acción del modelo." {
			t.Errorf("%s: text not written by the LLM: %+v", a.MeterID, a)
		}
		if a.Confidence <= 0 || a.PriorityScore <= 0 {
			t.Errorf("%s: scores lost", a.MeterID)
		}
	}
	if calls.Load() != 4 {
		t.Errorf("LLM calls = %d, want 4 (one per anomaly)", calls.Load())
	}
	r := decode[llmSummary](t, e.mustGet("/api/ai/analysis/latest", 200))
	if !r.Summary.LLMEnabled || r.Summary.LLMExplained != 4 {
		t.Errorf("summary = %+v", r.Summary)
	}
	// The detail keeps the engine's evidence next to the LLM text.
	var first struct {
		ID       int
		Evidence map[string]any
	}
	all := decode[[]struct{ ID int }](t, e.mustGet("/api/anomalies", 200))
	first.ID = all[0].ID
	first = decode[struct {
		ID       int
		Evidence map[string]any
	}](t, e.mustGet("/api/anomalies/"+itoa(first.ID), 200))
	if _, ok := first.Evidence["rules_fired"]; !ok {
		t.Errorf("evidence missing after LLM rewrite")
	}
}

// If the provider is down the demo still works: the run completes with the engine's templates.
func TestAnalysisFallsBackToTemplatesWhenTheLLMFails(t *testing.T) {
	cfg, calls := fakeLLM(t, http.StatusInternalServerError)
	e := newEnvLLM(t, 0, cfg)
	r := e.analyze()
	if r.Status != "COMPLETED" {
		t.Fatalf("analysis should complete even if the LLM fails: %+v", r)
	}
	list := decode[[]llmAnomaly](t, e.mustGet("/api/anomalies", 200))
	checkClassification(t, list)
	for _, a := range list {
		if a.Source != "TEMPLATE" || a.Reason == "Frase del modelo sobre el 12 sep." || a.Reason == "" || a.Explanation == "" {
			t.Errorf("%s: expected template text, got %+v", a.MeterID, a)
		}
	}
	if calls.Load() == 0 {
		t.Errorf("the LLM should have been tried")
	}
	s := decode[llmSummary](t, e.mustGet("/api/ai/analysis/latest", 200))
	if !s.Summary.LLMEnabled || s.Summary.LLMExplained != 0 {
		t.Errorf("summary = %+v", s.Summary)
	}
}

func TestAnalysisWithoutAPIKeyNeverCallsOut(t *testing.T) {
	cfg, calls := fakeLLM(t, http.StatusOK)
	cfg.APIKey = ""
	e := newEnvLLM(t, 0, cfg)
	e.analyze()
	list := decode[[]llmAnomaly](t, e.mustGet("/api/anomalies", 200))
	checkClassification(t, list)
	for _, a := range list {
		if a.Source != "TEMPLATE" {
			t.Errorf("%s source = %s", a.MeterID, a.Source)
		}
	}
	if calls.Load() != 0 {
		t.Errorf("LLM was called %d times without an API key", calls.Load())
	}
	s := decode[llmSummary](t, e.mustGet("/api/ai/analysis/latest", 200))
	if s.Summary.LLMEnabled || s.Summary.LLMExplained != 0 {
		t.Errorf("summary = %+v", s.Summary)
	}
}
