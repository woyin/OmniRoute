package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFHandler(t *testing.T) {
	t.Setenv(JWTSecretEnv, "test-secret-at-least-32-bytes-long")
	session, err := GenerateSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/auth/csrf", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: session})
	rec := httptest.NewRecorder()
	CSRFHandler().ServeHTTP(rec, req)

	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status=%d cache=%q", rec.Code, rec.Header().Get("Cache-Control"))
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(body.Token, "v1.") {
		t.Fatalf("unexpected token %q", body.Token)
	}
}
