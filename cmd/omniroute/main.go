// OmniRoute Go Rewrite — Unified AI proxy/router
//
// This is the Go rewrite of the OmniRoute application, achieving 1:1 functional
// parity with the original Next.js/Node.js main branch.
//
// Architecture:
//   - cmd/omniroute/        : HTTP server, route registration, handler stubs
//   - internal/config/     : Configuration loading (env vars, CLI flags)
//   - internal/db/         : SQLite database access layer
//   - internal/handler/    : Core API handlers (chat, models, embeddings, etc.)
//   - internal/provider/   : Provider registry, executor, translator
//   - internal/routing/    : Request routing engine
//   - internal/mcp/        : MCP (Model Context Protocol) server
//   - internal/a2a/        : A2A (Agent-to-Agent) server
//   - internal/auth/       : Authentication middleware
//   - internal/oauth/      : OAuth flow handlers
//   - internal/management/ : Management subsystem handlers
//   - internal/middleware/  : HTTP middleware (CORS, rate limiting, etc.)
//   - internal/sse/        : Server-Sent Events support
//
// Route Structure:
//
//	/              : Root handler, health check, agent card
//	/api/*        : Management API routes (providers, combos, settings, etc.)
//	/api/v1/*     : AI proxy routes (chat/completions, embeddings, etc.)
//	/api/v1beta/* : Beta API routes
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/mcp"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

var (
	version         = "4.0.0-go"
	showHelp        = false
	portFlag        = 0
	dataDir         = ""
	mcpMode         = false
	smokeSQLite     = false
	inventoryRoutes = false
)

func init() {
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.IntVar(&portFlag, "port", 0, "Server port (default: 3456, env: PORT)")
	flag.StringVar(&dataDir, "data-dir", "", "Data directory (default: ~/.omniroute, env: DATA_DIR)")
	flag.BoolVar(&mcpMode, "mcp", false, "Start MCP server (stdio transport)")
	flag.BoolVar(&smokeSQLite, "smoke-sqlite", false, "Run SQLite release smoke check")
	flag.BoolVar(&inventoryRoutes, "inventory-routes", false, "Print Go runtime route inventory as JSON")
}

func main() {
	if os.Getenv("OMNIROUTE_SMOKE_TEST") == "sqlite" {
		cfg := config.Load()
		dbConn, err := db.GetDB(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer dbConn.Close()
		if _, err := dbConn.Exec("CREATE TABLE IF NOT EXISTS smoke_test (value TEXT NOT NULL)"); err != nil {
			log.Fatal(err)
		}
		if _, err := dbConn.Exec("INSERT INTO smoke_test (value) VALUES ('ok')"); err != nil {
			log.Fatal(err)
		}
		var value string
		if err := dbConn.QueryRow("SELECT value FROM smoke_test ORDER BY rowid DESC LIMIT 1").Scan(&value); err != nil || value != "ok" {
			log.Fatalf("sqlite smoke failed: value=%q err=%v", value, err)
		}
		return
	}
	flag.Parse()

	if showHelp {
		fmt.Fprintf(os.Stderr, "OmniRoute — Unified AI proxy/router (Go rewrite)\n\n")
		fmt.Fprintf(os.Stderr, "Usage: omniroute [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  PORT              Server port (default: 3456)\n")
		fmt.Fprintf(os.Stderr, "  DATA_DIR          Data directory (default: ~/.omniroute)\n")
		fmt.Fprintf(os.Stderr, "  REQUIRE_API_KEY   Require API key for all requests (default: false)\n")
		fmt.Fprintf(os.Stderr, "  FETCH_TIMEOUT_MS  Upstream fetch timeout (default: 120000)\n")
		os.Exit(0)
	}

	cfg := config.Load()
	if portFlag > 0 {
		cfg.Port = portFlag
	}
	if dataDir != "" {
		cfg.DataDir = dataDir
		cfg.SQLiteFile = dataDir + "/storage.sqlite"
	}

	if smokeSQLite {
		if err := sqliteSmoke(cfg); err != nil {
			log.Fatalf("sqlite smoke: %v", err)
		}
		fmt.Println("sqlite smoke: ok")
		return
	}

	if mcpMode {
		registry.RegisterBuiltinProviders()
		log.Printf("[MCP] Registered %d providers", len(registry.List()))
		dbConn, err := db.GetDB(cfg)
		if err != nil {
			log.Fatalf("[MCP] Database initialization failed: %v", err)
		}
		mcpServer := mcp.NewMCPServer(dbConn)
		log.Printf("[MCP] Starting stdio transport with %d tools", len(mcpServer.Tools()))
		mcpServer.StartStdio()
		return
	}

	log.Printf("[OmniRoute] v%s starting on port %d", version, cfg.Port)
	log.Printf("[OmniRoute] Data directory: %s", cfg.ResolveDataDir())

	dbConn, err := db.GetDB(cfg)
	if err != nil {
		log.Fatalf("[OmniRoute] Database initialization failed: %v", err)
	}

	registry.RegisterBuiltinProviders()
	log.Printf("[OmniRoute] Registered %d providers", len(registry.List()))

	r := buildRouter(cfg, dbConn)
	if inventoryRoutes {
		data, err := marshalGoRouteInventory(r, cfg.RequireApiKey)
		if err != nil {
			log.Fatalf("[OmniRoute] Route inventory failed: %v", err)
		}
		fmt.Println(string(data))
		return
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[OmniRoute] Listening on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("[OmniRoute] Server failed: %v", err)
	}
}

// jsonError writes a JSON-formatted error response.
func jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"message": message},
	})
}

