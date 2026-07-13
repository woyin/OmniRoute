package migrations

const SchemaV7 = `
ALTER TABLE api_key_token_limits RENAME TO api_key_token_limits_legacy;
CREATE TABLE api_key_token_limits (
 id TEXT PRIMARY KEY,
 api_key_id TEXT NOT NULL,
 scope_type TEXT NOT NULL CHECK (scope_type IN ('model', 'provider', 'global')),
 scope_value TEXT NOT NULL DEFAULT '',
 token_limit INTEGER NOT NULL CHECK (token_limit > 0),
 reset_interval TEXT NOT NULL DEFAULT 'monthly' CHECK (reset_interval IN ('daily', 'weekly', 'monthly')),
 reset_time TEXT NOT NULL DEFAULT '00:00',
 enabled INTEGER NOT NULL DEFAULT 1,
 created_at TEXT NOT NULL DEFAULT (datetime('now')),
 updated_at TEXT NOT NULL DEFAULT (datetime('now')),
 UNIQUE (api_key_id, scope_type, scope_value)
);
CREATE INDEX idx_aktl_api_key_id ON api_key_token_limits(api_key_id);
CREATE TABLE api_key_token_counters (
 limit_id TEXT NOT NULL,
 window_start TEXT NOT NULL,
 tokens_used INTEGER NOT NULL DEFAULT 0,
 updated_at TEXT NOT NULL DEFAULT (datetime('now')),
 PRIMARY KEY (limit_id, window_start)
);
CREATE TABLE api_key_token_limit_reset_logs (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 limit_id TEXT NOT NULL,
 reset_at TEXT NOT NULL DEFAULT (datetime('now')),
 prev_tokens INTEGER NOT NULL DEFAULT 0,
 window_start TEXT NOT NULL
);
CREATE INDEX idx_aktlrl_limit_id ON api_key_token_limit_reset_logs(limit_id);
`
