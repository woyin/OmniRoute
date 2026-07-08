package executor

import (
	"testing"

	"github.com/omniroute/omniroute/internal/config"
	"github.com/omniroute/omniroute/internal/provider/registry"
)

func TestOpencodeBuildURL(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewOpencodeExecutor("opencode", cfg)

	// Default format: openai
	url := e.BuildURL("big-pickle", true, Credentials{})
	if url != "https://opencode.ai/zen/v1/chat/completions" {
		t.Errorf("expected zen chat completions URL, got %s", url)
	}

	// Claude format model
	e.requestFormat = "claude"
	url = e.BuildURL("qwen3.6-plus-free", false, Credentials{})
	if url != "https://opencode.ai/zen/v1/messages" {
		t.Errorf("expected messages URL, got %s", url)
	}

	// OpenCode Go variant
	e2 := NewOpencodeExecutor("opencode-go", cfg)
	url = e2.BuildURL("deepseek-v4-pro", true, Credentials{})
	if url != "https://opencode.ai/go/v1/chat/completions" {
		t.Errorf("expected go chat completions URL, got %s", url)
	}
}

func TestOpencodeBuildHeaders(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewOpencodeExecutor("opencode", cfg)

	// No auth (free tier)
	headers := e.BuildHeaders(Credentials{}, true, nil)
	if headers["Content-Type"] != "application/json" {
		t.Error("expected Content-Type application/json")
	}
	if _, hasAuth := headers["Authorization"]; hasAuth {
		t.Error("should not have Authorization header for free tier")
	}

	// With API key
	headers = e.BuildHeaders(Credentials{APIKey: "test-key"}, true, nil)
	if headers["Authorization"] != "Bearer test-key" {
		t.Errorf("expected Bearer test-key, got %s", headers["Authorization"])
	}

	// Claude format headers
	e.requestFormat = "claude"
	headers = e.BuildHeaders(Credentials{APIKey: "test-key"}, true, nil)
	if headers["x-api-key"] != "test-key" {
		t.Errorf("expected x-api-key test-key, got %s", headers["x-api-key"])
	}
	if headers["anthropic-version"] != "2023-06-01" {
		t.Errorf("expected anthropic-version 2023-06-01, got %s", headers["anthropic-version"])
	}
}

func TestOpencodeTransformRequest(t *testing.T) {
	registry.RegisterBuiltinProviders()
	cfg := config.Load()

	e := NewOpencodeExecutor("opencode", cfg)

	// Strip client_metadata
	body := map[string]interface{}{
		"model":           "big-pickle",
		"client_metadata": map[string]interface{}{"foo": "bar"},
	}
	result := e.TransformRequest("big-pickle", body, true, Credentials{})
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}
	if _, hasCM := resultMap["client_metadata"]; hasCM {
		t.Error("client_metadata should be stripped")
	}

	// DeepSeek V4 Pro effort level suffix
	body = map[string]interface{}{
		"model": "deepseek-v4-pro-low",
	}
	result = e.TransformRequest("deepseek-v4-pro-low", body, true, Credentials{})
	resultMap = result.(map[string]interface{})
	if resultMap["model"] != "deepseek-v4-pro" {
		t.Errorf("expected model deepseek-v4-pro, got %v", resultMap["model"])
	}
	if resultMap["reasoning_effort"] != "low" {
		t.Errorf("expected reasoning_effort low, got %v", resultMap["reasoning_effort"])
	}
}

func TestParseDeepSeekEffortLevel(t *testing.T) {
	tests := []struct {
		model    string
		expected *deepSeekEffort
	}{
		{"deepseek-v4-pro-low", &deepSeekEffort{BaseModel: "deepseek-v4-pro", Effort: "low"}},
		{"deepseek-v4-pro-medium", &deepSeekEffort{BaseModel: "deepseek-v4-pro", Effort: "medium"}},
		{"deepseek-v4-pro-high", &deepSeekEffort{BaseModel: "deepseek-v4-pro", Effort: "high"}},
		{"deepseek-v4-pro-max", &deepSeekEffort{BaseModel: "deepseek-v4-pro", Effort: "max"}},
		{"deepseek-v4-pro", nil},
		{"gpt-5.5-low", nil},
	}
	for _, test := range tests {
		result := parseDeepSeekEffortLevel(test.model)
		if test.expected == nil {
			if result != nil {
				t.Errorf("parseDeepSeekEffortLevel(%q) = %+v, want nil", test.model, result)
			}
		} else {
			if result == nil {
				t.Errorf("parseDeepSeekEffortLevel(%q) = nil, want %+v", test.model, test.expected)
			} else if result.BaseModel != test.expected.BaseModel || result.Effort != test.expected.Effort {
				t.Errorf("parseDeepSeekEffortLevel(%q) = %+v, want %+v", test.model, result, test.expected)
			}
		}
	}
}
