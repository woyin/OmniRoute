package registry

func RegisterPollinations() {
	Register(&RegistryEntry{
		ID:     "pollinations",
		Name:   "Pollinations",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://text.pollinations.ai/openai",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeNoAuth,
		DefaultContextLength: 131072,
		PassthroughModels: true,
		HasFree: true,
		Models: []RegistryModel{
			{ID: "openai", Name: "OpenAI (via Pollinations)", ContextLength: 131072},
			{ID: "mistral", Name: "Mistral (via Pollinations)", ContextLength: 131072},
			{ID: "llama", Name: "Llama (via Pollinations)", ContextLength: 131072},
			{ID: "deepseek", Name: "DeepSeek (via Pollinations)", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
