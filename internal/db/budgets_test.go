package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/omniroute/omniroute/internal/config"
)

func TestBudgetSummary(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	budget, err := UpsertBudget(conn, Budget{APIKeyID: "key-1", DailyLimitUSD: 10, WarningThreshold: .8, ResetInterval: "daily", ResetTime: "00:00"})
	if err != nil || budget.DailyLimitUSD != 10 {
		t.Fatalf("budget=%+v err=%v", budget, err)
	}
	if err := RecordUsage(conn, UsageEntry{Provider: "p", Model: "m", APIKey: "key-1", Cost: 3, Success: true}); err != nil {
		t.Fatal(err)
	}
	summary, err := BudgetSummary(conn, "key-1", time.Now())
	if err != nil || summary["totalCostToday"] != float64(3) {
		t.Fatalf("summary=%+v err=%v", summary, err)
	}
	check := BudgetCheck(summary)
	if check["remaining"] != float64(7) || check["allowed"] != true {
		t.Fatalf("check=%+v", check)
	}
}
