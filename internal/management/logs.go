package management

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// LogsHandler provides real DB-backed log query endpoints using the call_logs table.
type LogsHandler struct {
	DB *sql.DB
}

// List returns paginated call logs with optional filtering.
func (h *LogsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit := parseIntParam(r, "limit", 50)
	if limit > 200 {
		limit = 200
	}
	offset := parseIntParam(r, "offset", 0)
	providerFilter := r.URL.Query().Get("provider")
	modelFilter := r.URL.Query().Get("model")
	statusFilter := r.URL.Query().Get("status")
	searchQuery := r.URL.Query().Get("q")

	type logEntry struct {
		ID           int    `json:"id"`
		Provider     string `json:"provider"`
		Model        string `json:"model"`
		StatusCode   int    `json:"statusCode"`
		LatencyMs    int    `json:"latencyMs"`
		RequestID    string `json:"requestId"`
		APIKey       string `json:"apiKey"`
		ErrorMessage string `json:"errorMessage"`
		CreatedAt    string `json:"createdAt"`
	}

	where := []string{}
	var args []interface{}

	if providerFilter != "" {
		where = append(where, "provider = ?")
		args = append(args, providerFilter)
	}
	if modelFilter != "" {
		where = append(where, "model = ?")
		args = append(args, modelFilter)
	}
	if statusFilter != "" {
		code, err := strconv.Atoi(statusFilter)
		if err == nil {
			where = append(where, "status_code = ?")
			args = append(args, code)
		} else if statusFilter == "error" {
			where = append(where, "status_code >= 400")
		} else if statusFilter == "success" {
			where = append(where, "status_code >= 200 AND status_code < 300")
		}
	}
	if searchQuery != "" {
		where = append(where, "(error_message LIKE ? OR request_id LIKE ?)")
		args = append(args, "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	total := 0
	var entries []logEntry

	if h.DB != nil {
		countArgs := make([]interface{}, len(args))
		copy(countArgs, args)
		h.DB.QueryRow("SELECT COUNT(*) FROM call_logs"+whereClause, countArgs...).Scan(&total)

		query := `SELECT id, provider, model, status_code, latency_ms,
			request_id, api_key, COALESCE(error_message, ''), created_at
			FROM call_logs` + whereClause + `
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?`
		args = append(args, limit, offset)

		rows, err := h.DB.Query(query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var e logEntry
				if err := rows.Scan(&e.ID, &e.Provider, &e.Model, &e.StatusCode,
					&e.LatencyMs, &e.RequestID, &e.APIKey, &e.ErrorMessage, &e.CreatedAt); err == nil {
					entries = append(entries, e)
				}
			}
		}
	}

	if entries == nil {
		entries = []logEntry{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   entries,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Detail returns a single log entry by ID.
func (h *LogsHandler) Detail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from URL: /logs/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	idStr := ""
	for i, p := range parts {
		if p == "logs" && i+1 < len(parts) {
			idStr = parts[i+1]
			break
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "valid numeric log id required")
		return
	}

	if h.DB != nil {
		var provider, model, requestID, apiKey, errorMessage, createdAt string
		var statusCode, latencyMs int
		err := h.DB.QueryRow(`
			SELECT provider, model, status_code, latency_ms, request_id,
			       api_key, COALESCE(error_message, ''), created_at
			FROM call_logs WHERE id = ?
		`, id).Scan(&provider, &model, &statusCode, &latencyMs, &requestID,
			&apiKey, &errorMessage, &createdAt)
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "log entry not found")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to query log")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           id,
			"provider":     provider,
			"model":        model,
			"statusCode":   statusCode,
			"latencyMs":    latencyMs,
			"requestId":    requestID,
			"apiKey":       apiKey,
			"errorMessage": errorMessage,
			"createdAt":    createdAt,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// Export exports call logs as JSON.
func (h *LogsHandler) Export(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	providerFilter := r.URL.Query().Get("provider")
	limit := parseIntParam(r, "limit", 1000)
	if limit > 5000 {
		limit = 5000
	}

	type logEntry struct {
		ID           int    `json:"id"`
		Provider     string `json:"provider"`
		Model        string `json:"model"`
		StatusCode   int    `json:"statusCode"`
		LatencyMs    int    `json:"latencyMs"`
		RequestID    string `json:"requestId"`
		ErrorMessage string `json:"errorMessage"`
		CreatedAt    string `json:"createdAt"`
	}

	var entries []logEntry
	if h.DB != nil {
		query := `SELECT id, provider, model, status_code, latency_ms,
			request_id, COALESCE(error_message, ''), created_at
			FROM call_logs`
		var args []interface{}
		if providerFilter != "" {
			query += " WHERE provider = ?"
			args = append(args, providerFilter)
		}
		query += " ORDER BY created_at DESC LIMIT ?"
		args = append(args, limit)

		rows, err := h.DB.Query(query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var e logEntry
				if err := rows.Scan(&e.ID, &e.Provider, &e.Model, &e.StatusCode,
					&e.LatencyMs, &e.RequestID, &e.ErrorMessage, &e.CreatedAt); err == nil {
					entries = append(entries, e)
				}
			}
		}
	}

	if entries == nil {
		entries = []logEntry{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"exported":  true,
		"count":     len(entries),
		"data":      entries,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Console returns recent console-style log output (tail).
func (h *LogsHandler) Console(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit := parseIntParam(r, "limit", 100)
	if limit > 500 {
		limit = 500
	}

	type logLine struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"level"`
		Provider  string `json:"provider"`
		Model     string `json:"model"`
		Status    int    `json:"status"`
		Latency   int    `json:"latencyMs"`
		Message   string `json:"message"`
	}

	var lines []logLine
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT created_at, provider, model, status_code, latency_ms, COALESCE(error_message, '')
			FROM call_logs
			ORDER BY created_at DESC
			LIMIT ?
		`, limit)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var l logLine
				if err := rows.Scan(&l.Timestamp, &l.Provider, &l.Model, &l.Status,
					&l.Latency, &l.Message); err == nil {
					if l.Status >= 400 {
						l.Level = "error"
					} else {
						l.Level = "info"
					}
					if l.Message == "" {
						l.Message = fmt.Sprintf("%s %s -> %d (%dms)", l.Provider, l.Model, l.Status, l.Latency)
					}
					lines = append(lines, l)
				}
			}
		}
	}

	if lines == nil {
		lines = []logLine{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  lines,
		"total": len(lines),
	})
}
