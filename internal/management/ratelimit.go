package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// RateLimitHandler provides real DB-backed rate limit management endpoints.
// Uses the api_key_token_limits table for per-key rate limits and key_value for global config.
type RateLimitHandler struct {
	DB *sql.DB
}

// List returns all rate limit configurations.
func (h *RateLimitHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type rateLimit struct {
		ID                string `json:"id"`
		APIKeyID          string `json:"apiKeyId"`
		Model             string `json:"model"`
		MaxTokensPerDay   int    `json:"maxTokensPerDay"`
		MaxTokensPerMonth int    `json:"maxTokensPerMonth"`
		TokensUsedToday   int    `json:"tokensUsedToday"`
		TokensUsedMonth   int    `json:"tokensUsedMonth"`
		ResetDay          string `json:"resetDay"`
		CreatedAt         string `json:"createdAt"`
		UpdatedAt         string `json:"updatedAt"`
	}

	var limits []rateLimit
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT id, api_key_id, COALESCE(model, ''),
			       max_tokens_per_day, max_tokens_per_month,
			       tokens_used_today, tokens_used_month,
			       COALESCE(reset_day, ''), created_at, updated_at
			FROM api_key_token_limits
			ORDER BY updated_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var rl rateLimit
				if err := rows.Scan(&rl.ID, &rl.APIKeyID, &rl.Model,
					&rl.MaxTokensPerDay, &rl.MaxTokensPerMonth,
					&rl.TokensUsedToday, &rl.TokensUsedMonth,
					&rl.ResetDay, &rl.CreatedAt, &rl.UpdatedAt); err == nil {
					limits = append(limits, rl)
				}
			}
		}
	}

	// Also read global rate limit config from key_value store
	globalEnabled := true
	globalMaxRequests := 100
	globalWindowMs := 60000
	if h.DB != nil {
		var val string
		if err := h.DB.QueryRow("SELECT value FROM key_value WHERE namespace = 'rate_limit' AND key = 'config'").Scan(&val); err == nil {
			var cfg map[string]interface{}
			if json.Unmarshal([]byte(val), &cfg) == nil {
				if v, ok := cfg["enabled"].(bool); ok {
					globalEnabled = v
				}
				if v, ok := cfg["maxRequests"].(float64); ok {
					globalMaxRequests = int(v)
				}
				if v, ok := cfg["windowMs"].(float64); ok {
					globalWindowMs = int(v)
				}
			}
		}
	}

	if limits == nil {
		limits = []rateLimit{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   limits,
		"total":  len(limits),
		"global": map[string]interface{}{
			"enabled":     globalEnabled,
			"maxRequests": globalMaxRequests,
			"windowMs":    globalWindowMs,
		},
	})
}

