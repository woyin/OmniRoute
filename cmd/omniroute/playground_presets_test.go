package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestPlaygroundPresetHTTPCRUD(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	router := chi.NewRouter()
	handler := playgroundPresetsHandler(conn)
	router.Get("/api/playground/presets", handler)
	router.Post("/api/playground/presets", handler)
	router.Get("/api/playground/presets/{id}", handler)
	router.Put("/api/playground/presets/{id}", handler)
	router.Delete("/api/playground/presets/{id}", handler)

	create := httptest.NewRecorder()
	router.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/playground/presets", bytes.NewBufferString(`{"name":"Demo","endpoint":"/api/v1/chat/completions","model":"gpt-5","params":{}}`)))
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var preset db.PlaygroundPreset
	if err := json.Unmarshal(create.Body.Bytes(), &preset); err != nil {
		t.Fatal(err)
	}

	update := httptest.NewRecorder()
	router.ServeHTTP(update, httptest.NewRequest(http.MethodPut, "/api/playground/presets/"+preset.ID, bytes.NewBufferString(`{"name":"Updated"}`)))
	if update.Code != http.StatusOK || !bytes.Contains(update.Body.Bytes(), []byte(`"Updated"`)) {
		t.Fatalf("update status=%d body=%s", update.Code, update.Body.String())
	}

	remove := httptest.NewRecorder()
	router.ServeHTTP(remove, httptest.NewRequest(http.MethodDelete, "/api/playground/presets/"+preset.ID, nil))
	if remove.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", remove.Code, remove.Body.String())
	}
}
