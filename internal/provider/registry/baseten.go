package registry

// RegisterBaseten registers the Baseten provider.
func RegisterBaseten() {
	Register(&RegistryEntry{
		ID:     "baseten",
		Name:   "Baseten",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://model-api.baseten.co/production",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Api-Key ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{},
	})
}
