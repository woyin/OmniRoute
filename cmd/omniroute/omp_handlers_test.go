package main

import (
	"bytes"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestOMPSettingsPreserveOtherProviders(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".omp", "agent")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent.db"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	models := filepath.Join(dir, "models.yml")
	if err := os.WriteFile(models, []byte("version: 1\nproviders:\n  other:\n    baseUrl: https://example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	h := ompSettingsHandler(conn)
	apply := httptest.NewRecorder()
	h.ServeHTTP(apply, httptest.NewRequest(http.MethodPost, "/api/cli-tools/omp-settings", bytes.NewBufferString(`{"baseUrl":"http://localhost:3456","apiKey":"secret"}`)))
	if apply.Code != http.StatusOK {
		t.Fatalf("apply=%d %s", apply.Code, apply.Body.String())
	}
	raw, _ := os.ReadFile(models)
	if !bytes.Contains(raw, []byte("other:")) || !bytes.Contains(raw, []byte("omniroute:")) {
		t.Fatalf("models=%s", raw)
	}
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/cli-tools/omp-settings", nil))
	if !bytes.Contains(get.Body.Bytes(), []byte(`"hasOmniRoute":true`)) {
		t.Fatalf("get=%s", get.Body.String())
	}
	remove := httptest.NewRecorder()
	h.ServeHTTP(remove, httptest.NewRequest(http.MethodDelete, "/api/cli-tools/omp-settings", nil))
	raw, _ = os.ReadFile(models)
	if !bytes.Contains(raw, []byte("other:")) || bytes.Contains(raw, []byte("omniroute:")) {
		t.Fatalf("removed models=%s", raw)
	}
}
