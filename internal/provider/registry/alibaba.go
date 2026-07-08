package registry

// RegisterAlibaba registers the Alibaba/Qwen provider.
func RegisterAlibaba() {
	Register(&RegistryEntry{
		ID:     "alibaba",
		Alias:  "qwen",
		Name:   "Alibaba Qwen",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "qwen3.6-plus", Name: "Qwen 3.6 Plus", ContextLength: 131072, SupportsReasoning: true},
			{ID: "qwen3.6-max", Name: "Qwen 3.6 Max", ContextLength: 131072, SupportsReasoning: true},
			{ID: "qwen-max-latest", Name: "Qwen Max Latest", ContextLength: 131072},
			{ID: "qwen-max", Name: "Qwen Max", ContextLength: 32768},
			{ID: "qwen-plus", Name: "Qwen Plus", ContextLength: 131072},
			{ID: "qwen-turbo", Name: "Qwen Turbo", ContextLength: 131072},
			{ID: "qwen-long", Name: "Qwen Long", ContextLength: 10000000},
			{ID: "qwen-vl-max", Name: "Qwen VL Max", ContextLength: 32768, SupportsVision: true},
			{ID: "Qwen/Qwen3-Coder-480B-A35B-Instruct", Name: "Qwen3 Coder 480B", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
