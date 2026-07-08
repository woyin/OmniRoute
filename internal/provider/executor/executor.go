package executor

import (
	"context"
	"net/http"

	"github.com/omniroute/omniroute/internal/config"
)

// ExecuteResult holds the result of an executor execution.
type ExecuteResult struct {
	Response       *http.Response
	URL            string
	Headers        map[string]string
	TransformedBody interface{}
}

// Credentials holds provider authentication data.
type Credentials struct {
	AccessToken          string
	RefreshToken         string
	APIKey               string
	ProjectID            string
	ConnectionID         string
	RequestEndpointPath  string
	ProviderSpecificData map[string]interface{}
}

// ExecuteInput is the input for an executor's Execute method.
type ExecuteInput struct {
	Model               string
	Body                interface{}
	Stream              bool
	Credentials         Credentials
	UpstreamExtraHeaders map[string]string
	ClientHeaders       map[string]string
	SkipUpstreamRetry   bool
	EndpointPath        string // e.g. "/embeddings", "/images/generations", "/moderations"
}

// Executor is the interface that all provider executors must implement.
type Executor interface {
	// Execute sends a request to the upstream provider and returns the response.
	Execute(ctx context.Context, input ExecuteInput) (*ExecuteResult, error)
	// BuildURL constructs the upstream URL for a given model and stream mode.
	BuildURL(model string, stream bool, credentials Credentials) string
	// BuildHeaders constructs the upstream request headers.
	BuildHeaders(credentials Credentials, stream bool, clientHeaders map[string]string) map[string]string
	// TransformRequest modifies the request body before sending.
	TransformRequest(model string, body interface{}, stream bool, credentials Credentials) interface{}
	// RefreshCredentials attempts to refresh OAuth tokens.
	RefreshCredentials(ctx context.Context, credentials Credentials) (*Credentials, error)
}

// GetExecutor returns the appropriate executor for a provider ID.
// All executors receive the application config for timeout/feature-flag access.
func GetExecutor(providerID string, cfg *config.Config) Executor {
	switch providerID {
	case "opencode", "opencode-zen", "opencode-go":
		return NewOpencodeExecutor(providerID, cfg)
	case "codex":
		return NewCodexExecutor(cfg)
	case "command-code", "cmd":
		return NewCommandCodeExecutor(cfg)
	default:
		// openai-compatible-* and anthropic-compatible-* use DefaultExecutor
		return NewDefaultExecutor(providerID, cfg)
	}
}
