package registry

func RegisterKimi() {
	Register(&RegistryEntry{
		ID:     "kimi",
		Name:   "Kimi (Moonshot)",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.kimi.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 262144,
		Models: []RegistryModel{
			{ID: "kimi-k2.6", Name: "Kimi K2.6", ContextLength: 262144, SupportsReasoning: true},
			{ID: "kimi-latest", Name: "Kimi Latest", ContextLength: 131072},
		},
	})
}
