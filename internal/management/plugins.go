package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// PluginsHandler provides plugin management endpoints.
// Plugins are tracked in the key_value store since there's no dedicated plugins table.
type PluginsHandler struct {
	DB *sql.DB
}

// List returns all installed plugins from the key_value store.
func (h *PluginsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var plugins []interface{}
	if h.DB != nil {
		var val string
		if err := h.DB.QueryRow("SELECT value FROM key_value WHERE namespace = 'plugins' AND key = 'installed'").Scan(&val); err == nil {
			json.Unmarshal([]byte(val), &plugins)
		}
	}

	if plugins == nil {
		plugins = []interface{}{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"plugins": plugins,
		"total":   len(plugins),
	})
}

// Install registers a plugin installation.
func (h *PluginsHandler) Install(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Source  string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	if h.DB != nil {
		// Read existing plugins
		var existing []map[string]interface{}
		var val string
		if err := h.DB.QueryRow("SELECT value FROM key_value WHERE namespace = 'plugins' AND key = 'installed'").Scan(&val); err == nil {
			json.Unmarshal([]byte(val), &existing)
		}

		// Add new plugin
		plugin := map[string]interface{}{
			"name":        body.Name,
			"version":     body.Version,
			"source":      body.Source,
			"installedAt": time.Now().UTC().Format(time.RFC3339),
			"enabled":     true,
		}
		existing = append(existing, plugin)

		// Save back
		data, _ := json.Marshal(existing)
		_, err := h.DB.Exec(`
			INSERT INTO key_value (namespace, key, value)
			VALUES ('plugins', 'installed', ?)
			ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value
		`, string(data))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to save plugin")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"plugin":  plugin,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}
