package db

import (
	"github.com/omniroute/omniroute/internal/config"
	"path/filepath"
	"testing"
)

func TestModelCapabilityOverridePersistence(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := SetModelCapabilityOverride(conn, "openai", "gpt-5", "max_token", 4096); err != nil {
		t.Fatal(err)
	}
	items, err := ListModelCapabilityOverrides(conn)
	if err != nil || len(items) != 1 || items[0].Value != 4096 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
}
