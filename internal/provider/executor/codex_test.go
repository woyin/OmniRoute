package executor

import (
	"testing"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

func TestCodexBuildURL(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewCodexExecutor(cfg)

	url := e.BuildURL("gpt-5.5", true, Credentials{})
	if url != "https://chatgpt.com/backend-api/codex/responses" {
		t.Errorf("expected codex responses URL, got %s", url)
	}

	// With responses subpath
	url = e.BuildURL("gpt-5.5", true, Credentials{RequestEndpointPath: "responses"})
	if url != "https://chatgpt.com/backend-api/codex/responses" {
		t.Errorf("expected codex responses URL with subpath, got %s", url)
	}

	// Compact endpoint
	url = e.BuildURL("gpt-5.5", true, Credentials{RequestEndpointPath: "responses/compact"})
	if url != "https://chatgpt.com/backend-api/codex/responses/compact" {
		t.Errorf("expected compact URL, got %s", url)
	}
}

func TestCodexBuildHeaders(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewCodexExecutor(cfg)

	// With access token
	headers := e.BuildHeaders(Credentials{AccessToken: "test-token"}, true, nil)
	if headers["Authorization"] != "Bearer test-token" {
		t.Errorf("expected Bearer test-token, got %s", headers["Authorization"])
	}
	if headers["Version"] != codexClientVersion {
		t.Errorf("expected Version %s, got %s", codexClientVersion, headers["Version"])
	}
	if headers["Originator"] != "codex_cli_rs" {
		t.Errorf("expected Originator codex_cli_rs, got %s", headers["Originator"])
	}

	// With workspace ID
	headers = e.BuildHeaders(Credentials{
		AccessToken:          "test-token",
		ProviderSpecificData: map[string]interface{}{"workspaceId": "ws-123"},
	}, true, nil)
	if headers["chatgpt-account-id"] != "ws-123" {
		t.Errorf("expected chatgpt-account-id ws-123, got %s", headers["chatgpt-account-id"])
	}
}

func TestCodexTransformRequest(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewCodexExecutor(cfg)

	// Force stream=true
	body := map[string]interface{}{
		"model": "gpt-5.5",
		"stream": false,
	}
	result := e.TransformRequest("gpt-5.5", body, true, Credentials{})
	resultMap := result.(map[string]interface{})
	if resultMap["stream"] != true {
		t.Error("expected stream to be forced to true")
	}

	// Default instructions
	if instructions, ok := resultMap["instructions"].(string); ok {
		if instructions == "" {
			t.Error("expected default instructions")
		}
	} else {
		t.Error("expected instructions field")
	}

	// Default store=false
	if resultMap["store"] != false {
		t.Error("expected store to default to false")
	}

	// Strip unsupported fields
	body = map[string]interface{}{
		"model":       "gpt-5.5",
		"max_tokens":  100,
		"messages":    []interface{}{},
		"user":        "test",
		"stream_options": map[string]interface{}{"include_usage": true},
	}
	result = e.TransformRequest("gpt-5.5", body, true, Credentials{})
	resultMap = result.(map[string]interface{})
	if _, hasMaxTokens := resultMap["max_tokens"]; hasMaxTokens {
		t.Error("max_tokens should be stripped")
	}
	if _, hasMessages := resultMap["messages"]; hasMessages {
		t.Error("messages should be stripped")
	}
	if _, hasUser := resultMap["user"]; hasUser {
		t.Error("user should be stripped")
	}
	if _, hasStreamOptions := resultMap["stream_options"]; hasStreamOptions {
		t.Error("stream_options should be stripped")
	}

	// Allowlist enforcement
	if _, hasStore := resultMap["store"]; !hasStore {
		t.Error("store should be in allowlist")
	}
}

func TestSplitCodexReasoningSuffix(t *testing.T) {
	tests := []struct {
		model    string
		base     string
		effort   string
		hasEffort bool
	}{
		{"gpt-5.5", "gpt-5.5", "", false},
		{"gpt-5.5-xhigh", "gpt-5.5", "xhigh", true},
		{"gpt-5.5-high", "gpt-5.5", "high", true},
		{"gpt-5.5-medium", "gpt-5.5", "medium", true},
		{"gpt-5.5-low", "gpt-5.5", "low", true},
		{"gpt-5.5-none", "gpt-5.5", "none", true},
	}

	for _, test := range tests {
		result := splitCodexReasoningSuffix(test.model)
		if result.baseModel != test.base {
			t.Errorf("splitCodexReasoningSuffix(%q).baseModel = %q, want %q", test.model, result.baseModel, test.base)
		}
		if test.hasEffort && result.effort == nil {
			t.Errorf("splitCodexReasoningSuffix(%q) expected effort %q, got nil", test.model, test.effort)
		} else if !test.hasEffort && result.effort != nil {
			t.Errorf("splitCodexReasoningSuffix(%q) expected nil effort, got %v", test.model, *result.effort)
		} else if test.hasEffort && *result.effort != test.effort {
			t.Errorf("splitCodexReasoningSuffix(%q) effort = %q, want %q", test.model, *result.effort, test.effort)
		}
	}
}

func TestGetResponsesSubpath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
		nil      bool
	}{
		{"responses", "", false},
		{"responses/compact", "/compact", false},
		{"/responses/", "", false},
		{"chat", "", true}, // not a responses path
	}

	for _, test := range tests {
		result := getResponsesSubpath(test.path)
		if test.nil {
			if result != nil {
				t.Errorf("getResponsesSubpath(%q) = %q, want nil", test.path, *result)
			}
		} else {
			if result == nil {
				t.Errorf("getResponsesSubpath(%q) = nil, want %q", test.path, test.expected)
			} else if *result != test.expected {
				t.Errorf("getResponsesSubpath(%q) = %q, want %q", test.path, *result, test.expected)
			}
		}
	}
}
