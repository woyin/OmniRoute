// routes_providers_extended.go registers extended provider management routes.
//
// Handles provider authentication flows (Claude, AGY, Codex imports),
// bulk operations, health monitoring, and Zed provider discovery.
package main

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
)

// registerProvidersExtendedRoutes registers extended provider sub-routes inside the /api group.
func registerProvidersExtendedRoutes(r chi.Router, dbConn *sql.DB) {
	// Provider auth import routes (Claude, AGY, Codex)
	r.Get("/providers/claude-auth/import", placeholderHandler("providers/claude-auth/import"))
	r.Post("/providers/claude-auth/import", placeholderHandler("providers/claude-auth/import"))
	r.Post("/providers/claude-auth/import-bulk", placeholderHandler("providers/claude-auth/import-bulk"))
	r.Post("/providers/claude-auth/zip-extract", placeholderHandler("providers/claude-auth/zip-extract"))

	r.Get("/providers/agy-auth/import", placeholderHandler("providers/agy-auth/import"))
	r.Post("/providers/agy-auth/import", placeholderHandler("providers/agy-auth/import"))
	r.Post("/providers/agy-auth/import-bulk", placeholderHandler("providers/agy-auth/import-bulk"))
	r.Post("/providers/agy-auth/zip-extract", placeholderHandler("providers/agy-auth/zip-extract"))
	r.Post("/providers/agy-auth/apply-local", placeholderHandler("providers/agy-auth/apply-local"))

	r.Post("/providers/codex-auth/import", placeholderHandler("providers/codex-auth/import"))
	r.Post("/providers/codex-auth/import-bulk", placeholderHandler("providers/codex-auth/import-bulk"))
	r.Post("/providers/codex-auth/zip-extract", placeholderHandler("providers/codex-auth/zip-extract"))

	// Provider bulk operations
	r.Post("/providers/bulk", providerBulkHandler(dbConn))
	r.Post("/providers/bulk-web-session", placeholderHandler("providers/bulk-web-session"))

	// Provider client
	r.Get("/providers/client", placeholderHandler("providers/client"))

	// Command-code OAuth
	r.Post("/providers/command-code/auth/start", placeholderHandler("providers/command-code/auth/start"))
	r.Post("/providers/command-code/auth/callback", placeholderHandler("providers/command-code/auth/callback"))
	r.Post("/providers/command-code/auth/apply", placeholderHandler("providers/command-code/auth/apply"))
	r.Get("/providers/command-code/auth/status", placeholderHandler("providers/command-code/auth/status"))

	// Provider health / management
	r.Get("/providers/expiration", providerExpirationHandler(dbConn))
	r.Get("/providers/health-autopilot", providerHealthAutopilotHandler(dbConn))
	r.Post("/providers/health-autopilot/actions", placeholderHandler("providers/health-autopilot/actions"))
	r.Get("/providers/health-matrix", providerHealthMatrixHandler(dbConn))
	r.Get("/providers/quota-windows", providerQuotaWindowsHandler())
	r.Post("/providers/test-batch", providerBatchTestHandler(dbConn))
	r.Post("/providers/validate", providerValidateHandler())

	// Per-provider sub-routes
	r.Post("/providers/{id}/login", placeholderHandler("providers/login"))
	r.Get("/providers/{id}/models", providerModelsHandler(dbConn))
	r.Post("/providers/{id}/refresh", providerRefreshHandler(dbConn))
	r.Post("/providers/{id}/sync-models", providerSyncModelsHandler(dbConn))
	r.Get("/providers/{id}/claude-auth/apply-local", placeholderHandler("providers/claude-auth/apply-local"))
	r.Post("/providers/{id}/claude-auth/export", placeholderHandler("providers/claude-auth/export"))
	r.Get("/providers/{id}/codex-auth/apply-local", placeholderHandler("providers/codex-auth/apply-local"))
	r.Post("/providers/{id}/codex-auth/export", placeholderHandler("providers/codex-auth/export"))

	// Zed provider discovery
	r.Get("/providers/zed/discover", placeholderHandler("providers/zed/discover"))
	r.Post("/providers/zed/import", placeholderHandler("providers/zed/import"))
	r.Post("/providers/zed/manual-import", placeholderHandler("providers/zed/manual-import"))
}
