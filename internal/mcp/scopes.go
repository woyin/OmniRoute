package mcp

// MCPScope represents a permission scope for MCP tool access.
type MCPScope string

const (
	ScopeCore         MCPScope = "core"
	ScopeCache        MCPScope = "cache"
	ScopeCompression  MCPScope = "compression"
	ScopeOneProxy     MCPScope = "oneproxy"
	ScopeMemory       MCPScope = "memory"
	ScopeSkills       MCPScope = "skills"
	ScopeAgentSkills  MCPScope = "agent-skills"
	ScopePool         MCPScope = "pool"
	ScopeGamification MCPScope = "gamification"
	ScopePlugins      MCPScope = "plugins"
	ScopeNotion       MCPScope = "notion"
	ScopeObsidian     MCPScope = "obsidian"
	ScopeAdvanced     MCPScope = "advanced"
	ScopeAdmin        MCPScope = "admin"
)

// AllScopes lists all defined MCP scopes.
var AllScopes = []MCPScope{
	ScopeCore, ScopeCache, ScopeCompression, ScopeOneProxy,
	ScopeMemory, ScopeSkills, ScopeAgentSkills, ScopePool,
	ScopeGamification, ScopePlugins, ScopeNotion, ScopeObsidian,
	ScopeAdvanced, ScopeAdmin,
}

