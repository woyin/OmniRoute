package handler

import (
	"testing"

	"github.com/omniroute/omniroute/internal/provider/registry"
)

func TestResolveProvider(t *testing.T) {
	registry.RegisterBuiltinProviders()

	tests := []struct {
		model            string
		expectedProvider string
	}{
		{"gpt-5.5", "openai"},
		{"gpt-5.4", "openai"},
		{"gpt-4.1", "openai"},
		{"claude-opus-4-7", "anthropic"},
		{"claude-sonnet-4-6", "anthropic"},
		{"deepseek-chat", "deepseek"},
		{"gemini-2.5-pro", "gemini"},
		{"grok-3", "xai"},
		{"sonar-pro", "perplexity"},
		{"openai/gpt-5.5", "openai"},
		{"o3", "openai"},
		{"o3-mini", "openai"},
		{"o4-mini", "openai"},
		{"qwen-max", "alibaba"},
		{"llama-4-maverick", "meta-llama"},
		{"mixtral-8x22b", "mistral"},
		{"codestral-latest", "mistral"},
		{"kimi-latest", "kimi"},
		{"minimax-text-01", "minimax"},
		{"glm-4-plus", "glm"},
		{"moonshot-v1", "moonshot"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			pid, resolved := resolveProvider(tt.model)
			if pid != tt.expectedProvider {
				t.Errorf("resolveProvider(%q) = provider %q, want %q (resolved model: %q)", tt.model, pid, tt.expectedProvider, resolved)
			}
		})
	}
}
