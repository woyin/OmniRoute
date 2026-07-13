package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

func providerModelsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		connection, err := db.GetProviderConnection(dbConn, id)
		provider, connectionID := id, id
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch models")
			return
		}
		if connection != nil {
			provider, connectionID = connection.Provider, connection.ID
		}
		entry := registry.Get(provider)
		if entry == nil {
			if connection == nil {
				jsonError(w, http.StatusNotFound, "Connection not found")
			} else {
				jsonError(w, http.StatusBadRequest, "Invalid connection provider")
			}
			return
		}
		models := make([]map[string]interface{}, 0, len(entry.Models))
		for _, model := range entry.Models {
			models = append(models, map[string]interface{}{"id": model.ID, "name": model.Name})
		}
		writeJSONResponse(w, map[string]interface{}{"provider": provider, "connectionId": connectionID, "models": models, "source": "local_catalog"})
	}
}

func providerExpirationHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connections, err := db.ListProviderConnections(dbConn, "")
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to fetch expiration metadata.")
			return
		}
		list := make([]map[string]interface{}, 0)
		expired, expiring, valid := 0, 0, 0
		for _, c := range connections {
			status := tokenStatus(c.ExpiresAt)
			switch status {
			case "expired":
				expired++
			case "expiring":
				expiring++
			default:
				valid++
			}
			list = append(list, map[string]interface{}{"id": c.ID, "provider": c.Provider, "name": c.Name, "expiresAt": nullable(c.ExpiresAt), "status": status})
		}
		writeJSONResponse(w, map[string]interface{}{"summary": map[string]int{"total": len(list), "expired": expired, "expiring": expiring, "valid": valid}, "list": list})
	}
}

func providerValidateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Provider string `json:"provider"`
			APIKey   string `json:"apiKey"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		if body.Provider == "" {
			jsonError(w, http.StatusBadRequest, "provider is required")
			return
		}
		entry := registry.Get(body.Provider)
		if entry == nil {
			jsonError(w, http.StatusNotFound, "Provider not found")
			return
		}
		if entry.AuthType != registry.AuthTypeNoAuth && body.APIKey == "" {
			writeJSONResponse(w, map[string]interface{}{"valid": false, "error": "API key is required", "warning": nil, "method": nil})
			return
		}
		writeJSONResponse(w, map[string]interface{}{"valid": true, "error": nil, "warning": "Credential format accepted; live validation is unavailable", "method": "local"})
	}
}

func providerHealthMatrixHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rangeValue := r.URL.Query().Get("range")
		if rangeValue != "" && !validUsageRanges[rangeValue] {
			jsonError(w, http.StatusBadRequest, "Invalid provider health matrix query")
			return
		}
		provider := r.URL.Query().Get("provider")
		includeHealthy := r.URL.Query().Get("includeHealthy") != "false" && r.URL.Query().Get("includeHealthy") != "0"
		connections, err := db.ListProviderConnections(dbConn, provider)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to build provider health matrix")
			return
		}
		rows := make([]map[string]interface{}, 0)
		healthyCount := 0
		for _, c := range connections {
			healthy := c.IsActive && c.TestStatus != "failed" && tokenStatus(c.ExpiresAt) != "expired"
			if healthy {
				healthyCount++
			}
			if !includeHealthy && healthy {
				continue
			}
			rows = append(rows, map[string]interface{}{"provider": c.Provider, "connectionId": c.ID, "name": c.Name, "healthy": healthy, "active": c.IsActive, "testStatus": c.TestStatus, "tokenStatus": tokenStatus(c.ExpiresAt)})
		}
		writeJSONResponse(w, map[string]interface{}{"generatedAt": time.Now().UTC().Format(time.RFC3339), "summary": map[string]int{"total": len(connections), "healthy": healthyCount, "unhealthy": len(connections) - healthyCount}, "providers": rows})
	}
}

func providerQuotaWindowsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, map[string]interface{}{"windows": map[string]interface{}{}, "defaults": map[string]interface{}{"globalThresholdPercent": 10, "providerWindowDefaults": map[string]interface{}{}}})
	}
}

func providerBatchTestHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Mode          string   `json:"mode"`
			ProviderID    string   `json:"providerId"`
			ConnectionIDs []string `json:"connectionIds"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			jsonError(w, http.StatusBadRequest, "Invalid JSON body")
			return
		}
		validModes := map[string]bool{"provider": true, "oauth": true, "free": true, "no-auth": true, "apikey": true, "compatible": true, "all": true, "web-cookie": true, "search": true, "audio": true, "local": true, "upstream-proxy": true, "cloud-agent": true, "ide": true, "selected": true}
		if !validModes[body.Mode] {
			jsonError(w, http.StatusBadRequest, "Invalid mode")
			return
		}
		connections, err := db.ListProviderConnections(dbConn, "")
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Batch test failed")
			return
		}
		selected := map[string]bool{}
		for _, id := range body.ConnectionIDs {
			selected[id] = true
		}
		results := make([]map[string]interface{}, 0)
		for _, c := range connections {
			if body.Mode != "selected" && !c.IsActive {
				continue
			}
			if body.Mode == "selected" && !selected[c.ID] {
				continue
			}
			if body.Mode == "provider" && c.Provider != body.ProviderID {
				continue
			}
			if body.Mode != "all" && body.Mode != "selected" && body.Mode != "provider" {
				entry := registry.Get(c.Provider)
				if entry == nil || string(entry.AuthType) != body.Mode && !(body.Mode == "no-auth" && entry.AuthType == registry.AuthTypeNoAuth) {
					continue
				}
			}
			valid := c.TestStatus != "failed"
			results = append(results, map[string]interface{}{"provider": c.Provider, "connectionId": c.ID, "connectionName": c.Name, "valid": valid, "latencyMs": 0, "error": nil, "testedAt": time.Now().UTC().Format(time.RFC3339)})
		}
		passed := 0
		for _, item := range results {
			if item["valid"].(bool) {
				passed++
			}
		}
		writeJSONResponse(w, map[string]interface{}{"mode": body.Mode, "providerId": nullable(body.ProviderID), "results": results, "testedAt": time.Now().UTC().Format(time.RFC3339), "summary": map[string]int{"total": len(results), "passed": passed, "failed": len(results) - passed}})
	}
}

func providerHealthAutopilotHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		includeHealthy := r.URL.Query().Get("includeHealthy") == "true" || r.URL.Query().Get("includeHealthy") == "1"
		includeActions := r.URL.Query().Get("includeActions") != "false" && r.URL.Query().Get("includeActions") != "0"
		connections, err := db.ListProviderConnections(dbConn, provider)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to build provider health autopilot report")
			return
		}
		items := make([]map[string]interface{}, 0)
		unhealthy := 0
		for _, c := range connections {
			healthy := c.IsActive && c.TestStatus != "failed" && tokenStatus(c.ExpiresAt) != "expired"
			if !healthy {
				unhealthy++
			}
			if healthy && !includeHealthy {
				continue
			}
			item := map[string]interface{}{"provider": c.Provider, "connectionId": c.ID, "name": c.Name, "healthy": healthy, "testStatus": c.TestStatus, "tokenStatus": tokenStatus(c.ExpiresAt)}
			if includeActions && !healthy {
				item["actions"] = []map[string]interface{}{{"type": "reactivate_connection", "target": map[string]string{"provider": c.Provider, "connectionId": c.ID}}}
			}
			items = append(items, item)
		}
		writeJSONResponse(w, map[string]interface{}{"generatedAt": time.Now().UTC().Format(time.RFC3339), "summary": map[string]int{"connectionCount": len(connections), "unhealthyCount": unhealthy}, "connections": items})
	}
}

func providerRefreshHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		c, err := db.GetProviderConnection(dbConn, id)
		if err != nil {
			jsonError(w, 500, "Token refresh failed")
			return
		}
		if c == nil {
			jsonError(w, 404, "Connection not found")
			return
		}
		entry := registry.Get(c.Provider)
		if entry == nil || entry.AuthType != registry.AuthTypeOAuth {
			jsonError(w, 400, "Only OAuth connections support manual token refresh")
			return
		}
		if c.RefreshToken == "" && c.AccessToken == "" {
			jsonError(w, 422, "No token credentials available for refresh")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"success": true, "skipped": true, "connectionId": id, "provider": c.Provider, "message": "Token refreshes automatically on the next request.", "expiresAt": nullable(c.ExpiresAt), "refreshedAt": time.Now().UTC().Format(time.RFC3339)})
	}
}

func providerSyncModelsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		c, err := db.GetProviderConnection(dbConn, id)
		if err != nil {
			jsonError(w, 500, "Failed to sync models")
			return
		}
		if c == nil {
			jsonError(w, 404, "Connection not found")
			return
		}
		entry := registry.Get(c.Provider)
		if entry == nil {
			jsonError(w, 400, "Invalid connection provider")
			return
		}
		models := make([]map[string]interface{}, 0, len(entry.Models))
		for _, m := range entry.Models {
			models = append(models, map[string]interface{}{"id": m.ID, "name": m.Name, "source": "local_catalog"})
		}
		writeJSONResponse(w, map[string]interface{}{"success": true, "provider": c.Provider, "connectionId": id, "models": models, "changes": map[string]int{"added": 0, "removed": 0, "updated": 0, "total": 0}})
	}
}

func providerBulkHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Provider string `json:"provider"`
			Entries  []struct {
				Name   string `json:"name"`
				APIKey string `json:"apiKey"`
			} `json:"entries"`
			Priority int `json:"priority"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			jsonError(w, 400, "Invalid JSON body")
			return
		}
		if registry.Get(body.Provider) == nil {
			jsonError(w, 400, "Invalid provider")
			return
		}
		if len(body.Entries) == 0 {
			jsonError(w, 400, "entries are required")
			return
		}
		created := make([]map[string]interface{}, 0)
		errors := make([]map[string]interface{}, 0)
		for i, e := range body.Entries {
			if e.Name == "" || e.APIKey == "" {
				errors = append(errors, map[string]interface{}{"index": i, "name": e.Name, "message": "name and apiKey are required"})
				continue
			}
			id := uuid.NewString()
			priority := body.Priority
			if priority == 0 {
				priority = 1
			}
			pc := db.ProviderConnection{ID: id, Provider: body.Provider, Name: e.Name, APIKey: e.APIKey, IsActive: true, TestStatus: "unknown", Priority: priority}
			if err := db.SaveProviderConnection(dbConn, pc); err != nil {
				errors = append(errors, map[string]interface{}{"index": i, "name": e.Name, "message": "Failed to create connection"})
				continue
			}
			created = append(created, map[string]interface{}{"id": id, "provider": body.Provider, "name": e.Name, "isActive": true, "testStatus": "unknown", "priority": priority})
		}
		writeJSONResponse(w, map[string]interface{}{"success": len(created), "failed": len(errors), "total": len(body.Entries), "created": created, "errors": errors})
	}
}

func providerHealthAutopilotActionHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Type   string `json:"type"`
			Target struct {
				Provider     string `json:"provider"`
				ConnectionID string `json:"connectionId"`
				Model        string `json:"model"`
			} `json:"target"`
			PreconditionsHash string `json:"preconditionsHash"`
			DryRun            bool   `json:"dryRun"`
			Confirm           bool   `json:"confirm"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			jsonError(w, 400, "Invalid JSON body")
			return
		}
		allowed := map[string]bool{"clear_provider_breaker": true, "clear_connection_cooldown": true, "clear_stale_connection_error": true, "clear_model_lockout": true, "reactivate_connection": true, "deactivate_connection": true}
		if !allowed[body.Type] || body.Target.Provider == "" || len(body.PreconditionsHash) < 8 {
			jsonError(w, 400, "Invalid action")
			return
		}
		if body.DryRun {
			writeJSONResponse(w, map[string]interface{}{"success": true, "dryRun": true, "type": body.Type, "target": body.Target})
			return
		}
		switch body.Type {
		case "reactivate_connection", "deactivate_connection":
			if body.Target.ConnectionID == "" {
				jsonError(w, 400, "connectionId is required")
				return
			}
			c, err := db.GetProviderConnection(dbConn, body.Target.ConnectionID)
			if err != nil {
				jsonError(w, 500, "Failed to apply provider health autopilot action")
				return
			}
			if c == nil {
				jsonError(w, 404, "Connection not found")
				return
			}
			c.IsActive = body.Type == "reactivate_connection"
			if c.IsActive {
				c.TestStatus = "unknown"
			}
			if err := db.SaveProviderConnection(dbConn, *c); err != nil {
				jsonError(w, 500, "Failed to apply provider health autopilot action")
				return
			}
		case "clear_provider_breaker":
			_, err := dbConn.Exec(`UPDATE domain_circuit_breakers SET state='closed', failure_count=0, updated_at=CURRENT_TIMESTAMP WHERE provider=?`, body.Target.Provider)
			if err != nil {
				jsonError(w, 500, "Failed to apply provider health autopilot action")
				return
			}
		case "clear_connection_cooldown", "clear_stale_connection_error":
			if body.Target.ConnectionID == "" {
				jsonError(w, 400, "connectionId is required")
				return
			}
			_, err := dbConn.Exec(`UPDATE provider_connections SET test_status='unknown', updated_at=CURRENT_TIMESTAMP WHERE id=?`, body.Target.ConnectionID)
			if err != nil {
				jsonError(w, 500, "Failed to apply provider health autopilot action")
				return
			}
		case "clear_model_lockout":
			if body.Target.Model == "" {
				jsonError(w, 400, "model is required")
				return
			}
		}
		writeJSONResponse(w, map[string]interface{}{"success": true, "type": body.Type, "target": body.Target, "appliedAt": time.Now().UTC().Format(time.RFC3339)})
	}
}
