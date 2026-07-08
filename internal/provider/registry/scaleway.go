package registry

// RegisterScaleway registers the Scaleway provider.
func RegisterScaleway() {
	Register(&RegistryEntry{
		ID:     "scaleway",
		Name:   "Scaleway",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.scaleway.com/ generative-ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "llama-3.3-70b-instruct", Name: "Llama 3.3 70B (Scaleway)", ContextLength: 131072},
			{ID: "mistral-nemo-instruct-2407", Name: "Mistral Nemo (Scaleway)", ContextLength: 131072},
		},
	})
}
