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

func TestSearchProvidersCatalogAndStatus(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.SaveProviderConnection(conn, db.ProviderConnection{ID: "one", Provider: "brave-search", APIKey: "secret", IsActive: true}); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	searchProvidersHandler(conn).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/search/providers", nil))
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"brave-search"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"configured"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"search_provider"`)) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