// toolScopeMap maps tool name prefixes to required scopes.
var toolScopeMap = map[string]MCPScope{
	"get_health":            ScopeCore,
	"list_combos":           ScopeCore,
	"get_combo_metrics":     ScopeCore,
	"switch_combo":          ScopeCore,
	"check_quota":           ScopeCore,
	"route_request":         ScopeCore,
	"cost_report":           ScopeCore,
	"list_models_catalog":   ScopeCore,
	"web_search":            ScopeCore,
	"simulate_route":        ScopeCore,
	"set_budget_guard":      ScopeCore,
	"set_routing_strategy":  ScopeCore,
	"set_resilience_profile": ScopeCore,
	"test_combo":            ScopeCore,
	"get_provider_metrics":  ScopeCore,
	"best_combo_for_task":   ScopeCore,
	"explain_route":         ScopeCore,
	"get_session_snapshot":  ScopeCore,
	"db_health_check":       ScopeCore,
	"sync_pricing":          ScopeCore,
	"get_quota_snapshot":    ScopeCore,
	"set_quota_override":    ScopeCore,
	"omniroute_web_fetch":           ScopeCore,
	"omniroute_pick_fastest_model":  ScopeCore,
	"omniroute_tool_search":         ScopeCore,

	"cache_stats":  ScopeCache,
	"cache_flush":  ScopeCache,

	"compression_status":       ScopeCompression,
	"compression_configure":    ScopeCompression,
	"set_compression_engine":   ScopeCompression,
	"list_compression_combos":  ScopeCompression,
	"compression_combo_stats":  ScopeCompression,
	"omniroute_ccr_retrieve":   ScopeCompression,
	"omniroute_rtk_discover":   ScopeCompression,
	"omniroute_rtk_learn":      ScopeCompression,

	"oneproxy_fetch":  ScopeOneProxy,
	"oneproxy_rotate": ScopeOneProxy,
	"oneproxy_stats":  ScopeOneProxy,

	"memory_search": ScopeMemory,
	"memory_add":    ScopeMemory,
	"memory_clear":  ScopeMemory,

	"skills_list":       ScopeSkills,
	"skills_enable":     ScopeSkills,
	"skills_execute":    ScopeSkills,
	"skills_executions": ScopeSkills,

	"agent_skill_discover": ScopeAgentSkills,
	"agent_skill_invoke":   ScopeAgentSkills,
	"agent_skill_status":   ScopeAgentSkills,
	"omniroute_agent_skills_list":      ScopeAgentSkills,
	"omniroute_agent_skills_get":       ScopeAgentSkills,
	"omniroute_agent_skills_coverage":  ScopeAgentSkills,
	"omniroute_github_skills_search":   ScopeAgentSkills,
	"omniroute_github_skills_scan":     ScopeAgentSkills,
	"omniroute_github_skills_install":  ScopeAgentSkills,

	"pool_list":       ScopePool,
	"pool_status":     ScopePool,
	"pool_drain":      ScopePool,
	"pool_add":        ScopePool,
	"pool_remove":     ScopePool,
	"pool_rebalance":  ScopePool,
	"omniroute_pool_sessions":         ScopePool,
	"omniroute_pool_reset":            ScopePool,
	"omniroute_pool_warm":             ScopePool,
	"omniroute_pool_health":           ScopePool,
	"omniroute_browser_pool_status":   ScopePool,

	"gamification_levels":            ScopeGamification,
	"gamification_badges":            ScopeGamification,
	"gamification_leaderboard":       ScopeGamification,
	"gamification_federation_score":  ScopeGamification,
	"gamification_anomalies":         ScopeGamification,
	"gamification_notifications":     ScopeGamification,
	"gamification_servers":           ScopeGamification,
	"gamification_stream":            ScopeGamification,
	"gamification_rank":              ScopeGamification,
	"gamification_profile":           ScopeGamification,
	"gamification_transfer":          ScopeGamification,
	"gamification_invite":            ScopeGamification,

	"plugins_list":       ScopePlugins,
	"plugins_install":    ScopePlugins,
	"plugins_activate":   ScopePlugins,
	"plugins_deactivate": ScopePlugins,
	"plugins_marketplace": ScopePlugins,
	"plugins_scan":       ScopePlugins,
	"plugins_config":     ScopePlugins,
	"plugins_inspect":    ScopePlugins,
	"plugin_uninstall":   ScopePlugins,
	"plugin_executions":  ScopePlugins,

	"notion_search":          ScopeNotion,
	"notion_get_page":        ScopeNotion,
	"notion_create_page":     ScopeNotion,
	"notion_update_page":     ScopeNotion,
	"notion_list_databases":  ScopeNotion,
	"notion_query_database":  ScopeNotion,
	"notion_list_block_children": ScopeNotion,
	"notion_get_database":        ScopeNotion,
	"notion_append_blocks":       ScopeNotion,

	"obsidian_search":          ScopeObsidian,
	"obsidian_read_note":       ScopeObsidian,
	"obsidian_create_note":     ScopeObsidian,
	"obsidian_update_note":     ScopeObsidian,
	"obsidian_delete_note":     ScopeObsidian,
	"obsidian_list_notes":      ScopeObsidian,
	"obsidian_list_tags":       ScopeObsidian,
	"obsidian_list_links":      ScopeObsidian,
	"obsidian_get_backlinks":   ScopeObsidian,
	"obsidian_get_metadata":    ScopeObsidian,
	"obsidian_search_tags":     ScopeObsidian,
	"obsidian_append_note":     ScopeObsidian,
	"obsidian_move_note":       ScopeObsidian,
	"obsidian_get_graph":       ScopeObsidian,
	"obsidian_get_daily_note":  ScopeObsidian,
	"obsidian_list_templates":  ScopeObsidian,
	"obsidian_apply_template":  ScopeObsidian,
	"obsidian_webdav_status":   ScopeObsidian,
	"obsidian_webdav_sync":     ScopeObsidian,
	"obsidian_webdav_list":     ScopeObsidian,
	"obsidian_webdav_read":     ScopeObsidian,
	"obsidian_webdav_write":    ScopeObsidian,
	"obsidian_webdav_delete":   ScopeObsidian,
	"obsidian_webdav_mkdir":    ScopeObsidian,
	"obsidian_check_status":       ScopeObsidian,
	"obsidian_search_structured":  ScopeObsidian,
	"obsidian_get_document_map":   ScopeObsidian,
	"obsidian_get_active_file":    ScopeObsidian,
	"obsidian_execute_command":    ScopeObsidian,
}

// GetToolScope returns the required scope for a tool.
func GetToolScope(toolName string) MCPScope {
	if scope, ok := toolScopeMap[toolName]; ok {
		return scope
	}
	return ScopeAdvanced
}

// HasScope checks if a list of scopes includes the required scope.
// An empty scopes list means "all access" (admin key with no restrictions).
func HasScope(scopes []string, required MCPScope) bool {
	if len(scopes) == 0 {
		return true
	}
	for _, s := range scopes {
		if s == string(required) || s == string(ScopeAdmin) {
			return true
		}
	}
	return false
}
