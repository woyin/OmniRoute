package registry

// RegisterMiniMax registers the MiniMax provider.
func RegisterMiniMax() {
	Register(&RegistryEntry{
		ID:     "minimax",
		Name:   "MiniMax",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.minimax.chat/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 262144,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "minimax-m3", Name: "MiniMax M3", ContextLength: 1048576, SupportsVision: true},
			{ID: "minimax-m2.7", Name: "MiniMax M2.7", ContextLength: 262144},
			{ID: "minimax-text-01", Name: "MiniMax Text 01", ContextLength: 1048576},
			{ID: "MiniMaxAI/MiniMax-M2.7", Name: "MiniMax M2.7 (Alt)", ContextLength: 262144},
		},
	})
}

// RegisterMiniMaxCN registers the MiniMax China (Anthropic-compatible) provider.
func RegisterMiniMaxCN() {
	Register(&RegistryEntry{
		ID:       "minimax-cn",
		Alias:    "minimax-cn",
		Format:   FormatClaude,
		Executor: "default",
		BaseURL:  "https://api.minimaxi.com/anthropic/v1/messages",
		ModelsURL: "https://api.minimaxi.com/v1/models",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Headers: map[string]string{
			"Anthropic-Version": "2023-06-01",
		},
		Models: []RegistryModel{
			{ID: "MiniMax-M3", Name: "MiniMax M3", ContextLength: 1048576, SupportsVision: true},
			{ID: "MiniMax-M2.7", Name: "MiniMax M2.7"},
			{ID: "MiniMax-M2.7-highspeed", Name: "MiniMax M2.7 Highspeed"},
			{ID: "MiniMax-M2.5", Name: "MiniMax M2.5"},
			{ID: "MiniMax-M2.5-highspeed", Name: "MiniMax M2.5 Highspeed"},
		},
	})
}
