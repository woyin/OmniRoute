package registry

// RegisterOpenCode registers the OpenCode Free (zen) provider.
func RegisterOpenCode() {
	Register(&RegistryEntry{
		ID:     "opencode",
		Alias:  "oc",
		Name:   "OpenCode Free",
		Format: FormatOpenAI,
		Executor: "opencode",
		BaseURL: "https://opencode.ai/zen/v1",
		ModelsURL: "https://opencode.ai/zen/v1/models",
		AuthType: AuthTypeNoAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer",
		DefaultContextLength: 200000,
		PassthroughModels: true,
		HasFree: true,
		Models: []RegistryModel{
			{ID: "big-pickle", Name: "Big Pickle", SupportsReasoning: true},
			{ID: "deepseek-v4-flash-free", Name: "DeepSeek V4 Flash Free", SupportsReasoning: true},
			{ID: "minimax-m3-free", Name: "MiniMax M3 Free", ContextLength: 1048576, SupportsVision: true},
			{ID: "minimax-m2.5-free", Name: "MiniMax M2.5 Free", ContextLength: 204800},
			{ID: "ling-2.6-1t-free", Name: "Ling 2.6 Free", ContextLength: 262000},
			{ID: "trinity-large-preview-free", Name: "Trinity Large Preview Free", ContextLength: 131000},
			{ID: "nemotron-3-super-free", Name: "Nemotron 3 Super Free", ContextLength: 1000000},
			{ID: "qwen3.6-plus-free", Name: "Qwen3.6 Plus Free", TargetFormat: FormatClaude, ContextLength: 200000},
		},
	})
}

// RegisterOpenCodeGo registers the OpenCode Go provider.
func RegisterOpenCodeGo() {
	Register(&RegistryEntry{
		ID:     "opencode-go",
		Alias:  "ocgo",
		Name:   "OpenCode Go",
		Format: FormatOpenAI,
		Executor: "opencode",
		BaseURL: "https://opencode.ai/go/v1",
		ModelsURL: "https://opencode.ai/go/v1/models",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer",
		DefaultContextLength: 200000,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "deepseek-v4-pro", Name: "DeepSeek V4 Pro (Go)", SupportsReasoning: true},
			{ID: "deepseek-v4-flash", Name: "DeepSeek V4 Flash (Go)", SupportsReasoning: true},
			{ID: "kimi-k2.6", Name: "Kimi K2.6 (Go)"},
			{ID: "glm-5.1", Name: "GLM 5.1 (Go)"},
			{ID: "minimax-m3", Name: "MiniMax M3 (Go)", ContextLength: 1048576, SupportsVision: true},
		},
	})
}
