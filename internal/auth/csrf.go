package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const csrfTokenContext = "omniroute-dashboard-csrf-v1"

func CSRFHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		cookie, err := r.Cookie("auth_token")
		secret := GetJWTSecret()
		if err != nil || secret == "" || !validateSessionToken(cookie.Value) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"token": nil, "expiresAt": nil})
			return
		}

		expires := time.Now().Add(10 * time.Minute).Truncate(time.Second)
		hash := sha256.Sum256([]byte(cookie.Value))
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte(csrfTokenContext + "\n" + strconv.FormatInt(expires.Unix(), 10) + "\n" + base64.RawURLEncoding.EncodeToString(hash[:])))
		token := "v1." + strconv.FormatInt(expires.Unix(), 10) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		_ = json.NewEncoder(w).Encode(map[string]string{"token": token, "expiresAt": expires.UTC().Format("2006-01-02T15:04:05.000Z")})
	}
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		fetchSite := strings.ToLower(r.Header.Get("Sec-Fetch-Site"))
		if fetchSite != "" && fetchSite != "same-origin" && fetchSite != "same-site" && fetchSite != "none" {
			writeCSRFError(w)
			return
		}
		origin := r.Header.Get("Origin")
		if origin == "" || sameRequestOrigin(r, origin) || validateCSRFToken(r) {
			next.ServeHTTP(w, r)
			return
		}
		writeCSRFError(w)
	})
}

func validateCSRFToken(r *http.Request) bool {
	secret := GetJWTSecret()
	cookie, err := r.Cookie("auth_token")
	parts := strings.Split(r.Header.Get("x-omniroute-csrf"), ".")
	if secret == "" || err != nil || !validateSessionToken(cookie.Value) || len(parts) != 3 || parts[0] != "v1" {
		return false
	}
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || expires < time.Now().Unix() {
		return false
	}
	provided, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	hash := sha256.Sum256([]byte(cookie.Value))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(csrfTokenContext + "\n" + parts[1] + "\n" + base64.RawURLEncoding.EncodeToString(hash[:])))
	return hmac.Equal(provided, mac.Sum(nil))
}

func sameRequestOrigin(r *http.Request, rawOrigin string) bool {
	origin, err := url.Parse(rawOrigin)
	if err != nil || origin.Scheme == "" || origin.Host == "" {
		return false
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]); forwarded != "" {
		proto = strings.ToLower(forwarded)
	}
	host := r.Host
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0]); forwarded != "" {
		host = forwarded
	}
	return strings.EqualFold(origin.Scheme, proto) && strings.EqualFold(origin.Host, host)
}

func writeCSRFError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    "INVALID_ORIGIN",
			"message": "Invalid request origin. Same-origin dashboard writes must include a valid dashboard CSRF token.",
		},
	})
}
