package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// --- Admin handlers ---

func adminConcurrencyHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"activeRequests": 0,
			"maxConcurrent":  100,
			"queueDepth":     0,
		})
	}
}

// --- Assess handler ---

func assessHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"score": 0.0, "assessment": "not implemented"})
	}
}

// --- Copilot chat handler ---

func copilotChatHandler(dbConn *sql.DB, cfg interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Copilot chat not yet implemented in Go rewrite",
				"type":    "not_implemented",
			},
		})
	}
}

// --- Docs handlers ---

func docsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"version": "4.0.0-go",
			"endpoints": []string{
				"POST /api/v1/chat/completions",
				"POST /api/v1/embeddings",
				"GET  /api/v1/models",
				"POST /api/auth/login",
			},
		})
	}
}

func docsCodexCLIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"codexCLI": map[string]interface{}{
				"endpoint": "/api/v1/chat/completions",
				"format":   "openai",
			},
		})
	}
}

// --- Local Redis handlers ---

func localRedisStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"running": false, "port": 0})
	}
}

func localRedisStartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"started": false, "message": "Redis not available in Go rewrite"})
	}
}

func localRedisStopHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"stopped": true})
	}
}

// --- Middleware hooks handlers ---

func middlewareHooksListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

func middlewareHooksDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"name": name, "enabled": false})
	}
}

// --- OpenAPI handlers ---

func openapiSpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"openapi": "3.1.0",
			"info":    map[string]interface{}{"title": "OmniRoute API", "version": "4.0.0-go"},
			"paths":   map[string]interface{}{},
		})
	}
}

func openapiTryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "OpenAPI try not implemented"})
	}
}

// --- Policies handler ---

func policiesHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

// --- Proxy fallback test handler ---

func proxyFallbackTestHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"working": false, "latency": 0})
	}
}

// --- Session pools handlers ---

func sessionPoolsListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

func sessionPoolDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "provider")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"provider": provider, "sessions": []interface{}{}})
	}
}

// --- Synced available models handler ---

func syncedAvailableModelsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

// --- GitHub skills handler ---

func githubSkillsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"object": "list", "data": []interface{}{}})
	}
}

// --- Free provider rankings handler ---

func freeProviderRankingsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"rankings": []interface{}{}})
	}
}

