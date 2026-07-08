package registry

// RegisterAIMLAPI registers the AI/ML API provider.
func RegisterAIMLAPI() {
	Register(&RegistryEntry{
		ID:     "aimlapi",
		Alias:  "ai-ml-api",
		Name:   "AI/ML API",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.aimlapi.com/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "gpt-4o", Name: "GPT-4o (AI/ML)", ContextLength: 128000},
			{ID: "claude-3-5-sonnet-latest", Name: "Claude 3.5 Sonnet (AI/ML)", ContextLength: 200000},
		},
	})
}
