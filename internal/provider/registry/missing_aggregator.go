package registry

// RegisterAggregatorProviders registers aggregator provider entries.
func RegisterAggregatorProviders() {
	Register(&RegistryEntry{
		ID:       "api-airforce",
		Name:     "Grok-3 (Free)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.airforce/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "x-ai/grok-3", Name: "Grok-3 (Free)"},
			{ID: "x-ai/grok-2-1212", Name: "Grok-2 1212 (Free)"},
			{ID: "anthropic/claude-3.7-sonnet", Name: "Claude 3.7 Sonnet (Free)"},
			{ID: "qwen/qwen3-32b", Name: "Qwen3 32B (Free)"},
			{ID: "moonshot/kimi-k2.6", Name: "Kimi K2.6 (Free)"},
			{ID: "google/gemini-2.5-flash", Name: "Gemini 2.5 Flash (Free)"},
			{ID: "deepseek/deepseek-v3", Name: "DeepSeek V3 (Free)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "bazaarlink",
		Name:     "Auto Free (Zero Cost)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://bazaarlink.ai/api/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "auto:free", Name: "Auto Free (Zero Cost)"},
			{ID: "claude-opus-4.7", Name: "Claude Opus 4.7"},
			{ID: "claude-sonnet-4.6", Name: "Claude Sonnet 4.6"},
			{ID: "claude-haiku-4.5", Name: "Claude Haiku 4.5"},
			{ID: "gpt-5.5", Name: "GPT-5.5"},
			{ID: "gpt-5.4", Name: "GPT-5.4"},
			{ID: "gpt-5.4-mini", Name: "GPT-5.4 Mini"},
			{ID: "gpt-5.4-nano", Name: "GPT-5.4 Nano"},
			{ID: "grok-4.3", Name: "Grok 4.3"},
			{ID: "grok-4.20", Name: "Grok 4.20"},
		},
	})

	Register(&RegistryEntry{
		ID:       "bluesminds",
		Name:     "GPT-4o",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.bluesminds.com/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "gpt-4o", Name: "GPT-4o"},
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini"},
			{ID: "gpt-4.1", Name: "GPT-4.1"},
			{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini"},
			{ID: "gpt-4.1-nano", Name: "GPT-4.1 Nano"},
			{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5"},
			{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5"},
			{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash"},
			{ID: "gemini-2.0-flash-exp", Name: "Gemini 2.0 Flash (Exp)"},
			{ID: "deepseek-reasoner", Name: "DeepSeek Reasoner"},
		},
	})

	Register(&RegistryEntry{
		ID:       "freeaiapikey",
		Name:     "GPT-5 (via FreeAIAPIKey)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://freeaiapikey.com/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "openai/gpt-5", Name: "GPT-5 (via FreeAIAPIKey)"},
			{ID: "openai/gpt-4o", Name: "GPT-4o (via FreeAIAPIKey)"},
			{ID: "openai/gpt-5.2-codex", Name: "GPT-5.2 Codex (via FreeAIAPIKey)"},
			{ID: "anthropic/claude-opus-4.6", Name: "Claude Opus 4.6 (via FreeAIAPIKey)"},
			{ID: "anthropic/claude-sonnet-4.6", Name: "Claude Sonnet 4.6 (via FreeAIAPIKey)"},
			{ID: "Alibaba/qwen3.5", Name: "Qwen 3.5 (via FreeAIAPIKey)"},
			{ID: "Alibaba/qwen3-vl:235b", Name: "Qwen 3 VL 235B (via FreeAIAPIKey)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "freemodel-dev",
		Name:     "GPT-5.5",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.freemodel.dev/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "gpt-5.5", Name: "GPT-5.5"},
			{ID: "gpt-5.4", Name: "GPT-5.4"},
			{ID: "gpt-5.4-mini", Name: "GPT-5.4 Mini"},
			{ID: "gpt-5.3-codex", Name: "GPT-5.3 Codex"},
		},
	})

	Register(&RegistryEntry{
		ID:       "glhf",
		Name:     "DeepSeek 7B Chat",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.laf.run/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "deepseek-7b-chat", Name: "DeepSeek 7B Chat"},
		},
	})

}