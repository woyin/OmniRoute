package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
)

// MCPServer implements a Model Context Protocol server with JSON-RPC 2.0.
type MCPServer struct {
	DB           *sql.DB
	tools        []Tool
	toolsMu      sync.RWMutex
	sessionID    string
	apiKeyScopes []string
}

// Tool represents a registered MCP tool with its name, description, input schema, and handler.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
	Handler     func(args map[string]interface{}) (interface{}, error)
}

// JSONRPCRequest represents an incoming JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPServer creates a new MCPServer and registers all built-in tools.
func NewMCPServer(db *sql.DB) *MCPServer {
	s := &MCPServer{
		DB:        db,
		sessionID: uuid.New().String(),
	}
	s.registerBuiltinTools()
	return s
}

// RegisterTool adds a tool to the server's registry.
func (s *MCPServer) RegisterTool(tool Tool) {
	s.toolsMu.Lock()
	defer s.toolsMu.Unlock()
	s.tools = append(s.tools, tool)
}

// Tools returns the list of registered tools.
func (s *MCPServer) Tools() []Tool {
	s.toolsMu.RLock()
	defer s.toolsMu.RUnlock()
	return s.tools
}

// simpleTool is a convenience helper for registering tools with minimal boilerplate.
func (s *MCPServer) simpleTool(name, desc string, schema map[string]interface{}, handler func(map[string]interface{}) (interface{}, error)) {
	if schema == nil {
		schema = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
	}
	s.RegisterTool(Tool{Name: name, Description: desc, InputSchema: schema, Handler: handler})
}

// registerBuiltinTools registers all built-in tool groups.
func (s *MCPServer) registerBuiltinTools() {
	s.registerCoreTools()
	s.registerCacheTools()
	s.registerCompressionTools()
	s.registerMemoryTools()
	s.registerSkillTools()
	s.registerGamificationTools()
	s.registerPluginTools()
	s.registerNotionTools()
	s.registerObsidianTools()
	s.registerPoolTools()
	s.registerOneproxyTools()
	s.registerAgentTools()
	s.registerAgentSkillsTools()
	s.registerGithubSkillsTools()
}

// --- HTTP handlers ---

// HandleSSE serves the SSE transport endpoint.
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

// HandleStream serves the Streamable HTTP transport endpoint.
func (s *MCPServer) HandleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.HandleSSE(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if apiKey := extractMCPAPIKey(r); apiKey != "" && s.DB != nil {
		var scopesJSON string
		err := s.DB.QueryRow("SELECT scopes FROM api_keys WHERE key = ? AND is_active = 1", apiKey).Scan(&scopesJSON)
		if err == nil && scopesJSON != "" {
			var scopes []string
			json.Unmarshal([]byte(scopesJSON), &scopes)
			s.apiKeyScopes = scopes
		}
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

// HandleMCPStatus returns the MCP server status as JSON.
func (s *MCPServer) HandleMCPStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.toolsMu.RLock()
	toolCount := len(s.tools)
	s.toolsMu.RUnlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "running", "toolCount": toolCount, "sessionId": s.sessionID,
		"transports": []string{"stdio", "sse", "streamable-http"},
		"scopes":     s.ScopeInfo(),
	})
}

// HandleMCPTools returns the list of registered tools as JSON.
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

// HandleMCPAudit returns the MCP audit log as JSON.
func (s *MCPServer) HandleMCPAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.DB == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"entries": []interface{}{}, "total": 0})
		return
	}
	rows, err := s.DB.Query("SELECT id, tool_name, success, duration_ms, created_at FROM mcp_tool_audit ORDER BY id DESC LIMIT 100")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"entries": []interface{}{}, "total": 0})
		return
	}
	defer rows.Close()
	var entries []map[string]interface{}
	for rows.Next() {
		var id, dur int
		var name, createdAt string
		var success bool
		if rows.Scan(&id, &name, &success, &dur, &createdAt) == nil {
			entries = append(entries, map[string]interface{}{"id": id, "toolName": name, "success": success, "durationMs": dur, "createdAt": createdAt})
		}
	}
	if entries == nil {
		entries = []map[string]interface{}{}
	}
	var total int
	s.DB.QueryRow("SELECT COUNT(*) FROM mcp_tool_audit").Scan(&total)
	json.NewEncoder(w).Encode(map[string]interface{}{"entries": entries, "total": total})
}

// --- JSON-RPC dispatch ---

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

	requiredScope := GetToolScope(toolName)
	if !HasScope(s.apiKeyScopes, requiredScope) {
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{
			Code: -32603, Message: fmt.Sprintf("Insufficient scope: requires '%s'", requiredScope),
		}}
	}

	result, err := found.Handler(args)
	if err != nil {
		go s.auditToolCall(toolName, args, false, 0)
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{
			"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("Error: %s", err.Error())}},
			"isError": true,
		}}
	}
	go s.auditToolCall(toolName, args, true, 0)
	return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{
		"content": []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("%v", result)}},
	}}
}

func (s *MCPServer) auditToolCall(toolName string, args map[string]interface{}, success bool, durationMs int) {
	if s.DB == nil {
		return
	}
	argsJSON, _ := json.Marshal(args)
	s.DB.Exec(
		"INSERT INTO mcp_tool_audit (tool_name, args, result_summary, success, duration_ms) VALUES (?, ?, ?, ?, ?)",
		toolName, string(argsJSON), "", success, durationMs)
}

// --- Stdio transport ---

// StartStdio starts the MCP server over stdin/stdout using JSON-RPC 2.0.
func (s *MCPServer) StartStdio() {
	log.Println("[MCP] Starting stdio transport")
	decoder := json.NewDecoder(os.Stdin)
	for {
		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return
			}
			continue
		}
		resp := s.handleRequest(&req)
		json.NewEncoder(os.Stdout).Encode(resp)
	}
}

func extractMCPAPIKey(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return r.Header.Get("x-api-key")
}

// SetScopes sets the API key scopes for access control.
func (s *MCPServer) SetScopes(scopes []string) {
	s.apiKeyScopes = scopes
}

// ScopeInfo returns scope metadata for the status endpoint.
func (s *MCPServer) ScopeInfo() map[string]interface{} {
	scopeNames := make([]string, len(AllScopes))
	for i, sc := range AllScopes {
		scopeNames[i] = string(sc)
	}
	return map[string]interface{}{
		"definedScopes":      scopeNames,
		"activeScopes":       s.apiKeyScopes,
		"enforcementEnabled": len(s.apiKeyScopes) > 0,
	}
}
