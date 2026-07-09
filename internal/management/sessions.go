package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SessionsHandler provides real DB-backed session management endpoints.
type SessionsHandler struct {
	DB *sql.DB
}

// List returns sessions from the session_account_affinity table.
func (h *SessionsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	providerFilter := r.URL.Query().Get("provider")

	type session struct {
		ID           int    `json:"id"`
		SessionID    string `json:"sessionId"`
		Provider     string `json:"provider"`
		ConnectionID string `json:"connectionId"`
		Model        string `json:"model"`
		LastUsedAt   string `json:"lastUsedAt"`
	}

	var sessions []session
	total := 0
	if h.DB != nil {
		where := ""
		var countArgs []interface{}
		if providerFilter != "" {
			where = " WHERE provider = ?"
			countArgs = append(countArgs, providerFilter)
		}

		h.DB.QueryRow("SELECT COUNT(*) FROM session_account_affinity"+where, countArgs...).Scan(&total)

		query := `SELECT id, session_id, provider, connection_id, model, last_used_at
			FROM session_account_affinity` + where + `
			ORDER BY last_used_at DESC
			LIMIT ? OFFSET ?`
		queryArgs := append(countArgs, limit, offset)

		rows, err := h.DB.Query(query, queryArgs...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var s session
				if err := rows.Scan(&s.ID, &s.SessionID, &s.Provider, &s.ConnectionID,
					&s.Model, &s.LastUsedAt); err == nil {
					sessions = append(sessions, s)
				}
			}
		}
	}

	if sessions == nil {
		sessions = []session{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object":  "list",
		"data":    sessions,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// Delete removes a session and its affinity records.
func (h *SessionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from URL: /sessions/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := ""
	for i, p := range parts {
		if p == "sessions" && i+1 < len(parts) {
			id = parts[i+1]
			break
		}
	}

	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "session id required")
		return
	}

	if h.DB != nil {
		// Delete from both session_account_affinity and reasoning_cache
		tx, err := h.DB.Begin()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to start transaction")
			return
		}

		res1, _ := tx.Exec("DELETE FROM session_account_affinity WHERE session_id = ?", id)
		res2, _ := tx.Exec("DELETE FROM reasoning_cache WHERE session_id = ?", id)

		if err := tx.Commit(); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete session")
			return
		}

		affinityDeleted, _ := res1.RowsAffected()
		cacheDeleted, _ := res2.RowsAffected()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":          true,
			"id":               id,
			"affinityDeleted":  int(affinityDeleted),
			"cacheDeleted":     int(cacheDeleted),
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}
