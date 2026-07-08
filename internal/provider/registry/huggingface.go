package registry

func RegisterHuggingFace() {
	Register(&RegistryEntry{
		ID:     "huggingface",
		Name:   "HuggingFace Inference",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api-inference.huggingface.co/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "meta-llama/Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "Qwen/Qwen3-235B-A22B", Name: "Qwen3 235B", ContextLength: 131072},
		},
	})
}
