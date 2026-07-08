package executor

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

// DefaultExecutor handles most OpenAI-compatible providers including
// openai-compatible-* and anthropic-compatible-* custom providers.
type DefaultExecutor struct {
	providerID string
	config     *config.Config
	base       BaseExecutor
}

// NewDefaultExecutor creates a DefaultExecutor for a given provider.
func NewDefaultExecutor(providerID string, cfg *config.Config) *DefaultExecutor {
	return &DefaultExecutor{
		providerID: providerID,
		config:     cfg,
		base:       BaseExecutor{ProviderID: providerID, Config: cfg},
	}
}

// Execute builds URL/headers, transforms the body, and sends the request.
func (e *DefaultExecutor) Execute(ctx context.Context, input ExecuteInput) (*ExecuteResult, error) {
	var url string
	if input.EndpointPath != "" {
		url = e.buildURLForEndpoint(input.EndpointPath, input.Credentials)
	} else {
		url = e.BuildURL(input.Model, input.Stream, input.Credentials)
	}
	headers := e.BuildHeaders(input.Credentials, input.Stream, input.ClientHeaders)
	for k, v := range input.UpstreamExtraHeaders {
		headers[k] = v
	}

	transformedBody := e.TransformRequest(input.Model, input.Body, input.Stream, input.Credentials)
	bodyJSON, err := json.Marshal(transformedBody)
	if err != nil {
		return nil, err
	}

	resp, err := e.base.DoRequest(ctx, "POST", url, headers, bodyJSON, 3, input.SkipUpstreamRetry)
	if err != nil {
		return nil, err
	}

	return &ExecuteResult{
		Response:        resp,
		URL:             url,
		Headers:         headers,
		TransformedBody: transformedBody,
	}, nil
}

// BuildURL constructs the URL based on provider type.
func (e *DefaultExecutor) BuildURL(model string, stream bool, credentials Credentials) string {
	if strings.HasPrefix(e.providerID, "openai-compatible-") {
		return e.buildOpenAICompatibleURL(model, stream, credentials)
	}
	if strings.HasPrefix(e.providerID, "anthropic-compatible-") {
		return e.buildAnthropicCompatibleURL(model, stream, credentials)
	}

	entry := registry.Get(e.providerID)
	if entry != nil && entry.BaseURL != "" {
		baseURL := entry.BaseURL
		if entry.ChatPath != "" {
			return strings.TrimRight(baseURL, "/") + entry.ChatPath
		}
		return strings.TrimRight(baseURL, "/") + "/chat/completions"
	}

	return "https://api.openai.com/v1/chat/completions"
}

func (e *DefaultExecutor) buildOpenAICompatibleURL(model string, stream bool, credentials Credentials) string {
	baseURL := "https://api.openai.com/v1"
	customPath := ""
	if credentials.ProviderSpecificData != nil {
		if bu, ok := credentials.ProviderSpecificData["baseUrl"].(string); ok && bu != "" {
			baseURL = bu
		}
		if cp, ok := credentials.ProviderSpecificData["chatPath"].(string); ok && cp != "" {
			customPath = cp
		}
	}
	normalized := strings.TrimRight(baseURL, "/")
	if customPath != "" {
		return normalized + customPath
	}
	return normalized + "/chat/completions"
}

func (e *DefaultExecutor) buildAnthropicCompatibleURL(model string, stream bool, credentials Credentials) string {
	baseURL := "https://api.anthropic.com/v1"
	customPath := "/messages"
	if credentials.ProviderSpecificData != nil {
		if bu, ok := credentials.ProviderSpecificData["baseUrl"].(string); ok && bu != "" {
			baseURL = bu
		}
		if cp, ok := credentials.ProviderSpecificData["chatPath"].(string); ok && cp != "" {
			customPath = cp
		}
	}
	normalized := strings.TrimRight(baseURL, "/")
	return normalized + customPath
}

