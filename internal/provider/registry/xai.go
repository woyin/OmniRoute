package registry

func RegisterXAI() {
	Register(&RegistryEntry{
		ID:     "xai",
		Name:   "xAI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.x.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "grok-4", Name: "Grok 4", ContextLength: 262144, SupportsReasoning: true},
			{ID: "grok-3", Name: "Grok 3", ContextLength: 131072, SupportsReasoning: true},
			{ID: "grok-3-mini", Name: "Grok 3 Mini", ContextLength: 131072, SupportsReasoning: true},
			{ID: "grok-3-fast", Name: "Grok 3 Fast", ContextLength: 131072},
		},
	})
}
