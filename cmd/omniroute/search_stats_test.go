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

func TestSearchStatsUsesPersistedLogsAndCache(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.RecordCallLog(conn, db.CallLog{Provider: "tavily", Model: "search", StatusCode: 200, LatencyMs: 40}); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec("INSERT INTO semantic_cache(key_hash, response_body, hit_count) VALUES('hash','{}',3)"); err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	searchStatsHandler(conn).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/search/stats", nil))
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"tavily"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"hits":3`)) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
