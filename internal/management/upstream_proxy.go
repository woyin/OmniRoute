package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// UpstreamProxyHandler provides real DB-backed upstream proxy management endpoints.
type UpstreamProxyHandler struct {
	DB *sql.DB
}

// List returns all upstream proxy configurations.
func (h *UpstreamProxyHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type proxy struct {
		ID         string `json:"id"`
		ProviderID string `json:"providerId"`
		ProxyURL   string `json:"proxyUrl"`
		AuthType   string `json:"authType"`
		IsActive   bool   `json:"isActive"`
		CreatedAt  string `json:"createdAt"`
	}

	var proxies []proxy
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT id, provider_id, proxy_url, auth_type, is_active, created_at
			FROM upstream_proxy_config
			ORDER BY created_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var p proxy
				var active int
				if err := rows.Scan(&p.ID, &p.ProviderID, &p.ProxyURL, &p.AuthType, &active, &p.CreatedAt); err == nil {
					p.IsActive = active == 1
					proxies = append(proxies, p)
				}
			}
		}
	}

	if proxies == nil {
		proxies = []proxy{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   proxies,
		"total":  len(proxies),
	})
}

// Create creates a new upstream proxy configuration.
func (h *UpstreamProxyHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		ProviderID string `json:"providerId"`
		ProxyURL   string `json:"proxyUrl"`
		AuthType   string `json:"authType"`
		AuthValue  string `json:"authValue"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.ProviderID == "" || body.ProxyURL == "" {
		writeJSONError(w, http.StatusBadRequest, "providerId and proxyUrl are required")
		return
	}

	if h.DB != nil {
		id := generateID()
		_, err := h.DB.Exec(`
			INSERT INTO upstream_proxy_config (id, provider_id, proxy_url, auth_type, auth_value, is_active)
			VALUES (?, ?, ?, ?, ?, 1)
		`, id, body.ProviderID, body.ProxyURL, body.AuthType, body.AuthValue)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create upstream proxy")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         id,
			"providerId": body.ProviderID,
			"success":    true,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// Get returns a specific upstream proxy configuration by provider ID.
func (h *UpstreamProxyHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract providerId from URL: /upstream-proxy/{providerId}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	providerID := ""
	for i, p := range parts {
		if p == "upstream-proxy" && i+1 < len(parts) {
			providerID = parts[i+1]
			break
		}
	}

	if providerID == "" {
		writeJSONError(w, http.StatusBadRequest, "providerId required")
		return
	}

	if h.DB != nil {
		var id, pid, proxyURL, authType string
		var isActive int
		var createdAt string
		err := h.DB.QueryRow(`
			SELECT id, provider_id, proxy_url, auth_type, is_active, created_at
			FROM upstream_proxy_config
			WHERE provider_id = ?
		`, providerID).Scan(&id, &pid, &proxyURL, &authType, &isActive, &createdAt)
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "proxy not found")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to query proxy")
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         id,
			"providerId": pid,
			"proxyUrl":   proxyURL,
			"authType":   authType,
			"isActive":   isActive == 1,
			"createdAt":  createdAt,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// Delete removes an upstream proxy configuration.
func (h *UpstreamProxyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract providerId from URL: /upstream-proxy/{providerId}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	providerID := ""
	for i, p := range parts {
		if p == "upstream-proxy" && i+1 < len(parts) {
			providerID = parts[i+1]
			break
		}
	}

	if providerID == "" {
		writeJSONError(w, http.StatusBadRequest, "providerId required")
		return
	}

	if h.DB != nil {
		res, err := h.DB.Exec("DELETE FROM upstream_proxy_config WHERE provider_id = ?", providerID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete proxy")
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			writeJSONError(w, http.StatusNotFound, "proxy not found")
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"providerId": providerID,
			"deleted":    int(affected),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}
