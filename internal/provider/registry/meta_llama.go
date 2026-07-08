package registry

// RegisterMetaLlama registers the Meta Llama provider.
func RegisterMetaLlama() {
	Register(&RegistryEntry{
		ID:     "meta-llama",
		Alias:  "llama",
		Name:   "Meta Llama API",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.llama.com/compat/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "llama-4-scout-17b-16e-instruct", Name: "Llama 4 Scout", ContextLength: 131072},
			{ID: "llama-3.3-70b-instruct", Name: "Llama 3.3 70B", ContextLength: 131072},
			{ID: "llama-3.1-405b-instruct", Name: "Llama 3.1 405B", ContextLength: 131072},
			{ID: "llama-3.1-70b-instruct", Name: "Llama 3.1 70B", ContextLength: 131072},
			{ID: "llama-3.1-8b-instruct", Name: "Llama 3.1 8B", ContextLength: 131072},
		},
	})
}
