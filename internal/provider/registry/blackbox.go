package registry

// RegisterBlackbox registers the Blackbox AI provider.
func RegisterBlackbox() {
	Register(&RegistryEntry{
		ID:     "blackbox",
		Name:   "Blackbox AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://www.blackbox.ai",
		ChatPath: "/api/chat",
		AuthType: AuthTypeNoAuth,
		AuthHeader: "",
		AuthPrefix: "",
		DefaultContextLength: 32768,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "blackboxai", Name: "Blackbox AI", ContextLength: 32768},
		},
	})
}
