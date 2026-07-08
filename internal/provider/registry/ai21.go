package registry

func RegisterAI21() {
	Register(&RegistryEntry{
		ID:     "ai21",
		Name:   "AI21 Labs",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.ai21.com/studio/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "jamba-1.6-large", Name: "Jamba 1.6 Large", ContextLength: 262144},
			{ID: "jamba-1.6-mini", Name: "Jamba 1.6 Mini", ContextLength: 262144},
		},
	})
}
