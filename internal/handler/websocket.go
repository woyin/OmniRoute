package handler

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/executor"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// WSChatRequest is a WebSocket chat completion request.
type WSChatRequest struct {
	Type     string                   `json:"type"`
	ID       string                   `json:"id,omitempty"`
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
	Stream   bool                     `json:"stream"`
}

// WSHandler handles WebSocket connections for real-time streaming chat.
type WSHandler struct {
	DB     *sql.DB
	Config *config.Config
}

// ServeHTTP upgrades the HTTP connection to WebSocket and handles the lifecycle.
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("[WS] Connection established from %s", r.RemoteAddr)

	for {
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Read error: %v", err)
			}
			break
		}
		if mt == websocket.CloseMessage {
			break
		}

		var req WSChatRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			h.sendError(conn, "invalid_request", "Invalid JSON")
			continue
		}

		if req.Model == "" {
			h.sendError(conn, "invalid_request", "model is required")
			continue
		}

		h.handleChatRequest(conn, r, &req)
	}

	log.Printf("[WS] Connection closed from %s", r.RemoteAddr)
}

func (h *WSHandler) handleChatRequest(conn *websocket.Conn, r *http.Request, req *WSChatRequest) {
	providerID, resolvedModel := resolveProvider(req.Model)
	credentials := resolveCredentials(r, h.DB, providerID)

	body := map[string]interface{}{
		"model":    resolvedModel,
		"messages": req.Messages,
		"stream":   req.Stream,
	}

	exec := executor.GetExecutor(providerID, h.Config)
	ctx := r.Context()

	result, err := exec.Execute(ctx, executor.ExecuteInput{
		Model:       resolvedModel,
		Body:        body,
		Stream:      req.Stream,
		Credentials: credentials,
	})
	if err != nil {
		h.sendError(conn, "upstream_error", err.Error())
		return
	}

	if result.Response.StatusCode >= 400 {
		result.Response.Body.Close()
		h.sendError(conn, "upstream_error", fmt.Sprintf("upstream returned %d", result.Response.StatusCode))
		return
	}

	if req.Stream {
		h.forwardSSEToWS(conn, result.Response)
	} else {
		h.forwardJSONResponse(conn, result.Response)
	}
}

// forwardSSEToWS reads an upstream SSE stream and forwards each chunk as a WebSocket JSON message.
func (h *WSHandler) forwardSSEToWS(conn *websocket.Conn, resp *http.Response) {
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			_ = conn.WriteJSON(map[string]interface{}{
				"type": "chat.completion.done",
			})
			break
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		chunk["type"] = "chat.completion.chunk"
		if err := conn.WriteJSON(chunk); err != nil {
			log.Printf("[WS] Write error: %v", err)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[WS] SSE scan error: %v", err)
	}
}

// forwardJSONResponse reads a full upstream JSON response and sends it as a single WebSocket message.
func (h *WSHandler) forwardJSONResponse(conn *websocket.Conn, resp *http.Response) {
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.sendError(conn, "upstream_error", "Failed to decode upstream response")
		return
	}

	result["type"] = "chat.completion"
	if err := conn.WriteJSON(result); err != nil {
		log.Printf("[WS] Write error: %v", err)
	}
}

func (h *WSHandler) sendError(conn *websocket.Conn, errType, message string) {
	_ = conn.WriteJSON(map[string]interface{}{
		"type":    "error",
		"error":   message,
		"errType": errType,
	})
}
