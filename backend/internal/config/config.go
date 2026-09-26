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
}

// Load reads .env (if present, here or in the repo root) and returns the config.
func Load() Config {
	_ = godotenv.Load(".env", "../.env")
	return Config{
		DatabaseURL: get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/astrophage?sslmode=disable"),
		JWTSecret:   get("JWT_SECRET", "dev-secret-change-me"),
		Port:        get("PORT", "8080"),
		CORSOrigin:  get("CORS_ORIGIN", "http://localhost:5173"),
		StepDelay:   time.Duration(getInt("ANALYSIS_STEP_DELAY_MS", 600)) * time.Millisecond,
		Engine:      DefaultEngine(),
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
