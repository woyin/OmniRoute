package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestProviderParamFiltersCRUD(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	router := chi.NewRouter()
	handler := providerParamFiltersHandler(conn)
	router.Get("/api/providers/{id}/param-filters", handler)
	router.Put("/api/providers/{id}/param-filters", handler)
	router.Delete("/api/providers/{id}/param-filters", handler)
	put := httptest.NewRecorder()
	router.ServeHTTP(put, httptest.NewRequest(http.MethodPut, "/api/providers/openai/param-filters", bytes.NewBufferString(`{"block":["temperature"],"allow":[],"autoLearn":true}`)))
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/providers/openai/param-filters", nil))
	if get.Code != http.StatusOK || !bytes.Contains(get.Body.Bytes(), []byte(`"temperature"`)) {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	remove := httptest.NewRecorder()
	router.ServeHTTP(remove, httptest.NewRequest(http.MethodDelete, "/api/providers/openai/param-filters", nil))
	if remove.Code != http.StatusOK {
		t.Fatalf("delete status=%d", remove.Code)
	}
}
