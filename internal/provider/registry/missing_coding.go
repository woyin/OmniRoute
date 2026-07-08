package registry

// RegisterCodingProviders registers coding provider entries.
func RegisterCodingProviders() {
	Register(&RegistryEntry{
		ID:       "clinepass",
		Name:     "GLM-5.2 (ClinePass)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.cline.bot/api/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "cline-pass/glm-5.2", Name: "GLM-5.2 (ClinePass)"},
			{ID: "cline-pass/kimi-k2.7-code", Name: "Kimi K2.7 Code (ClinePass)"},
			{ID: "cline-pass/kimi-k2.6", Name: "Kimi K2.6 (ClinePass)"},
			{ID: "cline-pass/deepseek-v4-pro", Name: "DeepSeek V4 Pro (ClinePass)"},
			{ID: "cline-pass/deepseek-v4-flash", Name: "DeepSeek V4 Flash (ClinePass)"},
			{ID: "cline-pass/mimo-v2.5", Name: "MiMo-V2.5 (ClinePass)"},
			{ID: "cline-pass/mimo-v2.5-pro", Name: "MiMo-V2.5-Pro (ClinePass)"},
			{ID: "cline-pass/minimax-m3", Name: "MiniMax M3 (ClinePass)"},
			{ID: "cline-pass/qwen3.7-max", Name: "Qwen3.7 Max (ClinePass)"},
			{ID: "cline-pass/qwen3.7-plus", Name: "Qwen3.7 Plus (ClinePass)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "dit",
		Name:     "GPT-5.4 (DIT.ai)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.dit.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "gpt-5.4", Name: "GPT-5.4 (DIT.ai)"},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (DIT.ai)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "factory",
		Name:     "Factory Auto (best model)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.factory.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "auto", Name: "Factory Auto (best model)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "kie",
		Name:     "Claude 4.8 Opus",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.kie.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "claude-opus-4-8", Name: "Claude 4.8 Opus"},
			{ID: "claude-opus-4-7", Name: "Claude 4.7 Opus"},
			{ID: "claude-sonnet-4-6", Name: "Claude 4.6 Sonnet"},
			{ID: "claude-haiku-4-5", Name: "Claude 4.5 Haiku"},
			{ID: "gpt-5-5", Name: "GPT 5.5"},
			{ID: "gpt-5-4", Name: "GPT 5.4"},
			{ID: "gpt-5-2", Name: "GPT 5.2"},
			{ID: "gemini-3-1-pro", Name: "Gemini 3.1 Pro"},
			{ID: "gemini-2-5-pro", Name: "Gemini 2.5 Pro"},
			{ID: "gemini-3-flash", Name: "Gemini 3 Flash"},
		},
	})

	Register(&RegistryEntry{
		ID:       "pioneer",
		Name:     "Qwen3 32B",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.pioneer.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "Qwen/Qwen3-32B", Name: "Qwen3 32B"},
			{ID: "Qwen/Qwen3.6-27B", Name: "Qwen3.6 27B"},
			{ID: "Qwen/Qwen3.5-9B", Name: "Qwen3.5 9B"},
			{ID: "Qwen/Qwen3-8B", Name: "Qwen3 8B"},
			{ID: "Qwen/Qwen3-4B-Base", Name: "Qwen3 4B Base"},
			{ID: "Qwen/Qwen3-1.7B-Base", Name: "Qwen3 1.7B Base"},
			{ID: "meta-llama/Llama-3.1-8B-Instruct", Name: "Llama 3.1 8B Instruct"},
			{ID: "meta-llama/Llama-3.2-1B-Instruct", Name: "Llama 3.2 1B Instruct"},
			{ID: "google/gemma-3-4b-pt", Name: "Gemma 3 4B (Pretrained)"},
			{ID: "HuggingFaceTB/SmolLM3-3B-Base", Name: "SmolLM3 3B Base"},
		},
	})

}