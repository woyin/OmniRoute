package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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

func modelTestHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ProviderID   string `json:"providerId"`
			ModelID      string `json:"modelId"`
			ConnectionID string `json:"connectionId"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			w.WriteHeader(400)
			writeJSONResponse(w, map[string]interface{}{"status": "error", "error": "Invalid JSON body"})
			return
		}
		result, status := runModelProbe(r.Context(), dbConn, body.ProviderID, body.ModelID, body.ConnectionID)
		w.WriteHeader(status)
		writeJSONResponse(w, result)
	}
}

func modelTestAllHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ProviderID   string   `json:"providerId"`
			ModelIDs     []string `json:"modelIds"`
			ConnectionID string   `json:"connectionId"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.ProviderID == "" || len(body.ModelIDs) < 1 || len(body.ModelIDs) > 100 {
			jsonError(w, 400, "Invalid request")
			return
		}
		results := map[string]interface{}{}
		for _, model := range body.ModelIDs {
			result, _ := runModelProbe(r.Context(), dbConn, body.ProviderID, model, body.ConnectionID)
			results[model] = result
		}
		writeJSONResponse(w, map[string]interface{}{"results": results})
	}
}

func runModelProbe(ctx context.Context, dbConn *sql.DB, providerID, modelID, connectionID string) (map[string]interface{}, int) {
	if providerID == "" || modelID == "" {
		return map[string]interface{}{"status": "error", "error": "providerId and modelId are required"}, 400
	}
	entry := registry.Get(providerID)
	var connection *db.ProviderConnection
	if connectionID != "" {
		connection, _ = db.GetProviderConnection(dbConn, connectionID)
		if connection == nil {
			return map[string]interface{}{"status": "error", "error": "Connection not found"}, 404
		}
		entry = registry.Get(connection.Provider)
	}
	if entry == nil {
		return map[string]interface{}{"status": "error", "error": "Provider not found"}, 404
	}
	if entry.BaseURL == "" {
		return map[string]interface{}{"status": "error", "error": "Provider has no HTTP endpoint"}, 400
	}
	endpoint := strings.TrimRight(entry.BaseURL, "/") + entry.ChatPath
	if entry.ChatPath == "" {
		endpoint = strings.TrimRight(entry.BaseURL, "/") + "/chat/completions"
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" && u.Scheme != "http" {
		return map[string]interface{}{"status": "error", "error": "Invalid provider endpoint"}, 400
	}
	payload, _ := json.Marshal(map[string]interface{}{"model": modelID, "messages": []map[string]string{{"role": "user", "content": "ping"}}, "max_tokens": 1, "stream": false})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if connection != nil {
		key := connection.APIKey
		if key == "" {
			key = connection.AccessToken
		}
		if key != "" {
			header := entry.AuthHeader
			if header == "" {
				header = "Authorization"
			}
			prefix := entry.AuthPrefix
			if prefix == "" && header == "Authorization" {
				prefix = "Bearer "
			}
			req.Header.Set(header, prefix+key)
		}
	}
	start := time.Now()
	client := http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return map[string]interface{}{"status": "error", "latencyMs": latency, "error": err.Error()}, 502
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return map[string]interface{}{"status": "error", "latencyMs": latency, "error": strings.TrimSpace(string(raw)), "statusCode": resp.StatusCode, "rateLimited": resp.StatusCode == 429}, resp.StatusCode
	}
	return map[string]interface{}{"status": "ok", "latencyMs": latency, "responseText": strings.TrimSpace(string(raw))}, 200
}
