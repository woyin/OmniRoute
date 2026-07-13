package migrations

const SchemaV8 = `
CREATE TABLE IF NOT EXISTS domain_budgets (
 api_key_id TEXT PRIMARY KEY,
 daily_limit_usd REAL NOT NULL DEFAULT 0,
 weekly_limit_usd REAL NOT NULL DEFAULT 0,
 monthly_limit_usd REAL NOT NULL DEFAULT 0,
 warning_threshold REAL NOT NULL DEFAULT 0.8,
 reset_interval TEXT NOT NULL DEFAULT 'daily' CHECK (reset_interval IN ('daily', 'weekly', 'monthly')),
 reset_time TEXT NOT NULL DEFAULT '00:00',
 budget_reset_at INTEGER,
 last_budget_reset_at INTEGER,
 warning_emitted_at INTEGER,
 warning_period_start INTEGER
);
CREATE TABLE IF NOT EXISTS domain_budget_reset_logs (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 api_key_id TEXT NOT NULL,
 reset_interval TEXT NOT NULL,
 previous_spend REAL NOT NULL DEFAULT 0,
 reset_at INTEGER NOT NULL,
 next_reset_at INTEGER NOT NULL,
 period_start INTEGER NOT NULL,
 period_end INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_dbrl_key_reset ON domain_budget_reset_logs(api_key_id, reset_at DESC);
`
