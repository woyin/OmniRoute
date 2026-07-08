package registry

// RegisterCodex registers the Codex (OpenAI) OAuth provider.
func RegisterCodex() {
	Register(&RegistryEntry{
		ID:     "codex",
		Alias:  "cx",
		Name:   "OpenAI Codex",
		Format: FormatOpenAIResponses,
		Executor: "codex",
		BaseURL: "https://chatgpt.com/backend-api/codex/responses",
		AuthType: AuthTypeOAuth,
		AuthHeader: "bearer",
		DefaultContextLength: 400000,
		OAuth: &OAuthConfig{
			ClientIDEnv:       "CODEX_OAUTH_CLIENT_ID",
			ClientIDDefault:   "",
			ClientSecretEnv:   "CODEX_OAUTH_CLIENT_SECRET",
			ClientSecretDefault: "",
			TokenURL:          "https://auth.openai.com/oauth/token",
		},
		Models: []RegistryModel{
			{ID: "gpt-5.5", Name: "GPT 5.5", ContextLength: 400000, MaxInputTokens: 272000, MaxOutputTokens: 128000, SupportsReasoning: true, SupportsXHighEffort: true},
			{ID: "gpt-5.5-xhigh", Name: "GPT 5.5 (xHigh)", ContextLength: 400000, MaxInputTokens: 272000, MaxOutputTokens: 128000, SupportsReasoning: true, SupportsXHighEffort: true},
			{ID: "gpt-5.5-high", Name: "GPT 5.5 (High)", ContextLength: 400000, MaxInputTokens: 272000, MaxOutputTokens: 128000, SupportsReasoning: true},
			{ID: "gpt-5.5-medium", Name: "GPT 5.5 (Medium)", ContextLength: 400000, MaxInputTokens: 272000, MaxOutputTokens: 128000, SupportsReasoning: true},
			{ID: "gpt-5.5-low", Name: "GPT 5.5 (Low)", ContextLength: 400000, MaxInputTokens: 272000, MaxOutputTokens: 128000, SupportsReasoning: true},
			{ID: "gpt-5.4", Name: "GPT 5.4", SupportsReasoning: true},
			{ID: "gpt-5.3-codex", Name: "GPT 5.3 Codex", TargetFormat: FormatOpenAIResponses, SupportsReasoning: true, SupportsXHighEffort: true},
			{ID: "gpt-5.4-mini", Name: "GPT 5.4 Mini", TargetFormat: FormatOpenAIResponses},
		},
	})
}
