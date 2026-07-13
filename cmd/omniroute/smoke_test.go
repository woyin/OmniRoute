package main

import (
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestSQLiteSmokePreservesExistingValue(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")

	dbConn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dbConn.Exec(
		"INSERT INTO key_value (namespace, key, value) VALUES (?, ?, ?)",
		"smoke", "release", "original",
	); err != nil {
		t.Fatal(err)
	}
	if err := dbConn.Close(); err != nil {
		t.Fatal(err)
	}

	if err := sqliteSmoke(cfg); err != nil {
		t.Fatalf("sqliteSmoke() error = %v", err)
	}

	dbConn, err = db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	var got string
	if err := dbConn.QueryRow(
		"SELECT value FROM key_value WHERE namespace = ? AND key = ?",
		"smoke", "release",
	).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != "original" {
		t.Fatalf("smoke value = %q, want original", got)
	}
}
