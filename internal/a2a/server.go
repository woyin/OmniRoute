package a2a

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// A2AServer implements an Agent-to-Agent protocol server.
type A2AServer struct {
	DB      *sql.DB
	skills  []A2ASkill
	skillsMu sync.RWMutex
}

// A2ASkill describes a capability exposed via the A2A protocol.
type A2ASkill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        []string `json:"tags"`
}

// A2ATask represents an A2A task.
type A2ATask struct {
	ID        string `json:"id"`
	Status    string `json:"status"` // submitted, working, completed, failed, canceled
	Model     string `json:"model,omitempty"`
	Input     string `json:"input,omitempty"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	ExpiresAt string `json:"expiresAt,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewA2AServer creates a new A2A server.
func NewA2AServer(db *sql.DB) *A2AServer {
	s := &A2AServer{DB: db}
	s.registerBuiltinSkills()
	return s
}

// AgentCard returns the agent card for /.well-known/agent.json.
func (s *A2AServer) AgentCard() map[string]interface{} {
	s.skillsMu.RLock()
	defer s.skillsMu.RUnlock()
	var skills []map[string]interface{}
	for _, sk := range s.skills {
		skills = append(skills, map[string]interface{}{
			"id": sk.ID, "name": sk.Name, "description": sk.Description, "tags": sk.Tags,
		})
	}
	return map[string]interface{}{
		"name":        "OmniRoute",
		"description": "Unified AI proxy/router — route any LLM through one endpoint",
		"url":         "https://omniroute.dev",
		"version":     "4.0.0-go",
		"provider":    map[string]interface{}{"organization": "OmniRoute"},
		"capabilities": map[string]interface{}{
			"streaming":               true,
			"pushNotifications":       false,
			"stateTransitionHistory":  true,
		},
		"skills": skills,
	}
}

// HandleJSONRPC processes A2A JSON-RPC requests.
func (s *A2AServer) HandleJSONRPC(w http.ResponseWriter, r *http.Request) {
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

func (s *A2AServer) handleRequest(req *JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "message/send":
		return s.handleMessageSend(req)
	case "message/stream":
		// SSE streaming - for now, just do sync
		return s.handleMessageSend(req)
	case "tasks/get":
		return s.handleTasksGet(req)
	case "tasks/cancel":
		return s.handleTasksCancel(req)
	default:
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)}}
	}
}

func (s *A2AServer) handleMessageSend(req *JSONRPCRequest) JSONRPCResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Invalid params"}}
	}

	taskID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)
	task := A2ATask{
		ID: taskID, Status: "completed",
		Input: fmt.Sprintf("%v", params["message"]),
		Output: "Task processed by OmniRoute A2A server",
		CreatedAt: now, UpdatedAt: now,
	}

	// Persist to DB
	if s.DB != nil {
		s.DB.Exec(
			"INSERT INTO a2a_tasks (id, status, input, output, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			task.ID, task.Status, task.Input, task.Output, task.CreatedAt, task.UpdatedAt)
	}

	return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: task}
}

func (s *A2AServer) handleTasksGet(req *JSONRPCRequest) JSONRPCResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Invalid params"}}
	}
	taskID, _ := params["id"].(string)

	if s.DB != nil {
		var task A2ATask
		err := s.DB.QueryRow(
			"SELECT id, status, COALESCE(model,''), COALESCE(input,''), COALESCE(output,''), COALESCE(error,''), created_at, updated_at FROM a2a_tasks WHERE id = ?",
			taskID).Scan(&task.ID, &task.Status, &task.Model, &task.Input, &task.Output, &task.Error, &task.CreatedAt, &task.UpdatedAt)
		if err == nil {
			return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: task}
		}
	}

	return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Task not found"}}
}

func (s *A2AServer) handleTasksCancel(req *JSONRPCRequest) JSONRPCResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32602, Message: "Invalid params"}}
	}
	taskID, _ := params["id"].(string)

	if s.DB != nil {
		s.DB.Exec("UPDATE a2a_tasks SET status = 'canceled', updated_at = ? WHERE id = ?",
			time.Now().UTC().Format(time.RFC3339), taskID)
	}

	return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{"id": taskID, "status": "canceled"}}
}

// HandleStatus returns A2A server status.
func (s *A2AServer) HandleStatus(w http.ResponseWriter, r *http.Request) {
	s.skillsMu.RLock()
	skillCount := len(s.skills)
	s.skillsMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "running",
		"skillCount": skillCount,
		"protocol":   "a2a-v0.3",
	})
}

func (s *A2AServer) registerBuiltinSkills() {
	skills := []A2ASkill{
		{ID: "smartRouting", Name: "Smart Routing", Description: "Route requests to the best available provider", Tags: []string{"routing", "ai"}},
		{ID: "quotaManagement", Name: "Quota Management", Description: "Track and manage provider quotas", Tags: []string{"quota", "management"}},
		{ID: "providerDiscovery", Name: "Provider Discovery", Description: "Discover available providers and models", Tags: []string{"discovery", "providers"}},
		{ID: "costAnalysis", Name: "Cost Analysis", Description: "Analyze routing costs and optimize spending", Tags: []string{"cost", "analysis"}},
		{ID: "healthReport", Name: "Health Report", Description: "Generate provider health reports", Tags: []string{"health", "monitoring"}},
		{ID: "listCapabilities", Name: "List Capabilities", Description: "List all OmniRoute capabilities", Tags: []string{"capabilities"}},
	}
	s.skillsMu.Lock()
	s.skills = skills
	s.skillsMu.Unlock()
}

// Ensure log is used
var _ = log.Println