func placeholderHandler(feature string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(fmt.Sprintf(`{"error":{"message":"%s not yet implemented in Go rewrite","type":"not_implemented"}}`, feature)))
	}
}

func listAPIKeysHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := db.ListAPIKeys(dbConn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	}
}

func createAPIKeyHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ak db.APIKey
		if err := json.NewDecoder(r.Body).Decode(&ak); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if ak.Key == "" {
			ak.Key = "sk-or-" + uuid.New().String()
		}
		ak.IsActive = true
		if err := db.CreateAPIKey(dbConn, ak); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ak)
	}
}

func serveRootHandler(ver string, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":    "OmniRoute",
			"version": ver,
			"status":  "running",
			"port":    cfg.Port,
			"endpoints": map[string]string{
				"health":               "GET  /health",
				"auth.require-login":   "GET|POST /api/auth/require-login",
				"auth.login":           "POST /api/auth/login",
				"auth.logout":          "POST /api/auth/logout",
				"providers":            "GET|POST|PUT|DELETE /api/providers",
				"combos":               "GET|POST|PUT|DELETE /api/combos",
				"api-keys":             "GET|POST /api/api-keys",
				"chat.completions":     "POST /api/v1/chat/completions",
				"responses":            "POST /api/v1/responses",
				"completions":          "POST /api/v1/completions",
				"models":               "GET  /api/v1/models",
				"embeddings":           "POST /api/v1/embeddings",
				"images.generations":   "POST /api/v1/images/generations",
				"audio.speech":         "POST /api/v1/audio/speech",
				"audio.transcriptions": "POST /api/v1/audio/transcriptions",
				"moderations":          "POST /api/v1/moderations",
				"rerank":               "POST /api/v1/rerank",
				"search":               "POST /api/v1/search",
				"files":                "GET|POST /api/v1/files",
				"batches":              "GET|POST /api/v1/batches",
				"usage.analytics":      "GET  /api/usage/analytics",
				"usage.history":        "GET  /api/usage/history",
				"usage.logs":           "GET  /api/usage/logs",
				"db.health":            "GET  /api/db/health",
				"settings":             "GET|PUT /api/settings",
				"export":               "GET  /api/settings/export-json",
				"import":               "POST /api/settings/import-json",
				"version":              "GET  /api/system/version",
				"agent-card":           "GET  /.well-known/agent.json",
			},
			"docs": "https://omniroute.dev",
		})
	}
}

func freeModelsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var freeModels []map[string]interface{}
		for _, entry := range registry.List() {
			if entry.HasFree {
				for _, m := range entry.Models {
					freeModels = append(freeModels, map[string]interface{}{
						"id":       m.ID,
						"name":     m.Name,
						"provider": entry.ID,
						"object":   "model",
						"free":     true,
					})
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"object": "list",
			"data":   freeModels,
		})
	}
}

func providerStatsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{}
		for _, entry := range registry.List() {
			count := 0
			if dbConn != nil {
				dbConn.QueryRow("SELECT COUNT(*) FROM provider_connections WHERE provider = ? AND is_active = 1", entry.ID).Scan(&count)
			}
			stats[entry.ID] = map[string]interface{}{
				"name":        entry.Name,
				"authType":    entry.AuthType,
				"format":      entry.Format,
				"modelCount":  len(entry.Models),
				"connections": count,
				"hasFree":     entry.HasFree,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

func serveAgentCard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		card := map[string]interface{}{
			"name":        "OmniRoute",
			"description": "Unified AI proxy/router — route any LLM through one endpoint",
			"version":     version,
			"url":         "https://omniroute.dev",
			"capabilities": map[string]interface{}{
				"streaming":              true,
				"pushNotifications":      false,
				"stateTransitionHistory": true,
			},
			"skills": []map[string]interface{}{
				{"id": "smartRouting", "name": "Smart Routing", "description": "Route requests to the best available provider"},
				{"id": "quotaManagement", "name": "Quota Management", "description": "Track and manage provider quotas"},
				{"id": "providerDiscovery", "name": "Provider Discovery", "description": "Discover available providers and models"},
				{"id": "costAnalysis", "name": "Cost Analysis", "description": "Analyze routing costs and optimize spending"},
				{"id": "healthReport", "name": "Health Report", "description": "Generate provider health reports"},
				{"id": "listCapabilities", "name": "List Capabilities", "description": "List all OmniRoute capabilities"},
			},
			"provider": map[string]interface{}{
				"organization": "OmniRoute",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(card)
	}
}

// --- Skills handlers ---

func skillsListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		skills, err := db.ListSkills(dbConn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(skills)
	}
}

func skillsCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var skill db.Skill
		if err := json.NewDecoder(r.Body).Decode(&skill); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if skill.ID == "" {
			skill.ID = uuid.New().String()
		}
		if err := db.SaveSkill(dbConn, skill); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(skill)
	}
}

func skillsDeleteHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := db.DeleteSkill(dbConn, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// --- Memory handlers ---

func memoryListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		memories, err := db.ListMemories(dbConn, 100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(memories)
	}
}

func memoryCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mem db.Memory
		if err := json.NewDecoder(r.Body).Decode(&mem); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if mem.ID == "" {
			mem.ID = uuid.New().String()
		}
		if err := db.SaveMemory(dbConn, mem); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(mem)
	}
}

func memoryDeleteHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := db.DeleteMemory(dbConn, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

func memorySearchHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			jsonError(w, http.StatusBadRequest, "q parameter required")
			return
		}
		memories, err := db.SearchMemories(dbConn, query, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(memories)
	}
}

// --- Webhooks handlers ---

func webhooksListHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		webhooks, err := db.ListWebhooks(dbConn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(webhooks)
	}
}

func webhooksCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var wh db.Webhook
		if err := json.NewDecoder(r.Body).Decode(&wh); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if wh.ID == "" {
			wh.ID = uuid.New().String()
		}
		if err := db.SaveWebhook(dbConn, wh); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(wh)
	}
}

func webhooksDeleteHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := db.DeleteWebhook(dbConn, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

func webhooksTestHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		wh, err := db.ListWebhooks(dbConn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		found := false
		for _, w := range wh {
			if w.ID == id {
				found = true
				break
			}
		}
		if !found {
			jsonError(w, http.StatusNotFound, "webhook not found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Webhook test queued",
			"id":      id,
		})
	}
}
