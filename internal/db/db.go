package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db/migrations"
)

var (
	instance    *sql.DB
	instanceErr error
	once        sync.Once
)

// OpenDB opens a SQLite database connection and runs pending migrations.
func OpenDB(cfg *config.Config) (*sql.DB, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	dbConn, err := sql.Open("sqlite3", cfg.SQLiteFile+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	dbConn.SetMaxOpenConns(1) // SQLite single-writer constraint
	dbConn.SetMaxIdleConns(2)
	if err := runMigrations(dbConn); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	return dbConn, nil
}

// GetDB returns a singleton SQLite database connection with WAL mode.
func GetDB(cfg *config.Config) (*sql.DB, error) {
	once.Do(func() {
		instance, instanceErr = OpenDB(cfg)
	})
	return instance, instanceErr
}

// CloseDB closes the database connection.
func CloseDB() error {
	if instance != nil {
		return instance.Close()
	}
	return nil
}

// runMigrations applies all pending schema migrations inside a transaction.
func runMigrations(db *sql.DB) error {
	// Create migrations tracking table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS _omniroute_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Base schema (v1)
	migrations := []struct {
		name string
		sql  string
	}{
		{"001_initial_schema", schemaV1},
		{"002_provider_nodes", schemaV2},
		{"003_critical_tables", schemaV3},
		{"004_upstream_proxy_parity", schemaV4},
		{"005_combo_priority", schemaV5},
		{"006_sync_tokens", schemaV6},
		{"007_token_limits_parity", schemaV7},
		{"008_usage_budgets", schemaV8},
		{"009_playground_presets", schemaV9},
		{"010_files", schemaV10},
		{"011_batches", schemaV11},
	}

	for _, m := range migrations {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM _omniroute_migrations WHERE name = ?", m.name).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if count > 0 {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", m.name, err)
		}
		if _, err := tx.Exec(m.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", m.name, err)
		}
		if _, err := tx.Exec("INSERT INTO _omniroute_migrations (name) VALUES (?)", m.name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", m.name, err)
		}
	}
	return nil
}

// LoadSQL reads a .sql file from the migrations directory.
func LoadSQL(name string) (string, error) {
	path := filepath.Join("migrations", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(data), nil
}

var schemaV3 = migrations.SchemaV3
var schemaV4 = migrations.SchemaV4
var schemaV5 = migrations.SchemaV5
var schemaV6 = migrations.SchemaV6
var schemaV7 = migrations.SchemaV7
var schemaV8 = migrations.SchemaV8
var schemaV9 = migrations.SchemaV9
var schemaV10 = migrations.SchemaV10
var schemaV11 = migrations.SchemaV11

const schemaV1 = `
CREATE TABLE IF NOT EXISTS provider_connections (
	id TEXT PRIMARY KEY,
	provider TEXT NOT NULL,
	name TEXT DEFAULT '',
	api_key TEXT,
	access_token TEXT,
	refresh_token TEXT,
	project_id TEXT,
	expires_at TEXT,
	is_active INTEGER DEFAULT 1,
	test_status TEXT DEFAULT '',
	priority INTEGER DEFAULT 0,
	provider_specific_data TEXT DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS key_value (
	namespace TEXT NOT NULL,
	key TEXT NOT NULL,
	value TEXT,
	PRIMARY KEY (namespace, key)
);

CREATE TABLE IF NOT EXISTS combos (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	strategy TEXT DEFAULT 'priority',
	targets TEXT DEFAULT '[]',
	is_active INTEGER DEFAULT 1,
	domain TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_keys (
	key TEXT PRIMARY KEY,
	name TEXT DEFAULT '',
	is_active INTEGER DEFAULT 1,
	scopes TEXT DEFAULT '[]',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	last_used_at DATETIME
);

CREATE TABLE IF NOT EXISTS usage_history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	provider TEXT NOT NULL,
	model TEXT NOT NULL,
	api_key TEXT DEFAULT '',
	input_tokens INTEGER DEFAULT 0,
	output_tokens INTEGER DEFAULT 0,
	cost REAL DEFAULT 0,
	latency_ms INTEGER DEFAULT 0,
	success INTEGER DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS call_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	provider TEXT NOT NULL,
	model TEXT NOT NULL,
	status_code INTEGER DEFAULT 0,
	latency_ms INTEGER DEFAULT 0,
	request_id TEXT DEFAULT '',
	api_key TEXT DEFAULT '',
	error_message TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS semantic_cache (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	key_hash TEXT NOT NULL UNIQUE,
	response_body TEXT,
	model TEXT,
	provider TEXT,
	hit_count INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME
);

CREATE TABLE IF NOT EXISTS quota_snapshots (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	provider TEXT NOT NULL,
	api_key TEXT DEFAULT '',
	remaining INTEGER DEFAULT 0,
	limit_val INTEGER DEFAULT 0,
	reset_at TEXT DEFAULT '',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS db_meta (
	key TEXT PRIMARY KEY,
	value TEXT
);
`

const schemaV2 = `
CREATE TABLE IF NOT EXISTS provider_nodes (
	id TEXT PRIMARY KEY,
	provider TEXT NOT NULL,
	name TEXT DEFAULT '',
	base_url TEXT DEFAULT '',
	api_key TEXT,
	is_active INTEGER DEFAULT 1,
	priority INTEGER DEFAULT 0,
	max_concurrent INTEGER DEFAULT 0,
	provider_specific_data TEXT DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS model_combo_mappings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	model TEXT NOT NULL,
	combo_id TEXT NOT NULL,
	priority INTEGER DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (combo_id) REFERENCES combos(id)
);
CREATE INDEX IF NOT EXISTS idx_model_combo_mappings_model ON model_combo_mappings(model);
`
