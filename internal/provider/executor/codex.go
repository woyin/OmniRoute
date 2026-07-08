package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

// CodexExecutor handles requests to the OpenAI Codex Responses API.
type CodexExecutor struct {
	config *config.Config
	base   BaseExecutor
}

// NewCodexExecutor creates a CodexExecutor.
func NewCodexExecutor(cfg *config.Config) *CodexExecutor {
	return &CodexExecutor{
		config: cfg,
		base:   BaseExecutor{ProviderID: "codex", Config: cfg},
	}
}

const (
	codexClientVersion = "0.1.2506092151"
	codexUserAgent     = "CodexCLI/0.1.2506092151"
)

// Execute sends a request to the Codex Responses API.
func (e *CodexExecutor) Execute(ctx context.Context, input ExecuteInput) (*ExecuteResult, error) {
	url := e.BuildURL(input.Model, input.Stream, input.Credentials)
	headers := e.BuildHeaders(input.Credentials, input.Stream, input.ClientHeaders)
	for k, v := range input.UpstreamExtraHeaders {
		headers[k] = v
	}

	transformedBody := e.TransformRequest(input.Model, input.Body, input.Stream, input.Credentials)
	bodyJSON, err := json.Marshal(transformedBody)
	if err != nil {
		return nil, fmt.Errorf("marshal codex request: %w", err)
	}

	resp, err := e.base.DoRequest(ctx, "POST", url, headers, bodyJSON, 3, input.SkipUpstreamRetry)
	if err != nil {
		return nil, fmt.Errorf("codex: %w", err)
	}

	// If upstream returned an error status, wrap it properly
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
		Response:       resp,
		URL:            url,
		Headers:        headers,
		TransformedBody: transformedBody,
	}, nil
}

// BuildURL constructs the Codex Responses URL.
func (e *CodexExecutor) BuildURL(model string, stream bool, credentials Credentials) string {
	entry := registry.Get("codex")
	baseURL := "https://chatgpt.com/backend-api/codex/responses"
	if entry != nil && entry.BaseURL != "" {
		baseURL = entry.BaseURL
	}

	if credentials.RequestEndpointPath != "" {
		subpath := getResponsesSubpath(credentials.RequestEndpointPath)
		if subpath != nil {
			return strings.TrimRight(baseURL, "/") + *subpath
		}
	}

	return baseURL
}

// BuildHeaders constructs headers for the Codex Responses API.
func (e *CodexExecutor) BuildHeaders(credentials Credentials, stream bool, clientHeaders map[string]string) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "text/event-stream",
		"Version":      codexClientVersion,
		"User-Agent":   codexUserAgent,
		"Originator":   "codex_cli_rs",
	}

	key := credentials.AccessToken
	if key == "" {
		key = credentials.APIKey
	}
	if key != "" {
		headers["Authorization"] = "Bearer " + key
	}

	if credentials.ProviderSpecificData != nil {
		if wsID, ok := credentials.ProviderSpecificData["workspaceId"].(string); ok && wsID != "" {
			headers["chatgpt-account-id"] = wsID
		}
	}

	return headers
}

// TransformRequest modifies the request body for Codex-specific requirements.
func (e *CodexExecutor) TransformRequest(model string, body interface{}, stream bool, credentials Credentials) interface{} {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return body
	}

	bodyMap["stream"] = true

	if input, ok := bodyMap["input"].([]interface{}); ok {
		for _, item := range input {
			if msg, ok := item.(map[string]interface{}); ok {
				if role, ok := msg["role"].(string); ok && role == "system" {
					msg["role"] = "developer"
				}
			}
		}
	}

	if instructions, ok := bodyMap["instructions"].(string); !ok || instructions == "" {
		bodyMap["instructions"] = "Follow the developer instructions in the conversation."
	}

	if _, hasStore := bodyMap["store"]; !hasStore {
		bodyMap["store"] = false
	}

	stripStoredItemReferences(bodyMap)

	delete(bodyMap, "max_tokens")
	delete(bodyMap, "max_output_tokens")
	delete(bodyMap, "messages")
	delete(bodyMap, "prompt")
	delete(bodyMap, "user")
	delete(bodyMap, "truncation")

	if parsed := splitCodexReasoningSuffix(model); parsed.effort != nil {
		bodyMap["model"] = parsed.baseModel
		if _, hasReasoning := bodyMap["reasoning"]; !hasReasoning {
			bodyMap["reasoning"] = map[string]interface{}{
				"effort":  *parsed.effort,
				"summary": "auto",
			}
		}
	}

	delete(bodyMap, "stream_options")

	allowlist := map[string]bool{
		"model": true, "input": true, "instructions": true, "tools": true,
		"tool_choice": true, "stream": true, "store": true, "reasoning": true,
		"service_tier": true, "include": true, "previous_response_id": true,
		"prompt_cache_key": true, "client_metadata": true, "text": true,
	}
	filtered := make(map[string]interface{})
	for k, v := range bodyMap {
		if allowlist[k] {
			filtered[k] = v
		}
	}

	return filtered
}

