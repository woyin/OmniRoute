package registry

// RegisterOllamaCloud registers the Ollama Cloud provider.
func RegisterOllamaCloud() {
	Register(&RegistryEntry{
		ID:     "ollama-cloud",
		Alias:  "ollamacloud",
		Name:   "Ollama Cloud",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://ollama.com/v1/chat/completions",
		ModelsURL: "https://ollama.com/api/tags",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "bearer",
		DefaultContextLength: 200000,
		PassthroughModels: true,
		HasFree: true,
		Models: []RegistryModel{
			{ID: "deepseek-v4-pro", Name: "DeepSeek V4 Pro", SupportsReasoning: true},
			{ID: "deepseek-v4-flash", Name: "DeepSeek V4 Flash", SupportsReasoning: true},
			{ID: "kimi-k2.6", Name: "Kimi K2.6"},
			{ID: "glm-5.1", Name: "GLM 5.1"},
			{ID: "minimax-m3", Name: "MiniMax M3", ContextLength: 1048576, SupportsVision: true},
			{ID: "minimax-m2.7", Name: "MiniMax M2.7"},
			{ID: "gemma4:31b", Name: "Gemma 4 31B"},
			{ID: "nemotron-3-super", Name: "NVIDIA Nemotron 3 Super"},
			{ID: "qwen3.5:397b", Name: "Qwen 3.5 397B"},
		},
	})
}
