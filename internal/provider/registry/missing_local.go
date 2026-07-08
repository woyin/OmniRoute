package registry

// RegisterLocalProviders registers local/self-hosted provider entries.
func RegisterLocalProviders() {
	Register(&RegistryEntry{
		ID:       "ollama-local",
		Name:     "Ollama (Local)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeNoAuth,
		BaseURL:  "http://localhost:11434/v1/chat/completions",
	})
	Register(&RegistryEntry{
		ID:       "lm-studio",
		Name:     "LM Studio",
		Format:   FormatOpenAI,
		AuthType: AuthTypeNoAuth,
		BaseURL:  "http://localhost:1234/v1/chat/completions",
	})
	Register(&RegistryEntry{
		ID:       "llama-cpp",
		Name:     "llama.cpp Server",
		Format:   FormatOpenAI,
		AuthType: AuthTypeNoAuth,
		BaseURL:  "http://localhost:8080/v1/chat/completions",
	})
	Register(&RegistryEntry{
		ID:       "docker-model-runner",
		Name:     "Docker Model Runner",
		Format:   FormatOpenAI,
		AuthType: AuthTypeNoAuth,
		BaseURL:  "http://localhost:12434/engines/llama/v1/chat/completions",
	})
}
