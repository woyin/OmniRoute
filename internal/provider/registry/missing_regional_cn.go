package registry

// RegisterRegionalCnProviders registers regional_cn provider entries.
func RegisterRegionalCnProviders() {
	Register(&RegistryEntry{
		ID:       "bailian-coding-plan",
		Name:     "Qwen3.7 Plus(vision)",
		Format:   FormatClaude,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://coding-intl.dashscope.aliyuncs.com/apps/anthropic/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "qwen3.7-plus", Name: "Qwen3.7 Plus(vision)"},
			{ID: "qwen3-coder-plus", Name: "Qwen3 Coder Plus"},
			{ID: "qwen3-coder-next", Name: "Qwen3 Coder Next"},
			{ID: "glm-4.7", Name: "GLM 4.7"},
			{ID: "qwen3.6-plus", Name: "Qwen3.6 Plus(vision)"},
			{ID: "qwen3.5-plus", Name: "Qwen3.5 Plus(vision)"},
			{ID: "qwen3-max-2026-01-23", Name: "Qwen3 Max"},
			{ID: "kimi-k2.5", Name: "Kimi K2.5(vision)"},
			{ID: "glm-5", Name: "GLM 5"},
			{ID: "MiniMax-M2.5", Name: "MiniMax M2.5"},
		},
	})

	Register(&RegistryEntry{
		ID:       "byteplus",
		Name:     "Seed 2.0",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://ark.ap-southeast.bytepluses.com/api/v3/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "seed-2.0", Name: "Seed 2.0"},
			{ID: "kimi-k2-thinking", Name: "Kimi K2 Thinking"},
			{ID: "glm-4.7", Name: "GLM 4.7"},
			{ID: "gpt-oss-120b", Name: "GPT-OSS-120B"},
		},
	})





	Register(&RegistryEntry{
		ID:       "qianfan",
		Name:     "ERNIE 5.1",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://qianfan.baidubce.com/v2/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "ernie-5.1", Name: "ERNIE 5.1"},
			{ID: "ernie-5.0-thinking-latest", Name: "ERNIE 5.0 Thinking Latest"},
			{ID: "ernie-x1.1", Name: "ERNIE X1.1"},
		},
	})

}