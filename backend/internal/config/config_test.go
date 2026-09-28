package config

import (
	"os"
	"path/filepath"
	"testing"
)

func clearEnvVars(t *testing.T) {
	t.Helper()
	for _, k := range []string{"LLM_API_KEY", "LLM_MODEL", "PORT", "ANALYSIS_STEP_DELAY_MS", "JWT_SECRET", "DATABASE_URL"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

const testEnvFile = "LLM_API_KEY=from-file\nLLM_MODEL=from-file-model\nPORT=9999\nANALYSIS_STEP_DELAY_MS=0\nJWT_SECRET=from-env-file\n"

func checkLoadedFromFile(t *testing.T, cfg Config) {
	t.Helper()
	if cfg.LLM.APIKey != "from-file" || cfg.LLM.Model != "from-file-model" || cfg.Port != "9999" || cfg.StepDelay != 0 {
		t.Errorf("values from .env were not loaded: %+v", cfg)
	}
}

// The API runs from backend/ (as `make api` does) while .env lives in the repo root: it must still
// be found one level up.
func TestLoadFindsEnvInParentDir(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "backend")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(testEnvFile), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	clearEnvVars(t)
	t.Setenv("JWT_SECRET", "from-real-env") // real environment wins over the file

	cfg := Load()
	checkLoadedFromFile(t, cfg)
	if cfg.JWTSecret != "from-real-env" {
		t.Errorf("environment must win over .env, got %q", cfg.JWTSecret)
	}
}

// An editor's "run" button typically launches the binary with the source file's own directory as the
// working directory (e.g. backend/cmd/api), several levels below the repo root: .env must still be
// found by walking further up, not just one level.
func TestLoadFindsEnvSeveralLevelsUp(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "backend", "cmd", "api")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(testEnvFile), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.work"), []byte("go 1.24\n\nuse ./backend\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(deep)
	clearEnvVars(t)

	checkLoadedFromFile(t, Load())
}

// go.work marks the repo root: the search must not walk past it into unrelated parent directories
// (e.g. the user's home directory) looking for an .env that has nothing to do with this project.
func TestLoadDoesNotSearchPastTheRepoRoot(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "backend")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	// an unrelated .env one level above the repo root must be ignored
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("JWT_SECRET=unrelated-project\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "go.work"), []byte("go 1.24\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	clearEnvVars(t)

	if cfg := Load(); cfg.JWTSecret != "dev-secret-change-me" {
		t.Errorf("JWTSecret = %q, want the default (must not read past the repo root)", cfg.JWTSecret)
	}
}

func TestDefaultsWithoutEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	clearEnvVars(t)
	cfg := Load()
	if cfg.Port != "8080" || cfg.StepDelay.Milliseconds() != 600 || cfg.LLM.APIKey != "" || cfg.LLM.Model != "openai/gpt-oss-20b" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}
