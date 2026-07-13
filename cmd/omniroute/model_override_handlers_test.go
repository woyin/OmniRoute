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

func TestModelCapabilityOverrideCRUD(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	handler := modelCapabilityOverridesHandler(conn)
	patch := httptest.NewRecorder()
	handler.ServeHTTP(patch, httptest.NewRequest(http.MethodPatch, "/api/model-capability-overrides", bytes.NewBufferString(`{"target":"openai/gpt-5","key":"max_token","value":8192}`)))
	if patch.Code != http.StatusOK || !bytes.Contains(patch.Body.Bytes(), []byte(`"value":8192`)) {
		t.Fatalf("patch status=%d body=%s", patch.Code, patch.Body.String())
	}
	remove := httptest.NewRecorder()
	handler.ServeHTTP(remove, httptest.NewRequest(http.MethodDelete, "/api/model-capability-overrides?target=openai/gpt-5&key=max_token", nil))
	if remove.Code != http.StatusOK || bytes.Contains(remove.Body.Bytes(), []byte(`"value":8192`)) {
		t.Fatalf("delete status=%d body=%s", remove.Code, remove.Body.String())
	}
}
