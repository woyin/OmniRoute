package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestUsageBudgetAndBulk(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.CreateAPIKey(conn, db.APIKey{Key: "key-1", Name: "test", IsActive: true}); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordUsage(conn, db.UsageEntry{Provider: "openai", Model: "gpt-5", APIKey: "key-1", Cost: 2.5, Success: true}); err != nil {
		t.Fatal(err)
	}

	handler := usageBudgetHandler(conn)
	create := httptest.NewRecorder()
	body := []byte(`{"apiKeyId":"key-1","dailyLimitUsd":10,"warningThreshold":0.8,"resetInterval":"daily","resetTime":"00:00"}`)
	handler.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/usage/budget", bytes.NewReader(body)))
	if create.Code != http.StatusOK || !bytes.Contains(create.Body.Bytes(), []byte(`"success":true`)) {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}

	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/usage/budget?apiKeyId=key-1", nil))
	if get.Code != http.StatusOK || !bytes.Contains(get.Body.Bytes(), []byte(`"totalCostToday":2.5`)) || !bytes.Contains(get.Body.Bytes(), []byte(`"remaining":7.5`)) {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}

	bulk := httptest.NewRecorder()
	usageBudgetBulkHandler(conn).ServeHTTP(bulk, httptest.NewRequest(http.MethodGet, "/api/usage/budget/bulk", nil))
	if bulk.Code != http.StatusOK || !bytes.Contains(bulk.Body.Bytes(), []byte(`"key-1"`)) {
		t.Fatalf("bulk status=%d body=%s", bulk.Code, bulk.Body.String())
	}
}
