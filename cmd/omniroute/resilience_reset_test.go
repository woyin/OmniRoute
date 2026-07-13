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

func TestResilienceResetClosesBreakers(t *testing.T) {
	cfg := config.Load()
	cfg.DataDir = t.TempDir()
	cfg.SQLiteFile = filepath.Join(cfg.DataDir, "storage.sqlite")
	conn, err := db.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`INSERT INTO domain_circuit_breakers(id,provider,state,failure_count) VALUES('one','openai','open',5)`); err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	resilienceResetHandler(conn).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/resilience/reset", nil))
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"resetCount":1`)) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var state string
	var failures int
	if err := conn.QueryRow("SELECT state,failure_count FROM domain_circuit_breakers WHERE id='one'").Scan(&state, &failures); err != nil || state != "closed" || failures != 0 {
		t.Fatalf("state=%s failures=%d err=%v", state, failures, err)
	}
}
