package registry

func RegisterTogether() {
	Register(&RegistryEntry{
		ID:     "together",
		Name:   "Together AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.together.xyz/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "deepseek-r1", Name: "DeepSeek R1", ContextLength: 131072, SupportsReasoning: true},
			{ID: "meta-llama/Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "Qwen/Qwen3-235B-A22B", Name: "Qwen3 235B", ContextLength: 131072},
		},
	})
}
