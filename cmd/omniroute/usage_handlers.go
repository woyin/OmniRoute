package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
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

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
var validUsageRanges = map[string]bool{"1h": true, "24h": true, "7d": true, "30d": true}

func usageComboHealthHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rangeValue := r.URL.Query().Get("range")
		if !validUsageRanges[rangeValue] {
			jsonError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
		comboID := r.URL.Query().Get("comboId")
		if comboID != "" && !uuidPattern.MatchString(comboID) {
			jsonError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
		combos, err := db.ListCombos(dbConn)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch combo health")
			return
		}
		result := make([]map[string]interface{}, 0)
		for _, combo := range combos {
			if comboID != "" && combo.ID != comboID {
				continue
			}
			var requests, successes int
			var latency float64
			_ = dbConn.QueryRow(`SELECT COUNT(*), COALESCE(SUM(CASE WHEN success=1 THEN 1 ELSE 0 END),0), COALESCE(AVG(latency_ms),0) FROM usage_history WHERE model = ? AND created_at >= datetime('now', ?)`, combo.ID, rangeModifier(rangeValue)).Scan(&requests, &successes, &latency)
			successRate := 0.0
			if requests > 0 {
				successRate = float64(successes) / float64(requests)
			}
			result = append(result, map[string]interface{}{"id": combo.ID, "name": combo.Name, "strategy": combo.Strategy, "active": combo.IsActive, "requestCount": requests, "successRate": successRate, "averageLatencyMs": latency, "targets": combo.Targets})
		}
		if comboID != "" && len(result) == 0 {
			jsonError(w, http.StatusNotFound, "Combo not found")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"range": rangeValue, "generatedAt": time.Now().UTC().Format(time.RFC3339), "combos": result})
	}
}

func usageRouteExplainHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := chi.URLParam(r, "id")
		var log db.CallLog
		var created string
		err := dbConn.QueryRow(`SELECT id, provider, model, status_code, latency_ms, request_id, api_key, error_message, created_at FROM call_logs WHERE request_id = ? ORDER BY created_at DESC LIMIT 1`, requestID).Scan(&log.ID, &log.Provider, &log.Model, &log.StatusCode, &log.LatencyMs, &log.RequestID, &log.APIKey, &log.ErrorMessage, &created)
		if err == sql.ErrNoRows {
			jsonError(w, http.StatusNotFound, "Routing decision not found")
			return
		}
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to explain route")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"requestId": requestID, "provider": log.Provider, "model": log.Model, "status": log.StatusCode, "latencyMs": log.LatencyMs, "error": nullable(log.ErrorMessage), "timestamp": created})
	}
}

func rangeModifier(value string) string {
	switch value {
	case "1h":
		return "-1 hour"
	case "24h":
		return "-1 day"
	case "7d":
		return "-7 days"
	default:
		return "-30 days"
	}
}

func usageUtilizationHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rangeValue := r.URL.Query().Get("range")
		if !validUsageRanges[rangeValue] {
			jsonError(w, http.StatusBadRequest, "Invalid range. Must be one of: 1h, 24h, 7d, 30d")
			return
		}
		provider := r.URL.Query().Get("provider")
		bucket := map[string]int{"1h": 5, "24h": 60, "7d": 360, "30d": 1440}[rangeValue]
		query := `SELECT provider, created_at, remaining, limit_val, reset_at FROM quota_snapshots WHERE created_at >= datetime('now', ?)`
		args := []interface{}{rangeModifier(rangeValue)}
		if provider != "" {
			query += " AND provider = ?"
			args = append(args, provider)
		}
		query += " ORDER BY created_at"
		rows, err := dbConn.Query(query, args...)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch utilization data")
			return
		}
		defer rows.Close()
		data := make([]map[string]interface{}, 0)
		set := map[string]bool{}
		for rows.Next() {
			var p, created, reset string
			var remaining, limit int
			if rows.Scan(&p, &created, &remaining, &limit, &reset) == nil {
				set[p] = true
				data = append(data, map[string]interface{}{"provider": p, "timestamp": created, "remaining": remaining, "limit": limit, "resetAt": reset})
			}
		}
		providers := make([]string, 0, len(set))
		for p := range set {
			providers = append(providers, p)
		}
		sort.Strings(providers)
		writeJSONResponse(w, map[string]interface{}{"timeRange": rangeValue, "bucketSizeMinutes": bucket, "providers": providers, "data": data})
	}
}
