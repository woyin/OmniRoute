package registry

func RegisterNVIDIA() {
	Register(&RegistryEntry{
		ID:     "nvidia",
		Name:   "NVIDIA NIM",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://integrate.api.nvidia.com/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "meta/llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "deepseek-ai/deepseek-r1", Name: "DeepSeek R1", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
