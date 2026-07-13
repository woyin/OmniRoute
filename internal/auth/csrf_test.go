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

func TestCSRFMiddleware(t *testing.T) {
	t.Setenv(JWTSecretEnv, "test-secret-at-least-32-bytes-long")
	session, err := GenerateSessionToken()
	if err != nil {
		t.Fatal(err)
	}

	issueReq := httptest.NewRequest(http.MethodGet, "http://localhost/api/auth/csrf", nil)
	issueReq.AddCookie(&http.Cookie{Name: "auth_token", Value: session})
	issueRec := httptest.NewRecorder()
	CSRFHandler().ServeHTTP(issueRec, issueReq)
	var issued struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(issueRec.Body.Bytes(), &issued); err != nil {
		t.Fatal(err)
	}

	handler := CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	check := func(name, origin, fetchSite, token string, want int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "http://localhost/api/settings", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Sec-Fetch-Site", fetchSite)
		req.Header.Set("x-omniroute-csrf", token)
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: session})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("%s: status=%d body=%s", name, rec.Code, rec.Body.String())
		}
	}
	check("same origin", "http://localhost", "same-origin", "", http.StatusNoContent)
	check("invalid origin without token", "https://evil.example", "same-site", "", http.StatusForbidden)
	check("invalid origin with token", "https://proxy.example", "same-site", issued.Token, http.StatusNoContent)
	check("cross site metadata cannot bypass", "https://proxy.example", "cross-site", issued.Token, http.StatusForbidden)
	check("tampered token", "https://proxy.example", "same-site", issued.Token+"x", http.StatusForbidden)
}
