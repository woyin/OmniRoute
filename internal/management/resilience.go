package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ResilienceHandler provides real DB-backed resilience status endpoints.
type ResilienceHandler struct {
	DB *sql.DB
}

// Status returns the overall resilience status by aggregating circuit breaker states.
func (h *ResilienceHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type breakerSummary struct {
		Total    int `json:"total"`
		Closed   int `json:"closed"`
		Open     int `json:"open"`
		HalfOpen int `json:"halfOpen"`
	}

	summary := breakerSummary{}
	if h.DB != nil {
		rows, err := h.DB.Query("SELECT state, COUNT(*) FROM domain_circuit_breakers GROUP BY state")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var state string
				var count int
				if err := rows.Scan(&state, &count); err == nil {
					summary.Total += count
					switch strings.ToLower(state) {
					case "closed":
						summary.Closed += count
					case "open":
						summary.Open += count
					case "half-open", "halfopen":
						summary.HalfOpen += count
					}
				}
			}
		}
	}

	healthy := summary.Open == 0

	json.NewEncoder(w).Encode(map[string]interface{}{
		"layers":          []string{"circuit-breaker", "fallback", "retry"},
		"healthy":         healthy,
		"circuitBreakers": summary,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	})
}

// CircuitBreakersList returns all circuit breakers from the domain_circuit_breakers table.
func (h *ResilienceHandler) CircuitBreakersList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type breaker struct {
		ID               string  `json:"id"`
		Provider         string  `json:"provider"`
		State            string  `json:"state"`
		FailureCount     int     `json:"failureCount"`
		LastFailureAt    *string `json:"lastFailureAt"`
		LastStateChange  *string `json:"lastStateChangeAt"`
		CreatedAt        string  `json:"createdAt"`
		UpdatedAt        string  `json:"updatedAt"`
	}

	var breakers []breaker
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT id, provider, state, failure_count,
			       COALESCE(last_failure_at, ''), COALESCE(last_state_change_at, ''),
			       created_at, updated_at
			FROM domain_circuit_breakers
			ORDER BY updated_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var b breaker
				var lastFail, lastChange string
				if err := rows.Scan(&b.ID, &b.Provider, &b.State, &b.FailureCount,
					&lastFail, &lastChange, &b.CreatedAt, &b.UpdatedAt); err == nil {
					if lastFail != "" {
						b.LastFailureAt = &lastFail
					}
					if lastChange != "" {
						b.LastStateChange = &lastChange
					}
					breakers = append(breakers, b)
				}
			}
		}
	}

	if breakers == nil {
		breakers = []breaker{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   breakers,
		"total":  len(breakers),
	})
}

// CircuitBreakerReset resets a specific circuit breaker to closed state.
func (h *ResilienceHandler) CircuitBreakerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from URL path: /resilience/circuit-breakers/{id}/reset
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := ""
	for i, p := range parts {
		if p == "circuit-breakers" && i+1 < len(parts) {
			id = parts[i+1]
			break
		}
	}

	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "circuit breaker id required")
		return
	}

	if h.DB != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		res, err := h.DB.Exec(`
			UPDATE domain_circuit_breakers
			SET state = 'closed', failure_count = 0, last_state_change_at = ?, updated_at = ?
			WHERE id = ?
		`, now, now, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to reset circuit breaker")
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			writeJSONError(w, http.StatusNotFound, "circuit breaker not found")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
			"state":   "closed",
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// CircuitBreakerResetAll resets all circuit breakers to closed state.
func (h *ResilienceHandler) CircuitBreakerResetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.DB != nil {
		now := time.Now().UTC().Format(time.RFC3339)
		res, err := h.DB.Exec(`
			UPDATE domain_circuit_breakers
			SET state = 'closed', failure_count = 0, last_state_change_at = ?, updated_at = ?
			WHERE state != 'closed'
		`, now, now)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to reset circuit breakers")
			return
		}
		affected, _ := res.RowsAffected()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"reset":    int(affected),
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// FallbackChainsList returns all domain fallback chains.
func (h *ResilienceHandler) FallbackChainsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type chain struct {
		ID            string `json:"id"`
		Provider      string `json:"provider"`
		Model         string `json:"model"`
		FallbackOrder string `json:"fallbackOrder"`
		IsActive      bool   `json:"isActive"`
		CreatedAt     string `json:"createdAt"`
	}

	var chains []chain
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT id, provider, model, fallback_order, is_active, created_at
			FROM domain_fallback_chains
			ORDER BY created_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var c chain
				var active int
				if err := rows.Scan(&c.ID, &c.Provider, &c.Model, &c.FallbackOrder, &active, &c.CreatedAt); err == nil {
					c.IsActive = active == 1
					chains = append(chains, c)
				}
			}
		}
	}

	if chains == nil {
		chains = []chain{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   chains,
		"total":  len(chains),
	})
}

// FallbackChainsCreate creates a new domain fallback chain.
func (h *ResilienceHandler) FallbackChainsCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		ID            string `json:"id"`
		Provider      string `json:"provider"`
		Model         string `json:"model"`
		FallbackOrder string `json:"fallbackOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Provider == "" {
		writeJSONError(w, http.StatusBadRequest, "provider is required")
		return
	}

	if h.DB != nil {
		if body.ID == "" {
			body.ID = generateID()
		}
		if body.FallbackOrder == "" {
			body.FallbackOrder = "[]"
		}
		_, err := h.DB.Exec(`
			INSERT INTO domain_fallback_chains (id, provider, model, fallback_order, is_active)
			VALUES (?, ?, ?, ?, 1)
		`, body.ID, body.Provider, body.Model, body.FallbackOrder)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create fallback chain")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      body.ID,
			"success": true,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}
