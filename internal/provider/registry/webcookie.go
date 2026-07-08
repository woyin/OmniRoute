package registry

// RegisterWebCookieProviders adds web/cookie-based providers.
func RegisterWebCookieProviders() {
	Register(&RegistryEntry{
		ID:         "claude",
		Name:       "Claude Web",
		BaseURL:    "https://claude.ai",
		Format:     FormatClaude,
		AuthType:   AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "chatgpt-web",
		Name:       "ChatGPT Web",
		BaseURL:    "https://chatgpt.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "gemini-web",
		Name:       "Gemini Web",
		BaseURL:    "https://gemini.google.com",
		Format:     FormatGemini,
		AuthType:   AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "huggingchat",
		Name:       "HuggingChat",
		BaseURL:    "https://huggingface.co",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeOAuth,
		PassthroughModels: true,
	})
}
