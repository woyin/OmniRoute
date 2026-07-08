package registry

// RegisterLambdaAI registers the Lambda AI provider.
func RegisterLambdaAI() {
	Register(&RegistryEntry{
		ID:     "lambda",
		Alias:  "lambda-ai",
		Name:   "Lambda AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.lambdalabs.com/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick (Lambda)", ContextLength: 131072},
			{ID: "deepseek-r1", Name: "DeepSeek R1 (Lambda)", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
