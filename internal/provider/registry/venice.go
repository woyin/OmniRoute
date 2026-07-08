package registry

func RegisterVenice() {
	Register(&RegistryEntry{
		ID:     "venice",
		Name:   "Venice.ai",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.venice.ai/api/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "deepseek-r1-llama-70b", Name: "DeepSeek R1 70B", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
