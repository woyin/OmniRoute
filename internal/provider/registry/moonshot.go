package registry

// RegisterMoonshot registers the Moonshot AI provider.
func RegisterMoonshot() {
	Register(&RegistryEntry{
		ID:     "moonshot",
		Name:   "Moonshot AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.moonshot.cn/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "kimi-k2.6", Name: "Kimi K2.6", ContextLength: 262144, SupportsReasoning: true},
			{ID: "moonshot-v1-128k", Name: "Moonshot V1 128K", ContextLength: 131072},
			{ID: "moonshot-v1-32k", Name: "Moonshot V1 32K", ContextLength: 32768},
			{ID: "moonshot-v1-8k", Name: "Moonshot V1 8K", ContextLength: 8192},
			{ID: "moonshot-v1", Name: "Moonshot V1", ContextLength: 8192},
		},
	})
}
