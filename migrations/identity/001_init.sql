CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    org_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Hardcode one org + API key for v1
INSERT INTO organizations (id, name) VALUES ('org-1', 'Test Organization')
ON CONFLICT (id) DO NOTHING;
INSERT INTO api_keys (id, org_id, key) VALUES ('key-1', 'org-1', 'test-api-key-12345')
ON CONFLICT (id) DO NOTHING;

