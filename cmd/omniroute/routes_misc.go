// routes_misc.go registers miscellaneous management API routes.
//
// ⚠️  IMPORTANT: This file registers routes that coexist with main.go's /api
// group. chi router uses "last match wins", so any path here that duplicates
// a real handler in main.go would overwrite it with a placeholder. The rules are:
//  1. Any path registered in main.go (or in auth.LoginMiddleware group) → DO NOT register here.
//  2. Any path with a working handler (keys*, agent*, services*, tools*, etc.) → KEEP.
//  3. Any remaining placeholder routes → KEEP as stubs.
package main

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/auth"
)

// registerMiscRoutes registers miscellaneous sub-routes inside the /api group.
func registerMiscRoutes(r chi.Router, dbConn *sql.DB) {
	// --- Auth (public) ---
	r.With(auth.LoginMiddleware(dbConn)).Get("/auth/csrf", auth.CSRFHandler())

	// --- API Keys (v2 style) ---
	r.Get("/keys", keysListHandler(dbConn))
	r.Post("/keys", keysListHandler(dbConn))
	r.Get("/keys/{id}", keysDetailHandler(dbConn))
	r.Put("/keys/{id}", keysUpdateHandler(dbConn))
	r.Delete("/keys/{id}", keysDeleteHandler(dbConn))
	r.Get("/keys/{id}/devices", keysDevicesHandler(dbConn))
	r.Post("/keys/{id}/regenerate", keysRegenerateHandler(dbConn))
	r.Get("/keys/{id}/reveal", keysRevealHandler(dbConn))
	r.Get("/keys/{id}/usage-limits", keysUsageLimitsHandler(dbConn))
	r.Get("/keys/groups", keyGroupsListHandler(dbConn))
	r.Post("/keys/groups", keyGroupsCreateHandler(dbConn))
	r.Get("/keys/groups/{id}", keyGroupsDetailHandler(dbConn))
	r.Get("/keys/groups/{id}/keys", keyGroupsKeysHandler(dbConn))
	r.Get("/keys/groups/{id}/permissions", keyGroupsPermissionsHandler(dbConn))

	// --- Agent Skills ---
	r.Get("/agent-skills", agentSkillsListHandler(dbConn))
	r.Post("/agent-skills", agentSkillsCreateHandler(dbConn))
	r.Get("/agent-skills/coverage", agentSkillsCoverageHandler(dbConn))
	r.Post("/agent-skills/generate", agentSkillsGenerateHandler())
	r.Get("/agent-skills/{id}", agentSkillsDetailHandler(dbConn))
	r.Delete("/agent-skills/{id}", agentSkillsDeleteHandler(dbConn))
	r.Get("/agent-skills/{id}/raw", agentSkillsRawHandler(dbConn))

	// --- ACP Agents ---
	r.Get("/acp/agents", acpAgentsListHandler(dbConn))

	// --- Admin ---
	r.Get("/admin/concurrency", adminConcurrencyHandler(dbConn))

	// --- Assess ---
	r.Get("/assess", assessHandler(dbConn))

	// --- Batches ---
	r.Get("/batches", placeholderHandler("batches"))
	r.Post("/batches", placeholderHandler("batches"))

	// --- Cache ---
	r.Get("/cache", placeholderHandler("cache"))
	r.Delete("/cache", placeholderHandler("cache"))
	r.Get("/cache/stats", placeholderHandler("cache/stats"))
	r.Get("/cache/entries", placeholderHandler("cache/entries"))
	r.Get("/cache/reasoning", placeholderHandler("cache/reasoning"))

	// --- CLI connect/tokens/whoami ---
	r.Post("/cli/connect", cliConnectHandler())
	r.Get("/cli/tokens", cliTokensListHandler(dbConn))
	r.Get("/cli/tokens/{id}", cliTokenDetailHandler(dbConn))
	r.Get("/cli/whoami", cliWhoamiHandler(dbConn))

	// --- Codex connect ---
	r.Get("/codex/connect/{token}", codexConnectHandler(dbConn))

	// --- Copilot ---
	r.Post("/copilot/chat", copilotChatHandler(dbConn, nil))

	// --- DB backups ---
	r.Get("/db-backups", placeholderHandler("db-backups"))
	r.Post("/db-backups/export", placeholderHandler("db-backups/export"))
	r.Post("/db-backups/exportAll", placeholderHandler("db-backups/exportAll"))
	r.Post("/db-backups/import", placeholderHandler("db-backups/import"))

	// --- Discovery ---
	r.Post("/discovery/scan", discoveryScanHandler())
	r.Get("/discovery/results", discoveryResultsHandler())
	r.Get("/discovery/results/{id}", discoveryResultDetailHandler())
	r.Post("/discovery/verify/{id}", discoveryVerifyHandler())

	// --- Docs ---
	r.Get("/docs", docsHandler())
	r.Get("/docs/codex-cli", docsCodexCLIHandler())

	// --- Evals ---
	r.Get("/evals", placeholderHandler("evals"))
	r.Post("/evals", placeholderHandler("evals"))
	r.Get("/evals/suites", placeholderHandler("evals/suites"))
	r.Post("/evals/suites", placeholderHandler("evals/suites"))
	r.Get("/evals/suites/{suiteId}", placeholderHandler("evals/suites/detail"))
	r.Get("/evals/{suiteId}", placeholderHandler("evals/detail"))

	// --- Fallback chains ---
	r.Get("/fallback/chains", placeholderHandler("fallback/chains"))
	r.Post("/fallback/chains", placeholderHandler("fallback/chains"))

	// --- Files ---
	r.Post("/files", placeholderHandler("files"))
	r.Get("/files", placeholderHandler("files"))
	r.Get("/files/{id}", placeholderHandler("files/detail"))
	r.Get("/files/{id}/content", placeholderHandler("files/content"))

	// --- Free provider rankings (public) ---
	r.Get("/free-provider-rankings", freeProviderRankingsHandler(dbConn))

	// --- GitHub skills ---
	r.Get("/github-skills", githubSkillsHandler(dbConn))

	// --- Guardrails ---
	r.Get("/guardrails", placeholderHandler("guardrails"))
	r.Post("/guardrails", placeholderHandler("guardrails"))
	r.Post("/guardrails/test", placeholderHandler("guardrails/test"))

	// --- Health (degradation) ---
	r.Get("/health/degradation", placeholderHandler("health/degradation"))
	r.Get("/health/ping", healthPingHandler())

	// --- Intelligence ---
	r.Post("/intelligence/sync", intelligenceSyncHandler())

	// --- Internal ---
	r.Get("/internal/codex-responses-ws", internalCodexResponsesWSHandler())

	// --- Local Redis ---
	r.Post("/local/redis/start", localRedisStartHandler())
	r.Get("/local/redis/status", localRedisStatusHandler())
	r.Post("/local/redis/stop", localRedisStopHandler())

	// --- Logs ---
	r.Get("/logs/{id}", placeholderHandler("logs/detail"))
	r.Get("/logs/console", placeholderHandler("logs/console"))
	r.Get("/logs/detail", placeholderHandler("logs/detail"))
	r.Post("/logs/export", placeholderHandler("logs/export"))

	// --- MCP audit stats ---
	r.Get("/mcp/audit/stats", mcpAuditStatsHandler(nil))

	// --- Memory ---
	r.Get("/memory", memoryListHandler(dbConn))
	r.Post("/memory", memoryCreateHandler(dbConn))
	r.Get("/memory/{id}", memoryDeleteHandler(dbConn))

	// --- Middleware hooks ---
	r.Get("/middleware/hooks", middlewareHooksListHandler(dbConn))
	r.Get("/middleware/hooks/{name}", middlewareHooksDetailHandler(dbConn))
	r.Post("/middleware/hooks/{name}", middlewareHooksDetailHandler(dbConn))

	// --- Model combo mappings ---
	r.Get("/model-combo-mappings", modelComboMappingsListHandler(dbConn))
	r.Post("/model-combo-mapping", modelComboMappingsCreateHandler(dbConn))
	r.Get("/model-combo-mappings/{id}", placeholderHandler("model-combo-mapping/detail"))

	// --- Models (public and management) ---
	r.Get("/models/alias", placeholderHandler("models/alias"))
	r.Get("/models/openrouter-catalog", placeholderHandler("models/openrouter-catalog"))
	r.Post("/models/test", placeholderHandler("models/test"))
	r.Post("/models/test-all", placeholderHandler("models/test-all"))
	r.Get("/models/{model}", placeholderHandler("models/detail"))

	// --- Monitoring ---
	r.Get("/monitoring/health", placeholderHandler("monitoring/health"))

	// --- Network ---
	r.Get("/network/info", placeholderHandler("network/info"))

	// --- OpenAPI ---
	r.Get("/openapi/spec", openapiSpecHandler())
	r.Post("/openapi/try", openapiTryHandler())

	// --- Playground ---
	r.Post("/playground/improve-prompt", placeholderHandler("playground/improve-prompt"))
	r.Get("/playground/presets", playgroundPresetsHandler())
	r.Get("/playground/presets/{id}", placeholderHandler("playground/presets/detail"))
	r.Post("/playground/simulate-route", placeholderHandler("playground/simulate-route"))

	// --- Plugins ---
	r.Get("/plugins", placeholderHandler("plugins"))
	r.Post("/plugins", placeholderHandler("plugins"))
	r.Get("/plugins/{name}", placeholderHandler("plugins/detail"))
	r.Post("/plugins/{name}/activate", placeholderHandler("plugins/activate"))
	r.Post("/plugins/{name}/config", placeholderHandler("plugins/config"))
	r.Post("/plugins/{name}/deactivate", placeholderHandler("plugins/deactivate"))
	r.Get("/plugins/marketplace", placeholderHandler("plugins/marketplace"))
	r.Post("/plugins/scan", placeholderHandler("plugins/scan"))

	// --- Policies ---
	r.Get("/policies", policiesHandler(dbConn))

	// --- Pricing ---
	r.Get("/pricing", placeholderHandler("pricing"))
	r.Get("/pricing/defaults", placeholderHandler("pricing/defaults"))
	r.Get("/pricing/models", placeholderHandler("pricing/models"))
	r.Post("/pricing/sync", placeholderHandler("pricing/sync"))

	// --- Provider nodes ---
	r.Get("/provider-nodes", providerNodesListHandler(dbConn))
	r.Post("/provider-nodes", providerNodesCreateHandler(dbConn))
	r.Delete("/provider-nodes/{id}", providerNodesDeleteHandler(dbConn))
	r.Post("/provider-nodes/validate", placeholderHandler("provider-nodes/validate"))

	// --- Proxy fallback test ---
	r.Post("/proxy-fallback/test", proxyFallbackTestHandler(dbConn))

	// --- Quota ---
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

	// --- Rate limit ---
	r.Get("/rate-limit", placeholderHandler("rate-limit"))
	r.Get("/rate-limits", placeholderHandler("rate-limits"))

	// --- Relay tokens ---
	r.Get("/relay/tokens", relayTokensListHandler(dbConn))
	r.Get("/relay/tokens/{id}", relayTokenDetailHandler(dbConn))

	// --- Resilience ---
	r.Get("/resilience", placeholderHandler("resilience"))
	r.Get("/resilience/model-cooldowns", placeholderHandler("resilience/model-cooldowns"))
	r.Post("/resilience/reset", placeholderHandler("resilience/reset"))

	// --- Restart ---
	r.Post("/restart", restartHandler())

	// --- Routing decisions ---
	r.Get("/routing/decisions/{requestId}", routingDecisionDetailHandler(dbConn))

	// --- Search ---
	r.Get("/search/providers", placeholderHandler("search/providers"))
	r.Get("/search/stats", placeholderHandler("search/stats"))

	// --- Services (9router, bifrost, cliproxy, mux) ---
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

	// --- Session pools ---
	r.Get("/session-pools", sessionPoolsListHandler(dbConn))
	r.Get("/session-pools/{provider}", sessionPoolDetailHandler(dbConn))

	// --- Sessions ---
	r.Get("/sessions", placeholderHandler("sessions"))
	r.Post("/sessions", placeholderHandler("sessions"))
	r.Delete("/sessions/{id}", placeholderHandler("sessions/delete"))

	// --- Skills ---
	r.Get("/skills", skillsListHandler(dbConn))
	r.Post("/skills", skillsCreateHandler(dbConn))
	r.Get("/skills/{id}", skillsDeleteHandler(dbConn))
	r.Post("/skills/{id}", placeholderHandler("skills/update"))
	r.Get("/skills/executions", placeholderHandler("skills/executions"))
	r.Post("/skills/install", placeholderHandler("skills/install"))
	r.Get("/skills/marketplace", placeholderHandler("skills/marketplace"))
	r.Post("/skills/marketplace/install", placeholderHandler("skills/marketplace/install"))
	r.Get("/skills/skillssh", placeholderHandler("skills/skillssh"))
	r.Post("/skills/skillssh/install", placeholderHandler("skills/skillssh/install"))

	// --- Storage health ---
	r.Get("/storage/health", placeholderHandler("storage/health"))

	// --- Sync ---
	r.Get("/sync/bundle", placeholderHandler("sync/bundle"))
	r.Post("/sync/cloud", placeholderHandler("sync/cloud"))
	r.Post("/sync/initialize", syncInitializeHandler())
	r.Get("/sync/tokens", syncTokensListHandler())
	r.Get("/sync/tokens/{id}", placeholderHandler("sync/tokens/detail"))

	// --- Synced available models ---
	r.Get("/synced-available-models", syncedAvailableModelsHandler(dbConn))

	// --- System ---
	r.Get("/system/env/repair", placeholderHandler("system/env/repair"))
	r.Get("/system/version", placeholderHandler("system/version"))

	// --- Tags ---
	r.Get("/tags", tagsListHandler())

	// --- Telemetry ---
	r.Get("/telemetry/summary", placeholderHandler("telemetry/summary"))

	// --- Token health ---
	r.Get("/token-health", placeholderHandler("token-health"))

	// --- Tools / Agent Bridge ---
	r.Get("/tools/agent-bridge/agents", agentBridgeAgentsListHandler(dbConn))
	r.Get("/tools/agent-bridge/agents/{id}", agentBridgeAgentDetailHandler(dbConn))
	r.Post("/tools/agent-bridge/agents/{id}/detect", agentBridgeAgentDetectHandler(dbConn))
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

	// --- Tools / Traffic Inspector ---
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

	// --- Translator ---
	r.Post("/translator/translate", translatorTranslateHandler())
	r.Post("/translator/transform-stream", translatorTransformStreamHandler())
	r.Get("/translator/detect", translatorDetectHandler())
	r.Get("/translator/history", translatorHistoryHandler())
	r.Post("/translator/send", placeholderHandler("translator/send"))

	// --- Tunnels ---
	r.Post("/tunnels/cloudflared", placeholderHandler("tunnels/cloudflared"))
	r.Post("/tunnels/ngrok", placeholderHandler("tunnels/ngrok"))
	r.Get("/tunnels/tailscale", placeholderHandler("tunnels/tailscale"))
	r.Get("/tunnels/tailscale/check", placeholderHandler("tunnels/tailscale/check"))
	r.Post("/tunnels/tailscale/disable", placeholderHandler("tunnels/tailscale/disable"))
	r.Post("/tunnels/tailscale/enable", placeholderHandler("tunnels/tailscale/enable"))
	r.Post("/tunnels/tailscale/install", placeholderHandler("tunnels/tailscale/install"))
	r.Post("/tunnels/tailscale/login", placeholderHandler("tunnels/tailscale/login"))
	r.Post("/tunnels/tailscale/start-daemon", placeholderHandler("tunnels/tailscale/start-daemon"))
	r.Get("/tunnels/tailscale/status", placeholderHandler("tunnels/tailscale/status"))

	// --- Version manager ---
	r.Get("/version-manager/status", versionManagerStatusHandler())
	r.Post("/version-manager/check-update", versionManagerCheckUpdateHandler())
	r.Post("/version-manager/install", versionManagerInstallHandler())
	r.Post("/version-manager/restart", versionManagerRestartHandler())
	r.Post("/version-manager/start", versionManagerStartHandler())
	r.Post("/version-manager/stop", versionManagerStopHandler())

	// --- Webhooks ---
	r.Get("/webhooks", webhooksListHandler(dbConn))
	r.Post("/webhooks", webhooksCreateHandler(dbConn))
	r.Get("/webhooks/{id}", webhooksDeleteHandler(dbConn))
	r.Delete("/webhooks/{id}", webhooksDeleteHandler(dbConn))
	r.Get("/webhooks/{id}/deliveries", placeholderHandler("webhooks/deliveries"))
	r.Post("/webhooks/{id}/test", webhooksTestHandler(dbConn))
	r.Post("/webhooks/validate-url", placeholderHandler("webhooks/validate-url"))

	// --- Shutdown ---
	r.Post("/shutdown", restartHandler())

	// Cloud models alias
	r.Get("/cloud/models/alias", placeholderHandler("cloud/models/alias"))

	// Combo builder/reorder
	r.Get("/combos/builder/options", placeholderHandler("combos/builder/options"))
	r.Post("/combos/reorder", placeholderHandler("combos/reorder"))

	// Compression compare verify / retrieve
	r.Post("/compression/compare/verify", placeholderHandler("compression/compare/verify"))
	r.Post("/compression/retrieve", placeholderHandler("compression/retrieve"))
}
