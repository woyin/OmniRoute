package migrations

const SchemaV12 = `
CREATE TABLE IF NOT EXISTS model_capability_overrides (
 provider TEXT NOT NULL, model_id TEXT NOT NULL, override_key TEXT NOT NULL,
 override_value TEXT NOT NULL, refreshed_at TEXT NOT NULL DEFAULT (datetime('now')),
 PRIMARY KEY(provider,model_id,override_key)
);
`