// BuildHeaders constructs headers for the provider.
func (e *DefaultExecutor) BuildHeaders(credentials Credentials, stream bool, clientHeaders map[string]string) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	key := credentials.APIKey
	if key == "" {
		key = credentials.AccessToken
	}

	if strings.HasPrefix(e.providerID, "anthropic-compatible-") {
		if key != "" {
			headers["x-api-key"] = key
			baseURL := ""
			if credentials.ProviderSpecificData != nil {
				if bu, ok := credentials.ProviderSpecificData["baseUrl"].(string); ok {
					baseURL = bu
				}
			}
			if !isOfficialAnthropicURL(baseURL) {
				headers["Authorization"] = "Bearer " + key
			}
		}
		headers["anthropic-version"] = "2023-06-01"
	} else {
		if key != "" {
			entry := registry.Get(e.providerID)
			authHeader := "bearer"
			if entry != nil && entry.AuthHeader != "" {
				authHeader = strings.ToLower(entry.AuthHeader)
			}
			switch authHeader {
			case "x-api-key":
				headers["x-api-key"] = key
			default:
				headers["Authorization"] = "Bearer " + key
			}
		}
	}

	if stream {
		headers["Accept"] = "text/event-stream"
	} else {
		headers["Accept"] = "application/json"
	}

	if credentials.ProviderSpecificData != nil {
		if customHeaders, ok := credentials.ProviderSpecificData["customHeaders"].(map[string]interface{}); ok {
			for k, v := range customHeaders {
				if vs, ok := v.(string); ok && !isForbiddenHeader(k) {
					headers[k] = vs
				}
			}
		}
	}

	return headers
}

// TransformRequest modifies the request body for provider-specific requirements.
func (e *DefaultExecutor) TransformRequest(model string, body interface{}, stream bool, credentials Credentials) interface{} {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return body
	}

	if strings.HasPrefix(e.providerID, "anthropic-compatible-") {
		delete(bodyMap, "stream_options")
	}

	if stream && !strings.HasPrefix(e.providerID, "anthropic-compatible-") {
		if _, hasStreamOptions := bodyMap["stream_options"]; !hasStreamOptions {
			if bodyStream, ok := bodyMap["stream"].(bool); !ok || bodyStream {
				bodyMap["stream"] = true
				bodyMap["stream_options"] = map[string]interface{}{
					"include_usage": true,
				}
			}
		}
	}

	if e.providerID == "cerebras" || e.providerID == "mistral" || e.providerID == "nvidia" {
		delete(bodyMap, "client_metadata")
	}

	return bodyMap
}

// RefreshCredentials attempts to refresh OAuth tokens.
func (e *DefaultExecutor) RefreshCredentials(ctx context.Context, credentials Credentials) (*Credentials, error) {
	if credentials.RefreshToken == "" {
		return nil, nil
	}

	entry := registry.Get(e.providerID)
	if entry == nil || entry.OAuth == nil {
		return nil, nil
	}

	log.Printf("[DEFAULT] Attempting token refresh for %s", e.providerID)
	return nil, nil
}

func isOfficialAnthropicURL(baseURL string) bool {
	return baseURL == "" || strings.Contains(baseURL, "api.anthropic.com")
}

func isForbiddenHeader(name string) bool {
	lower := strings.ToLower(name)
	forbidden := []string{"host", "connection", "content-length", "transfer-encoding", "authorization"}
	for _, f := range forbidden {
		if lower == f {
			return true
		}
	}
	return false
}

// buildURLForEndpoint constructs a URL for a specific API endpoint path.
func (e *DefaultExecutor) buildURLForEndpoint(endpointPath string, credentials Credentials) string {
	baseURL := "https://api.openai.com/v1"
	if strings.HasPrefix(e.providerID, "openai-compatible-") {
		if credentials.ProviderSpecificData != nil {
			if bu, ok := credentials.ProviderSpecificData["baseUrl"].(string); ok && bu != "" {
				baseURL = bu
			}
		}
	} else {
		entry := registry.Get(e.providerID)
		if entry != nil && entry.BaseURL != "" {
			// Extract base up to /v1 or similar
			baseURL = entry.BaseURL
		}
	}
	normalized := strings.TrimRight(baseURL, "/")
	// Strip known chat paths to get the base
	for _, strip := range []string{"/chat/completions", "/messages", "/chat"} {
		if strings.HasSuffix(normalized, strip) {
			normalized = strings.TrimSuffix(normalized, strip)
			break
		}
	}
	return normalized + endpointPath
}
