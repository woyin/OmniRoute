package translator

import (
	"testing"
)

func TestNeedsTranslation(t *testing.T) {
	tests := []struct {
		source, target string
		expected       bool
	}{
		{"openai", "openai", false},
		{"openai", "claude", true},
		{"claude", "openai", true},
		{"", "openai", false},
		{"openai", "", false},
		{"openai", "openai-responses", true},
	}
	for _, test := range tests {
		result := NeedsTranslation(test.source, test.target)
		if result != test.expected {
			t.Errorf("NeedsTranslation(%q, %q) = %v, want %v", test.source, test.target, result, test.expected)
		}
	}
}

func TestOpenAIToClaude(t *testing.T) {
	body := map[string]interface{}{
		"model": "claude-sonnet-4-6",
		"messages": []interface{}{
			map[string]interface{}{"role": "system", "content": "You are helpful."},
			map[string]interface{}{"role": "user", "content": "Hello"},
		},
		"max_tokens": float64(1024),
		"stream":     true,
		"tools": []interface{}{
			map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        "get_weather",
					"description": "Get weather",
					"parameters":  map[string]interface{}{"type": "object"},
				},
			},
		},
	}

	result := TranslateRequest(body, FormatOpenAI, FormatClaude)

	// Check that system messages are extracted
	if _, hasSystem := result["system"]; !hasSystem {
		t.Error("expected system key in Claude format")
	}

	// Check that model is preserved
	if result["model"] != "claude-sonnet-4-6" {
		t.Errorf("expected model claude-sonnet-4-6, got %v", result["model"])
	}

	// Check that tools are translated
	if tools, ok := result["tools"].([]interface{}); ok {
		if len(tools) != 1 {
			t.Errorf("expected 1 tool, got %d", len(tools))
		}
	} else {
		t.Error("expected tools array")
	}

	// Check anthropic-version
	if result["anthropic_version"] != "2023-06-01" {
		t.Errorf("expected anthropic_version 2023-06-01, got %v", result["anthropic_version"])
	}
}

func TestClaudeToOpenAI(t *testing.T) {
	body := map[string]interface{}{
		"model":      "claude-sonnet-4-6",
		"max_tokens": float64(1024),
		"stream":     true,
		"system":     "You are helpful.",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Hello"},
		},
	}

	result := TranslateRequest(body, FormatClaude, FormatOpenAI)

	// Check that system is converted to messages
	messages, ok := result["messages"].([]interface{})
	if !ok {
		t.Fatal("expected messages array")
	}
	if len(messages) < 2 {
		t.Errorf("expected at least 2 messages (system + user), got %d", len(messages))
	}

	// First message should be system
	firstMsg, ok := messages[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected first message to be a map")
	}
	if firstMsg["role"] != "system" {
		t.Errorf("expected first message role=system, got %v", firstMsg["role"])
	}
}

func TestIdentityTranslation(t *testing.T) {
	body := map[string]interface{}{
		"model":    "gpt-5.5",
		"messages": []interface{}{},
	}

	result := TranslateRequest(body, FormatOpenAI, FormatOpenAI)
	if result["model"] != "gpt-5.5" {
		t.Errorf("expected model unchanged, got %v", result["model"])
	}
}
