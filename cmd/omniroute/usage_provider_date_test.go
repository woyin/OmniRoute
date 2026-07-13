package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestUsageRequestsByProviderDate(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.RecordUsage(conn, db.UsageEntry{Provider: "OpenAI", Model: "gpt-5", InputTokens: 10, OutputTokens: 5, Success: true}); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	usageRequestsByProviderDateHandler(conn).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/usage/requests-by-provider-date?range=1d", nil))
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"provider":"openai"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"totalTokens":15`)) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestUsageDateWindowValidation(t *testing.T) {
	if _, _, err := usageDateWindow("bad", "", "", "", time.Now()); err == nil {
		t.Fatal("invalid range accepted")
	}
	if start, end, err := usageDateWindow("30d", "2026-07-13", "", "", time.Now()); err != nil || start == "" || end == "" {
		t.Fatalf("start=%q end=%q err=%v", start, end, err)
	}
}
