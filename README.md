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

## Analysis engine

`backend/internal/engine` is a package of pure functions: `Run(readings, events, cfg) → []AnomalyResult`.
It has no HTTP or database dependencies, is deterministic and is tested against the real dataset
(`go test ./internal/engine/...`). All thresholds live in `backend/internal/config/engine.go`.

Pipeline: readings → baseline → detection → correlation → events → explanation → recommendation.

| Step | What it does | Why |
|---|---|---|
| Baseline | Median and scaled MAD of consumption, voltage, current and PF **per hour of day**, from the first 7 days | The load has a strong daily pattern, so a single average would flag every morning. Median/MAD are robust to outliers |
| Consumption detection | Robust z-score per reading; a **persistent change** is ≥ 6 consecutive hours with \|z\| > 4 in the same direction | Isolated 1–2 h spikes are normal noise and never become anomalies |
| Data quality | Flags voltage outside ±5% of 220 V, voltage jumps > 10 V, PF jumps > 0.2, kWh not matching V·I·PF (±30%). Recurrent = ≥ 5 flagged hours within 24 h, **and** no persistent consumption change | Quality problems are found in the readings. The `status` column is always `OK`, and the event type is never used as a label |
| Correlation | For each window: Δ current, Δ PF, Δ mean voltage, Δ voltage spread vs baseline | A rise in consumption with a PF drop points to a load or installation problem |
| Events | An event within 6 h before / 2 h after the window start explains it only if it is **compatible**: an outage must match the described duration and consumption must return to baseline; an operational change must be an upward step without electrical deterioration. `UNKNOWN` never explains. `DATA_QUALITY` only corroborates | Time coincidence alone is not an explanation |
| Classification | Persistent change + compatible outage → `FALSE_POSITIVE`/LOW · + compatible operational change → `EXPLAINABLE_ANOMALY`/MEDIUM · unexplained → `REAL_ANOMALY`/HIGH · recurrent quality flags with stable consumption → `DATA_QUALITY`/HIGH | Matches the cases in the brief |
| Priority (0–100) | Magnitude 25 + persistence 15 + electrical deterioration 20 + no explanation 20 + extra kWh 20; `FALSE_POSITIVE` capped below 20 | Ranks what to investigate first. The breakdown is stored in `evidence.priority_breakdown` |
| Confidence (0–0.99) | 0.40 base + signal strength 0.25 + independent signals 0.20 + explanation clarity 0.10 | Breakdown in `evidence.confidence_breakdown` |

Result on the dataset: M-109 `REAL_ANOMALY` (priority 100) > M-112 `DATA_QUALITY` (71.8) > M-104 `EXPLAINABLE_ANOMALY` (46.6)
> M-106 `FALSE_POSITIVE` (19). The other 8 meters produce no anomaly. The explanation texts come from
deterministic templates in Spanish (`explain.go`); an LLM can rewrite them later but never changes the classification.

Limitations: thresholds are tuned to this dataset; the 7-day baseline is short and assumes it contains no
anomalies (true here); one anomaly is reported per meter (the highest-priority window).

## LLM explanations

The engine decides type, severity, priority and confidence; the LLM only **writes** `reason`, `explanation` and
`recommended_action` from the evidence, during the "Explicación" stage. It never changes the classification.

- Any OpenAI-compatible API works; configure `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL` in `.env` (defaults: Groq,
  `openai/gpt-oss-20b`, `LLM_REASONING_EFFORT=low`, which is only for reasoning models — leave it empty otherwise).
- Without `LLM_API_KEY` the app makes no external calls and uses the deterministic templates in
  `engine/explain.go`. `explanation_source` (`LLM` or `TEMPLATE`) is stored per anomaly and shown in the UI.
- The answer is validated before it is used: valid JSON with the three fields, `reason` one sentence,
  `explanation` at most 4, and **every number it cites must come from the evidence** (allowing rounding), so the
  model cannot invent figures. If a call fails, times out (8 s), is rate limited, or the answer is rejected, that
  anomaly keeps its template text (after one retry). The analysis always completes.
- The prompt is small (~1k tokens) because free tiers limit tokens per minute (Groq: 8.000 for this model). One
  analysis fits; running several in the same minute makes some anomalies fall back to templates.
- Calls run in parallel, one per anomaly (typically about 1–2 s in total).

## API

REST under `/api`, documented in [backend/openapi.yaml](backend/openapi.yaml). JWT auth (`POST /api/auth/login`) on
everything except `/health`. Main flow: `POST /ai/analyze` starts a run in a goroutine (409 if one is already
running), `GET /ai/analysis/{id}` reports the stage and per-stage state, and `GET /anomalies` returns the
anomalies of the latest completed run ordered by priority. `GET /meters` supports `status`, `search`, `sort`
(`consumption`, `variation`, `severity`) and `order`; `GET /meters/{id}/readings` returns the series with the
baseline band for each point (`granularity=hour|day`).

The engine runs as one pure computation during the "Detección" stage; the seven stages of the stepper mirror the
pipeline and pause `ANALYSIS_STEP_DELAY_MS` (default 600) each so the progress is visible. Runs left in progress
by a previous process are marked `FAILED` on startup.

Meter status is derived from the latest analysis: `REAL_ANOMALY` → CRITICAL, `DATA_QUALITY` and
`EXPLAINABLE_ANOMALY` → ALERT, `FALSE_POSITIVE` and no anomaly → OK.

## Tests

`make test` runs the backend and frontend tests. The API tests need PostgreSQL: they create and use their own
`astrophage_test` database (override with `TEST_DATABASE_URL`) and skip themselves if Postgres is not reachable.

## Status

Phases 1 (scaffolding), 2 (data), 3 (analysis engine), 4 (API) and 5 (LLM layer) done. Next: frontend.
