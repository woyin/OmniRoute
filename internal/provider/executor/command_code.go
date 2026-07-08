package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

// CommandCodeExecutor handles requests to the Command Code API.
type CommandCodeExecutor struct {
	config *config.Config
	base   BaseExecutor
}

// NewCommandCodeExecutor creates a CommandCodeExecutor.
func NewCommandCodeExecutor(cfg *config.Config) *CommandCodeExecutor {
	return &CommandCodeExecutor{
		config: cfg,
		base:   BaseExecutor{ProviderID: "command-code", Config: cfg},
	}
}

const (
	commandCodeVersion   = "0.33.2"
	maxCommandCodeTokens = 200000
)

// Execute sends a request to the Command Code API.
func (e *CommandCodeExecutor) Execute(ctx context.Context, input ExecuteInput) (*ExecuteResult, error) {
	apiKey := input.Credentials.APIKey
	if apiKey == "" {
		apiKey = input.Credentials.AccessToken
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Command Code API key required")
	}

	url := e.BuildURL(input.Model, input.Stream, input.Credentials)
	headers := map[string]string{
		"Content-Type":           "application/json",
		"Authorization":          "Bearer " + apiKey,
		"x-command-code-version": commandCodeVersion,
		"x-cli-environment":      "external",
		"x-project-slug":         "pi-cc",
		"x-taste-learning":       "false",
		"x-co-flag":              "false",
		"x-session-id":           uuid.New().String(),
	}

	for k, v := range input.UpstreamExtraHeaders {
		headers[k] = v
	}

	transformedBody := e.TransformRequest(input.Model, input.Body, input.Stream, input.Credentials)
	bodyJSON, err := json.Marshal(transformedBody)
	if err != nil {
		return nil, fmt.Errorf("marshal command-code request: %w", err)
	}

	resp, err := e.base.DoRequest(ctx, "POST", url, headers, bodyJSON, 1, true)
	if err != nil {
		return nil, fmt.Errorf("command-code: %w", err)
	}

	// If upstream returned an error status
	if resp.StatusCode >= 400 {
		body := e.base.ReadErrorBody(resp)
		return &ExecuteResult{
			Response: &http.Response{
				StatusCode: resp.StatusCode,
				Header:     resp.Header,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			URL:             url,
			Headers:         headers,
			TransformedBody: transformedBody,
		}, nil
	}

	return &ExecuteResult{
		Response:        resp,
		URL:             url,
		Headers:         headers,
		TransformedBody: transformedBody,
	}, nil
}

// BuildURL returns the Command Code endpoint URL.
func (e *CommandCodeExecutor) BuildURL(model string, stream bool, credentials Credentials) string {
	baseURL := "https://api.commandcode.ai"
	entry := registry.Get("command-code")
	if entry != nil && entry.BaseURL != "" {
		baseURL = entry.BaseURL
	}
	chatPath := "/alpha/generate"
	if entry != nil && entry.ChatPath != "" {
		chatPath = entry.ChatPath
	}
	return strings.TrimRight(baseURL, "/") + chatPath
}

// BuildHeaders is handled in Execute for Command Code.
func (e *CommandCodeExecutor) BuildHeaders(credentials Credentials, stream bool, clientHeaders map[string]string) map[string]string {
	return map[string]string{"Content-Type": "application/json"}
}

// TransformRequest converts an OpenAI-format body into Command Code's internal format.
func (e *CommandCodeExecutor) TransformRequest(model string, body interface{}, stream bool, credentials Credentials) interface{} {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		bodyMap = map[string]interface{}{}
	}

	resolvedModel := model
	if m, ok := bodyMap["model"].(string); ok && m != "" {
		resolvedModel = m
	}

	converted := convertMessagesForCommandCode(bodyMap["messages"], resolvedModel)

	systemParts := []string{}
	if converted.System != "" {
		systemParts = append(systemParts, converted.System)
	}
	if s, ok := bodyMap["system"].(string); ok && s != "" {
		systemParts = append(systemParts, s)
	}
	system := strings.Join(systemParts, "\n\n")

	params := map[string]interface{}{
		"model":   resolvedModel,
		"messages": converted.Messages,
		"tools":   convertToolsForCommandCode(bodyMap["tools"]),
		"system":  system,
		"stream":  true,
	}

	if maxTokens := clampMaxTokens(bodyMap["max_tokens"], bodyMap["max_completion_tokens"]); maxTokens != nil {
		params["max_tokens"] = *maxTokens
	}

	passthroughFields := []string{"reasoning_effort", "reasoning", "thinking", "effort", "output_config", "extra_body"}
	for _, field := range passthroughFields {
		if v, ok := bodyMap[field]; ok && v != nil {
			params[field] = v
		}
	}

	return map[string]interface{}{
		"config": map[string]interface{}{
			"workingDir":    "/workspace",
			"date":          time.Now().Format("2006-01-02"),
			"environment":   "external",
			"structure":     []interface{}{},
			"isGitRepo":     false,
			"currentBranch": "",
			"mainBranch":    "",
			"gitStatus":     "",
			"recentCommits": []interface{}{},
		},
		"memory":         "",
		"taste":          "",
		"skills":         "",
		"permissionMode": "standard",
		"params":         params,
	}
}

