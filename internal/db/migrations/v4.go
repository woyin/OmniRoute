package migrations

// SchemaV4 aligns upstream proxy persistence with main branch migration 017.
const SchemaV4 = `
ALTER TABLE upstream_proxy_config RENAME TO upstream_proxy_config_legacy;

CREATE TABLE upstream_proxy_config (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	provider_id TEXT NOT NULL UNIQUE,
	mode TEXT NOT NULL DEFAULT 'native',
	cliproxyapi_model_mapping TEXT,
	native_priority INTEGER NOT NULL DEFAULT 1,
	cliproxyapi_priority INTEGER NOT NULL DEFAULT 2,
	enabled INTEGER NOT NULL DEFAULT 1,
	family TEXT NOT NULL DEFAULT 'auto',
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO upstream_proxy_config (provider_id, enabled, created_at, updated_at)
SELECT provider_id, MAX(is_active), MIN(created_at), MAX(created_at)
FROM upstream_proxy_config_legacy
GROUP BY provider_id;

DROP TABLE upstream_proxy_config_legacy;

CREATE INDEX idx_upc_provider ON upstream_proxy_config(provider_id);
CREATE INDEX idx_upc_mode ON upstream_proxy_config(mode);
`
