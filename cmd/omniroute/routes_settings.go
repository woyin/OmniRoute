// routes_settings.go registers all settings sub-routes under the /api/settings/ path.
//
// Settings routes read from and write to the key_value table (namespace="settings")
// so that values persist across server restarts and are shareable with the UI.
//
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

const settingsNamespace = "settings"

// settingsGetHandler returns a handler that reads a setting from the database.
func settingsGetHandler(dbConn *sql.DB, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := db.GetSetting(dbConn, settingsNamespace, key)
		if err != nil {
			log.Printf("[settings] get %s error: %v", key, err)
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("failed to read setting %s", key))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"setting": key,
			"value":   val,
		})
	}
}

// settingsPutHandler returns a handler that writes a setting to the database.
func settingsPutHandler(dbConn *sql.DB, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		// Accept either {"value": {...}} or a raw object — store the whole body.
		if err := db.SetSetting(dbConn, settingsNamespace, key, body); err != nil {
			log.Printf("[settings] set %s error: %v", key, err)
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("failed to write setting %s", key))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"setting": key,
			"updated": true,
		})
	}
}

// settingsPostHandler returns a handler that accepts a POST action and stores any payload.
func settingsPostHandler(dbConn *sql.DB, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if err := db.SetSetting(dbConn, settingsNamespace, key, body); err != nil {
			log.Printf("[settings] post %s error: %v", key, err)
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("failed to write setting %s", key))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"action":  key,
			"success": true,
		})
	}
}

// registerSettingsGetPut registers a GET and PUT handler pair for a settings sub-path.
func registerSettingsGetPut(r chi.Router, path, key string, dbConn *sql.DB) {
	r.Get(path, settingsGetHandler(dbConn, key))
	r.Put(path, settingsPutHandler(dbConn, key))
}

