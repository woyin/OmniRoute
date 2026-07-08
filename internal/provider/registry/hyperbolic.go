package registry

func RegisterHyperbolic() {
	Register(&RegistryEntry{
		ID:     "hyperbolic",
		Name:   "Hyperbolic",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.hyperbolic.xyz/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "meta-llama/Meta-Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "deepseek-ai/DeepSeek-R1", Name: "DeepSeek R1", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
