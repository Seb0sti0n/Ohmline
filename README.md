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

Requirements: Go 1.22+, Node 20+, a local PostgreSQL.

```bash
cp .env.example .env     # adjust DATABASE_URL if your Postgres credentials differ
make db                  # creates the `astrophage` database
make api                 # http://localhost:8080/api/health
make web                 # http://localhost:5173
```

If `design/tokens.json` changes, regenerate the Tailwind theme with `cd frontend && npm run tokens`.

## Status

Phase 1 (scaffolding) done. Next: data ingestion, analysis engine, API, LLM layer, frontend.
