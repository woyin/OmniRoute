package registry

func RegisterVertex() {
	Register(&RegistryEntry{
		ID:     "vertex",
		Name:   "Google Vertex AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://{location}-aiplatform.googleapis.com/v1/projects/{project}/locations/{location}/endpoints/openapi",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 1048576,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "google/gemini-2.5-pro", Name: "Gemini 2.5 Pro (Vertex)", ContextLength: 1048576, SupportsReasoning: true},
			{ID: "google/gemini-2.5-flash", Name: "Gemini 2.5 Flash (Vertex)", ContextLength: 1048576, SupportsReasoning: true},
		},
	})
}

// RegisterVertexPartner registers the Google Vertex AI Partner models provider.
func RegisterVertexPartner() {
	Register(&RegistryEntry{
		ID:       "vertex-partner",
		Alias:    "vp",
		Format:   FormatGemini,
		Executor: "vertex",
		BaseURL:  "https://us-central1-aiplatform.googleapis.com/v1/projects",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "DeepSeek-V4-Flash", Name: "DeepSeek V4 Flash"},
			{ID: "DeepSeek-V4-Pro", Name: "DeepSeek V4 Pro"},
			{ID: "Qwen3.6-35B-A3B", Name: "Qwen 3.6 35B A3B"},
			{ID: "GLM-5.1-FP8", Name: "GLM 5.1"},
			{ID: "claude-opus-4-8", Name: "Claude Opus 4.8"},
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7"},
			{ID: "claude-opus-4-6", Name: "Claude Opus 4.6"},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6"},
		},
	})
}
