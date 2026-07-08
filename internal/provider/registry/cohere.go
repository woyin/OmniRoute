package registry

func RegisterCohere() {
	Register(&RegistryEntry{
		ID:     "cohere",
		Name:   "Cohere",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.cohere.ai/v2",
		ChatPath: "/chat",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		Models: []RegistryModel{
			{ID: "command-r-plus", Name: "Command R+", ContextLength: 128000},
			{ID: "command-r", Name: "Command R", ContextLength: 128000},
			{ID: "command-a-03-2025", Name: "Command A", ContextLength: 128000},
		},
	})
}
