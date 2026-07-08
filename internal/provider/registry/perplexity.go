package registry

func RegisterPerplexity() {
	Register(&RegistryEntry{
		ID:     "perplexity",
		Name:   "Perplexity",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.perplexity.ai",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "sonar-pro", Name: "Sonar Pro", ContextLength: 200000},
			{ID: "sonar", Name: "Sonar", ContextLength: 131072},
			{ID: "sonar-reasoning-pro", Name: "Sonar Reasoning Pro", ContextLength: 131072, SupportsReasoning: true},
			{ID: "sonar-reasoning", Name: "Sonar Reasoning", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
