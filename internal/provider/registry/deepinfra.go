package registry

func RegisterDeepInfra() {
	Register(&RegistryEntry{
		ID:     "deepinfra",
		Name:   "DeepInfra",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.deepinfra.com/v1/openai",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "meta-llama/Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "deepseek-ai/DeepSeek-R1", Name: "DeepSeek R1", ContextLength: 131072, SupportsReasoning: true},
			{ID: "Qwen/Qwen3-235B-A22B", Name: "Qwen3 235B", ContextLength: 131072},
		},
	})
}
