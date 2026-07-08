package middleware

import (
	"log"
	"net/http"
	"strings"
)

// CORSHeaders returns the standard CORS headers for OmniRoute.
func CORSHeaders() map[string]string {
	return map[string]string{
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
		"Access-Control-Allow-Headers":     "Authorization, Content-Type, x-api-key, Accept, X-Requested-With, anthropic-version, anthropic-beta, x-opencode-*",
		"Access-Control-Max-Age":          "86400",
		"Access-Control-Expose-Headers":    "X-Correlation-Id, X-Request-Id",
	}
}

// CORS is middleware that handles CORS preflight requests.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range CORSHeaders() {
			w.Header().Set(k, v)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Recovery is middleware that recovers from panics and logs the error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				http.Error(w, `{"error":{"message":"Internal server error","type":"server_error"}}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// StripTrailingSlash is middleware that redirects trailing slash paths.
func StripTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			r.URL.Path = strings.TrimRight(path, "/")
		}
		next.ServeHTTP(w, r)
	})
}
