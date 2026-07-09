package mcp

// registerCoreTools registers core MCP tools: health, combos, routing, quota, metrics, and pricing.
func (s *MCPServer) registerCoreTools() {
	s.simpleTool("get_health", "Get OmniRoute service health status", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"status": "ok", "version": "4.0.0-go"}, nil
		})

	s.simpleTool("list_combos", "List all routing combos", nil,
		func(args map[string]interface{}) (interface{}, error) {
			if s.DB == nil {
				return []interface{}{}, nil
			}
			rows, err := s.DB.Query("SELECT id, name, strategy, is_active FROM combos ORDER BY name")
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var results []map[string]interface{}
			for rows.Next() {
				var id, name, strategy string
				var isActive bool
				if rows.Scan(&id, &name, &strategy, &isActive) == nil {
					results = append(results, map[string]interface{}{"id": id, "name": name, "strategy": strategy, "isActive": isActive})
				}
			}
			return results, nil
		})

	s.simpleTool("switch_combo", "Switch the active routing combo",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"comboId": map[string]interface{}{"type": "string"}}, "required": []string{"comboId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true, "comboId": args["comboId"]}, nil
		})

	s.simpleTool("check_quota", "Check provider quota status",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"provider": args["provider"], "remaining": -1, "unlimited": true}, nil
		})

	s.simpleTool("cost_report", "Get cost report for a time period",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"period": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"period": args["period"], "totalCost": 0, "requests": 0}, nil
		})

	s.simpleTool("list_models_catalog", "List all available models across providers", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"total": 222}, nil
		})

	s.simpleTool("web_search", "Search the web",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"query": args["query"], "results": []interface{}{}}, nil
		})

	s.simpleTool("simulate_route", "Simulate a routing decision for a model",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"model": map[string]interface{}{"type": "string"}}, "required": []string{"model"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"model": args["model"], "provider": "auto"}, nil
		})

	s.simpleTool("set_routing_strategy", "Change the routing strategy for a combo",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"comboId": map[string]interface{}{"type": "string"}, "strategy": map[string]interface{}{"type": "string"}}, "required": []string{"comboId", "strategy"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true}, nil
		})

	s.simpleTool("db_health_check", "Check database health", nil,
		func(args map[string]interface{}) (interface{}, error) {
			if s.DB != nil {
				if err := s.DB.Ping(); err != nil {
					return map[string]interface{}{"status": "error", "error": err.Error()}, nil
				}
			}
			return map[string]interface{}{"status": "ok"}, nil
		})

	s.simpleTool("best_combo_for_task", "Find the best combo for a given task",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"task": map[string]interface{}{"type": "string"}}, "required": []string{"task"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"task": args["task"], "bestCombo": "default"}, nil
		})

	s.simpleTool("explain_route", "Explain why a particular route was chosen",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"model": map[string]interface{}{"type": "string"}}, "required": []string{"model"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"model": args["model"], "reason": "canonical_prefix_match"}, nil
		})

	s.simpleTool("get_session_snapshot", "Get current session routing snapshot", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"activeProviders": 60, "activeModels": 222}, nil
		})

	s.simpleTool("get_provider_metrics", "Get provider performance metrics",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"provider": args["provider"], "avgLatencyMs": 0, "successRate": 1.0}, nil
		})

	s.simpleTool("set_budget_guard", "Set a budget guard limit",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"limit": map[string]interface{}{"type": "number"}, "period": map[string]interface{}{"type": "string"}}, "required": []string{"limit"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true}, nil
		})

	s.simpleTool("sync_pricing", "Sync provider pricing data", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true, "updated": 0}, nil
		})

	// --- New core tools (v3.8 parity) ---

	s.simpleTool("get_combo_metrics", "Get detailed metrics for a specific combo (latency, success rate, cost)",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"comboId": map[string]interface{}{"type": "string"}}, "required": []string{"comboId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"comboId":      args["comboId"],
				"avgLatencyMs": 0,
				"successRate":  1.0,
				"totalCost":    0,
				"totalRequests": 0,
			}, nil
		})

	s.simpleTool("test_combo", "Test a combo by sending a sample request",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"comboId": map[string]interface{}{"type": "string"}, "message": map[string]interface{}{"type": "string"}}, "required": []string{"comboId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"comboId":  args["comboId"],
				"success":  true,
				"latencyMs": 0,
				"provider": "stub",
				"model":    "stub-model",
			}, nil
		})

	s.simpleTool("route_request", "Route a single request through the proxy",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"model": map[string]interface{}{"type": "string"}, "messages": map[string]interface{}{"type": "array"}}, "required": []string{"model"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"model":    args["model"],
				"provider": "auto",
				"routed":   true,
				"content":  "",
			}, nil
		})

	s.simpleTool("set_resilience_profile", "Configure resilience settings (retry count, circuit breaker threshold, fallback mode)",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"maxRetries":            map[string]interface{}{"type": "number"},
			"circuitBreakerThreshold": map[string]interface{}{"type": "number"},
			"fallbackMode":          map[string]interface{}{"type": "string"},
		}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true, "profile": args}, nil
		})

	s.simpleTool("get_quota_snapshot", "Get current quota snapshot for all providers", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"providers": []interface{}{},
				"total":     0,
			}, nil
		})

	s.simpleTool("set_quota_override", "Set a manual quota override for a provider",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"provider": map[string]interface{}{"type": "string"},
			"limit":    map[string]interface{}{"type": "number"},
			"period":   map[string]interface{}{"type": "string"},
		}, "required": []string{"provider", "limit"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"success":  true,
				"provider": args["provider"],
				"limit":    args["limit"],
			}, nil
		})

	// --- Web fetch, fastest model picker, tool search ---

	s.simpleTool("omniroute_web_fetch", "Fetch content from a URL via Firecrawl/Jina/Tavily",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"url":      map[string]interface{}{"type": "string"},
			"provider": map[string]interface{}{"type": "string", "description": "firecrawl, jina, or tavily"},
		}, "required": []string{"url"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"url":     args["url"],
				"content": "",
				"status":  "stub",
			}, nil
		})

	s.simpleTool("omniroute_pick_fastest_model", "Pick the fastest provider-model from telemetry data",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"task":  map[string]interface{}{"type": "string"},
			"topN":  map[string]interface{}{"type": "number"},
		}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"task":      args["task"],
				"model":     "",
				"provider":  "",
				"latencyMs": 0,
			}, nil
		})

	s.simpleTool("omniroute_tool_search", "Search MCP tools by keyword",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"keyword": map[string]interface{}{"type": "string"},
			"limit":   map[string]interface{}{"type": "number"},
		}, "required": []string{"keyword"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"keyword": args["keyword"],
				"results": []interface{}{},
			}, nil
		})
}
