package registry

func RegisterOpenRouter() {
	Register(&RegistryEntry{
		ID:     "openrouter",
		Name:   "OpenRouter",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://openrouter.ai/api/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "openai/gpt-5.5", Name: "GPT-5.5 (via OpenRouter)", ContextLength: 128000},
			{ID: "anthropic/claude-opus-4-7", Name: "Claude Opus 4.7 (via OpenRouter)", ContextLength: 200000, TargetFormat: FormatClaude},
			{ID: "google/gemini-2.5-pro", Name: "Gemini 2.5 Pro (via OpenRouter)", ContextLength: 1048576},
			{ID: "deepseek/deepseek-v4-pro", Name: "DeepSeek V4 Pro (via OpenRouter)", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
