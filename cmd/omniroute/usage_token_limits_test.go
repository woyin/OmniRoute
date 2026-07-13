package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
)

func TestUsageTokenLimitsCRUD(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	handler := usageTokenLimitsHandler(conn)

	body := []byte(`{"apiKeyId":"key-1","scopeType":"global","tokenLimit":1000}`)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/usage/token-limits", bytes.NewReader(body)))
	if create.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		Limit db.TokenLimit `json:"limit"`
	}
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Limit.ID == "" || !created.Limit.Enabled {
		t.Fatalf("created=%+v", created.Limit)
	}

	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/usage/token-limits?apiKeyId=key-1", nil))
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte(`"remaining":1000`)) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	remove := httptest.NewRecorder()
	handler.ServeHTTP(remove, httptest.NewRequest(http.MethodDelete, "/api/usage/token-limits?id="+created.Limit.ID, nil))
	if remove.Code != http.StatusOK || !bytes.Contains(remove.Body.Bytes(), []byte(`"success":true`)) {
		t.Fatalf("delete status=%d body=%s", remove.Code, remove.Body.String())
	}
}
