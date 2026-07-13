package management

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/omniroute/omniroute/internal/db/migrations"
)

func TestUpstreamProxyParity(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(migrations.SchemaV3); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO upstream_proxy_config (id, provider_id, is_active) VALUES
		('old-1', 'legacy', 0), ('old-2', 'legacy', 1)
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(migrations.SchemaV4); err != nil {
		t.Fatal(err)
	}

	var migrated int
	if err := db.QueryRow("SELECT COUNT(*) FROM upstream_proxy_config WHERE provider_id = 'legacy' AND enabled = 1").Scan(&migrated); err != nil || migrated != 1 {
		t.Fatalf("legacy migration: count=%d err=%v", migrated, err)
	}

	h := &UpstreamProxyHandler{DB: db}
	r := chi.NewRouter()
	r.Get("/{providerId}", h.Get)
	r.Put("/{providerId}", h.Upsert)
	r.Delete("/{providerId}", h.Delete)

	assertStatusBody(t, r, http.MethodGet, "/claude", "", http.StatusOK, `"enabled":false`, `"mode":"native"`)
	assertStatusBody(t, r, http.MethodPut, "/claude", `{"mode":"fallback","enabled":true}`, http.StatusOK, `"providerId":"claude"`, `"mode":"fallback"`)
	assertStatusBody(t, r, http.MethodGet, "/claude", "", http.StatusOK, `"enabled":true`, `"mode":"fallback"`)
	assertStatusBody(t, r, http.MethodDelete, "/claude", "", http.StatusOK, `"deleted":true`)
}

func assertStatusBody(t *testing.T, handler http.Handler, method, path, body string, status int, fragments ...string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s: status %d, body %s", method, path, rec.Code, rec.Body.String())
	}
	for _, fragment := range fragments {
		if !strings.Contains(rec.Body.String(), fragment) {
			t.Errorf("%s %s: body %s missing %s", method, path, rec.Body.String(), fragment)
		}
	}
}
