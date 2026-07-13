package main

import (
	"bytes"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestChaosConfigCRUD(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	h := chaosConfigHandler(conn)
	put := httptest.NewRecorder()
	h.ServeHTTP(put, httptest.NewRequest(http.MethodPut, "/api/chaos/config", bytes.NewBufferString(`{"enabled":true,"defaultMode":"parallel","providerOverrides":[],"timeoutMs":5000,"maxTokens":256}`)))
	if put.Code != http.StatusOK || !bytes.Contains(put.Body.Bytes(), []byte(`"enabled":true`)) {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/chaos/config", nil))
	if !bytes.Contains(get.Body.Bytes(), []byte(`"enabled":true`)) {
		t.Fatalf("get body=%s", get.Body.String())
	}
	reset := httptest.NewRecorder()
	h.ServeHTTP(reset, httptest.NewRequest(http.MethodDelete, "/api/chaos/config", nil))
	if !bytes.Contains(reset.Body.Bytes(), []byte(`"enabled":false`)) {
		t.Fatalf("reset body=%s", reset.Body.String())
	}
}
