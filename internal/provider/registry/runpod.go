package registry

// RegisterRunPod registers the RunPod provider.
func RegisterRunPod() {
	Register(&RegistryEntry{
		ID:     "runpod",
		Name:   "RunPod Serverless",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.runpod.ai/v2",
		ChatPath: "/openai/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{},
	})
}
