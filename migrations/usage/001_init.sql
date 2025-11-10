CREATE TABLE IF NOT EXISTS usage_events (
    id BIGSERIAL PRIMARY KEY,
    org_id TEXT NOT NULL,
    metric TEXT NOT NULL,
    quantity BIGINT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    idempotency_key TEXT
);

-- Create partial unique index for idempotency (only when idempotency_key is NOT NULL)
CREATE UNIQUE INDEX IF NOT EXISTS usage_events_org_idempotency_unique 
    ON usage_events (org_id, idempotency_key) 
    WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS usage_aggregates (
    org_id TEXT NOT NULL,
    metric TEXT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    total BIGINT NOT NULL,
    PRIMARY KEY (org_id, metric, period_start, period_end)
);
