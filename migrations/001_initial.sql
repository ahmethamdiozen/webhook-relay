CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE endpoints (
    id         TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE targets (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id TEXT NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE events (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint_id TEXT NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    headers     TEXT NOT NULL,
    body        TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE deliveries (
    id           TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id     TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    target_id    TEXT NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    target_url   TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    attempts     INT NOT NULL DEFAULT 0,
    last_error   TEXT,
    delivered_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_deliveries_status ON deliveries(status);
CREATE INDEX idx_events_endpoint_id ON events(endpoint_id);
