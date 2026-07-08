package registry

func RegisterFireworks() {
	Register(&RegistryEntry{
		ID:     "fireworks",
		Name:   "Fireworks AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.fireworks.ai/inference/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "accounts/fireworks/models/llama4-maverick-instruct-basic", Name: "Llama 4 Maverick", ContextLength: 131072},
			{ID: "accounts/fireworks/models/qwen3-235b-a22b", Name: "Qwen3 235B", ContextLength: 131072},
		},
	})
}
