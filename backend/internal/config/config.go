// Package config loads runtime settings from environment variables.
package config

import (
	"os"
	"path/filepath"
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

// Load reads .env (searching upward from the working directory to the repo root, so it is found
// regardless of how or from where the binary is launched) and returns the config. Variables already
// set in the environment win over the file.
func Load() Config {
	loadEnv()
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

// loadEnv looks for .env starting at the working directory and walking up through its parents,
// stopping once it finds one, once it reaches the repo root (marked by go.work, which sits next to
// .env), or after a handful of levels as a safety net. This way LLM_API_KEY and the rest of .env are
// found the same way whether the binary is launched with `go run ./cmd/api` from backend/ (make
// api's working directory), from the repo root, or by an editor's "run" button, which typically uses
// the source file's own directory (backend/cmd/api) as the working directory instead.
func loadEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for range 8 {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			_ = godotenv.Load(filepath.Join(dir, ".env"))
			return
		}
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return // reached the repo root; no .env to load
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return // reached the filesystem root
		}
		dir = parent
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
