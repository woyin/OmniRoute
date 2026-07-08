package migrations

// SchemaV3 contains critical tables needed for webhooks, skills, memory,
// MCP audit, circuit breakers, reasoning cache, rate limiting, and more.
const SchemaV3 = `
-- Webhooks
CREATE TABLE IF NOT EXISTS webhooks (
	id TEXT PRIMARY KEY,
	url TEXT NOT NULL,
	events TEXT DEFAULT '[]',
	secret TEXT DEFAULT '',
	is_active INTEGER DEFAULT 1,
	last_delivery_at TEXT,
	failure_count INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	webhook_id TEXT NOT NULL,
	event TEXT NOT NULL,
	payload TEXT DEFAULT '',
	status_code INTEGER DEFAULT 0,
	response_body TEXT DEFAULT '',
	success INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Registered keys (for key rotation/management)
CREATE TABLE IF NOT EXISTS registered_keys (
	id TEXT PRIMARY KEY,
	provider TEXT NOT NULL,
	key_fingerprint TEXT DEFAULT '',
	scopes TEXT DEFAULT '[]',
	is_active INTEGER DEFAULT 1,
	expires_at TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Files (for uploads)
CREATE TABLE IF NOT EXISTS files (
	id TEXT PRIMARY KEY,
	filename TEXT DEFAULT '',
	purpose TEXT DEFAULT '',
	size INTEGER DEFAULT 0,
	status TEXT DEFAULT 'uploaded',
	provider TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Batches (for batch API)
CREATE TABLE IF NOT EXISTS batches (
	id TEXT PRIMARY KEY,
	status TEXT DEFAULT 'pending',
	request_count INTEGER DEFAULT 0,
	completed_count INTEGER DEFAULT 0,
	failed_count INTEGER DEFAULT 0,
	endpoint TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at TEXT
);

-- Skills
CREATE TABLE IF NOT EXISTS skills (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT DEFAULT '',
	version TEXT DEFAULT '1.0.0',
	is_enabled INTEGER DEFAULT 1,
	is_builtin INTEGER DEFAULT 0,
	skill_type TEXT DEFAULT 'custom',
	config TEXT DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS skill_executions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	skill_id TEXT NOT NULL,
	input TEXT DEFAULT '',
	output TEXT DEFAULT '',
	success INTEGER DEFAULT 1,
	duration_ms INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Memory
CREATE TABLE IF NOT EXISTS memories (
	id TEXT PRIMARY KEY,
	content TEXT NOT NULL,
	tags TEXT DEFAULT '[]',
	embedding TEXT DEFAULT '',
	provider TEXT DEFAULT '',
	session_id TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- MCP audit
CREATE TABLE IF NOT EXISTS mcp_tool_audit (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	tool_name TEXT NOT NULL,
	args TEXT DEFAULT '{}',
	result_summary TEXT DEFAULT '',
	success INTEGER DEFAULT 1,
	api_key_id TEXT DEFAULT '',
	duration_ms INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Domain fallback chains
CREATE TABLE IF NOT EXISTS domain_fallback_chains (
	id TEXT PRIMARY KEY,
	provider TEXT NOT NULL,
	model TEXT DEFAULT '',
	fallback_order TEXT DEFAULT '[]',
	is_active INTEGER DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Domain circuit breakers
CREATE TABLE IF NOT EXISTS domain_circuit_breakers (
	id TEXT PRIMARY KEY,
	provider TEXT NOT NULL,
	state TEXT DEFAULT 'closed',
	failure_count INTEGER DEFAULT 0,
	last_failure_at TEXT,
	last_state_change_at TEXT DEFAULT CURRENT_TIMESTAMP,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Reasoning cache
CREATE TABLE IF NOT EXISTS reasoning_cache (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL,
	model TEXT NOT NULL,
	reasoning_content TEXT DEFAULT '',
	turn_index INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reasoning_cache_session ON reasoning_cache(session_id);

-- Session account affinity
CREATE TABLE IF NOT EXISTS session_account_affinity (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL,
	provider TEXT NOT NULL,
	connection_id TEXT DEFAULT '',
	model TEXT DEFAULT '',
	last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(session_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_session_affinity_session ON session_account_affinity(session_id);

-- API key token limits
CREATE TABLE IF NOT EXISTS api_key_token_limits (
	id TEXT PRIMARY KEY,
	api_key_id TEXT NOT NULL,
	model TEXT DEFAULT '',
	max_tokens_per_day INTEGER DEFAULT 0,
	max_tokens_per_month INTEGER DEFAULT 0,
	tokens_used_today INTEGER DEFAULT 0,
	tokens_used_month INTEGER DEFAULT 0,
	reset_day TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Daily usage summary
CREATE TABLE IF NOT EXISTS daily_usage_summary (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date TEXT NOT NULL,
	provider TEXT NOT NULL,
	model TEXT NOT NULL,
	request_count INTEGER DEFAULT 0,
	input_tokens INTEGER DEFAULT 0,
	output_tokens INTEGER DEFAULT 0,
	cost REAL DEFAULT 0,
	success_count INTEGER DEFAULT 0,
	error_count INTEGER DEFAULT 0,
	UNIQUE(date, provider, model)
);

CREATE INDEX IF NOT EXISTS idx_daily_usage_date ON daily_usage_summary(date);

-- Middleware hooks
CREATE TABLE IF NOT EXISTS middleware_hooks (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	hook_type TEXT DEFAULT 'pre-request',
	is_active INTEGER DEFAULT 1,
	priority INTEGER DEFAULT 0,
	config TEXT DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Relay tokens
CREATE TABLE IF NOT EXISTS relay_tokens (
	id TEXT PRIMARY KEY,
	token TEXT NOT NULL UNIQUE,
	provider TEXT DEFAULT '',
	model TEXT DEFAULT '',
	scopes TEXT DEFAULT '[]',
	is_active INTEGER DEFAULT 1,
	expires_at TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Provider quota reset events
CREATE TABLE IF NOT EXISTS provider_quota_reset_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	provider TEXT NOT NULL,
	connection_id TEXT DEFAULT '',
	reset_at TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Version manager
CREATE TABLE IF NOT EXISTS version_manager (
	key TEXT PRIMARY KEY,
	value TEXT
);

-- Upstream proxy config
CREATE TABLE IF NOT EXISTS upstream_proxy_config (
	id TEXT PRIMARY KEY,
	provider_id TEXT NOT NULL,
	proxy_url TEXT DEFAULT '',
	auth_type TEXT DEFAULT '',
	auth_value TEXT DEFAULT '',
	is_active INTEGER DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Compression combos
CREATE TABLE IF NOT EXISTS compression_combos (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	mode TEXT DEFAULT 'lite',
	is_active INTEGER DEFAULT 1,
	config TEXT DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Guardrails config
CREATE TABLE IF NOT EXISTS guardrails_config (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	guardrail_type TEXT DEFAULT '',
	is_enabled INTEGER DEFAULT 0,
	config TEXT DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- A2A tasks
CREATE TABLE IF NOT EXISTS a2a_tasks (
	id TEXT PRIMARY KEY,
	status TEXT DEFAULT 'submitted',
	model TEXT DEFAULT '',
	input TEXT DEFAULT '',
	output TEXT DEFAULT '',
	error TEXT DEFAULT '',
	expires_at TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS a2a_task_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	task_id TEXT NOT NULL,
	event_type TEXT NOT NULL,
	data TEXT DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_a2a_task_events_task ON a2a_task_events(task_id);
`
