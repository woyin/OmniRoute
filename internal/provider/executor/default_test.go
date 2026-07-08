package executor

import (
	"strings"
	"testing"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

func TestDefaultBuildURLOpenAICompatible(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewDefaultExecutor("openai-compatible-mycustom", cfg)

	// Default base URL
	url := e.BuildURL("my-model", true, Credentials{})
	if !strings.HasSuffix(url, "/chat/completions") {
		t.Errorf("expected chat completions URL, got %s", url)
	}

	// Custom base URL
	url = e.BuildURL("my-model", true, Credentials{
		ProviderSpecificData: map[string]interface{}{
			"baseUrl": "https://my-custom-llm.example.com/v1",
		},
	})
	if url != "https://my-custom-llm.example.com/v1/chat/completions" {
		t.Errorf("expected custom URL, got %s", url)
	}

	// Custom chat path
	url = e.BuildURL("my-model", true, Credentials{
		ProviderSpecificData: map[string]interface{}{
			"baseUrl":  "https://my-custom-llm.example.com/v1",
			"chatPath": "/v2/chat",
		},
	})
	if url != "https://my-custom-llm.example.com/v1/v2/chat" {
		t.Errorf("expected custom path URL, got %s", url)
	}
}

func TestDefaultBuildURLAnthropicCompatible(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewDefaultExecutor("anthropic-compatible-mycustom", cfg)

	// Default Anthropic URL
	url := e.BuildURL("my-model", true, Credentials{})
	if !strings.HasSuffix(url, "/messages") {
		t.Errorf("expected messages URL, got %s", url)
	}

	// Custom base URL
	url = e.BuildURL("my-model", true, Credentials{
		ProviderSpecificData: map[string]interface{}{
			"baseUrl": "https://my-anthropic-proxy.example.com",
		},
	})
	if url != "https://my-anthropic-proxy.example.com/messages" {
		t.Errorf("expected custom anthropic URL, got %s", url)
	}
}

func TestDefaultBuildHeadersAnthropic(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewDefaultExecutor("anthropic-compatible-mycustom", cfg)

	headers := e.BuildHeaders(Credentials{APIKey: "sk-ant-test"}, true, nil)
	if headers["x-api-key"] != "sk-ant-test" {
		t.Errorf("expected x-api-key, got %s", headers["x-api-key"])
	}
	if headers["anthropic-version"] != "2023-06-01" {
		t.Errorf("expected anthropic-version, got %s", headers["anthropic-version"])
	}

	// Third-party gateway: should also set Authorization header
	headers = e.BuildHeaders(Credentials{
		APIKey: "sk-ant-test",
		ProviderSpecificData: map[string]interface{}{
			"baseUrl": "https://my-gateway.example.com",
		},
	}, true, nil)
	if headers["Authorization"] != "Bearer sk-ant-test" {
		t.Errorf("expected Authorization header for third-party gateway, got %s", headers["Authorization"])
	}
}

func TestDefaultBuildHeadersOpenAI(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewDefaultExecutor("openai-compatible-mycustom", cfg)

	headers := e.BuildHeaders(Credentials{APIKey: "sk-test"}, true, nil)
	if headers["Authorization"] != "Bearer sk-test" {
		t.Errorf("expected Bearer token, got %s", headers["Authorization"])
	}
}

func TestDefaultTransformRequest(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	// Anthropic-compatible should strip stream_options
	e := NewDefaultExecutor("anthropic-compatible-mycustom", cfg)
	body := map[string]interface{}{
		"model":          "my-model",
		"stream_options": map[string]interface{}{"include_usage": true},
	}
	result := e.TransformRequest("my-model", body, true, Credentials{})
	resultMap := result.(map[string]interface{})
	if _, hasSO := resultMap["stream_options"]; hasSO {
		t.Error("stream_options should be stripped for anthropic-compatible")
	}

	// OpenAI-compatible should add stream_options for streaming
	e2 := NewDefaultExecutor("openai-compatible-mycustom", cfg)
	body = map[string]interface{}{
		"model":  "my-model",
		"stream": true,
	}
	result = e2.TransformRequest("my-model", body, true, Credentials{})
	resultMap = result.(map[string]interface{})
	if so, ok := resultMap["stream_options"].(map[string]interface{}); ok {
		if so["include_usage"] != true {
			t.Error("expected include_usage=true in stream_options")
		}
	} else {
		t.Error("expected stream_options for streaming openai-compatible")
	}
}

func TestIsForbiddenHeader(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"Host", true},
		{"Connection", true},
		{"Content-Length", true},
		{"Authorization", true},
		{"X-Custom-Header", false},
		{"Accept", false},
	}
	for _, test := range tests {
		result := isForbiddenHeader(test.name)
		if result != test.expected {
			t.Errorf("isForbiddenHeader(%q) = %v, want %v", test.name, result, test.expected)
		}
	}
}
