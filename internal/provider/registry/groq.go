package registry

func RegisterGroq() {
	Register(&RegistryEntry{
		ID:     "groq",
		Name:   "Groq",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.groq.com/openai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "llama-4-scout-17b-16e-instruct", Name: "Llama 4 Scout 17B", ContextLength: 131072},
			{ID: "llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick 17B", ContextLength: 131072},
			{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B", ContextLength: 131072},
			{ID: "llama-3.1-8b-instant", Name: "Llama 3.1 8B", ContextLength: 131072},
			{ID: "mixtral-8x7b-32768", Name: "Mixtral 8x7B", ContextLength: 32768},
			{ID: "gemma2-9b-it", Name: "Gemma 2 9B", ContextLength: 8192},
		},
	})
}
