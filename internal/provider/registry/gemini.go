package registry

// RegisterGemini registers the Google Gemini provider.
func RegisterGemini() {
	Register(&RegistryEntry{
		ID:     "gemini",
		Name:   "Google Gemini",
		Format: FormatGemini,
		Executor: "default",
		BaseURL: "https://generativelanguage.googleapis.com/v1beta",
		ChatPath: "/models/{model}:generateContent",
		ModelsURL: "https://generativelanguage.googleapis.com/v1beta/models",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "x-goog-api-key",
		AuthPrefix: "",
		DefaultContextLength: 1048576,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro", ContextLength: 1048576, MaxOutputTokens: 65536, SupportsReasoning: true, SupportsVision: true},
			{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash", ContextLength: 1048576, MaxOutputTokens: 65536, SupportsReasoning: true, SupportsVision: true},
			{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", ContextLength: 1048576, SupportsVision: true},
			{ID: "gemini-2.0-flash-lite", Name: "Gemini 2.0 Flash Lite", ContextLength: 1048576},
			{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", ContextLength: 2097152, SupportsVision: true},
			{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", ContextLength: 1048576, SupportsVision: true},
		},
	})
}
