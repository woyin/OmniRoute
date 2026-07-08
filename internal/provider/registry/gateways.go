package registry

// RegisterGatewayProviders adds gateway/router/specialty providers.
func RegisterGatewayProviders() {
	Register(&RegistryEntry{
		ID:         "openrouter",
		Name:       "OpenRouter",
		BaseURL:    "https://openrouter.ai/api",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "morph",
		Name:       "Morph",
		BaseURL:    "https://api.morph.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "upstage",
		Name:       "Upstage",
		BaseURL:    "https://api.upstage.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "maritalk",
		Name:       "Maritalk",
		BaseURL:    "https://chat.maritaca.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "codestral",
		Name:       "Codestral",
		BaseURL:    "https://codestral.mistral.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "poe",
		Name:       "Poe",
		BaseURL:    "https://api.poe.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "empower",
		Name:       "Empower",
		BaseURL:    "https://app.empower.dev",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "agentrouter",
		Name:       "AgentRouter",
		BaseURL:    "https://api.agentrouter.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "requesty",
		Name:       "Requesty",
		BaseURL:    "https://router.requesty.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "orcarouter",
		Name:       "OrcaRouter",
		BaseURL:    "https://api.orcarouter.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "zenmux",
		Name:       "ZenMux",
		BaseURL:    "https://api.zenmux.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "longcat",
		Name:       "Longcat",
		BaseURL:    "https://api.longcat.io",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "predibase",
		Name:       "Predibase",
		BaseURL:    "https://api.predibase.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "publicai",
		Name:       "PublicAI",
		BaseURL:    "https://api.publicai.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "bytez",
		Name:       "Bytez",
		BaseURL:    "https://api.bytez.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "modal",
		Name:       "Modal",
		BaseURL:    "https://api.modal.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "chutes",
		Name:       "Chutes",
		BaseURL:    "https://chutes.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "kluster",
		Name:       "Kluster",
		BaseURL:    "https://api.kluster.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "galadriel",
		Name:       "Galadriel",
		BaseURL:    "https://api.galadriel.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "novita",
		Name:       "Novita AI",
		BaseURL:    "https://api.novita.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "friendliai",
		Name:       "FriendliAI",
		BaseURL:    "https://api.friendli.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "gigachat",
		Name:       "GigaChat (Sber)",
		BaseURL:    "https://gigachat.devices.sberbank.ru",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "crof",
		Name:       "CrofAI",
		BaseURL:    "https://api.crofai.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})
}
