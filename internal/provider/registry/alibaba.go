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

// RegisterAlibabaCN registers the Alibaba China (DashScope) provider.
func RegisterAlibabaCN() {
	Register(&RegistryEntry{
		ID:                "alibaba-cn",
		Alias:             "ali-cn",
		Format:            FormatOpenAI,
		Executor:          "default",
		BaseURL:           "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		ModelsURL:         "https://dashscope.aliyuncs.com/compatible-mode/v1/models",
		AuthType:          AuthTypeAPIKey,
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "qwen-max", Name: "Qwen Max"},
			{ID: "qwen-max-2025-01-25", Name: "Qwen Max (2025-01-25)"},
			{ID: "qwen-plus", Name: "Qwen Plus"},
			{ID: "qwen-plus-2025-07-14", Name: "Qwen Plus (2025-07-14)"},
			{ID: "qwen-turbo", Name: "Qwen Turbo"},
			{ID: "qwen-turbo-2025-11-01", Name: "Qwen Turbo (2025-11-01)"},
			{ID: "qwen3-coder-plus", Name: "Qwen3 Coder Plus"},
			{ID: "qwen3-coder-flash", Name: "Qwen3 Coder Flash"},
			{ID: "qwq-plus", Name: "QwQ Plus (Reasoning)"},
			{ID: "qwq-32b", Name: "QwQ 32B"},
			{ID: "qwen3-32b", Name: "Qwen3 32B"},
			{ID: "qwen3-235b-a22b", Name: "Qwen3 235B A22B"},
		},
	})
}
