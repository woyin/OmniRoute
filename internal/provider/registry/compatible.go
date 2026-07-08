package registry

// RegisterOpenAICompatible registers the openai-compatible-* custom provider.
func RegisterOpenAICompatible() {
	Register(&RegistryEntry{
		ID:     "openai-compatible",
		Name:   "OpenAI Compatible",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.openai.com/v1",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "custom-openai-model", Name: "Custom OpenAI Model (configure via connection)", ContextLength: 128000},
		},
	})
}

// RegisterAnthropicCompatible registers the anthropic-compatible-* custom provider.
func RegisterAnthropicCompatible() {
	Register(&RegistryEntry{
		ID:     "anthropic-compatible",
		Name:   "Anthropic Compatible",
		Format: FormatClaude,
		Executor: "default",
		BaseURL: "https://api.anthropic.com/v1",
		ChatPath: "/messages",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "x-api-key",
		DefaultContextLength: 200000,
		PassthroughModels: true,
		Headers: map[string]string{
			"anthropic-version": "2023-06-01",
		},
		Models: []RegistryModel{
			{ID: "custom-anthropic-model", Name: "Custom Anthropic Model (configure via connection)", ContextLength: 200000},
		},
	})
}

// IsOpenAICompatible returns true if the provider ID starts with "openai-compatible-".
func IsOpenAICompatible(providerID string) bool {
	return len(providerID) > len("openai-compatible-") && providerID[:len("openai-compatible-")] == "openai-compatible-"
}

// IsAnthropicCompatible returns true if the provider ID starts with "anthropic-compatible-".
func IsAnthropicCompatible(providerID string) bool {
	return len(providerID) > len("anthropic-compatible-") && providerID[:len("anthropic-compatible-")] == "anthropic-compatible-"
}

// IsClaudeCodeCompatible returns true if the provider ID starts with "anthropic-compatible-cc-".
func IsClaudeCodeCompatible(providerID string) bool {
	return len(providerID) > len("anthropic-compatible-cc-") && providerID[:len("anthropic-compatible-cc-")] == "anthropic-compatible-cc-"
}
