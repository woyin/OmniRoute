package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

func usageCallLogsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := queryInt(r, "limit", 100, 1, 1000)
		logs, err := db.GetCallLogs(dbConn, r.URL.Query().Get("provider"), limit)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch call logs")
			return
		}
		writeJSONResponse(w, logs)
	}
}

func usageCallLogDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			jsonError(w, http.StatusNotFound, "Log not found")
			return
		}
		var log db.CallLog
		var created string
		err = dbConn.QueryRow(`SELECT id, provider, model, status_code, latency_ms, request_id, api_key, error_message, created_at FROM call_logs WHERE id = ?`, id).Scan(
			&log.ID, &log.Provider, &log.Model, &log.StatusCode, &log.LatencyMs, &log.RequestID, &log.APIKey, &log.ErrorMessage, &created)
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "Log not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch log")
			return
		}
		log.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		writeJSONResponse(w, log)
	}
}

func usageQuotaHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		providerFilter, connectionFilter := r.URL.Query().Get("provider"), r.URL.Query().Get("connectionId")
		connections, err := db.ListProviderConnections(dbConn, providerFilter)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch quota data")
			return
		}
		providers := make([]map[string]interface{}, 0)
		for _, connection := range connections {
			if !connection.IsActive || connectionFilter != "" && connection.ID != connectionFilter {
				continue
			}
			providers = append(providers, map[string]interface{}{
				"name": connection.Name, "provider": connection.Provider, "connectionId": connection.ID,
				"quotaUsed": 0, "quotaTotal": nil, "percentRemaining": 100, "resetAt": nil,
				"tokenStatus": tokenStatus(connection.ExpiresAt),
			})
		}
		writeJSONResponse(w, map[string]interface{}{
			"providers": providers,
			"meta":      map[string]interface{}{"generatedAt": time.Now().UTC().Format(time.RFC3339), "filters": map[string]interface{}{"provider": nullable(providerFilter), "connectionId": nullable(connectionFilter)}},
		})
	}
}

func usageConnectionHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection, err := db.GetProviderConnection(dbConn, chi.URLParam(r, "connectionId"))
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch usage")
			return
		}
		if connection == nil {
			jsonError(w, http.StatusNotFound, "Provider connection not found")
			return
		}
		var remaining, limit int
		var resetAt string
		_ = dbConn.QueryRow(`SELECT remaining, limit_val, reset_at FROM quota_snapshots WHERE provider = ? AND api_key = ? ORDER BY created_at DESC LIMIT 1`, connection.Provider, connection.APIKey).Scan(&remaining, &limit, &resetAt)
		writeJSONResponse(w, map[string]interface{}{"provider": connection.Provider, "connectionId": connection.ID, "remaining": remaining, "limit": limit, "resetAt": resetAt})
	}
}

func queryInt(r *http.Request, key string, fallback, min, max int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return fallback
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
func nullable(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
func tokenStatus(expires string) string {
	if expires == "" {
		return "valid"
	}
	t, err := time.Parse(time.RFC3339, expires)
	if err != nil {
		return "valid"
	}
	if time.Now().After(t) {
		return "expired"
	}
	if time.Until(t) <= 15*time.Minute {
		return "expiring"
	}
	return "valid"
}
func writeJSONResponse(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

var providerNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,80}$`)

func usageRequestLogsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs, err := db.GetCallLogs(dbConn, "", 200)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch logs")
			return
		}
		writeJSONResponse(w, logs)
	}
}

func usageProviderWindowCostsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("provider")))
		if !providerNamePattern.MatchString(provider) {
			jsonError(w, http.StatusBadRequest, "provider query param is required")
			return
		}
		connectionID := strings.TrimSpace(r.URL.Query().Get("connectionId"))
		var totalCost float64
		var totalRequests, inputTokens, outputTokens int
		err := dbConn.QueryRow(`
			SELECT COALESCE(SUM(cost),0), COUNT(*), COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0)
			FROM usage_history WHERE provider = ?
		`, provider).Scan(&totalCost, &totalRequests, &inputTokens, &outputTokens)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch provider USD costs")
			return
		}
		writeJSONResponse(w, map[string]interface{}{
			"provider": provider, "connectionId": nullable(connectionID), "totalCostUsd": totalCost,
			"totalRequests": totalRequests, "inputTokens": inputTokens, "outputTokens": outputTokens,
		})
	}
}
