package main

import (
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
)

func TestSQLiteSmoke(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")

	if err := sqliteSmoke(cfg); err != nil {
		t.Fatalf("sqliteSmoke() error = %v", err)
	}
}