// RefreshCredentials refreshes Codex OAuth tokens.
func (e *CodexExecutor) RefreshCredentials(ctx context.Context, credentials Credentials) (*Credentials, error) {
	if credentials.RefreshToken == "" {
		return nil, nil
	}

	tokenURL := e.config.CodexOAuthTokenURL
	if tokenURL == "" {
		tokenURL = "https://auth.openai.com/oauth/token"
	}

	clientID := e.config.CodexOAuthClientID
	clientSecret := e.config.CodexOAuthClientSecret

	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": credentials.RefreshToken,
		"client_id":     clientID,
	}
	if clientSecret != "" {
		payload["client_secret"] = clientSecret
	}

	payloadJSON, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewReader(payloadJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("codex token refresh: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	if errMsg, ok := result["error"].(string); ok {
		return nil, fmt.Errorf("token refresh error: %s", errMsg)
	}

	newCreds := &Credentials{
		AccessToken:          credentials.AccessToken,
		RefreshToken:         credentials.RefreshToken,
		APIKey:               credentials.APIKey,
		ProjectID:            credentials.ProjectID,
		ConnectionID:         credentials.ConnectionID,
		ProviderSpecificData: credentials.ProviderSpecificData,
	}

	if at, ok := result["access_token"].(string); ok {
		newCreds.AccessToken = at
	}
	if rt, ok := result["refresh_token"].(string); ok {
		newCreds.RefreshToken = rt
	}

	return newCreds, nil
}

// --- Codex helpers ---

type codexModelSplit struct {
	baseModel string
	effort    *string
}

func splitCodexReasoningSuffix(model string) codexModelSplit {
	levels := []string{"none", "low", "medium", "high", "xhigh"}
	for _, level := range levels {
		suffix := "-" + level
		if strings.HasSuffix(model, suffix) {
			base := model[:len(model)-len(suffix)]
			return codexModelSplit{baseModel: base, effort: &level}
		}
	}
	return codexModelSplit{baseModel: model}
}

func getResponsesSubpath(endpointPath string) *string {
	normalized := strings.TrimRight(endpointPath, "/")
	lower := strings.ToLower(normalized)

	if lower == "responses" || strings.HasSuffix(lower, "/responses") {
		empty := ""
		return &empty
	}

	responsesSlash := "/responses/"
	idx := strings.LastIndex(lower, responsesSlash)
	if idx != -1 {
		sub := normalized[idx+len("/responses"):]
		return &sub
	}

	if strings.HasPrefix(lower, "responses/") {
		sub := normalized[len("responses"):]
		return &sub
	}

	return nil
}

func stripStoredItemReferences(body map[string]interface{}) {
	input, ok := body["input"].([]interface{})
	if !ok || len(input) == 0 {
		body["input"] = []interface{}{
			map[string]interface{}{
				"type":    "message",
				"role":    "user",
				"content": []interface{}{map[string]interface{}{"type": "input_text", "text": "continue"}},
			},
		}
		return
	}

	serverIDPrefixes := []string{"rs_", "fc_", "resp_", "msg_"}
	var filtered []interface{}
	for _, item := range input {
		if s, ok := item.(string); ok {
			isServerRef := false
			for _, prefix := range serverIDPrefixes {
				if strings.HasPrefix(s, prefix) {
					isServerRef = true
					break
				}
			}
			if !isServerRef {
				filtered = append(filtered, item)
			}
			continue
		}

		if m, ok := item.(map[string]interface{}); ok {
			if t, ok := m["type"].(string); ok && t == "item_reference" {
				continue
			}
			if t, ok := m["type"].(string); ok && t == "reasoning" {
				continue
			}
			if id, ok := m["id"].(string); ok {
				for _, prefix := range serverIDPrefixes {
					if strings.HasPrefix(id, prefix) {
						delete(m, "id")
						break
					}
				}
			}
		}
		filtered = append(filtered, item)
	}
	body["input"] = filtered
}
