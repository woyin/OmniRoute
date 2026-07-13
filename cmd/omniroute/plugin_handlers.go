package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

func loadPlugins(dbConn *sql.DB) ([]map[string]interface{}, error) {
	setting, err := db.GetSetting(dbConn, "plugins", "installed")
	if err != nil {
		return nil, err
	}
	raw, ok := setting["items"]
	if !ok {
		return []map[string]interface{}{}, nil
	}
	data, _ := json.Marshal(raw)
	var plugins []map[string]interface{}
	_ = json.Unmarshal(data, &plugins)
	if plugins == nil {
		plugins = []map[string]interface{}{}
	}
	return plugins, nil
}
func savePlugins(dbConn *sql.DB, plugins []map[string]interface{}) error {
	return db.SetSetting(dbConn, "plugins", "installed", map[string]interface{}{"items": plugins})
}
func pluginByName(plugins []map[string]interface{}, name string) (int, map[string]interface{}) {
	for i, p := range plugins {
		if p["name"] == name {
			return i, p
		}
	}
	return -1, nil
}

func pluginDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		plugins, err := loadPlugins(dbConn)
		if err != nil {
			jsonError(w, 500, "Failed to load plugins")
			return
		}
		i, p := pluginByName(plugins, name)
		if i < 0 {
			jsonError(w, 404, "Plugin '"+name+"' not found")
			return
		}
		if r.Method == http.MethodDelete {
			plugins = append(plugins[:i], plugins[i+1:]...)
			if savePlugins(dbConn, plugins) != nil {
				jsonError(w, 500, "Failed to uninstall plugin")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true, "message": "Plugin '" + name + "' uninstalled"})
			return
		}
		writeJSONResponse(w, map[string]interface{}{"plugin": p})
	}
}
func pluginActivationHandler(dbConn *sql.DB, enabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		plugins, err := loadPlugins(dbConn)
		if err != nil {
			jsonError(w, 500, "Failed to load plugins")
			return
		}
		i, _ := pluginByName(plugins, name)
		if i < 0 {
			jsonError(w, 404, "Plugin '"+name+"' not found")
			return
		}
		plugins[i]["enabled"] = enabled
		if enabled {
			plugins[i]["status"] = "active"
		} else {
			plugins[i]["status"] = "inactive"
		}
		if savePlugins(dbConn, plugins) != nil {
			jsonError(w, 500, "Failed to update plugin")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"success": true})
	}
}
func pluginConfigHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		plugins, err := loadPlugins(dbConn)
		if err != nil {
			jsonError(w, 500, "Failed to load plugins")
			return
		}
		i, p := pluginByName(plugins, name)
		if i < 0 {
			jsonError(w, 404, "Plugin '"+name+"' not found")
			return
		}
		if r.Method == http.MethodGet {
			writeJSONResponse(w, map[string]interface{}{"config": p["config"], "configSchema": p["configSchema"]})
			return
		}
		var body struct {
			Config map[string]interface{} `json:"config"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.Config == nil {
			jsonError(w, 400, "Invalid request")
			return
		}
		plugins[i]["config"] = body.Config
		if savePlugins(dbConn, plugins) != nil {
			jsonError(w, 500, "Failed to update plugin config")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"success": true, "config": body.Config})
	}
}
func pluginMarketplaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSONResponse(w, map[string]interface{}{"plugins": []interface{}{}})
	}
}
func pluginScanHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plugins, err := loadPlugins(dbConn)
		if err != nil {
			jsonError(w, 500, "Failed to scan plugin directory")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"discovered": plugins, "errors": []interface{}{}})
	}
}
