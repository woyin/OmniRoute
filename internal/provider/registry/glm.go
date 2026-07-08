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
