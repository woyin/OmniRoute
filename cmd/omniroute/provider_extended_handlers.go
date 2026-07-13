package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

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
