package registry

// RegisterDeepSeek registers the DeepSeek provider.
func RegisterDeepSeek() {
	Register(&RegistryEntry{
		ID:     "deepseek",
		Name:   "DeepSeek",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.deepseek.com/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "deepseek-chat", Name: "DeepSeek Chat (V3)", ContextLength: 131072, MaxOutputTokens: 8192},
			{ID: "deepseek-reasoner", Name: "DeepSeek Reasoner (R2)", ContextLength: 131072, MaxOutputTokens: 8192, SupportsReasoning: true},
			{ID: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", ContextLength: 1000000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "deepseek-v4-flash", Name: "DeepSeek V4 Flash", ContextLength: 1000000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "deepseek-coder-v2", Name: "DeepSeek Coder V2", ContextLength: 131072, MaxOutputTokens: 8192},
			{ID: "deepseek-v3", Name: "DeepSeek V3", ContextLength: 131072, MaxOutputTokens: 8192},
		},
	})
}
