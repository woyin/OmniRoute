package registry

// RegisterCommandCode registers the Command Code provider.
func RegisterCommandCode() {
	Register(&RegistryEntry{
		ID:     "command-code",
		Alias:  "cmd",
		Name:   "Command Code",
		Format: FormatOpenAI,
		Executor: "command-code",
		BaseURL: "https://api.commandcode.ai",
		ChatPath: "/alpha/generate",
		ModelsURL: "https://api.commandcode.ai/provider/v1/models",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 200000,
		Models: []RegistryModel{
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7 (CC)", SupportsReasoning: true, ContextLength: 200000, MaxOutputTokens: 32000},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (CC)", SupportsReasoning: true, ContextLength: 200000, MaxOutputTokens: 16384},
			{ID: "gpt-5.5", Name: "GPT-5.5 (CC)", SupportsReasoning: true, ContextLength: 256000, MaxOutputTokens: 128000},
			{ID: "gpt-5.4", Name: "GPT-5.4 (CC)", SupportsReasoning: true, ContextLength: 256000, MaxOutputTokens: 128000},
			{ID: "gpt-5.3-codex", Name: "GPT-5.3 Codex (CC)", SupportsReasoning: true, ContextLength: 256000, MaxOutputTokens: 128000},
			{ID: "deepseek/deepseek-v4-pro", Name: "DeepSeek V4 Pro (CC)", SupportsReasoning: true, ContextLength: 1000000, MaxOutputTokens: 131072},
			{ID: "deepseek/deepseek-v4-flash", Name: "DeepSeek V4 Flash (CC)", SupportsReasoning: true, ContextLength: 1000000, MaxOutputTokens: 131072},
			{ID: "moonshotai/Kimi-K2.6", Name: "Kimi K2.6 (CC)", SupportsReasoning: true, ContextLength: 262144, MaxOutputTokens: 65536},
			{ID: "MiniMaxAI/MiniMax-M2.7", Name: "MiniMax M2.7 (CC)", SupportsReasoning: true, ContextLength: 1048576, MaxOutputTokens: 65536},
			{ID: "Qwen/Qwen3.6-Plus", Name: "Qwen 3.6 Plus (CC)", SupportsReasoning: true, ContextLength: 1000000, MaxOutputTokens: 32768},
		},
	})
}
