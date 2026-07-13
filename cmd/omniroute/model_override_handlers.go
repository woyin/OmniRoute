package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

func modelCapabilityOverridesHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
		case http.MethodPatch:
			var body struct {
				Target string `json:"target"`
				Key    string `json:"key"`
				Value  int    `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				jsonError(w, http.StatusBadRequest, "Invalid JSON body")
				return
			}
			provider, model, ok := canonicalModelTarget(body.Target)
			if !ok || body.Key != "max_token" || body.Value <= 0 {
				jsonError(w, http.StatusBadRequest, "Invalid model capability override")
				return
			}
			if err := db.SetModelCapabilityOverride(dbConn, provider, model, body.Key, body.Value); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to set model capability override")
				return
			}
		case http.MethodDelete:
			provider, model, ok := canonicalModelTarget(r.URL.Query().Get("target"))
			key := r.URL.Query().Get("key")
			if !ok || key != "max_token" {
				jsonError(w, http.StatusBadRequest, "target and key are required")
				return
			}
			if err := db.DeleteModelCapabilityOverride(dbConn, provider, model, key); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to delete model capability override")
				return
			}
		}
		overrides, err := db.ListModelCapabilityOverrides(dbConn)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to list model capability overrides")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"overrides": overrides})
	}
}

func canonicalModelTarget(target string) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(target), "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}
	provider := strings.TrimSpace(parts[0])
	if entry := registry.Get(provider); entry != nil {
		provider = entry.ID
	}
	return provider, strings.TrimSpace(parts[1]), true
}
