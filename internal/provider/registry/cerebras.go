package registry

func RegisterCerebras() {
	Register(&RegistryEntry{
		ID:     "cerebras",
		Name:   "Cerebras",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.cerebras.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "llama-4-scout-17b-16e-instruct", Name: "Llama 4 Scout 17B", ContextLength: 131072},
			{ID: "llama3.1-8b", Name: "Llama 3.1 8B", ContextLength: 131072},
		},
	})
}
