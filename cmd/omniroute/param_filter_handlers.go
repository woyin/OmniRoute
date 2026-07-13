package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type paramFilterConfig struct {
	Block     []string                       `json:"block"`
	Allow     []string                       `json:"allow"`
	Models    map[string]map[string][]string `json:"models,omitempty"`
	AutoLearn bool                           `json:"autoLearn"`
}

func providerParamFiltersHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "id")
		switch r.Method {
		case http.MethodGet:
			var raw string
			err := dbConn.QueryRow("SELECT value FROM key_value WHERE namespace='provider_param_filters' AND key=?", provider).Scan(&raw)
			if err == sql.ErrNoRows {
				writeJSONResponse(w, paramFilterConfig{Block: []string{}, Allow: []string{}})
				return
			}
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to load parameter filters")
				return
			}
			var config paramFilterConfig
			if json.Unmarshal([]byte(raw), &config) != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to load parameter filters")
				return
			}
			if config.Block == nil {
				config.Block = []string{}
			}
			if config.Allow == nil {
				config.Allow = []string{}
			}
			writeJSONResponse(w, config)
		case http.MethodPut:
			var config paramFilterConfig
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&config); err != nil {
				jsonError(w, http.StatusBadRequest, "Invalid request")
				return
			}
			if config.Block == nil {
				config.Block = []string{}
			}
			if config.Allow == nil {
				config.Allow = []string{}
			}
			raw, err := json.Marshal(config)
			if err != nil {
				jsonError(w, http.StatusBadRequest, "Invalid request")
				return
			}
			_, err = dbConn.Exec(`INSERT INTO key_value(namespace,key,value) VALUES('provider_param_filters',?,?) ON CONFLICT(namespace,key) DO UPDATE SET value=excluded.value`, provider, string(raw))
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to save parameter filters")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true})
		case http.MethodDelete:
			if _, err := dbConn.Exec("DELETE FROM key_value WHERE namespace='provider_param_filters' AND key=?", provider); err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to delete parameter filters")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true})
		}
	}
}
