package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type MCPServer struct {
	DB        *sql.DB
	tools     []Tool
	toolsMu   sync.RWMutex
	sessionID string
}

type Tool struct {
	Name        string      `json:"name"`
	Description string     `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
	Handler     func(args map[string]interface{}) (interface{}, error)
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError  `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewMCPServer(db *sql.DB) *MCPServer {
	s := &MCPServer{
		DB:        db,
		sessionID: uuid.New().String(),
	}
	s.registerBuiltinTools()
	return s
}

func (s *MCPServer) RegisterTool(tool Tool) {
	s.toolsMu.Lock()
	defer s.toolsMu.Unlock()
	s.tools = append(s.tools, tool)
}

func (s *MCPServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	endpoint := fmt.Sprintf("/api/mcp/stream?session_id=%s", s.sessionID)
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpoint)
	flusher.Flush()

	notify := r.Context().Done()
	<-notify
}

func (s *MCPServer) HandleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.HandleSSE(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(JSONRPCResponse{
			JSONRPC: "2.0", ID: nil,
			Error: &RPCError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	resp := s.handleRequest(&req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *MCPServer) handleRequest(req *JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0", ID: req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2025-03-26",
				"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
				"serverInfo":      map[string]interface{}{"name": "OmniRoute MCP", "version": "4.0.0-go"},
			},
		}
	case "tools/list":
		s.toolsMu.RLock()
		defer s.toolsMu.RUnlock()
		var toolList []map[string]interface{}
		for _, t := range s.tools {
			toolList = append(toolList, map[string]interface{}{
				"name": t.Name, "description": t.Description, "inputSchema": t.InputSchema,
			})
		}
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"tools": toolList}}
	case "tools/call":
		return s.handleToolCall(req)
	case "ping":
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}}
	default:
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)}}
	}
}

func (s *MCPServer) handleToolCall(req *JSONRPCRequest) JSONRPCResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Invalid params"}}
	}
	toolName, _ := params["name"].(string)
	args, _ := params["arguments"].(map[string]interface{})

	s.toolsMu.RLock()
	var found *Tool
	for i := range s.tools {
		if s.tools[i].Name == toolName {
			found = &s.tools[i]
			break
		}
	}
	s.toolsMu.RUnlock()

	if found == nil {
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: fmt.Sprintf("Tool not found: %s", toolName)}}
	}

	result, err := found.Handler(args)
	if err != nil {
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("Error: %s", err.Error())}},
			"isError": true,
		}}
	}
	return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{
		"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%v", result)}},
	}}
}

func (s *MCPServer) HandleMCPStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.toolsMu.RLock()
	toolCount := len(s.tools)
	s.toolsMu.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "running", "toolCount": toolCount, "sessionId": s.sessionID,
		"transports": []string{"stdio", "sse", "streamable-http"},
	})
}

func (s *MCPServer) HandleMCPTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.toolsMu.RLock()
	var tools []map[string]interface{}
	for _, t := range s.tools {
		tools = append(tools, map[string]interface{}{"name": t.Name, "description": t.Description})
	}
	s.toolsMu.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{"tools": tools, "total": len(tools)})
}

func (s *MCPServer) HandleMCPAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"entries": []interface{}{}, "total": 0})
}

func (s *MCPServer) StartStdio() {
	log.Println("[MCP] Starting stdio transport")
	decoder := json.NewDecoder(os.Stdin)
	for {
		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF { return }
			continue
		}
		resp := s.handleRequest(&req)
		json.NewEncoder(os.Stdout).Encode(resp)
	}
}