// Create creates a new per-key rate limit rule.
func (h *RateLimitHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		APIKeyID          string `json:"apiKeyId"`
		Model             string `json:"model"`
		MaxTokensPerDay   int    `json:"maxTokensPerDay"`
		MaxTokensPerMonth int    `json:"maxTokensPerMonth"`
		ResetDay          string `json:"resetDay"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.APIKeyID == "" {
		writeJSONError(w, http.StatusBadRequest, "apiKeyId is required")
		return
	}

	if h.DB != nil {
		id := generateID()
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := h.DB.Exec(`
			INSERT INTO api_key_token_limits
			(id, api_key_id, model, max_tokens_per_day, max_tokens_per_month, reset_day, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, id, body.APIKeyID, body.Model, body.MaxTokensPerDay, body.MaxTokensPerMonth, body.ResetDay, now, now)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create rate limit")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      id,
			"success": true,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// GetConfig returns the global rate limit configuration.
func (h *RateLimitHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	enabled := true
	maxRequests := 100
	windowMs := 60000
	if h.DB != nil {
		var val string
		if err := h.DB.QueryRow("SELECT value FROM key_value WHERE namespace = 'rate_limit' AND key = 'config'").Scan(&val); err == nil {
			var cfg map[string]interface{}
			if json.Unmarshal([]byte(val), &cfg) == nil {
				if v, ok := cfg["enabled"].(bool); ok {
					enabled = v
				}
				if v, ok := cfg["maxRequests"].(float64); ok {
					maxRequests = int(v)
				}
				if v, ok := cfg["windowMs"].(float64); ok {
					windowMs = int(v)
				}
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":     enabled,
		"maxRequests": maxRequests,
		"windowMs":    windowMs,
	})
}

// SetConfig sets the global rate limit configuration in the key_value store.
func (h *RateLimitHandler) SetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if h.DB != nil {
		data, _ := json.Marshal(body)
		_, err := h.DB.Exec(`
			INSERT INTO key_value (namespace, key, value)
			VALUES ('rate_limit', 'config', ?)
			ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value
		`, string(data))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to save rate limit config")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"config":  body,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// Usage returns current rate limit usage for a given API key.
func (h *RateLimitHandler) Usage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := r.URL.Query().Get("apiKey")
	if apiKey == "" {
		// Return aggregate usage across all keys
		type usageRow struct {
			APIKeyID          string `json:"apiKeyId"`
			Model             string `json:"model"`
			TokensUsedToday   int    `json:"tokensUsedToday"`
			TokensUsedMonth   int    `json:"tokensUsedMonth"`
			MaxTokensPerDay   int    `json:"maxTokensPerDay"`
			MaxTokensPerMonth int    `json:"maxTokensPerMonth"`
		}
		var rows []usageRow
		if h.DB != nil {
			result, err := h.DB.Query(`
				SELECT api_key_id, COALESCE(model, ''), tokens_used_today, tokens_used_month,
				       max_tokens_per_day, max_tokens_per_month
				FROM api_key_token_limits
				ORDER BY api_key_id
			`)
			if err == nil {
				defer result.Close()
				for result.Next() {
					var u usageRow
					if err := result.Scan(&u.APIKeyID, &u.Model, &u.TokensUsedToday,
						&u.TokensUsedMonth, &u.MaxTokensPerDay, &u.MaxTokensPerMonth); err == nil {
						rows = append(rows, u)
					}
				}
			}
		}
		if rows == nil {
			rows = []usageRow{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data":   rows,
			"total":  len(rows),
		})
		return
	}

	// Return usage for specific key
	todayTokens := 0
	monthTokens := 0
	maxDay := 0
	maxMonth := 0
	if h.DB != nil {
		h.DB.QueryRow(`
			SELECT COALESCE(SUM(tokens_used_today), 0), COALESCE(SUM(tokens_used_month), 0),
			       COALESCE(SUM(max_tokens_per_day), 0), COALESCE(SUM(max_tokens_per_month), 0)
			FROM api_key_token_limits WHERE api_key_id = ?
		`, apiKey).Scan(&todayTokens, &monthTokens, &maxDay, &maxMonth)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"apiKey":            apiKey,
		"tokensUsedToday":   todayTokens,
		"tokensUsedMonth":   monthTokens,
		"maxTokensPerDay":   maxDay,
		"maxTokensPerMonth": maxMonth,
		"dailyPercent":      percentOf(todayTokens, maxDay),
		"monthlyPercent":    percentOf(monthTokens, maxMonth),
	})
}

// Delete removes a rate limit rule by ID.
func (h *RateLimitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id query parameter required")
		return
	}

	if h.DB != nil {
		res, err := h.DB.Exec("DELETE FROM api_key_token_limits WHERE id = ?", id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete rate limit")
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			writeJSONError(w, http.StatusNotFound, "rate limit not found")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
			"deleted": int(affected),
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// percentOf calculates what percentage used is of limit.
func percentOf(used, limit int) float64 {
	if limit <= 0 {
		return 0
	}
	return float64(used) / float64(limit) * 100
}

// parseIntParam safely parses a query parameter as int with a default value.
func parseIntParam(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
