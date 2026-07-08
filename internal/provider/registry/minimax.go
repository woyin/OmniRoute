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
