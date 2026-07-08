package registry

// RegisterAzureOpenAI registers the Azure OpenAI provider.
func RegisterAzureOpenAI() {
	Register(&RegistryEntry{
		ID:     "azure-openai",
		Alias:  "azure",
		Name:   "Azure OpenAI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://{resource}.openai.azure.com/openai",
		ChatPath: "/deployments/{deployment}/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "api-key",
		AuthPrefix: "",
		DefaultContextLength: 128000,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "gpt-5.5", Name: "GPT-5.5 (Azure)", ContextLength: 400000, SupportsReasoning: true},
			{ID: "gpt-4o", Name: "GPT-4o (Azure)", ContextLength: 128000, SupportsVision: true},
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini (Azure)", ContextLength: 128000, SupportsVision: true},
			{ID: "gpt-4", Name: "GPT-4 (Azure)", ContextLength: 8192},
		},
	})
}
