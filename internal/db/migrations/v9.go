package migrations

const SchemaV9 = `
CREATE TABLE IF NOT EXISTS playground_presets (
 id TEXT PRIMARY KEY,
 name TEXT NOT NULL,
 endpoint TEXT NOT NULL,
 model TEXT NOT NULL,
 system TEXT,
 params_json TEXT NOT NULL DEFAULT '{}',
 created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_playground_presets_created ON playground_presets(created_at DESC);
`
