package registry

// RegisterMistral registers the Mistral provider.
func RegisterMistral() {
	Register(&RegistryEntry{
		ID:     "mistral",
		Name:   "Mistral",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.mistral.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "mistral-large-latest", Name: "Mistral Large", ContextLength: 128000},
			{ID: "mistral-medium-latest", Name: "Mistral Medium", ContextLength: 32000},
			{ID: "mistral-small-latest", Name: "Mistral Small", ContextLength: 32000},
			{ID: "codestral-latest", Name: "Codestral", ContextLength: 32000},
			{ID: "open-mistral-nemo", Name: "Mistral Nemo", ContextLength: 128000},
			{ID: "mixtral-8x22b", Name: "Mixtral 8x22B", ContextLength: 65536},
			{ID: "mixtral-8x7b", Name: "Mixtral 8x7B", ContextLength: 32768},
			{ID: "mistral-embed", Name: "Mistral Embed", ContextLength: 8192},
		},
	})
}
