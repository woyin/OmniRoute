package registry

// RegisterNScale registers the nScale provider.
func RegisterNScale() {
	Register(&RegistryEntry{
		ID:     "nscale",
		Name:   "nScale",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.nscale.com/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "meta-llama/Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick (nScale)", ContextLength: 131072},
		},
	})
}
