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

func TestManagementBatchesListAndDetail(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`INSERT INTO batches(id,endpoint,completion_window,status,input_file_id,created_at) VALUES('batch-one','/v1/chat/completions','24h','completed','file-one',1)`); err != nil {
		t.Fatal(err)
	}
	handler := managementBatchesHandler(conn)
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/batches", nil))
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte(`"batch-one"`)) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	router := chi.NewRouter()
	router.Get("/api/batches/{id}", handler)
	detail := httptest.NewRecorder()
	router.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/batches/batch-one", nil))
	if detail.Code != http.StatusOK || !bytes.Contains(detail.Body.Bytes(), []byte(`"batch"`)) {
		t.Fatalf("detail status=%d body=%s", detail.Code, detail.Body.String())
	}
}
