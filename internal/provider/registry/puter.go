package registry

func RegisterPuter() {
	Register(&RegistryEntry{
		ID:     "puter",
		Name:   "Puter",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.puter.com/drivers/openai/completions",
		ChatPath: "",
		AuthType: AuthTypeNoAuth,
		DefaultContextLength: 131072,
		PassthroughModels: true,
		HasFree: true,
		Models: []RegistryModel{
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini (Puter)", ContextLength: 131072},
			{ID: "claude-3-5-sonnet", Name: "Claude 3.5 Sonnet (Puter)", ContextLength: 200000, TargetFormat: FormatClaude},
		},
	})
}