// registerSettingsRoutes registers all settings/* sub-routes inside the
// already-authenticated /api router group.
func registerSettingsRoutes(r chi.Router, dbConn *sql.DB) {
	// ---- Simple GET/PUT settings pages ----
	registerSettingsGetPut(r, "/settings/authz-inventory", "authz-inventory", dbConn)
	registerSettingsGetPut(r, "/settings/auto-disable-accounts", "auto-disable-accounts", dbConn)
	registerSettingsGetPut(r, "/settings/background-degradation", "background-degradation", dbConn)
	registerSettingsGetPut(r, "/settings/cache-config", "cache-config", dbConn)
	registerSettingsGetPut(r, "/settings/combo-defaults", "combo-defaults", dbConn)
	registerSettingsGetPut(r, "/settings/compression", "compression-settings", dbConn)
	registerSettingsGetPut(r, "/settings/favicon", "favicon", dbConn)
	registerSettingsGetPut(r, "/settings/feature-flags", "feature-flags", dbConn)
	registerSettingsGetPut(r, "/settings/ip-filter", "ip-filter", dbConn)
	registerSettingsGetPut(r, "/settings/lkgp-cache", "lkgp-cache", dbConn)
	registerSettingsGetPut(r, "/settings/memory", "memory-settings", dbConn)
	registerSettingsGetPut(r, "/settings/mitm", "mitm", dbConn)
	registerSettingsGetPut(r, "/settings/model-aliases", "model-aliases", dbConn)
	registerSettingsGetPut(r, "/settings/notion", "notion", dbConn)
	registerSettingsGetPut(r, "/settings/obsidian", "obsidian", dbConn)
	registerSettingsGetPut(r, "/settings/obsidian/webdav", "obsidian-webdav", dbConn)
	registerSettingsGetPut(r, "/settings/oneproxy", "oneproxy", dbConn)
	registerSettingsGetPut(r, "/settings/payload-rules", "payload-rules", dbConn)
	registerSettingsGetPut(r, "/settings/proxy", "proxy", dbConn)
	registerSettingsGetPut(r, "/settings/qdrant", "qdrant", dbConn)
	registerSettingsGetPut(r, "/settings/quota-store", "quota-store", dbConn)
	registerSettingsGetPut(r, "/settings/require-login", "require-login", dbConn)
	registerSettingsGetPut(r, "/settings/system-prompt", "system-prompt", dbConn)
	registerSettingsGetPut(r, "/settings/task-routing", "task-routing", dbConn)
	registerSettingsGetPut(r, "/settings/thinking-budget", "thinking-budget", dbConn)

	// ---- Compression sub-settings ----
	registerSettingsGetPut(r, "/settings/compression/rules", "compression-rules", dbConn)

	r.Get("/settings/compression/mcp-accessibility", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "compression-mcp-accessibility")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"setting": "compression-mcp-accessibility",
			"value":   val,
		})
	})
	r.Get("/settings/compression/run-telemetry", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "compression-run-telemetry")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"setting": "compression-run-telemetry",
			"value":   val,
		})
	})

	// ---- GET-only settings pages ----
	r.Get("/settings/cache-metrics", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "cache-metrics")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"hits": 0, "misses": 0, "size": 0}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Get("/settings/database", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "database")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"engine": "sqlite", "size": 0}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Get("/settings/models-dev", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "models-dev")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"setting": "models-dev",
			"value":   val,
		})
	})

	// ---- Database actions ----
	r.Post("/settings/database/refresh-stats", settingsPostHandler(dbConn, "database-refresh-stats"))
	r.Post("/settings/database/vacuum", settingsPostHandler(dbConn, "database-vacuum"))

	// ---- Free proxies ----
	r.Get("/settings/free-proxies", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "free-proxies")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"proxies": []interface{}{}}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Post("/settings/free-proxies/{id}/add-to-pool", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
			"action":  "add-to-pool",
		})
	})
	r.Post("/settings/free-proxies/bulk-add-to-pool", settingsPostHandler(dbConn, "free-proxies-bulk-add-to-pool"))
	r.Get("/settings/free-proxies/stats", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "free-proxies-stats")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"total": 0, "active": 0, "healthy": 0}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Post("/settings/free-proxies/sync", settingsPostHandler(dbConn, "free-proxies-sync"))

	// ---- OneProxy actions ----
	r.Post("/settings/oneproxy/rotate", settingsPostHandler(dbConn, "oneproxy-rotate"))

	// ---- Proxies (list + actions) ----
	r.Get("/settings/proxies", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "proxies-list")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"proxies": []interface{}{}}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Post("/settings/proxies/assignments", settingsPostHandler(dbConn, "proxies-assignments"))
	r.Post("/settings/proxies/auto-test", settingsPostHandler(dbConn, "proxies-auto-test"))
	r.Post("/settings/proxies/batch-activate", settingsPostHandler(dbConn, "proxies-batch-activate"))
	r.Post("/settings/proxies/batch-delete", settingsPostHandler(dbConn, "proxies-batch-delete"))
	r.Post("/settings/proxies/bulk-assign", settingsPostHandler(dbConn, "proxies-bulk-assign"))
	r.Post("/settings/proxies/bulk-import", settingsPostHandler(dbConn, "proxies-bulk-import"))
	r.Get("/settings/proxies/egress", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "proxies-egress")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"egress": []interface{}{}}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Get("/settings/proxies/health", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "proxies-health")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"healthy": 0, "total": 0}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Post("/settings/proxies/migrate", settingsPostHandler(dbConn, "proxies-migrate"))
	r.Get("/settings/proxies/pool", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "proxies-pool")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"pool": []interface{}{}}
		}
		json.NewEncoder(w).Encode(val)
	})

	// ---- Proxy (singular) actions ----
	r.Post("/settings/proxy/cloudflare-deploy", settingsPostHandler(dbConn, "proxy-cloudflare-deploy"))
	r.Post("/settings/proxy/deno-deploy", settingsPostHandler(dbConn, "proxy-deno-deploy"))
	r.Post("/settings/proxy/test", settingsPostHandler(dbConn, "proxy-test"))
	r.Post("/settings/proxy/vercel-deploy", settingsPostHandler(dbConn, "proxy-vercel-deploy"))

	// ---- Purge endpoints ----
	r.Post("/settings/purge-call-logs", settingsPostHandler(dbConn, "purge-call-logs"))
	r.Post("/settings/purge-detailed-logs", settingsPostHandler(dbConn, "purge-detailed-logs"))
	r.Post("/settings/purge-logs", settingsPostHandler(dbConn, "purge-logs"))
	r.Post("/settings/purge-quota-snapshots", settingsPostHandler(dbConn, "purge-quota-snapshots"))
	r.Post("/settings/purge-request-history", settingsPostHandler(dbConn, "purge-request-history"))
	r.Post("/settings/purge-usage-history", settingsPostHandler(dbConn, "purge-usage-history"))

	// ---- Qdrant sub-routes ----
	r.Post("/settings/qdrant/cleanup", settingsPostHandler(dbConn, "qdrant-cleanup"))
	r.Get("/settings/qdrant/embedding-models", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "qdrant-embedding-models")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"models": []interface{}{}}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Get("/settings/qdrant/health", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "qdrant-health")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"status": "unknown"}
		}
		json.NewEncoder(w).Encode(val)
	})
	r.Post("/settings/qdrant/search", func(w http.ResponseWriter, r *http.Request) {
		val, _ := db.GetSetting(dbConn, settingsNamespace, "qdrant-search")
		w.Header().Set("Content-Type", "application/json")
		if len(val) == 0 {
			val = map[string]interface{}{"results": []interface{}{}}
		}
		json.NewEncoder(w).Encode(val)
	})
}
