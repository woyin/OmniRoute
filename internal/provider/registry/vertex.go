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
