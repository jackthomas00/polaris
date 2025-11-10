CREATE TABLE IF NOT EXISTS plans (
    id TEXT PRIMARY KEY,
    org_id TEXT NOT NULL,
    name TEXT NOT NULL,
    metric TEXT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    free_quota BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoices (
    id TEXT PRIMARY KEY,
    org_id TEXT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft', -- 'draft', 'finalized'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Simple plan for v1: $0.01 per unit after 1000 free
INSERT INTO plans (id, org_id, name, metric, unit_price, free_quota) 
VALUES ('plan-1', 'org-1', 'Default Plan', 'api_calls', 0.01, 1000)
ON CONFLICT (id) DO NOTHING;

