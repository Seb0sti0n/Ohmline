package llm

import "testing"

func TestPromptIsCompact(t *testing.T) {
	_, user, _, err := buildPrompt(sampleResult())
	if err != nil {
		t.Fatal(err)
	}
	approxTokens := (len(systemPrompt) + len(user)) / 3 // conservative for Spanish + JSON
	t.Logf("prompt: %d chars ≈ %d tokens", len(systemPrompt)+len(user), approxTokens)
	if approxTokens > 1500 {
		t.Errorf("prompt too large (≈%d tokens): free-tier limits are per token", approxTokens)
	}
}
