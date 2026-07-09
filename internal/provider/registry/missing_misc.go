package registry

// RegisterMiscProviders registers misc provider entries.
func RegisterMiscProviders() {
	Register(&RegistryEntry{
		ID:       "bai",
		Name:     "bai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.b.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "charm-hyper",
		Name:     "Charm Hyper Auto",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://hyper.charm.land/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "hyper/auto", Name: "Charm Hyper Auto"},
		},
	})

	Register(&RegistryEntry{
		ID:       "featherless-ai",
		Name:     "featherless-ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.featherless.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "gitlawb",
		Name:     "gitlawb",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://opengateway.gitlawb.com/v1/xiaomi-mimo",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:                "gitlawb-gmi",
		Alias:             "glb-gmi",
		Format:            FormatOpenAI,
		Executor:          "default",
		BaseURL:           "https://opengateway.gitlawb.com/v1/gmi-cloud",
		AuthType:          AuthTypeAPIKey,
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		Headers: map[string]string{
			"User-Agent":   "OpenClaude/1.0 (linux; x86_64)",
			"X-Title":      "OpenClaude CLI",
			"HTTP-Referer": "https://github.com/Gitlawb/openclaude",
		},
		Models: []RegistryModel{
			{ID: "XiaomiMiMo/MiMo-V2.5-Pro", Name: "MiMo-V2.5-Pro (GMI)", ContextLength: 1050000, MaxOutputTokens: 131072},
			{ID: "XiaomiMiMo/MiMo-V2.5", Name: "MiMo-V2.5 (GMI)", ContextLength: 1050000, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.5", Name: "GPT-5.5", ContextLength: 1050000, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.4-pro", Name: "GPT-5.4 Pro", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.4", Name: "GPT-5.4", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.4-mini", Name: "GPT-5.4 Mini", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.4-nano", Name: "GPT-5.4 Nano", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.3-codex", Name: "GPT-5.3 Codex", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.2-codex", Name: "GPT-5.2 Codex", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.2", Name: "GPT-5.2", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5.1", Name: "GPT-5.1", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-5", Name: "GPT-5", ContextLength: 409600, MaxOutputTokens: 131072},
			{ID: "openai/gpt-4o", Name: "GPT-4o", ContextLength: 131072, MaxOutputTokens: 16384},
			{ID: "openai/gpt-4o-mini", Name: "GPT-4o Mini", ContextLength: 131072, MaxOutputTokens: 16384},
			{ID: "anthropic/claude-opus-4.7", Name: "Claude Opus 4.7", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "anthropic/claude-opus-4.6", Name: "Claude Opus 4.6", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "anthropic/claude-opus-4.5", Name: "Claude Opus 4.5", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "anthropic/claude-opus-4.1", Name: "Claude Opus 4.1", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "anthropic/claude-sonnet-4.6", Name: "Claude Sonnet 4.6", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "anthropic/claude-sonnet-4.5", Name: "Claude Sonnet 4.5", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "anthropic/claude-sonnet-4", Name: "Claude Sonnet 4", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "anthropic/claude-haiku-4.5", Name: "Claude Haiku 4.5", ContextLength: 409600, MaxOutputTokens: 131072, TargetFormat: FormatClaude},
			{ID: "deepseek-ai/DeepSeek-V4-Pro", Name: "DeepSeek V4 Pro", ContextLength: 1048576, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "deepseek-ai/DeepSeek-V4-Flash", Name: "DeepSeek V4 Flash", ContextLength: 1048575, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "deepseek-ai/DeepSeek-R1-0528", Name: "DeepSeek R1", ContextLength: 163840, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "deepseek-ai/DeepSeek-V3.2", Name: "DeepSeek V3.2", ContextLength: 163840, MaxOutputTokens: 131072},
			{ID: "google/gemini-3.1-pro-preview", Name: "Gemini 3.1 Pro", ContextLength: 1048576, MaxOutputTokens: 131072},
			{ID: "google/gemini-3.1-flash-lite-preview", Name: "Gemini 3.1 Flash Lite", ContextLength: 1048576, MaxOutputTokens: 131072},
			{ID: "google/gemini-3-flash-preview", Name: "Gemini 3 Flash", ContextLength: 1048576, MaxOutputTokens: 131072},
			{ID: "zai-org/GLM-5.1-FP8", Name: "GLM-5.1", ContextLength: 202752, MaxOutputTokens: 131072},
			{ID: "zai-org/GLM-5-FP8", Name: "GLM-5", ContextLength: 202752, MaxOutputTokens: 131072},
			{ID: "moonshotai/Kimi-K2.6", Name: "Kimi K2.6", ContextLength: 65536, MaxOutputTokens: 131072},
			{ID: "moonshotai/Kimi-K2.5", Name: "Kimi K2.5", ContextLength: 262144, MaxOutputTokens: 131072},
			{ID: "MiniMaxAI/MiniMax-M2.7", Name: "MiniMax M2.7", ContextLength: 196608, MaxOutputTokens: 131072},
			{ID: "MiniMaxAI/MiniMax-M2.5", Name: "MiniMax M2.5", ContextLength: 196608, MaxOutputTokens: 131072},
			{ID: "Qwen/Qwen3.6-Max-Preview", Name: "Qwen3.6 Max", ContextLength: 262144, MaxOutputTokens: 131072},
			{ID: "Qwen/Qwen3.6-Plus", Name: "Qwen3.6 Plus", ContextLength: 262144, MaxOutputTokens: 131072},
			{ID: "Qwen/Qwen3.5-397B-A17B", Name: "Qwen3.5 397B", ContextLength: 262144, MaxOutputTokens: 131072},
			{ID: "Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8", Name: "Qwen3 Coder 480B", ContextLength: 262128, MaxOutputTokens: 131072},
			{ID: "nvidia/NVIDIA-Nemotron-3-Nano-Omni", Name: "Nemotron 3 Nano", ContextLength: 262144, MaxOutputTokens: 131072},
		},
	})

	Register(&RegistryEntry{
		ID:                 "github-models",
		Alias:              "ghm",
		Format:             FormatOpenAI,
		Executor:           "default",
		BaseURL:            "https://models.github.ai/inference/chat/completions",
		ModelsURL:          "https://models.github.ai/inference/models",
		AuthType:           AuthTypeAPIKey,
		AuthHeader:         "Authorization",
		AuthPrefix:         "Bearer ",
		DefaultContextLength: 128000,
		Headers: map[string]string{
			"X-GitHub-Api-Version": "2022-11-28",
			"Accept":              "application/vnd.github+json",
		},
		Models: []RegistryModel{
			{ID: "openai/gpt-4.1", Name: "GPT-4.1 (Free)", ContextLength: 1047576},
			{ID: "openai/gpt-4o", Name: "GPT-4o (Free)", ContextLength: 128000},
			{ID: "openai/gpt-4o-mini", Name: "GPT-4o Mini (Free)", ContextLength: 128000},
			{ID: "openai/o1", Name: "o1 (Free)", ContextLength: 200000},
			{ID: "openai/o3", Name: "o3 (Free)", ContextLength: 200000},
			{ID: "openai/o4-mini", Name: "o4-mini (Free)", ContextLength: 200000},
			{ID: "deepseek/DeepSeek-R1", Name: "DeepSeek R1 (Free)", ContextLength: 131072},
			{ID: "meta/Llama-4-Maverick-17B-128E-Instruct", Name: "Llama 4 Maverick (Free)", ContextLength: 131072},
			{ID: "xai/grok-3", Name: "Grok 3 (Free)", ContextLength: 131072},
			{ID: "mistral-ai/Mistral-Medium-3", Name: "Mistral Medium 3 (Free)", ContextLength: 128000},
			{ID: "cohere/Cohere-command-a", Name: "Cohere Command A (Free)", ContextLength: 128000},
			{ID: "microsoft/Phi-4", Name: "Phi-4 (Free)", ContextLength: 16384},
			{ID: "openai/text-embedding-3-large", Name: "Text Embedding 3 Large (Free)"},
			{ID: "openai/text-embedding-3-small", Name: "Text Embedding 3 Small (Free)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "hcnsec",
		Name:     "hcnsec",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.hcnsec.cn/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "inclusionai",
		Name:     "Inclusion Model",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.inclusionai.tech/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "inclusion-model", Name: "Inclusion Model"},
		},
	})

	Register(&RegistryEntry{
		ID:       "inference-net",
		Name:     "inference-net",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.inference.net/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "kenari",
		Name:     "kenari",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://kenari.id/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "nanogpt",
		Name:     "nanogpt",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://nano-gpt.com/api/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "nube",
		Name:     "nube",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://ai.nube.sh/api/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "sumopod",
		Name:     "sumopod",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://ai.sumopod.com/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "uncloseai",
		Name:     "Hermes 3 Llama 3.1 8B (🆓 Free)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://hermes.ai.unturf.com/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "adamo1139/Hermes-3-Llama-3.1-8B-FP8-Dynamic", Name: "Hermes 3 Llama 3.1 8B (🆓 Free)"},
			{ID: "qwen3.6:27b", Name: "Qwen3 Coder 27B (🆓 Free)"},
			{ID: "gemma4:31b", Name: "Gemma 4 31B (🆓 Free)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "v0-vercel",
		Name:     "v0-vercel",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.v0.dev/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "wafer",
		Name:     "DeepSeek V4 Pro",
		Format:   FormatClaude,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://pass.wafer.ai/v1/messages",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "DeepSeek-V4-Pro", Name: "DeepSeek V4 Pro"},
			{ID: "MiniMax-M2.7", Name: "MiniMax M2.7"},
			{ID: "Qwen3.5-397B-A17B", Name: "Qwen3.5 397B A17B"},
			{ID: "GLM-5.1", Name: "GLM 5.1"},
		},
	})

	Register(&RegistryEntry{
		ID:       "x5lab",
		Name:     "x5lab",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.x5lab.dev/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "zai",
		Name:     "GLM 5.2",
		Format:   FormatClaude,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.z.ai/api/anthropic/v1/messages",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "glm-5.2", Name: "GLM 5.2"},
			{ID: "glm-5.1", Name: "GLM 5.1"},
			{ID: "glm-5", Name: "GLM 5"},
			{ID: "glm-5-turbo", Name: "GLM 5 Turbo"},
			{ID: "glm-4.7-flash", Name: "GLM 4.7 Flash"},
			{ID: "glm-4.7", Name: "GLM 4.7"},
		},
	})

}