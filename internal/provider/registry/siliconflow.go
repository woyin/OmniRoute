package registry

func RegisterSiliconFlow() {
	Register(&RegistryEntry{
		ID:     "siliconflow",
		Name:   "SiliconFlow",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.siliconflow.cn/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "deepseek-ai/DeepSeek-V4-Pro", Name: "DeepSeek V4 Pro", ContextLength: 131072, SupportsReasoning: true},
			{ID: "deepseek-ai/DeepSeek-V4-Flash", Name: "DeepSeek V4 Flash", ContextLength: 131072, SupportsReasoning: true},
			{ID: "Qwen/Qwen3-235B-A22B", Name: "Qwen3 235B", ContextLength: 131072},
		},
	})
}
