package registry

func RegisterNebius() {
	Register(&RegistryEntry{
		ID:     "nebius",
		Name:   "Nebius",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.studio.nebius.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "meta-llama/Meta-Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "Qwen/Qwen3-235B-A22B", Name: "Qwen3 235B", ContextLength: 131072},
		},
	})
}
