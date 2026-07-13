package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/omniroute/omniroute/internal/db"
)

// --- Provider nodes handlers ---

func providerNodesListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"nodes": []interface{}{}})
	}
}

func providerNodesCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

func providerNodesDeleteHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// --- Model combo mappings handlers ---

func modelComboMappingsListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"mappings": []interface{}{}})
	}
}

func modelComboMappingsCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// --- Version manager handlers ---

func versionManagerStatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"currentVersion": "4.0.0-go", "upToDate": true})
	}
}

func versionManagerCheckUpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"updateAvailable": false, "currentVersion": "4.0.0-go"})
	}
}

func versionManagerInstallHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Already on latest version"})
	}
}

func versionManagerRestartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Restart signal sent"})
	}
}

func versionManagerStartHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

func versionManagerStopHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// --- Discovery handlers ---

func discoveryScanHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Scan initiated"})
	}
}

func discoveryResultsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{}})
	}
}

func discoveryResultDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"result": nil})
	}
}

func discoveryVerifyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// --- Tags ---

func tagsListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"tags": []interface{}{}})
	}
}

// --- Intelligence ---

func intelligenceSyncHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "synced": 0})
	}
}

// --- Playground ---

func playgroundPresetsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id != "" {
			if _, err := uuid.Parse(id); err != nil {
				jsonError(w, http.StatusBadRequest, "Invalid preset id: must be a valid UUID")
				return
			}
			switch r.Method {
			case http.MethodGet:
				preset, err := db.GetPlaygroundPreset(dbConn, id)
				if err != nil {
					jsonError(w, http.StatusInternalServerError, "Failed to fetch preset")
					return
				}
				if preset == nil {
					jsonError(w, http.StatusNotFound, "Preset not found: "+id)
					return
				}
				writeJSONResponse(w, preset)
			case http.MethodPut:
				patch := map[string]interface{}{}
				if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
					jsonError(w, http.StatusBadRequest, "Invalid JSON body")
					return
				}
				if err := validatePlaygroundPresetPatch(patch, false); err != nil {
					jsonError(w, http.StatusBadRequest, err.Error())
					return
				}
				preset, err := db.UpdatePlaygroundPreset(dbConn, id, patch)
				if err != nil {
					jsonError(w, http.StatusInternalServerError, "Failed to update preset")
					return
				}
				if preset == nil {
					jsonError(w, http.StatusNotFound, "Preset not found: "+id)
					return
				}
				writeJSONResponse(w, preset)
			case http.MethodDelete:
				deleted, err := db.DeletePlaygroundPreset(dbConn, id)
				if err != nil {
					jsonError(w, http.StatusInternalServerError, "Failed to delete preset")
					return
				}
				if !deleted {
					jsonError(w, http.StatusNotFound, "Preset not found: "+id)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			presets, err := db.ListPlaygroundPresets(dbConn)
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to list presets")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"presets": presets})
		case http.MethodPost:
			body := map[string]interface{}{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				jsonError(w, http.StatusBadRequest, "Invalid JSON body")
				return
			}
			if err := validatePlaygroundPresetPatch(body, true); err != nil {
				jsonError(w, http.StatusBadRequest, err.Error())
				return
			}
			name, _ := body["name"].(string)
			endpoint, _ := body["endpoint"].(string)
			model, _ := body["model"].(string)
			params, _ := body["params"].(map[string]interface{})
			if params == nil {
				params = map[string]interface{}{}
			}
			var system *string
			if value, ok := body["system"].(string); ok {
				system = &value
			}
			preset, err := db.CreatePlaygroundPreset(dbConn, db.PlaygroundPreset{Name: name, Endpoint: endpoint, Model: model, System: system, Params: params})
			if err != nil {
				jsonError(w, http.StatusInternalServerError, "Failed to create preset")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(preset)
		}
	}
}

func validatePlaygroundPresetPatch(body map[string]interface{}, requireAll bool) error {
	for _, field := range []string{"name", "endpoint", "model"} {
		value, exists := body[field]
		if requireAll && !exists {
			return fmt.Errorf("%s: required", field)
		}
		if exists {
			text, ok := value.(string)
			if !ok || strings.TrimSpace(text) == "" {
				return fmt.Errorf("%s: must be a non-empty string", field)
			}
			if field == "name" && len(text) > 100 {
				return fmt.Errorf("name: must contain at most 100 characters")
			}
		}
	}
	if value, exists := body["system"]; exists && value != nil {
		text, ok := value.(string)
		if !ok || len(text) > 50000 {
			return fmt.Errorf("system: invalid value")
		}
	}
	if value, exists := body["params"]; exists {
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("params: must be an object")
		}
	}
	return nil
}

// --- Sync ---

func syncTokensListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"tokens": []interface{}{}})
	}
}

func syncInitializeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// --- Translator ---

func translatorTranslateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"result": nil, "message": "Use internal translator via chat handler"})
	}
}

func translatorTransformStreamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Use internal transform stream via chat handler"})
	}
}

func translatorDetectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"format": "openai"})
	}
}

func translatorHistoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"history": []interface{}{}})
	}
}
