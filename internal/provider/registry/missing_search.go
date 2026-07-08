package registry

// RegisterSearchProviders registers search provider entries.
func RegisterSearchProviders() {
	Register(&RegistryEntry{
		ID:        "brave-search",
		Name:      "Brave Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://api.search.brave.com/res/v1/web/search",
		AuthHeader: "X-Subscription-Token",
		AuthPrefix: "",
	})
	Register(&RegistryEntry{
		ID:        "exa-search",
		Name:      "Exa Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://api.exa.ai/search",
		AuthHeader: "x-api-key",
		AuthPrefix: "",
	})
	Register(&RegistryEntry{
		ID:        "serper-search",
		Name:      "Serper Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://google.serper.dev/search",
		AuthHeader: "X-API-KEY",
		AuthPrefix: "",
	})
	Register(&RegistryEntry{
		ID:        "tavily-search",
		Name:      "Tavily Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://api.tavily.com/search",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})
	Register(&RegistryEntry{
		ID:        "searchapi-search",
		Name:      "SearchAPI",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://api.searchapi.io/v1/search",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})
	Register(&RegistryEntry{
		ID:        "youcom-search",
		Name:      "You.com Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://api.ydc-search.io/search",
		AuthHeader: "X-API-Key",
		AuthPrefix: "",
	})
	Register(&RegistryEntry{
		ID:        "google-pse-search",
		Name:      "Google PSE Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://www.googleapis.com/customsearch/v1",
		AuthHeader: "key",
		AuthPrefix: "",
	})
	Register(&RegistryEntry{
		ID:        "linkup-search",
		Name:      "Linkup Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://api.linkup.so/v1/search",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})
	Register(&RegistryEntry{
		ID:        "searxng-search",
		Name:      "SearXNG Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeNoAuth,
		BaseURL:   "http://localhost:8080/search",
	})
	Register(&RegistryEntry{
		ID:        "perplexity-search",
		Name:      "Perplexity Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeAPIKey,
		BaseURL:   "https://api.perplexity.ai/search",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})
	Register(&RegistryEntry{
		ID:        "ollama-search",
		Name:      "Ollama Search",
		Format:    FormatOpenAI,
		AuthType:  AuthTypeNoAuth,
		BaseURL:   "http://localhost:11434/api/search",
	})
}
