// Package llm asks an OpenAI-compatible chat API to rewrite the engine's explanation.
//
// The engine decides type, severity, priority and confidence. The model only writes the text,
// from the evidence, and its answer is validated (shape, length, and that every number it cites
// comes from the evidence). Any failure falls back to the engine's template text.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Seb0sti0n/Ohmline/backend/internal/config"
	"github.com/Seb0sti0n/Ohmline/backend/internal/engine"
)

type Client struct {
	cfg  config.LLM
	http *http.Client
}

func New(cfg config.LLM) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 8 * time.Second
	}
	return &Client{cfg: cfg, http: &http.Client{}}
}

// Enabled is false without an API key: the app then uses templates only and never calls out.
func (c *Client) Enabled() bool { return c != nil && c.cfg.APIKey != "" }

// Rewrite is the validated text produced by the model.
type Rewrite struct {
	Reason            string `json:"reason"`
	Explanation       string `json:"explanation"`
	RecommendedAction string `json:"recommended_action"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string            `json:"model"`
	Messages       []message         `json:"messages"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens"`
	ResponseFormat map[string]string `json:"response_format"`
	Reasoning      string            `json:"reasoning_effort,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

// maxAttempts: the model is not deterministic, so a failed call or an invalid answer gets one retry.
const maxAttempts = 2

// Explain asks the model to write the texts for one anomaly. It returns an error (and the caller
// keeps the template) if every attempt fails on HTTP, parsing or validation, or on timeout.
func (c *Client) Explain(ctx context.Context, r engine.AnomalyResult) (Rewrite, error) {
	system, user, allowed, err := buildPrompt(r)
	if err != nil {
		return Rewrite{}, err
	}
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		rw, err := c.attempt(ctx, system, user, allowed)
		if err == nil {
			return rw, nil
		}
		lastErr = err
		if ctx.Err() != nil || errors.Is(err, errPermanent) {
			break
		}
		var rl *rateLimitError
		if errors.As(err, &rl) {
			if rl.after > maxRetryWait { // not worth holding the analysis: use the template
				break
			}
			select {
			case <-time.After(rl.after):
			case <-ctx.Done():
			}
		}
	}
	return Rewrite{}, lastErr
}

// maxRetryWait is the longest Retry-After honored after a 429; beyond it the template is used.
const maxRetryWait = 3 * time.Second

type rateLimitError struct {
	after time.Duration
	body  string
}

func (e *rateLimitError) Error() string {
	return fmt.Sprintf("HTTP 429 (retry in %s): %s", e.after, e.body)
}

// errPermanent marks failures a retry cannot fix (bad key, unknown model).
var errPermanent = errors.New("permanent error")

func (c *Client) attempt(ctx context.Context, system, user string, allowed *numberSet) (Rewrite, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	body, err := json.Marshal(chatRequest{
		Model:          c.cfg.Model,
		Messages:       []message{{"system", system}, {"user", user}},
		Temperature:    0.2,
		MaxTokens:      1500, // reasoning models spend tokens thinking before they answer
		ResponseFormat: map[string]string{"type": "json_object"},
		Reasoning:      c.cfg.ReasoningEffort,
	})
	if err != nil {
		return Rewrite{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Rewrite{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Rewrite{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Rewrite{}, fmt.Errorf("reading response: %w", err)
	}
	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound:
		return Rewrite{}, fmt.Errorf("HTTP %d: %s: %w", resp.StatusCode, truncate(string(raw), 200), errPermanent)
	case resp.StatusCode == http.StatusTooManyRequests:
		after := time.Second // no usable Retry-After header: one short pause
		if secs, err := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64); err == nil && secs >= 0 {
			after = time.Duration(secs * float64(time.Second))
		}
		return Rewrite{}, &rateLimitError{after, truncate(string(raw), 160)}
	case resp.StatusCode != http.StatusOK:
		return Rewrite{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var cr chatResponse
	if err := json.Unmarshal(raw, &cr); err != nil || len(cr.Choices) == 0 {
		return Rewrite{}, fmt.Errorf("unexpected response shape: %s", truncate(string(raw), 200))
	}
	return parseAndValidate(cr.Choices[0].Message.Content, allowed)
}

// Enhance rewrites the explanation of every result in place, in parallel. A result whose call
// fails validation keeps its template text. It returns how many were written by the model.
// Classification, severity, priority and confidence are never touched.
func (c *Client) Enhance(ctx context.Context, results []engine.AnomalyResult) int {
	if !c.Enabled() {
		return 0
	}
	var wg sync.WaitGroup
	var n atomic.Int32
	for i := range results {
		wg.Add(1)
		go func(r *engine.AnomalyResult) {
			defer wg.Done()
			rw, err := c.Explain(ctx, *r)
			if err != nil {
				log.Printf("llm: %s: %v (using template)", r.MeterID, err)
				return
			}
			r.Reason, r.Explanation, r.RecommendedAction = rw.Reason, rw.Explanation, rw.RecommendedAction
			r.ExplanationSource = engine.SourceLLM
			n.Add(1)
		}(&results[i])
	}
	wg.Wait()
	return int(n.Load())
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
