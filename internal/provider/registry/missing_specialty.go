package registry

// RegisterSpecialtyProviders registers specialty provider entries.
func RegisterSpecialtyProviders() {
	Register(&RegistryEntry{
		ID:       "haiper",
		Name:     "Gen 2 Video",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.haiper.ai/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "gen2", Name: "Gen 2 Video"},
			{ID: "gen2-image", Name: "Gen 2 Image"},
		},
	})

	Register(&RegistryEntry{
		ID:       "suno",
		Name:     "Chirp V5.5",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://studio-api.suno.ai/api/generate/v2/",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "chirp-fenix", Name: "Chirp V5.5"},
			{ID: "chirp-crow", Name: "Chirp V5"},
			{ID: "chirp-v4", Name: "Chirp V4"},
			{ID: "chirp-v3-5", Name: "Chirp V3.5"},
		},
	})

	Register(&RegistryEntry{
		ID:       "udio",
		Name:     "Udio Default",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://www.udio.com/api/generate-proxy",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "udio-default", Name: "Udio Default"},
		},
	})

}