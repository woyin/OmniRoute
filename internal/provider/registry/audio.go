package registry

// RegisterAudioProviders adds audio/speech/media providers.
func RegisterAudioProviders() {
	Register(&RegistryEntry{
		ID:         "elevenlabs",
		Name:       "ElevenLabs",
		BaseURL:    "https://api.elevenlabs.io",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "xi-api-key",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "cartesia",
		Name:       "Cartesia",
		BaseURL:    "https://api.cartesia.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "X-API-Key",
		AuthPrefix: "",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "playht",
		Name:       "PlayHT",
		BaseURL:    "https://api.play.ht",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "deepgram",
		Name:       "Deepgram",
		BaseURL:    "https://api.deepgram.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Token ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "assemblyai",
		Name:       "AssemblyAI",
		BaseURL:    "https://api.assemblyai.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "inworld",
		Name:       "Inworld",
		BaseURL:    "https://api.inworld.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})
}
