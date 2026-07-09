package registry

// RegisterNoAuthProviders registers all free/no-auth providers.
// These providers do not require an API key or OAuth token.
func RegisterNoAuthProviders() {
	RegisterDuckDuckGoWeb()
	RegisterTheOldLLM()
	RegisterChipotle()
	RegisterVeoAIFreeWeb()
}

// RegisterDuckDuckGoWeb registers the DuckDuckGo AI Chat noauth provider.
func RegisterDuckDuckGoWeb() {
	Register(&RegistryEntry{
		ID:        "duckduckgo-web",
		Alias:     "ddgw",
		Name:      "DuckDuckGo AI Chat",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeNoAuth,
		HasFree:   true,
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterTheOldLLM registers The Old LLM free provider.
func RegisterTheOldLLM() {
	Register(&RegistryEntry{
		ID:        "theoldllm",
		Alias:     "tllm",
		Name:      "The Old LLM (Free)",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeNoAuth,
		HasFree:   true,
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterChipotle registers the Chipotle Pepper AI noauth provider.
func RegisterChipotle() {
	Register(&RegistryEntry{
		ID:        "chipotle",
		Alias:     "pepper",
		Name:      "Chipotle Pepper AI (Free)",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeNoAuth,
		HasFree:   true,
		DefaultContextLength: 131072,
	})
}

// RegisterVeoAIFreeWeb registers the Veo AI Free video generation noauth provider.
func RegisterVeoAIFreeWeb() {
	Register(&RegistryEntry{
		ID:        "veoaifree-web",
		Alias:     "veo-free",
		Name:      "Veo AI Free",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeNoAuth,
		HasFree:   true,
		DefaultContextLength: 131072,
	})
}
