package registry

func RegisterSambaNova() {
	Register(&RegistryEntry{
		ID:     "sambanova",
		Name:   "SambaNova",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.sambanova.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "Meta-Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "DeepSeek-R1", Name: "DeepSeek R1", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
