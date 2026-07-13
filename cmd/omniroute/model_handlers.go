package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

func modelAliasHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aliases, err := db.GetSetting(dbConn, "models", "aliases")
		if err != nil {
			jsonError(w, 500, "Failed to fetch aliases")
			return
		}
		switch r.Method {
		case http.MethodGet:
			alias := r.URL.Query().Get("alias")
			if alias == "" {
				writeJSONResponse(w, map[string]interface{}{"aliases": aliases})
				return
			}
			target, ok := aliases[alias].(string)
			if !ok {
				jsonError(w, 404, "Alias not found")
				return
			}
			parts := strings.SplitN(target, "/", 2)
			resolved := map[string]interface{}{"qualifiedId": target, "model": target, "source": "custom", "target": target}
			if len(parts) == 2 {
				resolved["provider"] = parts[0]
				resolved["model"] = parts[1]
			}
			writeJSONResponse(w, map[string]interface{}{"alias": alias, "resolved": resolved})
		case http.MethodPut:
			var body struct {
				Model string `json:"model"`
				Alias string `json:"alias"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				jsonError(w, 400, "Invalid JSON body")
				return
			}
			if body.Model == "" || body.Alias == "" {
				jsonError(w, 400, "model and alias are required")
				return
			}
			aliases[body.Alias] = body.Model
			if db.SetSetting(dbConn, "models", "aliases", aliases) != nil {
				jsonError(w, 500, "Failed to update alias")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true, "model": body.Model, "alias": body.Alias})
		case http.MethodDelete:
			alias := r.URL.Query().Get("alias")
			if alias == "" {
				jsonError(w, 400, "Alias required")
				return
			}
			delete(aliases, alias)
			if db.SetSetting(dbConn, "models", "aliases", aliases) != nil {
				jsonError(w, 500, "Failed to delete alias")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true})
		}
	}
}

func modelDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		modelID := strings.TrimPrefix(r.URL.Path, "/api/models/")
		for _, entry := range registry.List() {
			for _, model := range entry.Models {
				if model.ID == modelID {
					writeJSONResponse(w, map[string]interface{}{"id": model.ID, "name": model.Name, "provider": entry.ID, "contextLength": model.ContextLength, "maxInputTokens": model.MaxInputTokens, "maxOutputTokens": model.MaxOutputTokens, "supportsReasoning": model.SupportsReasoning, "supportsVision": model.SupportsVision})
					return
				}
			}
		}
		jsonError(w, 404, "Model not found")
	}
}

func openRouterCatalogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entry := registry.Get("openrouter")
		data := make([]registry.RegistryModel, 0)
		if entry != nil {
			data = append(data, entry.Models...)
		}
		writeJSONResponse(w, map[string]interface{}{"object": "list", "data": data, "meta": map[string]interface{}{"source": "local_catalog", "count": len(data)}})
	}
}
