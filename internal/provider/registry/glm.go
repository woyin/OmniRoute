package registry

func RegisterGLM() {
	Register(&RegistryEntry{
		ID:     "glm",
		Name:   "GLM (Zhipu AI)",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "glm-5.1", Name: "GLM 5.1", ContextLength: 131072, SupportsReasoning: true},
			{ID: "glm-4-plus", Name: "GLM 4 Plus", ContextLength: 131072},
		},
	})
}

// RegisterGLMCN registers the GLM China (Zhipu AI coding) provider.
func RegisterGLMCN() {
	Register(&RegistryEntry{
		ID:                 "glm-cn",
		Alias:              "glmcn",
		Format:             FormatOpenAI,
		Executor:           "glm",
		BaseURL:            "https://open.bigmodel.cn/api/coding/paas/v4/chat/completions",
		AuthType:           AuthTypeAPIKey,
		AuthHeader:         "Authorization",
		AuthPrefix:         "Bearer ",
		DefaultContextLength: 200000,
		PassthroughModels:  true,
		Models: []RegistryModel{
			{ID: "glm-5.2", Name: "GLM 5.2", ContextLength: 1000000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-5.2-high", Name: "GLM 5.2 High", ContextLength: 1000000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-5.2-max", Name: "GLM 5.2 Max", ContextLength: 1000000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-5.1", Name: "GLM 5.1", ContextLength: 204800, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-5", Name: "GLM 5", ContextLength: 200000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-5-turbo", Name: "GLM 5 Turbo", ContextLength: 200000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-4.7-flash", Name: "GLM 4.7 Flash", ContextLength: 200000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-4.7", Name: "GLM 4.7", ContextLength: 200000, MaxOutputTokens: 131072, SupportsReasoning: true},
			{ID: "glm-4.6v", Name: "GLM 4.6V (Vision)", ContextLength: 128000, MaxOutputTokens: 32768, SupportsReasoning: true, SupportsVision: true},
			{ID: "glm-4.6", Name: "GLM 4.6", ContextLength: 200000, MaxOutputTokens: 32768, SupportsReasoning: true},
			{ID: "glm-4.5v", Name: "GLM 4.5V (Vision)", ContextLength: 16000, MaxOutputTokens: 32768, SupportsReasoning: true, SupportsVision: true},
			{ID: "glm-4.5", Name: "GLM 4.5", ContextLength: 128000, MaxOutputTokens: 32768, SupportsReasoning: true},
			{ID: "glm-4.5-air", Name: "GLM 4.5 Air", ContextLength: 128000, MaxOutputTokens: 32768, SupportsReasoning: true},
		},
	})
}
