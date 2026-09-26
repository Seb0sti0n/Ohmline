// Package config loads runtime settings from environment variables.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	CORSOrigin  string
}

// Load reads .env (if present, here or in the repo root) and returns the config.
func Load() Config {
	_ = godotenv.Load(".env", "../.env")
	return Config{
		DatabaseURL: get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/astrophage?sslmode=disable"),
		JWTSecret:   get("JWT_SECRET", "dev-secret-change-me"),
		Port:        get("PORT", "8080"),
		CORSOrigin:  get("CORS_ORIGIN", "http://localhost:5173"),
	}
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
