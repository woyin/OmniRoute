package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/omniroute/omniroute/internal/db"
)

type contextKey string

const apiKeyContextKey contextKey = "api_key"

// ExtractAPIKey extracts an API key from the request.
func ExtractAPIKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		key := strings.TrimPrefix(auth, "Bearer ")
		if strings.HasPrefix(key, "sk-") {
			return key
		}
	}
	key := r.Header.Get("x-api-key")
	if key != "" {
		return key
	}
	return ""
}

// RequireAPIKey is middleware that enforces API key authentication.
func RequireAPIKey(dbConn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ExtractAPIKey(r)
			if key == "" {
				writeAuthError(w, "Missing API key")
				return
			}
			apiKey, err := db.ValidateAPIKey(dbConn, key)
			if err != nil {
				writeAuthError(w, "API key validation failed")
				return
			}
			if apiKey == nil {
				writeAuthError(w, "Invalid API key")
				return
			}
			ctx := context.WithValue(r.Context(), apiKeyContextKey, apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAPIKey is middleware that extracts and validates an API key if present.
func OptionalAPIKey(dbConn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ExtractAPIKey(r)
			if key != "" {
				apiKey, err := db.ValidateAPIKey(dbConn, key)
				if err == nil && apiKey != nil {
					ctx := context.WithValue(r.Context(), apiKeyContextKey, apiKey)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetAPIKeyFromContext retrieves the API key from the request context.
func GetAPIKeyFromContext(ctx context.Context) *db.APIKey {
	val := ctx.Value(apiKeyContextKey)
	if val == nil {
		return nil
	}
	ak, ok := val.(*db.APIKey)
	if !ok {
		return nil
	}
	return ak
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)
	w.Write([]byte(`{"error":{"message":"` + message + `","type":"authentication_error","code":"invalid_api_key"}}`))
}

// LoginMiddleware checks for a valid session cookie and enforces authentication.
// During bootstrap (no password set yet), requests pass through so the user
// can complete initial setup. After a password is set, unauthenticated
// requests are rejected with 401.
func LoginMiddleware(dbConn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for session cookie
			if cookie, err := r.Cookie("auth_token"); err == nil && validateSessionToken(cookie.Value) {
				// Session token exists — add to context
				ctx := context.WithValue(r.Context(), contextKey("session"), cookie.Value)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			// No session — allow through during bootstrap window (first run, no password yet)
			if IsBootstrapWindow(dbConn) {
				next.ServeHTTP(w, r)
				return
			}

			// Not authenticated and not in bootstrap — reject
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"message":"Authentication required","type":"authentication_error"}}`))
		})
	}
}

// IsAuthenticated checks if the request has a valid session.
func IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("auth_token")
	if err != nil || cookie.Value == "" {
		return false
	}
	return validateSessionToken(cookie.Value)
}