// RefreshCredentials returns nil for Command Code (API key auth).
func (e *CommandCodeExecutor) RefreshCredentials(ctx context.Context, credentials Credentials) (*Credentials, error) {
	return nil, nil
}

// --- Command Code helpers ---

func clampMaxTokens(values ...interface{}) *int {
	for _, v := range values {
		if n, ok := v.(float64); ok && n > 0 {
			clamped := int(n)
			if clamped > maxCommandCodeTokens {
				clamped = maxCommandCodeTokens
			}
			return &clamped
		}
		if n, ok := v.(int); ok && n > 0 {
			clamped := n
			if clamped > maxCommandCodeTokens {
				clamped = maxCommandCodeTokens
			}
			return &clamped
		}
	}
	return nil
}

type commandCodeConverted struct {
	System   string
	Messages []interface{}
}

func convertMessagesForCommandCode(messages interface{}, model string) commandCodeConverted {
	arr, ok := messages.([]interface{})
	if !ok {
		return commandCodeConverted{}
	}
	_ = model

	var systemParts []string
	var out []interface{}
	pairedCallIDs := findPairedToolCallIDs(arr)

	for _, item := range arr {
		msg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)

		if role == "system" || role == "developer" {
			text := normalizeContentText(msg["content"])
			if text != "" {
				systemParts = append(systemParts, text)
			}
			continue
		}

		if role == "user" {
			out = append(out, map[string]interface{}{
				"role":    "user",
				"content": normalizeContentText(msg["content"]),
			})
			continue
		}

		if role == "assistant" {
			var parts []interface{}
			text := normalizeContentText(msg["content"])
			if text != "" {
				parts = append(parts, map[string]interface{}{"type": "text", "text": text})
			}
			if toolCalls, ok := msg["tool_calls"].([]interface{}); ok {
				for _, call := range toolCalls {
					cc, ok := call.(map[string]interface{})
					if !ok {
						continue
					}
					id, _ := cc["id"].(string)
					if id == "" || !pairedCallIDs[id] {
						continue
					}
					fn, _ := cc["function"].(map[string]interface{})
					name, _ := fn["name"].(string)
					args := recordOrEmpty(fn["arguments"])
					parts = append(parts, map[string]interface{}{
						"type":       "tool-call",
						"toolCallId": id,
						"toolName":   name,
						"input":      args,
					})
				}
			}
			if len(parts) > 0 {
				out = append(out, map[string]interface{}{"role": "assistant", "content": parts})
			}
			continue
		}

		if role == "tool" {
			toolCallID, _ := msg["tool_call_id"].(string)
			if toolCallID == "" || !pairedCallIDs[toolCallID] {
				continue
			}
			name, _ := msg["name"].(string)
			out = append(out, map[string]interface{}{
				"role":    "tool",
				"content": []interface{}{
					map[string]interface{}{
						"type":       "tool-result",
						"toolCallId": toolCallID,
						"toolName":   name,
						"output":     map[string]interface{}{"type": "text", "value": normalizeContentText(msg["content"])},
					},
				},
			})
		}
	}

	return commandCodeConverted{
		System:   strings.Join(systemParts, "\n\n"),
		Messages: out,
	}
}

func findPairedToolCallIDs(messages []interface{}) map[string]bool {
	callIDs := map[string]bool{}
	resultIDs := map[string]bool{}

	for _, item := range messages {
		msg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		if role == "assistant" {
			if toolCalls, ok := msg["tool_calls"].([]interface{}); ok {
				for _, call := range toolCalls {
					cc, ok := call.(map[string]interface{})
					if !ok {
						continue
					}
					id, _ := cc["id"].(string)
					if id != "" {
						callIDs[id] = true
					}
				}
			}
		} else if role == "tool" {
			id, _ := msg["tool_call_id"].(string)
			if id != "" {
				resultIDs[id] = true
			}
		}
	}

	paired := map[string]bool{}
	for id := range callIDs {
		if resultIDs[id] {
			paired[id] = true
		}
	}
	return paired
}

func normalizeContentText(content interface{}) string {
	if s, ok := content.(string); ok {
		return s
	}
	if arr, ok := content.([]interface{}); ok {
		var parts []string
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				if t, ok := m["type"].(string); ok && t == "text" {
					if text, ok := m["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

func recordOrEmpty(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	if s, ok := v.(string); ok && s != "" {
		var parsed map[string]interface{}
		if json.Unmarshal([]byte(s), &parsed) == nil {
			return parsed
		}
	}
	return map[string]interface{}{}
}

func convertToolsForCommandCode(tools interface{}) []interface{} {
	arr, ok := tools.([]interface{})
	if !ok {
		return []interface{}{}
	}
	var result []interface{}
	for _, tool := range arr {
		t, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}
		fn, ok := t["function"].(map[string]interface{})
		if !ok {
			fn = t
		}
		name, _ := fn["name"].(string)
		desc, _ := fn["description"].(string)
		params, _ := fn["parameters"].(map[string]interface{})
		if params == nil {
			params = map[string]interface{}{}
		}
		result = append(result, map[string]interface{}{
			"type":         "function",
			"name":         name,
			"description":  desc,
			"input_schema": params,
		})
	}
	return result
}
