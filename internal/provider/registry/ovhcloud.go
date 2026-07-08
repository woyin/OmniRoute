package registry

// RegisterOVHcloud registers the OVHcloud AI provider.
func RegisterOVHcloud() {
	Register(&RegistryEntry{
		ID:     "ovhcloud",
		Name:   "OVHcloud AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://gra.ai.cloud.ovh.net/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{},
	})
}
