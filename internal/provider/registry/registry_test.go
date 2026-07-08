package registry

import (
	"testing"
)

func TestRegisterAndGet(t *testing.T) {
	RegisterBuiltinProviders()

	// Test OpenCode
	entry := Get("opencode")
	if entry == nil {
		t.Fatal("opencode provider not found")
	}
	if entry.ID != "opencode" {
		t.Errorf("expected ID opencode, got %s", entry.ID)
	}
	if entry.Format != FormatOpenAI {
		t.Errorf("expected Format openai, got %s", entry.Format)
	}
	if entry.Executor != "opencode" {
		t.Errorf("expected Executor opencode, got %s", entry.Executor)
	}

	// Test alias
	aliasEntry := Get("oc")
	if aliasEntry == nil {
		t.Fatal("opencode alias 'oc' not found")
	}
	if aliasEntry.ID != "opencode" {
		t.Errorf("expected alias to resolve to opencode, got %s", aliasEntry.ID)
	}
}

func TestOpenCodeGo(t *testing.T) {
	RegisterBuiltinProviders()

	entry := Get("opencode-go")
	if entry == nil {
		t.Fatal("opencode-go provider not found")
	}
	if entry.AuthType != AuthTypeAPIKey {
		t.Errorf("expected AuthType apikey, got %s", entry.AuthType)
	}
}

func TestOllamaCloud(t *testing.T) {
	RegisterBuiltinProviders()

	entry := Get("ollama-cloud")
	if entry == nil {
		t.Fatal("ollama-cloud provider not found")
	}
	if entry.Executor != "default" {
		t.Errorf("expected Executor default, got %s", entry.Executor)
	}
	if entry.HasFree != true {
		t.Error("expected HasFree=true")
	}
}

func TestCodex(t *testing.T) {
	RegisterBuiltinProviders()

	entry := Get("codex")
	if entry == nil {
		t.Fatal("codex provider not found")
	}
	if entry.Format != FormatOpenAIResponses {
		t.Errorf("expected Format openai-responses, got %s", entry.Format)
	}
	if entry.AuthType != AuthTypeOAuth {
		t.Errorf("expected AuthType oauth, got %s", entry.AuthType)
	}
	if entry.OAuth == nil {
		t.Fatal("expected OAuth config")
	}
	if entry.OAuth.TokenURL != "https://auth.openai.com/oauth/token" {
		t.Errorf("unexpected token URL: %s", entry.OAuth.TokenURL)
	}
}

func TestCommandCode(t *testing.T) {
	RegisterBuiltinProviders()

	entry := Get("command-code")
	if entry == nil {
		t.Fatal("command-code provider not found")
	}
	if entry.ChatPath != "/alpha/generate" {
		t.Errorf("expected ChatPath /alpha/generate, got %s", entry.ChatPath)
	}
}

func TestOpenAICompatible(t *testing.T) {
	RegisterBuiltinProviders()

	if !IsOpenAICompatible("openai-compatible-mycustom") {
		t.Error("expected openai-compatible-mycustom to be detected as OpenAI compatible")
	}
	if IsOpenAICompatible("openai") {
		t.Error("openai should not be detected as OpenAI compatible")
	}
}

func TestAnthropicCompatible(t *testing.T) {
	RegisterBuiltinProviders()

	if !IsAnthropicCompatible("anthropic-compatible-mycustom") {
		t.Error("expected anthropic-compatible-mycustom to be detected as Anthropic compatible")
	}
	if !IsClaudeCodeCompatible("anthropic-compatible-cc-mycc") {
		t.Error("expected anthropic-compatible-cc-mycc to be Claude Code compatible")
	}
	if IsAnthropicCompatible("anthropic") {
		t.Error("anthropic should not be detected as Anthropic compatible")
	}
}

func TestGetModel(t *testing.T) {
	RegisterBuiltinProviders()

	entry := Get("codex")
	if entry == nil {
		t.Fatal("codex provider not found")
	}

	model := entry.GetModel("gpt-5.5")
	if model == nil {
		t.Fatal("gpt-5.5 model not found in codex")
	}
	if model.ContextLength != 400000 {
		t.Errorf("expected ContextLength 400000, got %d", model.ContextLength)
	}
	if !model.SupportsReasoning {
		t.Error("expected SupportsReasoning=true")
	}

	missing := entry.GetModel("nonexistent-model")
	if missing != nil {
		t.Error("expected nil for nonexistent model")
	}
}

func TestList(t *testing.T) {
	RegisterBuiltinProviders()

	entries := List()
	if len(entries) < 7 {
		t.Errorf("expected at least 7 providers, got %d", len(entries))
	}

	// Verify no duplicates
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry.ID] {
			t.Errorf("duplicate provider ID: %s", entry.ID)
		}
		seen[entry.ID] = true
	}
}
