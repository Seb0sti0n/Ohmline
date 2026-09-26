// Package config loads runtime settings from environment variables.
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	CORSOrigin  string
	// StepDelay is an artificial pause after each analysis stage so the progress is visible in the UI.
	StepDelay time.Duration
	Engine    Engine
	LLM       LLM
}

// LLM configures an OpenAI-compatible chat completions API (Groq by default). Without an API key
// the app uses the deterministic template explanations.
type LLM struct {
	BaseURL string
	APIKey  string
	Model   string
	// ReasoningEffort (low|medium|high) is sent only if set. Reasoning models such as gpt-oss need
	// "low" to answer within the token budget; leave empty for models that do not support it.
	ReasoningEffort string
	Timeout         time.Duration // per request; on timeout the template is used
}

// Load reads .env (here or in the repo root) and returns the config. Variables already set in the
// environment win over the file. Each file is loaded on its own because godotenv.Load stops at the
// first file it cannot open.
func Load() Config {
	for _, f := range []string{".env", "../.env"} {
		_ = godotenv.Load(f)
	}
	return Config{
		DatabaseURL: get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/astrophage?sslmode=disable"),
		JWTSecret:   get("JWT_SECRET", "dev-secret-change-me"),
		Port:        get("PORT", "8080"),
		CORSOrigin:  get("CORS_ORIGIN", "http://localhost:5173"),
		StepDelay:   time.Duration(getInt("ANALYSIS_STEP_DELAY_MS", 600)) * time.Millisecond,
		Engine:      DefaultEngine(),
		LLM: LLM{
			BaseURL:         get("LLM_BASE_URL", "https://api.groq.com/openai/v1"),
			APIKey:          os.Getenv("LLM_API_KEY"),
			Model:           get("LLM_MODEL", "openai/gpt-oss-20b"),
			ReasoningEffort: os.Getenv("LLM_REASONING_EFFORT"),
			Timeout:         8 * time.Second,
		},
	}
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v >= 0 {
		return v
	}
	return fallback
}
