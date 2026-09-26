package config

import (
	"os"
	"path/filepath"
	"testing"
)

// The API runs from backend/ while .env lives in the repo root: it must still be found.
func TestLoadFindsEnvInParentDir(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "backend")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	env := "LLM_API_KEY=from-file\nLLM_MODEL=from-file-model\nPORT=9999\nANALYSIS_STEP_DELAY_MS=0\nJWT_SECRET=from-env-file\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	for _, k := range []string{"LLM_API_KEY", "LLM_MODEL", "PORT", "ANALYSIS_STEP_DELAY_MS"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
	t.Setenv("JWT_SECRET", "from-real-env") // real environment wins over the file

	cfg := Load()
	if cfg.LLM.APIKey != "from-file" || cfg.LLM.Model != "from-file-model" || cfg.Port != "9999" || cfg.StepDelay != 0 {
		t.Errorf("values from ../.env were not loaded: %+v", cfg)
	}
	if cfg.JWTSecret != "from-real-env" {
		t.Errorf("environment must win over .env, got %q", cfg.JWTSecret)
	}
}

func TestDefaultsWithoutEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, k := range []string{"LLM_API_KEY", "PORT", "ANALYSIS_STEP_DELAY_MS", "JWT_SECRET", "DATABASE_URL", "LLM_MODEL"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
	cfg := Load()
	if cfg.Port != "8080" || cfg.StepDelay.Milliseconds() != 600 || cfg.LLM.APIKey != "" || cfg.LLM.Model != "openai/gpt-oss-20b" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}