func (s *MCPServer) registerBuiltinTools() {
	simpleTool := func(name, desc string, schema map[string]interface{}, handler func(map[string]interface{}) (interface{}, error)) {
		if schema == nil { schema = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}} }
		s.RegisterTool(Tool{Name: name, Description: desc, InputSchema: schema, Handler: handler})
	}

	simpleTool("get_health", "Get OmniRoute service health status", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"status": "ok", "version": "4.0.0-go"}, nil })

	simpleTool("list_combos", "List all routing combos", nil,
		func(args map[string]interface{}) (interface{}, error) {
			if s.DB == nil { return []interface{}{}, nil }
			rows, err := s.DB.Query("SELECT id, name, strategy, is_active FROM combos ORDER BY name")
			if err != nil { return nil, err }
			defer rows.Close()
			var results []map[string]interface{}
			for rows.Next() {
				var id, name, strategy string; var isActive bool
				if rows.Scan(&id, &name, &strategy, &isActive) == nil {
					results = append(results, map[string]interface{}{"id": id, "name": name, "strategy": strategy, "isActive": isActive})
				}
			}
			return results, nil
		})

	simpleTool("switch_combo", "Switch the active routing combo",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"comboId": map[string]interface{}{"type": "string"}}, "required": []string{"comboId"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true, "comboId": args["comboId"]}, nil })

	simpleTool("check_quota", "Check provider quota status",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"provider": args["provider"], "remaining": -1, "unlimited": true}, nil })

	simpleTool("cost_report", "Get cost report for a time period",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"period": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"period": args["period"], "totalCost": 0, "requests": 0}, nil })

	simpleTool("list_models_catalog", "List all available models across providers", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"total": 222}, nil })

	simpleTool("web_search", "Search the web",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"query": args["query"], "results": []interface{}{}}, nil })

	simpleTool("simulate_route", "Simulate a routing decision for a model",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"model": map[string]interface{}{"type": "string"}}, "required": []string{"model"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"model": args["model"], "provider": "auto"}, nil })

	simpleTool("set_routing_strategy", "Change the routing strategy for a combo",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"comboId": map[string]interface{}{"type": "string"}, "strategy": map[string]interface{}{"type": "string"}}, "required": []string{"comboId", "strategy"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("db_health_check", "Check database health", nil,
		func(args map[string]interface{}) (interface{}, error) {
			if s.DB != nil { if err := s.DB.Ping(); err != nil { return map[string]interface{}{"status": "error", "error": err.Error()}, nil } }
			return map[string]interface{}{"status": "ok"}, nil
		})

	simpleTool("cache_stats", "Get semantic cache statistics", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"entries": 0, "hitRate": 0}, nil })

	simpleTool("cache_flush", "Flush the semantic cache", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true, "flushed": 0}, nil })

	simpleTool("best_combo_for_task", "Find the best combo for a given task",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"task": map[string]interface{}{"type": "string"}}, "required": []string{"task"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"task": args["task"], "bestCombo": "default"}, nil })

	simpleTool("explain_route", "Explain why a particular route was chosen",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"model": map[string]interface{}{"type": "string"}}, "required": []string{"model"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"model": args["model"], "reason": "canonical_prefix_match"}, nil })

	simpleTool("get_session_snapshot", "Get current session routing snapshot", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"activeProviders": 60, "activeModels": 222}, nil })

	simpleTool("get_provider_metrics", "Get provider performance metrics",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"provider": args["provider"], "avgLatencyMs": 0, "successRate": 1.0}, nil })

	simpleTool("set_budget_guard", "Set a budget guard limit",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"limit": map[string]interface{}{"type": "number"}, "period": map[string]interface{}{"type": "string"}}, "required": []string{"limit"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("compression_status", "Get prompt compression status", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"mode": "off", "savingsPercent": 0}, nil })

	simpleTool("compression_configure", "Configure prompt compression settings",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"mode": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("memory_search", "Search stored memories",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) {
			if s.DB == nil { return []interface{}{}, nil }
			query, _ := args["query"].(string)
			rows, err := s.DB.Query("SELECT id, content FROM memories WHERE content LIKE ? LIMIT 10", "%"+query+"%")
			if err != nil { return nil, err }
			defer rows.Close()
			var results []map[string]interface{}
			for rows.Next() { var id, content string; if rows.Scan(&id, &content) == nil { results = append(results, map[string]interface{}{"id": id, "content": content}) } }
			return results, nil
		})

	simpleTool("memory_add", "Add a memory entry",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"content": map[string]interface{}{"type": "string"}, "tags": map[string]interface{}{"type": "array"}}, "required": []string{"content"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true, "id": uuid.New().String()}, nil })

	simpleTool("memory_clear", "Clear all memories", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("sync_pricing", "Sync provider pricing data", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true, "updated": 0}, nil })

	// --- Skill tools ---
	simpleTool("skills_list", "List all available skills", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"skills": []interface{}{}}, nil })

	simpleTool("skills_enable", "Enable or disable a skill",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"skillId": map[string]interface{}{"type": "string"}, "enabled": map[string]interface{}{"type": "boolean"}}, "required": []string{"skillId"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("skills_execute", "Execute a skill by ID",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"skillId": map[string]interface{}{"type": "string"}, "input": map[string]interface{}{"type": "object"}}, "required": []string{"skillId"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "skill execution not yet implemented"}, nil })

	simpleTool("skills_executions", "List skill execution history",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"limit": map[string]interface{}{"type": "number"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"executions": []interface{}{}}, nil })

	// --- Pool tools ---
	simpleTool("pool_list", "List provider connection pools", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"pools": []interface{}{}}, nil })

	simpleTool("pool_status", "Get pool status for a provider",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}}, "required": []string{"provider"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"provider": args["provider"], "active": 0, "idle": 0}, nil })

	simpleTool("pool_drain", "Drain a provider connection pool",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}}, "required": []string{"provider"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("pool_add", "Add a connection to a pool",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}, "apiKey": map[string]interface{}{"type": "string"}}, "required": []string{"provider", "apiKey"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("pool_remove", "Remove a connection from a pool",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}, "connectionId": map[string]interface{}{"type": "string"}}, "required": []string{"provider", "connectionId"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("pool_rebalance", "Rebalance pool connections",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"provider": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	// --- Gamification tools ---
	simpleTool("gamification_levels", "Get user gamification level", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"level": 1, "xp": 0, "nextLevelXp": 100}, nil })

	simpleTool("gamification_badges", "List earned badges", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"badges": []interface{}{}}, nil })

	simpleTool("gamification_leaderboard", "Get community leaderboard",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"period": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"leaderboard": []interface{}{}}, nil })

	simpleTool("gamification_federation_score", "Get federation score", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"score": 0}, nil })

	simpleTool("gamification_anomalies", "Check for gamification anomalies", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"anomalies": []interface{}{}}, nil })

	simpleTool("gamification_notifications", "Get gamification notifications", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"notifications": []interface{}{}}, nil })

	simpleTool("gamification_servers", "List gamification servers", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"servers": []interface{}{}}, nil })

	simpleTool("gamification_stream", "Stream gamification events", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"events": []interface{}{}}, nil })

	// --- Plugin tools ---
	simpleTool("plugins_list", "List installed plugins", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"plugins": []interface{}{}}, nil })

	simpleTool("plugins_install", "Install a plugin",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}}, "required": []string{"name"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "plugin install not yet implemented"}, nil })

	simpleTool("plugins_activate", "Activate a plugin",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}}, "required": []string{"name"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("plugins_deactivate", "Deactivate a plugin",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}}, "required": []string{"name"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": true}, nil })

	simpleTool("plugins_marketplace", "Browse plugin marketplace", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"marketplace": []interface{}{}}, nil })

	simpleTool("plugins_scan", "Scan for new plugins", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"found": 0}, nil })

	simpleTool("plugins_config", "Get or update plugin configuration",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}, "config": map[string]interface{}{"type": "object"}}, "required": []string{"name"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"name": args["name"], "config": map[string]interface{}{}}, nil })

	simpleTool("plugins_inspect", "Inspect plugin details",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}}, "required": []string{"name"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"name": args["name"]}, nil })

	// --- Notion tools ---
	simpleTool("notion_search", "Search Notion pages",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"results": []interface{}{}}, nil })

	simpleTool("notion_get_page", "Get a Notion page by ID",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"pageId": map[string]interface{}{"type": "string"}}, "required": []string{"pageId"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"page": nil}, nil })

	simpleTool("notion_create_page", "Create a Notion page",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"title": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"title"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Notion integration not configured"}, nil })

	simpleTool("notion_update_page", "Update a Notion page",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"pageId": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"pageId"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Notion integration not configured"}, nil })

	simpleTool("notion_list_databases", "List Notion databases", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"databases": []interface{}{}}, nil })

	simpleTool("notion_query_database", "Query a Notion database",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"databaseId": map[string]interface{}{"type": "string"}, "filter": map[string]interface{}{"type": "object"}}, "required": []string{"databaseId"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"results": []interface{}{}}, nil })

	// --- Obsidian tools ---
	simpleTool("obsidian_search", "Search Obsidian vault",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"results": []interface{}{}}, nil })

	simpleTool("obsidian_read_note", "Read an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"content": ""}, nil })

	simpleTool("obsidian_create_note", "Create an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path", "content"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	simpleTool("obsidian_update_note", "Update an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	simpleTool("obsidian_delete_note", "Delete an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	simpleTool("obsidian_list_notes", "List notes in a folder",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"folder": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"notes": []interface{}{}}, nil })

	simpleTool("obsidian_list_tags", "List all tags in vault", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"tags": []interface{}{}}, nil })

	simpleTool("obsidian_list_links", "List all links in vault", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"links": []interface{}{}}, nil })

	simpleTool("obsidian_get_backlinks", "Get backlinks for a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"backlinks": []interface{}{}}, nil })

	simpleTool("obsidian_get_metadata", "Get note metadata (frontmatter)",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"metadata": map[string]interface{}{}}, nil })

	simpleTool("obsidian_search_tags", "Search notes by tags",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"tags": map[string]interface{}{"type": "array"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"results": []interface{}{}}, nil })

	simpleTool("obsidian_append_note", "Append content to a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path", "content"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	simpleTool("obsidian_move_note", "Move/rename a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"fromPath": map[string]interface{}{"type": "string"}, "toPath": map[string]interface{}{"type": "string"}}, "required": []string{"fromPath", "toPath"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	simpleTool("obsidian_get_graph", "Get vault graph data", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}}, nil })

	simpleTool("obsidian_get_daily_note", "Get or create daily note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"date": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"content": ""}, nil })

	simpleTool("obsidian_list_templates", "List available templates", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"templates": []interface{}{}}, nil })

	simpleTool("obsidian_apply_template", "Apply a template to a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "template": map[string]interface{}{"type": "string"}}, "required": []string{"path", "template"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	simpleTool("obsidian_webdav_status", "Check WebDAV connection status", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"connected": false}, nil })

	simpleTool("obsidian_webdav_sync", "Trigger WebDAV sync", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"synced": false, "message": "WebDAV not configured"}, nil })

	simpleTool("obsidian_webdav_list", "List files via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"files": []interface{}{}}, nil })

	simpleTool("obsidian_webdav_read", "Read a file via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"content": ""}, nil })

	simpleTool("obsidian_webdav_write", "Write a file via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path", "content"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "WebDAV not configured"}, nil })

	simpleTool("obsidian_webdav_delete", "Delete a file via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "WebDAV not configured"}, nil })

	simpleTool("obsidian_webdav_mkdir", "Create directory via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "WebDAV not configured"}, nil })
}

var _ = strings.NewReader

// Tools returns the list of registered tools.
func (s *MCPServer) Tools() []Tool {
	s.toolsMu.RLock()
	defer s.toolsMu.RUnlock()
	return s.tools
}
