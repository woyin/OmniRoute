package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/omniroute/omniroute/internal/config"
)

func TestTokenLimitCRUDAndWindow(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	created, err := UpsertTokenLimit(conn, TokenLimit{APIKeyID: "key-1", ScopeType: "model", ScopeValue: "gpt-5", TokenLimit: 1000, ResetInterval: "daily", ResetTime: "03:00", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	limits, err := ListTokenLimits(conn, "key-1")
	if err != nil || len(limits) != 1 || limits[0].ID != created.ID {
		t.Fatalf("limits=%+v err=%v", limits, err)
	}
	start, next, err := TokenLimitWindow(created, time.Date(2026, 7, 13, 2, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 7, 12, 3, 0, 0, 0, time.UTC); !start.Equal(want) {
		t.Fatalf("start=%s want=%s", start, want)
	}
	if next.Sub(start) != 24*time.Hour {
		t.Fatalf("window=%s", next.Sub(start))
	}
	deleted, err := DeleteTokenLimit(conn, created.ID)
	if err != nil || !deleted {
		t.Fatalf("deleted=%v err=%v", deleted, err)
	}
}
