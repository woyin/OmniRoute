// routes_cli_tools.go registers CLI tool management routes.
//
// Provides GET/PUT endpoints for each CLI tools settings (Claude, Cline,
// Codex, etc.), plus backup, apply, and Antigravity MITM proxy routes.
package main

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/auth"
)

// registerCLIToolsRoutes registers CLI tool management sub-routes inside the /api group.
func registerCLIToolsRoutes(r chi.Router, dbConn *sql.DB) {
	// CLI tool settings routes
	protected := r.With(auth.LoginMiddleware(dbConn))
	protected.Get("/cli-tools/letta-settings", lettaSettingsHandler())
	protected.Post("/cli-tools/letta-settings", lettaSettingsHandler())
	protected.Delete("/cli-tools/letta-settings", lettaSettingsHandler())
	protected.Get("/cli-tools/omp-settings", ompSettingsHandler(dbConn))
	protected.Post("/cli-tools/omp-settings", ompSettingsHandler(dbConn))
	protected.Delete("/cli-tools/omp-settings", ompSettingsHandler(dbConn))
	r.Get("/cli-tools/claude-settings", placeholderHandler("cli-tools/claude-settings"))
	r.Put("/cli-tools/claude-settings", placeholderHandler("cli-tools/claude-settings"))

	r.Get("/cli-tools/cline-settings", placeholderHandler("cli-tools/cline-settings"))
	r.Put("/cli-tools/cline-settings", placeholderHandler("cli-tools/cline-settings"))

	r.Get("/cli-tools/codewhale-settings", placeholderHandler("cli-tools/codewhale-settings"))
	r.Put("/cli-tools/codewhale-settings", placeholderHandler("cli-tools/codewhale-settings"))

	r.Get("/cli-tools/codex-profiles", placeholderHandler("cli-tools/codex-profiles"))
	r.Put("/cli-tools/codex-profiles", placeholderHandler("cli-tools/codex-profiles"))

	r.Get("/cli-tools/codex-settings", placeholderHandler("cli-tools/codex-settings"))
	r.Put("/cli-tools/codex-settings", placeholderHandler("cli-tools/codex-settings"))

	r.Get("/cli-tools/config", placeholderHandler("cli-tools/config"))
	r.Put("/cli-tools/config", placeholderHandler("cli-tools/config"))

	r.Get("/cli-tools/crush-settings", placeholderHandler("cli-tools/crush-settings"))
	r.Put("/cli-tools/crush-settings", placeholderHandler("cli-tools/crush-settings"))

	r.Get("/cli-tools/deepseek-tui-settings", placeholderHandler("cli-tools/deepseek-tui-settings"))
	r.Put("/cli-tools/deepseek-tui-settings", placeholderHandler("cli-tools/deepseek-tui-settings"))

	r.Post("/cli-tools/detect", placeholderHandler("cli-tools/detect"))

	r.Get("/cli-tools/droid-settings", placeholderHandler("cli-tools/droid-settings"))
	r.Put("/cli-tools/droid-settings", placeholderHandler("cli-tools/droid-settings"))

	r.Get("/cli-tools/forge-settings", placeholderHandler("cli-tools/forge-settings"))
	r.Put("/cli-tools/forge-settings", placeholderHandler("cli-tools/forge-settings"))

	r.Get("/cli-tools/guide-settings/{toolId}", placeholderHandler("cli-tools/guide-settings"))
	r.Get("/cli-tools/hermes-agent-settings", placeholderHandler("cli-tools/hermes-agent-settings"))
	r.Put("/cli-tools/hermes-agent-settings", placeholderHandler("cli-tools/hermes-agent-settings"))

	r.Get("/cli-tools/jcode-settings", placeholderHandler("cli-tools/jcode-settings"))
	r.Put("/cli-tools/jcode-settings", placeholderHandler("cli-tools/jcode-settings"))

	r.Get("/cli-tools/keys", placeholderHandler("cli-tools/keys"))
	r.Get("/cli-tools/kilo-settings", placeholderHandler("cli-tools/kilo-settings"))
	r.Put("/cli-tools/kilo-settings", placeholderHandler("cli-tools/kilo-settings"))

	r.Get("/cli-tools/logs", placeholderHandler("cli-tools/logs"))
	r.Get("/cli-tools/openclaw-settings", placeholderHandler("cli-tools/openclaw-settings"))
	r.Put("/cli-tools/openclaw-settings", placeholderHandler("cli-tools/openclaw-settings"))
	r.Post("/cli-tools/openclaw/auto-order", placeholderHandler("cli-tools/openclaw/auto-order"))

	r.Get("/cli-tools/pi-settings", placeholderHandler("cli-tools/pi-settings"))
	r.Put("/cli-tools/pi-settings", placeholderHandler("cli-tools/pi-settings"))

	r.Get("/cli-tools/qwen-settings", placeholderHandler("cli-tools/qwen-settings"))
	r.Put("/cli-tools/qwen-settings", placeholderHandler("cli-tools/qwen-settings"))

	r.Get("/cli-tools/runtime/{toolId}", placeholderHandler("cli-tools/runtime"))

	r.Get("/cli-tools/smelt-settings", placeholderHandler("cli-tools/smelt-settings"))
	r.Put("/cli-tools/smelt-settings", placeholderHandler("cli-tools/smelt-settings"))

	// CLI tool backups/apply
	r.Get("/cli-tools/backups", placeholderHandler("cli-tools/backups"))
	r.Post("/cli-tools/apply", placeholderHandler("cli-tools/apply"))

	// Antigravity MITM
	r.Get("/cli-tools/antigravity-mitm", placeholderHandler("cli-tools/antigravity-mitm"))
	r.Post("/cli-tools/antigravity-mitm/alias", placeholderHandler("cli-tools/antigravity-mitm/alias"))
}
