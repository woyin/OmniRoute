// missing_specialty.go registers specialty providers ported from the main branch.
//
// Contains image/video generation, music, and specialty AI service providers.
// Also includes the 23 providers added for 1:1 parity with main branch:
//   - 360ai, arcee-ai, auto, aws-polly, azure-ai, black-forest-labs
//   - cablyai, codex-cloud, fal-ai, fenayai, getgoapi
//   - jina-ai, jina-reader, laozhang, nomic, piapi
//   - stability-ai, thebai, tinyfish, topaz, voyage-ai
//
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
// --- Missing providers from main branch (ported for 1:1 parity) ---

// Register360AI registers the 360 AI provider.
func Register360AI() {
	Register(&RegistryEntry{
		ID:                "360ai",
		Alias:             "360ai",
		Name:              "360 AI",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.ai.360.cn/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		HasFree:           true,
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterArceeAI registers the Arcee AI provider.
func RegisterArceeAI() {
	Register(&RegistryEntry{
		ID:                "arcee-ai",
		Alias:             "arcee",
		Name:              "Arcee AI",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.arcee.ai/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		HasFree:           true,
		PassthroughModels: true,
		DefaultContextLength: 262144,
	})
}

// RegisterAuto registers the Auto (Zero-Config) system provider.
func RegisterAuto() {
	Register(&RegistryEntry{
		ID:       "auto",
		Alias:    "auto",
		Name:     "Auto (Zero-Config)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeNoAuth,
		SystemOnly: true,
		DefaultContextLength: 131072,
	})
}

// RegisterAWSPolly registers the AWS Polly text-to-speech provider.
func RegisterAWSPolly() {
	Register(&RegistryEntry{
		ID:          "aws-polly",
		Alias:       "polly",
		Name:        "AWS Polly",
		Format:      FormatOpenAI,
		AuthType:    AuthTypeAPIKey,
		BaseURL:     "https://polly.us-east-1.amazonaws.com",
		AuthHeader:  "Authorization",
		AuthPrefix:  "AWS4-HMAC-SHA256 ",
		DefaultContextLength: 4096,
	})
}

// RegisterAzureAI registers the Azure AI Foundry provider.
func RegisterAzureAI() {
	Register(&RegistryEntry{
		ID:                "azure-ai",
		Alias:             "azure-ai",
		Name:              "Azure AI Foundry",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		AuthHeader:        "api-key",
		AuthPrefix:        "",
		PassthroughModels: true,
		DefaultContextLength: 128000,
	})
}

// RegisterBlackForestLabs registers the Black Forest Labs (FLUX) image provider.
func RegisterBlackForestLabs() {
	Register(&RegistryEntry{
		ID:       "black-forest-labs",
		Alias:    "bfl",
		Name:     "Black Forest Labs",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.bfl.ml/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 4096,
	})
}

// RegisterCablyAI registers the CablyAI (deprecated) provider.
func RegisterCablyAI() {
	Register(&RegistryEntry{
		ID:                "cablyai",
		Alias:             "cablyai",
		Name:              "CablyAI (Deprecated)",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.cablyai.com/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		Deprecated:        true,
		DefaultContextLength: 131072,
	})
}

// RegisterCodexCloud registers the Codex Cloud provider.
func RegisterCodexCloud() {
	Register(&RegistryEntry{
		ID:       "codex-cloud",
		Alias:    "codex-cloud",
		Name:     "Codex Cloud",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.openai.com/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
	})
}

// RegisterFalAI registers the Fal.ai image/video provider.
func RegisterFalAI() {
	Register(&RegistryEntry{
		ID:       "fal-ai",
		Alias:    "fal",
		Name:     "Fal.ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://fal.run",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 4096,
	})
}

// RegisterFenayAI registers the FenayAI provider.
func RegisterFenayAI() {
	Register(&RegistryEntry{
		ID:                "fenayai",
		Alias:             "fenayai",
		Name:              "FenayAI",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.fenayai.com/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterGetGoAPI registers the GoAPI provider.
func RegisterGetGoAPI() {
	Register(&RegistryEntry{
		ID:                "getgoapi",
		Alias:             "ggo",
		Name:              "GoAPI",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.getgoapi.com/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterJinaAI registers the Jina AI rerank/embed provider.
func RegisterJinaAI() {
	Register(&RegistryEntry{
		ID:       "jina-ai",
		Alias:    "jina",
		Name:     "Jina AI",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.jina.ai/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		HasFree:   true,
		DefaultContextLength: 8192,
	})
}

// RegisterJinaReader registers the Jina Reader web-fetch provider.
func RegisterJinaReader() {
	Register(&RegistryEntry{
		ID:       "jina-reader",
		Alias:    "jr",
		Name:     "Jina Reader",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://r.jina.ai",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		HasFree:   true,
		DefaultContextLength: 8192,
	})
}

// RegisterLaoZhang registers the LaoZhang AI provider.
func RegisterLaoZhang() {
	Register(&RegistryEntry{
		ID:                "laozhang",
		Alias:             "lz",
		Name:              "LaoZhang AI",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.laozhang.ai/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterNomic registers the Nomic embeddings provider.
func RegisterNomic() {
	Register(&RegistryEntry{
		ID:                "nomic",
		Alias:             "nomic",
		Name:              "Nomic",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api-atlas.nomic.ai/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		HasFree:           true,
		PassthroughModels: true,
		DefaultContextLength: 8192,
	})
}

// RegisterPiAPI registers the PiAPI provider.
func RegisterPiAPI() {
	Register(&RegistryEntry{
		ID:                "piapi",
		Alias:             "pi",
		Name:              "PiAPI",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.piapi.ai/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterStabilityAI registers the Stability AI image provider.
func RegisterStabilityAI() {
	Register(&RegistryEntry{
		ID:       "stability-ai",
		Alias:    "stability",
		Name:     "Stability AI",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.stability.ai/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 4096,
	})
}

// RegisterTheBAI registers the TheB.AI provider.
func RegisterTheBAI() {
	Register(&RegistryEntry{
		ID:                "thebai",
		Alias:             "thebai",
		Name:              "TheB.AI",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeAPIKey,
		BaseURL:           "https://api.theb.ai/v1",
		AuthHeader:        "Authorization",
		AuthPrefix:        "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 131072,
	})
}

// RegisterTinyFish registers the TinyFish Fetch web provider.
func RegisterTinyFish() {
	Register(&RegistryEntry{
		ID:       "tinyfish",
		Alias:    "tf",
		Name:     "TinyFish Fetch",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.tinyfish.ai/v1",
		AuthHeader: "X-API-Key",
		AuthPrefix: "",
		DefaultContextLength: 8192,
	})
}

// RegisterTopaz registers the Topaz image enhancement provider.
func RegisterTopaz() {
	Register(&RegistryEntry{
		ID:       "topaz",
		Alias:    "topaz",
		Name:     "Topaz",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.topazlabs.com/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
		DefaultContextLength: 4096,
	})
}

// RegisterVoyageAI registers the Voyage AI embeddings/rerank provider.
func RegisterVoyageAI() {
	Register(&RegistryEntry{
		ID:       "voyage-ai",
		Alias:    "voyage",
		Name:     "Voyage AI",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.voyageai.com/v1",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		HasFree:   true,
		DefaultContextLength: 32768,
	})
}
