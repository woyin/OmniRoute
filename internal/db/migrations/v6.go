package migrations

const SchemaV6 = `
CREATE TABLE IF NOT EXISTS sync_tokens (
 id TEXT PRIMARY KEY, name TEXT NOT NULL, token_hash TEXT NOT NULL UNIQUE,
 sync_api_key_id TEXT DEFAULT '', revoked_at TEXT, last_used_at TEXT,
 created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`
