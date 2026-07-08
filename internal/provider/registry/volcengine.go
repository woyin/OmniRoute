package registry

func RegisterVolcengine() {
	Register(&RegistryEntry{
		ID:     "volcengine",
		Name:   "Volcengine (ByteDance)",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "doubao-1.5-pro-256k", Name: "Doubao 1.5 Pro 256K", ContextLength: 262144},
			{ID: "doubao-1.5-pro-32k", Name: "Doubao 1.5 Pro 32K", ContextLength: 32768},
		},
	})
}
