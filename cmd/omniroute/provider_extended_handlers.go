package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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
