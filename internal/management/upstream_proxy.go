package management

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UpstreamProxyHandler struct {
	DB *sql.DB
}

type upstreamProxyConfig struct {
	ID                      int64                  `json:"id"`
	ProviderID              string                 `json:"providerId"`
	Mode                    string                 `json:"mode"`
	CLIProxyAPIModelMapping map[string]interface{} `json:"cliproxyapiModelMapping"`
	NativePriority          int                    `json:"nativePriority"`
	CLIProxyAPIPriority     int                    `json:"cliproxyapiPriority"`
	Enabled                 bool                   `json:"enabled"`
	Family                  string                 `json:"family"`
	CreatedAt               string                 `json:"createdAt"`
	UpdatedAt               string                 `json:"updatedAt"`
}

func (h *UpstreamProxyHandler) Get(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "providerId")
	if providerID == "" {
		writeJSONError(w, http.StatusBadRequest, "providerId required")
		return
	}

	config, err := h.get(providerID)
	if err == sql.ErrNoRows {
		writeJSON(w, map[string]interface{}{"enabled": false, "mode": "native"})
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to query upstream proxy")
		return
	}
	writeJSON(w, config)
}

func (h *UpstreamProxyHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "providerId")
	if providerID == "" {
		writeJSONError(w, http.StatusBadRequest, "providerId required")
		return
	}

	var body struct {
		Mode    string `json:"mode"`
		Enabled *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Mode == "" {
		body.Mode = "native"
	}
	if body.Mode != "native" && body.Mode != "cliproxyapi" && body.Mode != "fallback" {
		writeJSONError(w, http.StatusBadRequest, "mode must be native, cliproxyapi, or fallback")
		return
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	if h.DB == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not available")
		return
	}

	_, err := h.DB.Exec(`
		INSERT INTO upstream_proxy_config (provider_id, mode, enabled)
		VALUES (?, ?, ?)
		ON CONFLICT(provider_id) DO UPDATE SET
			mode = excluded.mode,
			enabled = excluded.enabled,
			updated_at = datetime('now')
	`, providerID, body.Mode, enabled)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save upstream proxy")
		return
	}

	config, err := h.get(providerID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to query upstream proxy")
		return
	}
	writeJSON(w, config)
}

func (h *UpstreamProxyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "providerId")
	if providerID == "" {
		writeJSONError(w, http.StatusBadRequest, "providerId required")
		return
	}
	if h.DB == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not available")
		return
	}

	result, err := h.DB.Exec("DELETE FROM upstream_proxy_config WHERE provider_id = ?", providerID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to delete upstream proxy")
		return
	}
	deleted, _ := result.RowsAffected()
	writeJSON(w, map[string]bool{"deleted": deleted > 0})
}

func (h *UpstreamProxyHandler) get(providerID string) (upstreamProxyConfig, error) {
	var config upstreamProxyConfig
	var mapping sql.NullString
	var enabled int
	if h.DB == nil {
		return config, sql.ErrConnDone
	}
	err := h.DB.QueryRow(`
		SELECT id, provider_id, mode, cliproxyapi_model_mapping,
		       native_priority, cliproxyapi_priority, enabled, family,
		       created_at, updated_at
		FROM upstream_proxy_config WHERE provider_id = ?
	`, providerID).Scan(
		&config.ID, &config.ProviderID, &config.Mode, &mapping,
		&config.NativePriority, &config.CLIProxyAPIPriority, &enabled, &config.Family,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return config, err
	}
	config.Enabled = enabled != 0
	if mapping.Valid {
		_ = json.Unmarshal([]byte(mapping.String), &config.CLIProxyAPIModelMapping)
	}
	return config, nil
}

func writeJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
