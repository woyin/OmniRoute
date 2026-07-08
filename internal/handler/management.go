package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/omniroute/omniroute/internal/auth"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

// --- Settings Handler ---

// SettingsHandler handles GET/PUT /api/settings.
type SettingsHandler struct {
	DB *sql.DB
}

func (h *SettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings := map[string]interface{}{
			"requireApiKey": false,
			"fetchTimeoutMs": 120000,
			"sseHeartbeatMs": 15000,
		}
		// Read from key_value store (exclude sensitive keys)
		sensitiveKeys := map[string]bool{"password": true, "jwtSecret": true}
		if h.DB != nil {
			rows, err := h.DB.Query("SELECT key, value FROM key_value WHERE namespace = 'settings'")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var key, value string
					if rows.Scan(&key, &value) == nil && !sensitiveKeys[key] {
						settings[key] = value
					}
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(settings)

	case http.MethodPut:
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		for key, value := range body {
			if strVal, ok := value.(string); ok {
				h.DB.Exec(
					"INSERT INTO key_value (namespace, key, value) VALUES ('settings', ?, ?) "+
						"ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value",
					key, strVal,
				)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Version Handler ---

// VersionHandler handles GET /api/system/version.
type VersionHandler struct{}

func (h *VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":    "4.0.0-go",
		"build":      "go-rewrite",
		"hostname":   hostname,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// --- Provider Detail/Update/Delete Handler ---

// ProviderDetailHandler handles GET/PUT/DELETE /api/providers/{id}.
type ProviderDetailHandler struct {
	DB *sql.DB
}

func (h *ProviderDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		// Try path param
		parts := splitPath(r.URL.Path)
		for i, p := range parts {
			if p == "providers" && i+1 < len(parts) {
				id = parts[i+1]
				break
			}
		}
	}
	if id == "" {
		http.Error(w, `{"error":{"message":"provider id required"}}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		conn, err := db.GetProviderConnection(h.DB, id)
		if err != nil || conn == nil {
			http.Error(w, `{"error":{"message":"provider not found"}}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conn)

	case http.MethodPut:
		var conn db.ProviderConnection
		if err := json.NewDecoder(r.Body).Decode(&conn); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		conn.ID = id
		if err := db.SaveProviderConnection(h.DB, conn); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(conn)

	case http.MethodDelete:
		if err := db.DeleteProviderConnection(h.DB, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Combo Detail/Delete Handler ---

// ComboDetailHandler handles GET/PUT/DELETE /api/combos/{id}.
type ComboDetailHandler struct {
	DB *sql.DB
}

func (h *ComboDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		parts := splitPath(r.URL.Path)
		for i, p := range parts {
			if p == "combos" && i+1 < len(parts) {
				id = parts[i+1]
				break
			}
		}
	}
	if id == "" {
		http.Error(w, `{"error":{"message":"combo id required"}}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		combo, err := db.GetCombo(h.DB, id)
		if err != nil || combo == nil {
			http.Error(w, `{"error":{"message":"combo not found"}}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(combo)

	case http.MethodPut:
		var combo db.Combo
		if err := json.NewDecoder(r.Body).Decode(&combo); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		combo.ID = id
		if err := db.SaveCombo(h.DB, combo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(combo)

	case http.MethodDelete:
		if err := db.DeleteCombo(h.DB, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- API Key Detail/Delete Handler ---

// APIKeyDetailHandler handles GET/DELETE /api/api-keys/{key}.
type APIKeyDetailHandler struct {
	DB *sql.DB
}

func (h *APIKeyDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := ""
	parts := splitPath(r.URL.Path)
	for i, p := range parts {
		if p == "api-keys" && i+1 < len(parts) {
			key = parts[i+1]
			break
		}
	}
	if key == "" {
		http.Error(w, `{"error":{"message":"api key required"}}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := db.DeleteAPIKey(h.DB, key); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Init Handler ---

// InitHandler handles POST /api/init — runs first-time initialization.
type InitHandler struct {
	DB *sql.DB
}

func (h *InitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure DB is reachable
	if h.DB != nil {
		if err := h.DB.Ping(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Database not reachable: " + err.Error(),
			})
			return
		}

		// Check table count
		var tableCount int
		h.DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&tableCount)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"message":    "Initialization complete",
			"tables":     tableCount,
			"providers":  len(registry.List()),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Initialization complete",
	})
}

// --- Shutdown Handler ---

// ShutdownHandler handles POST /api/shutdown — graceful shutdown.
type ShutdownHandler struct{}

func (h *ShutdownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	// TODO: signal graceful shutdown
}

// --- Generate UUID helper ---

func newUUID() string {
	return uuid.New().String()
}

// --- Path splitter ---

func splitPath(path string) []string {
	var parts []string
	for _, p := range splitString(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// --- Provider Test Handler ---

// ProviderTestHandler handles POST /api/providers/test — tests a provider connection.
type ProviderTestHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *ProviderTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Provider string `json:"provider"`
		APIKey   string `json:"apiKey"`
		BaseURL  string `json:"baseUrl,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Determine the models URL from the registry
	entry := registry.Get(body.Provider)
	modelsURL := ""
	if entry != nil && entry.ModelsURL != "" {
		modelsURL = entry.ModelsURL
	} else if entry != nil && entry.BaseURL != "" {
		modelsURL = strings.TrimRight(entry.BaseURL, "/") + "/models"
	}

	if modelsURL == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "No models URL configured for this provider",
		})
		return
	}

	// Try to fetch the models list
	req, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	if body.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+body.APIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    resp.StatusCode >= 200 && resp.StatusCode < 300,
		"statusCode": resp.StatusCode,
		"message":    fmt.Sprintf("Provider %s returned HTTP %d", body.Provider, resp.StatusCode),
	})
}

// --- Auth Status Handler ---

// AuthStatusHandler handles GET /api/auth/status.
type AuthStatusHandler struct {
	DB *sql.DB
}

func (h *AuthStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hasPassword":    auth.HasManagementPassword(h.DB),
		"setupComplete":  auth.IsSetupComplete(h.DB),
		"bootstrapWindow": auth.IsBootstrapWindow(h.DB),
		"requireLogin":   true,
	})
}

// --- Token Health Handler ---

// TokenHealthHandler handles GET /api/token-health.
type TokenHealthHandler struct {
	DB *sql.DB
}

func (h *TokenHealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Gather provider connection health info
	connections, err := db.ListProviderConnections(h.DB, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var tokens []map[string]interface{}
	for _, conn := range connections {
		entry := registry.Get(conn.Provider)
		token := map[string]interface{}{
			"provider":   conn.Provider,
			"connectionId": conn.ID,
			"hasApiKey":  conn.APIKey != "",
			"hasToken":   conn.AccessToken != "",
			"status":     "unknown",
		}
		if entry != nil {
			token["name"] = entry.Name
			token["authType"] = string(entry.AuthType)
		}
		if conn.APIKey != "" || conn.AccessToken != "" {
			token["status"] = "active"
		}
		tokens = append(tokens, token)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tokens": tokens,
		"total":  len(tokens),
	})
}

// --- Models Catalog Handler ---

// ModelsCatalogHandler handles GET /api/models/catalog.
type ModelsCatalogHandler struct {
	DB *sql.DB
}

func (h *ModelsCatalogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return all models from all providers with metadata
	var models []map[string]interface{}
	for _, entry := range registry.List() {
		for _, m := range entry.Models {
			model := map[string]interface{}{
				"id":       m.ID,
				"name":     m.Name,
				"provider": entry.ID,
				"object":   "model",
				"created":  time.Now().Unix(),
				"owned_by": entry.ID,
			}
			if m.ContextLength > 0 {
				model["context_length"] = m.ContextLength
			}
			if m.MaxOutputTokens > 0 {
				model["max_output_tokens"] = m.MaxOutputTokens
			}
			if m.SupportsReasoning {
				model["supports_reasoning"] = true
			}
			if m.SupportsVision {
				model["supports_vision"] = true
			}
			if m.SupportsXHighEffort {
				model["supports_xhigh_effort"] = true
			}
			models = append(models, model)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   models,
		"total":  len(models),
	})
}

// --- Provider Limits Handler ---

// ProviderLimitsHandler handles GET /api/usage/provider-limits.
type ProviderLimitsHandler struct {
	DB *sql.DB
}

func (h *ProviderLimitsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return provider rate limits from registry
	var limits []map[string]interface{}
	for _, entry := range registry.List() {
		limit := map[string]interface{}{
			"provider":          entry.ID,
			"authType":          string(entry.AuthType),
			"defaultContextLength": entry.DefaultContextLength,
			"hasFree":           entry.HasFree,
		}
		limits = append(limits, limit)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"limits": limits,
		"total":  len(limits),
	})
}

// --- Free Tier Summary Handler ---

// FreeTierSummaryHandler handles GET /api/free-tier/summary.
type FreeTierSummaryHandler struct{}

func (h *FreeTierSummaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var freeProviders []map[string]interface{}
	for _, entry := range registry.List() {
		if entry.HasFree || entry.AuthType == registry.AuthTypeNoAuth {
			freeProviders = append(freeProviders, map[string]interface{}{
				"provider": entry.ID,
				"name":     entry.Name,
				"hasFree":  entry.HasFree,
				"models":   len(entry.Models),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": freeProviders,
		"total":     len(freeProviders),
	})
}

// --- Combo Test Handler ---

// ComboTestHandler handles POST /api/combos/test — tests a combo configuration.
type ComboTestHandler struct {
	DB     *sql.DB
	Config *config.Config
}

func (h *ComboTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		ComboID string `json:"comboId"`
		Model   string `json:"model"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find the combo
	var combo *db.Combo
	if body.ComboID != "" {
		var err error
		combo, err = db.GetCombo(h.DB, body.ComboID)
		if err != nil || combo == nil {
			http.Error(w, "Combo not found", http.StatusNotFound)
			return
		}
	} else {
		// Try to find a combo that handles this model
		combos, err := db.ListCombos(h.DB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, c := range combos {
			for _, t := range c.Targets {
				if t.Model == body.Model {
					combo = &c
					break
				}
			}
			if combo != nil {
				break
			}
		}
	}

	if combo == nil {
		http.Error(w, "No matching combo found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comboId": combo.ID,
		"targets": len(combo.Targets),
		"status":  "available",
	})
}

// --- Combo Metrics Handler ---

// ComboMetricsHandler handles GET /api/combos/metrics.
type ComboMetricsHandler struct {
	DB *sql.DB
}

func (h *ComboMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	combos, err := db.ListCombos(h.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var metrics []map[string]interface{}
	for _, combo := range combos {
		metrics = append(metrics, map[string]interface{}{
			"comboId":  combo.ID,
			"name":     combo.Name,
			"targets":  len(combo.Targets),
			"strategy": combo.Strategy,
			"active":   combo.IsActive,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"combos": metrics,
		"total":  len(metrics),
	})
}

// --- Combo Auto Handler ---

// ComboAutoHandler handles POST /api/combos/auto — auto-generates a combo for a model.
type ComboAutoHandler struct {
	DB *sql.DB
}

func (h *ComboAutoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Model    string `json:"model"`
		Strategy string `json:"strategy,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	// Find all providers that support this model
	var targets []map[string]interface{}
	for _, entry := range registry.List() {
		if entry.GetModel(body.Model) != nil {
			targets = append(targets, map[string]interface{}{
				"provider": entry.ID,
				"model":    body.Model,
				"priority": len(targets) + 1,
			})
		}
	}

	// Check stored connections
	connections, _ := db.ListProviderConnections(h.DB, "")
	var availableTargets []map[string]interface{}
	for _, target := range targets {
		for _, conn := range connections {
			if conn.Provider == target["provider"] && (conn.APIKey != "" || conn.AccessToken != "") {
				target["hasCredentials"] = true
				break
			}
		}
		availableTargets = append(availableTargets, target)
	}

	strategy := body.Strategy
	if strategy == "" {
		strategy = "priority"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model":    body.Model,
		"strategy": strategy,
		"targets":  availableTargets,
		"total":    len(availableTargets),
	})
}
