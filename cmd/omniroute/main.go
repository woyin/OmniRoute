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
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"


	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/db"
	"github.com/omniroute/omniroute/internal/handler"
	"github.com/omniroute/omniroute/internal/middleware"
	"github.com/omniroute/omniroute/internal/a2a"
	"github.com/omniroute/omniroute/internal/oauth"
	"github.com/omniroute/omniroute/internal/mcp"
	"github.com/omniroute/omniroute/internal/auth"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

var (
	version   = "4.0.0-go"
	showHelp  = false
	portFlag  = 0
	dataDir   = ""
	mcpMode   = false
)

func init() {
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.IntVar(&portFlag, "port", 0, "Server port (default: 3456, env: PORT)")
	flag.StringVar(&dataDir, "data-dir", "", "Data directory (default: ~/.omniroute, env: DATA_DIR)")
	flag.BoolVar(&mcpMode, "mcp", false, "Start MCP server (stdio transport)")
}

func main() {
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

	if mcpMode {
		log.Println("[MCP] MCP server mode not yet implemented in Go rewrite")
		os.Exit(1)
	}

	cfg := config.Load()
	if portFlag > 0 {
		cfg.Port = portFlag
	}
	if dataDir != "" {
		cfg.DataDir = dataDir
		cfg.SQLiteFile = dataDir + "/storage.sqlite"
	}

	log.Printf("[OmniRoute] v%s starting on port %d", version, cfg.Port)
	log.Printf("[OmniRoute] Data directory: %s", cfg.ResolveDataDir())

	dbConn, err := db.GetDB(cfg)
	if err != nil {
		log.Fatalf("[OmniRoute] Database initialization failed: %v", err)
	}

	registry.RegisterBuiltinProviders()
	log.Printf("[OmniRoute] Registered %d providers", len(registry.List()))

	// Initialize MCP server
	mcpServer := mcp.NewMCPServer(dbConn)

	// Initialize A2A server
	a2aServer := a2a.NewA2AServer(dbConn)

	// Initialize OAuth handler
	oauthHandler := oauth.NewOAuthHandler(dbConn)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)
	r.Use(middleware.StripTrailingSlash)
	r.Use(chimw.Compress(5))


	// Root — API info landing page
	r.Get("/", webFileServer())

	r.Get("/health", (&handler.HealthHandler{DB: dbConn}).ServeHTTP)

	r.Get("/.well-known/agent.json", serveAgentCard())

	// Management routes (no auth required for initial setup)
	r.Route("/api", func(r chi.Router) {
		// Auth management
		r.Get("/auth/require-login", (&auth.RequireLoginHandler{DB: dbConn}).ServeHTTP)
		r.Post("/auth/require-login", (&auth.RequireLoginHandler{DB: dbConn}).ServeHTTP)
		r.Post("/auth/login", (&auth.LoginHandler{DB: dbConn}).ServeHTTP)
		r.Post("/auth/logout", (&auth.LogoutHandler{}).ServeHTTP)
		r.Get("/auth/status", (&handler.AuthStatusHandler{DB: dbConn}).ServeHTTP)

		// Login middleware for protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.LoginMiddleware(dbConn))

			// Providers CRUD
			r.Get("/providers", (&handler.ProvidersHandler{DB: dbConn}).ServeHTTP)
			r.Post("/providers", (&handler.ProvidersHandler{DB: dbConn}).ServeHTTP)
			r.Get("/providers/{id}", (&handler.ProviderDetailHandler{DB: dbConn}).ServeHTTP)
			r.Put("/providers/{id}", (&handler.ProviderDetailHandler{DB: dbConn}).ServeHTTP)
			r.Delete("/providers/{id}", (&handler.ProviderDetailHandler{DB: dbConn}).ServeHTTP)
			r.Post("/providers/test", (&handler.ProviderTestHandler{DB: dbConn, Config: cfg}).ServeHTTP)

			// Combos CRUD
			r.Get("/combos", (&handler.CombosHandler{DB: dbConn}).ServeHTTP)
			r.Post("/combos", (&handler.CombosHandler{DB: dbConn}).ServeHTTP)
			r.Get("/combos/{id}", (&handler.ComboDetailHandler{DB: dbConn}).ServeHTTP)
			r.Put("/combos/{id}", (&handler.ComboDetailHandler{DB: dbConn}).ServeHTTP)
			r.Delete("/combos/{id}", (&handler.ComboDetailHandler{DB: dbConn}).ServeHTTP)
			r.Post("/combos/test", (&handler.ComboTestHandler{DB: dbConn, Config: cfg}).ServeHTTP)
			r.Get("/combos/metrics", (&handler.ComboMetricsHandler{DB: dbConn}).ServeHTTP)
			r.Post("/combos/auto", (&handler.ComboAutoHandler{DB: dbConn}).ServeHTTP)

			// API Keys CRUD
			r.Get("/api-keys", listAPIKeysHandler(dbConn))
			r.Post("/api-keys", createAPIKeyHandler(dbConn))
			r.Delete("/api-keys/{key}", (&handler.APIKeyDetailHandler{DB: dbConn}).ServeHTTP)

			// Settings
			r.Get("/settings", (&handler.SettingsHandler{DB: dbConn}).ServeHTTP)
			r.Put("/settings", (&handler.SettingsHandler{DB: dbConn}).ServeHTTP)

			// Usage & analytics
			r.Get("/usage/analytics", (&handler.UsageHandler{DB: dbConn}).ServeHTTP)
			r.Get("/usage/history", (&handler.UsageHistoryHandler{DB: dbConn}).ServeHTTP)
			r.Get("/usage/logs", (&handler.CallLogsHandler{DB: dbConn}).ServeHTTP)
			r.Get("/token-health", (&handler.TokenHealthHandler{DB: dbConn}).ServeHTTP)
			r.Get("/usage/provider-limits", (&handler.ProviderLimitsHandler{DB: dbConn}).ServeHTTP)

			// DB health
			r.Get("/db/health", (&handler.DBHealthHandler{DB: dbConn}).ServeHTTP)

			// Import/Export
			r.Get("/settings/export-json", (&handler.ExportHandler{DB: dbConn}).ServeHTTP)
			r.Post("/settings/import-json", (&handler.ImportHandler{DB: dbConn}).ServeHTTP)

			// Init
			r.Post("/init", (&handler.InitHandler{DB: dbConn}).ServeHTTP)
		})

		// Public management routes
		r.Get("/free-models", freeModelsHandler())
		r.Get("/provider-stats", providerStatsHandler(dbConn))
		r.Get("/free-tier/summary", (&handler.FreeTierSummaryHandler{}).ServeHTTP)
		r.Get("/models/catalog", (&handler.ModelsCatalogHandler{DB: dbConn}).ServeHTTP)
		r.Get("/providers/registry", registryProvidersHandler())

		// System routes (public)
		r.Get("/system/version", (&handler.VersionHandler{}).ServeHTTP)
		r.Post("/shutdown", (&handler.ShutdownHandler{}).ServeHTTP)

		// MCP Server routes
		r.Get("/mcp/status", mcpServer.HandleMCPStatus)
		r.Get("/mcp/tools", mcpServer.HandleMCPTools)
		r.Get("/mcp/sse", mcpServer.HandleSSE)
		r.Post("/mcp/stream", mcpServer.HandleStream)
		r.Get("/mcp/stream", mcpServer.HandleStream)
		r.Get("/mcp/audit", mcpServer.HandleMCPAudit)

		// A2A routes
		r.Post("/a2a", a2aServer.HandleJSONRPC)
		r.Get("/a2a/status", a2aServer.HandleStatus)
		r.Get("/a2a/tasks", a2aTasksListHandler(dbConn))
		r.Get("/a2a/tasks/{id}", a2aTaskGetHandler(dbConn))

		// OAuth routes
		r.Get("/oauth/{provider}/{action}", oauthHandler.ServeHTTP)
		r.Post("/oauth/{provider}/{action}", oauthHandler.ServeHTTP)
		r.Post("/oauth/{provider}/paste-credentials", oauthHandler.HandlePasteCredentials)
		r.Post("/oauth/cliproxy-import", oauthHandler.HandleCLIProxyImport)
		r.Post("/oauth/codex/import", oauthHandler.HandleCLIProxyImport)
		r.Post("/oauth/codex/import-token", oauthHandler.HandleCLIProxyImport)
		r.Post("/oauth/cursor/import", oauthHandler.HandleCLIProxyImport)
		r.Post("/oauth/cursor/auto-import", oauthHandler.HandleCLIProxyImport)
		r.Post("/oauth/kiro/auto-import", oauthHandler.HandleCLIProxyImport)

		// Cache management routes
		r.Get("/cache", cacheStatusHandler(dbConn))
		r.Delete("/cache", cacheFlushHandler(dbConn))
		r.Get("/cache/stats", cacheStatsHandler(dbConn))
		r.Get("/cache/entries", cacheEntriesHandler(dbConn))
		r.Get("/cache/reasoning", cacheReasoningHandler(dbConn))

		// Guardrails routes
		r.Get("/guardrails", guardrailsListHandler(dbConn))
		r.Post("/guardrails", guardrailsCreateHandler(dbConn))
		r.Post("/guardrails/test", guardrailsTestHandler())

		// Fallback chains
		r.Get("/fallback/chains", fallbackChainsListHandler(dbConn))
		r.Post("/fallback/chains", fallbackChainsCreateHandler(dbConn))

		// Compression routes
		r.Get("/compression/engines", compressionEnginesHandler())
		r.Post("/compression/preview", compressionPreviewHandler())
		r.Get("/compression/compare", compressionCompareHandler())
		r.Get("/compression/language-packs", compressionLanguagePacksHandler())
		r.Get("/compression/rules", compressionRulesHandler())

		// Context / compression combos
		r.Get("/context/combos", compressionCombosListHandler(dbConn))
		r.Post("/context/combos", compressionCombosCreateHandler(dbConn))
		r.Get("/context/analytics", contextAnalyticsHandler())
		r.Get("/context/rtk/config", contextRTKConfigHandler())
		r.Get("/context/rtk/filters", contextRTKFiltersHandler())

		// Provider nodes
		r.Get("/provider-nodes", providerNodesListHandler(dbConn))
		r.Post("/provider-nodes", providerNodesCreateHandler(dbConn))
		r.Delete("/provider-nodes/{id}", providerNodesDeleteHandler(dbConn))

		// Model combo mappings
		r.Get("/model-combo-mappings", modelComboMappingsListHandler(dbConn))
		r.Post("/model-combo-mappings", modelComboMappingsCreateHandler(dbConn))

		// Version manager
		r.Get("/version-manager/status", versionManagerStatusHandler())
		r.Post("/version-manager/check-update", versionManagerCheckUpdateHandler())
		r.Post("/version-manager/install", versionManagerInstallHandler())
		r.Post("/version-manager/restart", versionManagerRestartHandler())
		r.Post("/version-manager/start", versionManagerStartHandler())
		r.Post("/version-manager/stop", versionManagerStopHandler())

		// DB backups
		r.Get("/db-backups", dbBackupsListHandler())
		r.Post("/db-backups/export", dbBackupsExportHandler(dbConn))
		r.Get("/db-backups/exportAll", dbBackupsExportAllHandler(dbConn))
		r.Post("/db-backups/import", dbBackupsImportHandler(dbConn))

		// Tunnels
		r.Get("/tunnels/cloudflared", tunnelsCloudflaredHandler())
		r.Get("/tunnels/ngrok", tunnelsNgrokHandler())
		r.Post("/tunnels/tailscale/enable", tunnelsTailscaleEnableHandler())
		r.Post("/tunnels/tailscale/disable", tunnelsTailscaleDisableHandler())
		r.Get("/tunnels/tailscale/status", tunnelsTailscaleStatusHandler())

		// Discovery
		r.Post("/discovery/scan", discoveryScanHandler())
		r.Get("/discovery/results", discoveryResultsHandler())
		r.Get("/discovery/results/{id}", discoveryResultDetailHandler())
		r.Post("/discovery/verify/{id}", discoveryVerifyHandler())

		// Resilience
		r.Get("/resilience", resilienceStatusHandler(dbConn))
		r.Get("/resilience/circuit-breakers", circuitBreakersListHandler(dbConn))
		r.Post("/resilience/circuit-breakers/{id}/reset", circuitBreakerResetHandler(dbConn))

		// Sessions
		r.Get("/sessions", sessionsListHandler())
		r.Delete("/sessions/{id}", sessionsDeleteHandler())

		// Rate limits
		r.Get("/rate-limits", rateLimitsListHandler())
		r.Post("/rate-limits", rateLimitsCreateHandler())

		// Upstream proxy
		r.Get("/upstream-proxy", upstreamProxyListHandler(dbConn))
		r.Post("/upstream-proxy", upstreamProxyCreateHandler(dbConn))
		r.Get("/upstream-proxy/{providerId}", upstreamProxyGetHandler(dbConn))
		r.Delete("/upstream-proxy/{providerId}", upstreamProxyDeleteHandler(dbConn))

		// Plugins
		r.Get("/plugins", pluginsListHandler())
		r.Post("/plugins/install", pluginsInstallHandler())

		// Evals
		r.Get("/evals", evalsListHandler())
		r.Get("/evals/suites", evalsSuitesHandler())

		// Cloud agents
		r.Get("/cloud/auth", cloudAuthHandler())
		r.Post("/cloud/credentials/update", cloudCredentialsUpdateHandler())
		r.Get("/cloud/model/resolve", cloudModelResolveHandler())

		// Provider metrics/models
		r.Get("/provider-metrics", providerMetricsHandler(dbConn))
		r.Get("/provider-models", providerModelsHandler(dbConn))

		// Compliance
		r.Get("/compliance/audit-log", complianceAuditLogHandler())

		// Monitoring
		r.Get("/monitoring/health", monitoringHealthHandler())
		r.Get("/health/degradation", healthDegradationHandler())

		// Network info
		r.Get("/network/info", networkInfoHandler())

		// Storage health
		r.Get("/storage/health", storageHealthHandler())

		// Tags
		r.Get("/tags", tagsListHandler())

		// Telemetry
		r.Get("/telemetry/summary", telemetrySummaryHandler())

		// Intelligence sync
		r.Post("/intelligence/sync", intelligenceSyncHandler())

		// Playground
		r.Get("/playground/presets", playgroundPresetsHandler())

		// Headroom
		r.Post("/headroom/start", headroomStartHandler())
		r.Get("/headroom/status", headroomStatusHandler())
		r.Post("/headroom/stop", headroomStopHandler())

		// Gamification
		r.Get("/gamification/level", gamificationLevelHandler())
		r.Get("/gamification/badges", gamificationBadgesHandler())
		r.Get("/gamification/leaderboard", gamificationLeaderboardHandler())

		// CLI tools status
		r.Get("/cli-tools/status", cliToolsStatusHandler())
		r.Get("/cli-tools/all-statuses", cliToolsAllStatusesHandler())

		// Sync
		r.Get("/sync/tokens", syncTokensListHandler())
		r.Post("/sync/initialize", syncInitializeHandler())

		// Search (management)
		r.Get("/search/analytics", searchAnalyticsHandler())

		// Translator
		r.Post("/translator/translate", translatorTranslateHandler())
		r.Post("/translator/transform-stream", translatorTransformStreamHandler())
		r.Get("/translator/detect", translatorDetectHandler())
		r.Get("/translator/history", translatorHistoryHandler())

		// Skills routes
		r.Get("/skills", skillsListHandler(dbConn))
		r.Post("/skills", skillsCreateHandler(dbConn))
		r.Delete("/skills/{id}", skillsDeleteHandler(dbConn))

		// Memory routes
		r.Get("/memory", memoryListHandler(dbConn))
		r.Post("/memory", memoryCreateHandler(dbConn))
		r.Delete("/memory/{id}", memoryDeleteHandler(dbConn))
		r.Get("/memory/search", memorySearchHandler(dbConn))

		// Webhooks routes
		r.Get("/webhooks", webhooksListHandler(dbConn))
		r.Post("/webhooks", webhooksCreateHandler(dbConn))
		r.Delete("/webhooks/{id}", webhooksDeleteHandler(dbConn))
		r.Post("/webhooks/{id}/test", webhooksTestHandler(dbConn))

		// Keys management routes
		r.Get("/keys", keysListHandler(dbConn))
		r.Post("/keys", keysListHandler(dbConn))
		r.Get("/keys/{id}", keysDetailHandler(dbConn))
		r.Put("/keys/{id}", keysUpdateHandler(dbConn))
		r.Delete("/keys/{id}", keysDeleteHandler(dbConn))
		r.Get("/keys/{id}/reveal", keysRevealHandler(dbConn))
		r.Post("/keys/{id}/regenerate", keysRegenerateHandler(dbConn))
		r.Get("/keys/{id}/usage-limits", keysUsageLimitsHandler(dbConn))
		r.Get("/keys/{id}/devices", keysDevicesHandler(dbConn))

		// Key groups
		r.Get("/keys/groups", keyGroupsListHandler(dbConn))
		r.Post("/keys/groups", keyGroupsCreateHandler(dbConn))
		r.Get("/keys/groups/{id}", keyGroupsDetailHandler(dbConn))
		r.Get("/keys/groups/{id}/keys", keyGroupsKeysHandler(dbConn))
		r.Get("/keys/groups/{id}/permissions", keyGroupsPermissionsHandler(dbConn))

		// Pricing routes
		r.Get("/pricing", pricingListHandler(dbConn))
		r.Get("/pricing/defaults", pricingDefaultsHandler())
		r.Get("/pricing/models", pricingModelsHandler(dbConn))
		r.Post("/pricing/sync", pricingSyncHandler())

		// Relay tokens
		r.Get("/relay/tokens", relayTokensListHandler(dbConn))
		r.Get("/relay/tokens/{id}", relayTokenDetailHandler(dbConn))

		// Routing decisions
		r.Get("/routing/decisions/{requestId}", routingDecisionDetailHandler(dbConn))

		// ACP (Agent Communication Protocol)
		r.Get("/acp/agents", acpAgentsListHandler(dbConn))

		// Agent skills
		r.Get("/agent-skills", agentSkillsListHandler(dbConn))
		r.Post("/agent-skills", agentSkillsCreateHandler(dbConn))
		r.Get("/agent-skills/{id}", agentSkillsDetailHandler(dbConn))
		r.Delete("/agent-skills/{id}", agentSkillsDeleteHandler(dbConn))
		r.Get("/agent-skills/coverage", agentSkillsCoverageHandler(dbConn))
		r.Post("/agent-skills/generate", agentSkillsGenerateHandler())
		r.Get("/agent-skills/{id}/raw", agentSkillsRawHandler(dbConn))

		// Admin
		r.Get("/admin/concurrency", adminConcurrencyHandler(dbConn))

		// Assess
		r.Post("/assess", assessHandler(dbConn))

		// Copilot
		r.Post("/copilot/chat", copilotChatHandler(dbConn, cfg))

		// Docs
		r.Get("/docs", docsHandler())
		r.Get("/docs/codex-cli", docsCodexCLIHandler())

		// Local Redis
		r.Get("/local/redis/status", localRedisStatusHandler())
		r.Post("/local/redis/start", localRedisStartHandler())
		r.Post("/local/redis/stop", localRedisStopHandler())

		// Middleware hooks
		r.Get("/middleware/hooks", middlewareHooksListHandler(dbConn))
		r.Get("/middleware/hooks/{name}", middlewareHooksDetailHandler(dbConn))

		// OpenAPI
		r.Get("/openapi/spec", openapiSpecHandler())
		r.Post("/openapi/try", openapiTryHandler())

		// Policies
		r.Get("/policies", policiesHandler(dbConn))

		// Proxy fallback
		r.Post("/proxy-fallback/test", proxyFallbackTestHandler(dbConn))

		// Session pools
		r.Get("/session-pools", sessionPoolsListHandler(dbConn))
		r.Get("/session-pools/{provider}", sessionPoolDetailHandler(dbConn))

		// Synced available models
		r.Get("/synced-available-models", syncedAvailableModelsHandler(dbConn))

		// GitHub skills
		r.Get("/github-skills", githubSkillsHandler(dbConn))

		// Free provider rankings
		r.Get("/free-provider-rankings", freeProviderRankingsHandler(dbConn))

		// Services management (9router, bifrost, cliproxy, mux)
		r.Get("/services/9router/status", servicesStatusHandler("9router"))
		r.Post("/services/9router/install", servicesInstallHandler("9router"))
		r.Post("/services/9router/start", servicesStartHandler("9router"))
		r.Post("/services/9router/stop", servicesStopHandler("9router"))
		r.Post("/services/9router/restart", servicesRestartHandler("9router"))
		r.Post("/services/9router/update", servicesUpdateHandler("9router"))
		r.Post("/services/9router/auto-start", servicesAutoStartHandler("9router"))
		r.Get("/services/9router/models", nineRouterModelsHandler())
		r.Post("/services/9router/rotate-key", nineRouterRotateKeyHandler())
		r.Post("/services/9router/provider-expose", nineRouterProviderExposeHandler())
		r.Get("/services/bifrost/status", servicesStatusHandler("bifrost"))
		r.Post("/services/bifrost/install", servicesInstallHandler("bifrost"))
		r.Post("/services/bifrost/start", servicesStartHandler("bifrost"))
		r.Post("/services/bifrost/stop", servicesStopHandler("bifrost"))
		r.Post("/services/bifrost/restart", servicesRestartHandler("bifrost"))
		r.Post("/services/bifrost/update", servicesUpdateHandler("bifrost"))
		r.Post("/services/bifrost/auto-start", servicesAutoStartHandler("bifrost"))
		r.Get("/services/cliproxy/status", servicesStatusHandler("cliproxy"))
		r.Post("/services/cliproxy/install", servicesInstallHandler("cliproxy"))
		r.Post("/services/cliproxy/start", servicesStartHandler("cliproxy"))
		r.Post("/services/cliproxy/stop", servicesStopHandler("cliproxy"))
		r.Post("/services/cliproxy/restart", servicesRestartHandler("cliproxy"))
		r.Post("/services/cliproxy/update", servicesUpdateHandler("cliproxy"))
		r.Post("/services/cliproxy/auto-start", servicesAutoStartHandler("cliproxy"))
		r.Get("/services/mux/status", servicesStatusHandler("mux"))
		r.Post("/services/mux/install", servicesInstallHandler("mux"))
		r.Post("/services/mux/start", servicesStartHandler("mux"))
		r.Post("/services/mux/stop", servicesStopHandler("mux"))
		r.Post("/services/mux/restart", servicesRestartHandler("mux"))
		r.Post("/services/mux/update", servicesUpdateHandler("mux"))
		r.Post("/services/mux/auto-start", servicesAutoStartHandler("mux"))
		r.Get("/services/{name}/logs", servicesLogsHandler())

		// Internal
		r.Post("/internal/codex-responses-ws", internalCodexResponsesWSHandler())

		// Logs
		r.Get("/logs", logsListHandler(dbConn))
		r.Get("/logs/{id}", logsDetailHandler(dbConn))
		r.Get("/logs/console", logsConsoleHandler())
		r.Get("/logs/detail", logsListHandler(dbConn))
		r.Post("/logs/export", logsExportHandler(dbConn))

		// Rate limit (singular, matches main branch)
		r.Get("/rate-limit", rateLimitHandler())

		// MCP audit stats
		r.Get("/mcp/audit/stats", mcpAuditStatsHandler(nil))

		// Restart
		r.Post("/restart", restartHandler())

		// Health ping
		r.Get("/health/ping", healthPingHandler())

		// CLI routes
		r.Post("/cli/connect", cliConnectHandler())
		r.Get("/cli/whoami", cliWhoamiHandler(dbConn))
		r.Get("/cli/tokens", cliTokensListHandler(dbConn))
		r.Get("/cli/tokens/{id}", cliTokenDetailHandler(dbConn))

		// Analytics
		r.Get("/analytics/auto-routing", analyticsAutoRoutingHandler(dbConn))
		r.Get("/analytics/compression", analyticsCompressionHandler(dbConn))
		r.Get("/analytics/diversity", analyticsDiversityHandler(dbConn))

		// Codex connect
		r.Get("/codex/connect/{token}", codexConnectHandler(dbConn))

		// Quota routes
		r.Get("/quota/groups", quotaGroupsListHandler(dbConn))
		r.Get("/quota/groups/{id}", quotaGroupsDetailHandler(dbConn))
		r.Get("/quota/keys/{id}/models", quotaKeysModelsHandler(dbConn))
		r.Get("/quota/plans", quotaPlansListHandler(dbConn))
		r.Get("/quota/plans/{connectionId}", quotaPlansDetailHandler(dbConn))
		r.Get("/quota/pools", quotaPoolsListHandler(dbConn))
		r.Get("/quota/pools/{id}", quotaPoolsDetailHandler(dbConn))
		r.Get("/quota/pools/{id}/log", quotaPoolsLogHandler(dbConn))
		r.Get("/quota/pools/{id}/usage", quotaPoolsUsageHandler(dbConn))
		r.Post("/quota/preview", quotaPreviewHandler(dbConn))

		// Tools / Agent Bridge
		r.Get("/tools/agent-bridge/agents", agentBridgeAgentsListHandler(dbConn))
		r.Get("/tools/agent-bridge/agents/{id}", agentBridgeAgentDetailHandler(dbConn))
		r.Get("/tools/agent-bridge/agents/{id}/detect", agentBridgeAgentDetectHandler(dbConn))
		r.Get("/tools/agent-bridge/agents/{id}/dns", agentBridgeAgentDnsHandler(dbConn))
		r.Get("/tools/agent-bridge/agents/{id}/mappings", agentBridgeAgentMappingsHandler(dbConn))
		r.Post("/tools/agent-bridge/bypass", agentBridgeBypassHandler(dbConn))
		r.Get("/tools/agent-bridge/cert", agentBridgeCertHandler(dbConn))
		r.Get("/tools/agent-bridge/cert/download", agentBridgeCertDownloadHandler(dbConn))
		r.Post("/tools/agent-bridge/cert/regenerate", agentBridgeCertRegenerateHandler(dbConn))
		r.Get("/tools/agent-bridge/config", agentBridgeConfigHandler(dbConn))
		r.Post("/tools/agent-bridge/diagnose", agentBridgeDiagnoseHandler(dbConn))
		r.Post("/tools/agent-bridge/repair", agentBridgeRepairHandler(dbConn))
		r.Get("/tools/agent-bridge/server", agentBridgeServerHandler(dbConn))
		r.Get("/tools/agent-bridge/state", agentBridgeStateHandler(dbConn))
		r.Post("/tools/agent-bridge/tproxy", agentBridgeTproxyHandler(dbConn))
		r.Get("/tools/agent-bridge/upstream-ca", agentBridgeUpstreamCAHandler(dbConn))
		r.Post("/tools/agent-bridge/upstream-ca/test", agentBridgeUpstreamCATestHandler(dbConn))

		// Tools / Traffic Inspector
		r.Get("/tools/traffic-inspector/capture-modes", trafficInspectorCaptureModesHandler())
		r.Get("/tools/traffic-inspector/capture-modes/http-proxy", trafficInspectorHTTPProxyHandler())
		r.Get("/tools/traffic-inspector/capture-modes/system-proxy", trafficInspectorSystemProxyHandler())
		r.Get("/tools/traffic-inspector/capture-modes/tls-intercept", trafficInspectorTLSInterceptHandler())
		r.Get("/tools/traffic-inspector/export.har", trafficInspectorExportHARHandler())
		r.Get("/tools/traffic-inspector/hosts", trafficInspectorHostsHandler())
		r.Get("/tools/traffic-inspector/hosts/{host}", trafficInspectorHostDetailHandler())
		r.Post("/tools/traffic-inspector/internal/ingest", trafficInspectorInternalIngestHandler())
		r.Get("/tools/traffic-inspector/requests", trafficInspectorRequestsHandler())
		r.Get("/tools/traffic-inspector/requests/{id}", trafficInspectorRequestDetailHandler())
		r.Post("/tools/traffic-inspector/requests/{id}/annotation", trafficInspectorRequestAnnotationHandler())
		r.Post("/tools/traffic-inspector/requests/{id}/replay", trafficInspectorRequestReplayHandler())
		r.Get("/tools/traffic-inspector/sessions", trafficInspectorSessionsHandler())
		r.Get("/tools/traffic-inspector/sessions/{id}", trafficInspectorSessionDetailHandler())
		r.Get("/tools/traffic-inspector/sessions/{id}/export.har", trafficInspectorSessionExportHARHandler())
		r.Get("/tools/traffic-inspector/sessions/{id}/requests", trafficInspectorSessionRequestsHandler())
		r.Get("/tools/traffic-inspector/ws", trafficInspectorWSHandler())
	})

	// API v1 routes (optional API key auth — enforce when REQUIRE_API_KEY=true)
	r.Route("/api/v1", func(r chi.Router) {
		if cfg.RequireApiKey {
			r.Use(auth.RequireAPIKey(dbConn))
		} else {
			r.Use(auth.OptionalAPIKey(dbConn))
		}
		// Chat / Responses
		r.Post("/chat/completions", (&handler.ChatHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Post("/responses", (&handler.ChatHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Post("/responses/{path}", (&handler.ChatHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Post("/completions", (&handler.CompletionsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Models
		r.Get("/models", (&handler.ModelsHandler{DB: dbConn}).ServeHTTP)
		r.Get("/models/{model}", (&handler.ModelsHandler{DB: dbConn}).ServeHTTP)

		// Embeddings
		r.Post("/embeddings", (&handler.EmbeddingsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Images
		r.Post("/images/generations", (&handler.ImageGenerationsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Audio
		r.Post("/audio/speech", (&handler.AudioSpeechHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Post("/audio/transcriptions", (&handler.AudioTranscriptionsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Moderations
		r.Post("/moderations", (&handler.ModerationsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Rerank
		r.Post("/rerank", (&handler.RerankHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Search
		r.Post("/search", (&handler.SearchHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Files
		r.Post("/files", (&handler.FilesHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Get("/files", (&handler.FilesHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Get("/files/{id}", (&handler.FilesHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Batches
		r.Post("/batches", (&handler.BatchesHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Get("/batches", (&handler.BatchesHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Get("/batches/{id}", (&handler.BatchesHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Videos
		r.Post("/videos/generations", (&handler.VideoGenerationsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Music
		r.Post("/music/generations", (&handler.MusicGenerationsHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Get("/music/generations", (&handler.MusicGenerationsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Messages (Anthropic format)
		r.Post("/messages", (&handler.MessagesHandler{DB: dbConn, Config: cfg}).ServeHTTP)
		r.Post("/messages/count_tokens", (&handler.MessagesCountTokensHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// OCR
		r.Post("/ocr", (&handler.OCRHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Image edits
		r.Post("/images/edits", (&handler.ImageEditsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Audio translations
		r.Post("/audio/translations", (&handler.AudioTranslationsHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Provider-specific routing
		r.Post("/providers/{provider}/chat/completions", (&handler.ProviderChatHandler{DB: dbConn, Config: cfg}).ServeHTTP)

		// Quota check
		r.Post("/quotas/check", (&handler.QuotaCheckHandler{DB: dbConn}).ServeHTTP)

		// Registered keys
		r.Get("/registered-keys", (&handler.RegisteredKeysHandler{DB: dbConn}).ServeHTTP)
	})

	// V1Beta routes
	r.Route("/api/v1beta", func(r chi.Router) {
		r.Get("/models", v1betaModelsListHandler(dbConn))
		r.Get("/models/{path}", v1betaModelsDetailHandler(dbConn))
	})

	log.Printf("[OmniRoute] MCP server initialized (%d tools)", len(mcpServer.Tools()))
	log.Printf("[OmniRoute] A2A server initialized")
	log.Printf("[OmniRoute] Routes configured")

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("[OmniRoute] Listening on http://localhost%s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("[OmniRoute] Server failed: %v", err)
	}
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
				"health":              "GET  /health",
				"auth.require-login":  "GET|POST /api/auth/require-login",
				"auth.login":          "POST /api/auth/login",
				"auth.logout":         "POST /api/auth/logout",
				"providers":           "GET|POST|PUT|DELETE /api/providers",
				"combos":              "GET|POST|PUT|DELETE /api/combos",
				"api-keys":            "GET|POST /api/api-keys",
				"chat.completions":    "POST /api/v1/chat/completions",
				"responses":           "POST /api/v1/responses",
				"completions":         "POST /api/v1/completions",
				"models":              "GET  /api/v1/models",
				"embeddings":          "POST /api/v1/embeddings",
				"images.generations":  "POST /api/v1/images/generations",
				"audio.speech":        "POST /api/v1/audio/speech",
				"audio.transcriptions": "POST /api/v1/audio/transcriptions",
				"moderations":         "POST /api/v1/moderations",
				"rerank":              "POST /api/v1/rerank",
				"search":              "POST /api/v1/search",
				"files":               "GET|POST /api/v1/files",
				"batches":             "GET|POST /api/v1/batches",
				"usage.analytics":     "GET  /api/usage/analytics",
			"usage.history":      "GET  /api/usage/history",
			"usage.logs":          "GET  /api/usage/logs",
			"db.health":           "GET  /api/db/health",
			"settings":            "GET|PUT /api/settings",
			"export":              "GET  /api/settings/export-json",
			"import":              "POST /api/settings/import-json",
			"version":             "GET  /api/system/version",
			"agent-card":          "GET  /.well-known/agent.json",
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
				"name":           entry.Name,
				"authType":       entry.AuthType,
				"format":         entry.Format,
				"modelCount":     len(entry.Models),
				"connections":    count,
				"hasFree":        entry.HasFree,
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
				"streaming": true,
				"pushNotifications": false,
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
			http.Error(w, `{"error":{"message":"q parameter required"}}`, http.StatusBadRequest)
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
			http.Error(w, `{"error":{"message":"webhook not found"}}`, http.StatusNotFound)
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
