CREATE TABLE meters (
    id         SERIAL PRIMARY KEY,
    meter_id   TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    location   TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'OK' CHECK (status IN ('OK', 'ALERT', 'CRITICAL')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE readings (
    id              BIGSERIAL PRIMARY KEY,
    meter_id        TEXT NOT NULL REFERENCES meters (meter_id),
    timestamp       TIMESTAMPTZ NOT NULL,
    consumption_kwh DOUBLE PRECISION NOT NULL,
    voltage_v       DOUBLE PRECISION NOT NULL,
    current_a       DOUBLE PRECISION NOT NULL,
    power_factor    DOUBLE PRECISION NOT NULL,
    status          TEXT NOT NULL,
    UNIQUE (meter_id, timestamp)
);

CREATE TABLE events (
    id          SERIAL PRIMARY KEY,
    meter_id    TEXT NOT NULL REFERENCES meters (meter_id),
    timestamp   TIMESTAMPTZ NOT NULL,
    type        TEXT NOT NULL,
    description TEXT NOT NULL
);
CREATE INDEX events_meter_ts ON events (meter_id, timestamp);

CREATE TABLE analysis_runs (
    id           SERIAL PRIMARY KEY,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at  TIMESTAMPTZ,
    status       TEXT NOT NULL CHECK (status IN ('PENDING', 'RUNNING', 'COMPLETED', 'FAILED')),
    current_step TEXT,
    steps        JSONB NOT NULL DEFAULT '[]',
    summary      JSONB
);

CREATE TABLE anomalies (
    id                 SERIAL PRIMARY KEY,
    meter_id           TEXT NOT NULL REFERENCES meters (meter_id),
    analysis_run_id    INT NOT NULL REFERENCES analysis_runs (id),
    detected_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    type               TEXT NOT NULL CHECK (type IN ('REAL_ANOMALY', 'EXPLAINABLE_ANOMALY', 'FALSE_POSITIVE', 'DATA_QUALITY')),
    severity           TEXT NOT NULL CHECK (severity IN ('HIGH', 'MEDIUM', 'LOW')),
    confidence         DOUBLE PRECISION NOT NULL,
    priority_score     DOUBLE PRECISION NOT NULL,
    reason             TEXT NOT NULL,
    recommended_action TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'RESOLVED')),
    window_start       TIMESTAMPTZ,
    window_end         TIMESTAMPTZ,
    evidence           JSONB NOT NULL DEFAULT '{}',
    explanation_source TEXT NOT NULL DEFAULT 'TEMPLATE' CHECK (explanation_source IN ('LLM', 'TEMPLATE'))
);
CREATE INDEX anomalies_run ON anomalies (analysis_run_id);

CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);
