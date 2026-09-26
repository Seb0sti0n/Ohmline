# Astrophage

AI energy management MVP: turns hourly readings from 12 electric meters into operational decisions
(data → analysis → anomaly → explanation → prioritization → action).

The original brief is in [docs/new_project.pdf](docs/new_project.pdf);
the UI design system and mockups are in [design/](design/).

## Stack

- **Backend:** Go, chi, pgx, PostgreSQL
- **Frontend:** Vue 3, Vite, TypeScript, Pinia, Vue Router, Tailwind (theme generated from `design/tokens.json`), ECharts
- **LLM (optional):** any OpenAI-compatible API (Groq by default); deterministic templates as fallback

## Layout

```
backend/    Go API (cmd/api, internal/...)
frontend/   Vue app
data/       readings.csv, events.csv
design/     design system and mockups
docs/       original brief
```

## Run locally

Requirements: Go 1.24+ (the toolchain downloads the exact version in `go.mod` automatically), Node 20+, a local PostgreSQL.

```bash
cp .env.example .env     # adjust DATABASE_URL if your Postgres credentials differ
make db                  # creates the `astrophage` database
make seed                # applies migrations, loads data/*.csv and the demo user
make api                 # http://localhost:8080/api/health
make web                 # http://localhost:5173
```

### Data and demo user

`make seed` is idempotent (it truncates and reloads). It validates the CSVs (duplicates, gaps in the hourly grid,
negative or empty values) and prints any issues; the current dataset has none. Result: 12 meters, 4.032 readings,
4 events. Demo login: `demo@energy.io` / `demo123`. The seed does not run the analysis.

Migrations live in `backend/migrations` (embedded in the binary and applied by the seed command).

If `design/tokens.json` changes, regenerate the Tailwind theme with `cd frontend && npm run tokens`.

## Status

Phases 1 (scaffolding) and 2 (data) done. Next: analysis engine, API, LLM layer, frontend.
