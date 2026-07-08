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