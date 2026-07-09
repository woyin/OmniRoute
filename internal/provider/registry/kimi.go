package registry

func RegisterKimi() {
	Register(&RegistryEntry{
		ID:     "kimi",
		Name:   "Kimi (Moonshot)",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.kimi.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 262144,
		Models: []RegistryModel{
			{ID: "kimi-k2.6", Name: "Kimi K2.6", ContextLength: 262144, SupportsReasoning: true},
			{ID: "kimi-latest", Name: "Kimi Latest", ContextLength: 131072},
		},
	})
}

// RegisterKimiCodingAPIKey registers the Kimi Coding (API key) provider.
func RegisterKimiCodingAPIKey() {
	Register(&RegistryEntry{
		ID:                 "kimi-coding-apikey",
		Alias:              "kmca",
		Format:             FormatClaude,
		Executor:           "default",
		BaseURL:            "https://api.kimi.com/coding/v1/messages",
		AuthType:           AuthTypeAPIKey,
		AuthHeader:         "x-api-key",
		DefaultContextLength: 262144,
		Headers: map[string]string{
			"Anthropic-Version": "2023-06-01",
		},
		Models: []RegistryModel{
			{ID: "kimi-k2.6", Name: "Kimi K2.6", ContextLength: 262144, MaxOutputTokens: 262144, SupportsVision: true},
			{ID: "kimi-k2.6-thinking", Name: "Kimi K2.6 Thinking", ContextLength: 262144, MaxOutputTokens: 262144},
			{ID: "kimi-k2.7-code", Name: "Kimi K2.7 Code", ContextLength: 262144, MaxOutputTokens: 262144, SupportsVision: true, SupportsReasoning: true},
			{ID: "kimi-k2.7-code-highspeed", Name: "Kimi K2.7 Code (High Speed)", ContextLength: 262144, MaxOutputTokens: 262144, SupportsVision: true, SupportsReasoning: true},
			{ID: "moonshotai/kimi-k2.7-code", Name: "Kimi K2.7 Code", ContextLength: 262144, MaxOutputTokens: 262144},
		},
	})
}
