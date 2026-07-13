// routes_usage.go registers usage analytics sub-routes.
//
// Provides endpoints for:
//   - Usage budget tracking
//   - Call log access
//   - Combo health monitoring
//   - Provider cost analytics
//   - Quota and utilization reporting
package main

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
)

// registerUsageRoutes registers usage/analytics sub-routes inside the /api group.
func registerUsageRoutes(r chi.Router, dbConn *sql.DB) {
	// Usage budget
	r.Get("/usage/budget", placeholderHandler("usage/budget"))
	r.Post("/usage/budget/bulk", placeholderHandler("usage/budget/bulk"))

	// Call logs
	r.Get("/usage/call-logs", usageCallLogsHandler(dbConn))
	r.Get("/usage/call-logs/{id}", usageCallLogDetailHandler(dbConn))

	// Codex reset credit
	r.Post("/usage/codex-reset-credit", placeholderHandler("usage/codex-reset-credit"))

	// Combo health / forecast
	r.Get("/usage/combo-forecast", placeholderHandler("usage/combo-forecast"))
	r.Get("/usage/combo-health", placeholderHandler("usage/combo-health"))
	r.Get("/usage/combo-health-autopilot", placeholderHandler("usage/combo-health-autopilot"))
	r.Get("/usage/combo-health-dashboard", placeholderHandler("usage/combo-health-dashboard"))
	r.Get("/usage/combo-scoring-inspector", placeholderHandler("usage/combo-scoring-inspector"))

	// OM usage
	r.Get("/usage/om-usage", placeholderHandler("usage/om-usage"))

	// Provider window costs
	r.Get("/usage/provider-window-costs", placeholderHandler("usage/provider-window-costs"))

	// Proxy logs
	r.Get("/usage/proxy-logs", placeholderHandler("usage/proxy-logs"))

	// Quota
	r.Get("/usage/quota", usageQuotaHandler(dbConn))

	// Request logs
	r.Get("/usage/request-logs", placeholderHandler("usage/request-logs"))

	// Route explain
	r.Get("/usage/route-explain/{id}", placeholderHandler("usage/route-explain"))

	// Token limits
	r.Get("/usage/token-limits", placeholderHandler("usage/token-limits"))

	// Utilization
	r.Get("/usage/utilization", placeholderHandler("usage/utilization"))

	// Per-connection usage
	r.Get("/usage/{connectionId}", usageConnectionHandler(dbConn))
}
