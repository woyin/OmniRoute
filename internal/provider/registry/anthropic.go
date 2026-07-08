package registry

// RegisterAnthropic registers the Anthropic provider.
func RegisterAnthropic() {
	Register(&RegistryEntry{
		ID:     "anthropic",
		Name:   "Anthropic",
		Format: FormatClaude,
		Executor: "default",
		BaseURL: "https://api.anthropic.com/v1",
		ChatPath: "/messages",
		ModelsURL: "https://api.anthropic.com/v1/models",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "x-api-key",
		AuthPrefix: "",
		DefaultContextLength: 200000,
		PassthroughModels: true,
		Headers: map[string]string{
			"anthropic-version": "2023-06-01",
		},
		Models: []RegistryModel{
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7", ContextLength: 200000, MaxOutputTokens: 32000, SupportsReasoning: true, SupportsVision: true},
			{ID: "claude-opus-4-5", Name: "Claude Opus 4.5", ContextLength: 200000, MaxOutputTokens: 32000, SupportsReasoning: true, SupportsVision: true},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6", ContextLength: 200000, MaxOutputTokens: 16384, SupportsReasoning: true, SupportsVision: true},
			{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", ContextLength: 200000, MaxOutputTokens: 16384, SupportsReasoning: true, SupportsVision: true},
			{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5", ContextLength: 200000, MaxOutputTokens: 8192, SupportsVision: true},
			{ID: "claude-3-5-sonnet-latest", Name: "Claude 3.5 Sonnet", ContextLength: 200000, MaxOutputTokens: 8192, SupportsVision: true},
			{ID: "claude-3-5-haiku-latest", Name: "Claude 3.5 Haiku", ContextLength: 200000, MaxOutputTokens: 8192},
			{ID: "claude-3-opus-latest", Name: "Claude 3 Opus", ContextLength: 200000, MaxOutputTokens: 4096, SupportsVision: true},
		},
	})
}
