package registry

// RegisterExtendedProviders adds IDE/CLI, search, cloud-agent, and miscellaneous providers.
func RegisterExtendedProviders() {
	// IDE / CLI tools
	Register(&RegistryEntry{
		ID:       "cline",
		Name:     "Cline",
		BaseURL:  "https://api.cline.bot",
		Format:   FormatOpenAI,
		AuthType: AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:       "kilocode",
		Name:     "Kilo Code",
		BaseURL:  "https://api.kilocode.ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:       "trae",
		Name:     "Trae",
		BaseURL:  "https://api.trae.ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:       "gitlab",
		Name:     "GitLab Duo",
		BaseURL:  "https://gitlab.com",
		Format:   FormatOpenAI,
		AuthType: AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:       "zed",
		Name:     "Zed",
		BaseURL:  "https://api.zed.dev",
		Format:   FormatOpenAI,
		AuthType: AuthTypeOAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:       "mimocode",
		Name:     "MimoCode",
		BaseURL:  "https://api.mimocode.ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeNoAuth,
		PassthroughModels: true,
		HasFree:  true,
	})

	Register(&RegistryEntry{
		ID:       "auggie",
		Name:     "Auggie",
		BaseURL:  "https://api.auggie.ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeNoAuth,
		PassthroughModels: true,
	})

	// Search providers
	Register(&RegistryEntry{
		ID:         "firecrawl",
		Name:       "Firecrawl",
		BaseURL:    "https://api.firecrawl.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	// Cloud agents
	Register(&RegistryEntry{
		ID:       "devin",
		Name:     "Devin",
		BaseURL:  "https://api.devin.ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:       "jules",
		Name:     "Jules",
		BaseURL:  "https://jules.google.com",
		Format:   FormatOpenAI,
		AuthType: AuthTypeOAuth,
		PassthroughModels: true,
	})

	// Additional important API-key providers
	Register(&RegistryEntry{
		ID:         "reka",
		Name:       "Reka",
		BaseURL:    "https://api.reka.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "recraft",
		Name:       "Recraft",
		BaseURL:    "https://api.recraft.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "leonardo",
		Name:       "Leonardo AI",
		BaseURL:    "https://cloud.leonardo.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "ideogram",
		Name:       "Ideogram",
		BaseURL:    "https://api.ideogram.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "runwayml",
		Name:       "RunwayML",
		BaseURL:    "https://api.runwayml.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "coze",
		Name:       "Coze",
		BaseURL:    "https://api.coze.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "dify",
		Name:       "Dify",
		BaseURL:    "https://api.dify.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "clarifai",
		Name:       "Clarifai",
		BaseURL:    "https://api.clarifai.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "nlpcloud",
		Name:       "NLP Cloud",
		BaseURL:    "https://api.nlpcloud.io",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Token ",
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "datarobot",
		Name:       "DataRobot",
		BaseURL:    "https://api.datarobot.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "liquid",
		Name:       "Liquid AI",
		BaseURL:    "https://api.liquid.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "llamagate",
		Name:       "LlamaGate",
		BaseURL:    "https://api.llamagate.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "modelscope",
		Name:       "ModelScope",
		BaseURL:    "https://api.modelscope.cn",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "doubao",
		Name:       "Doubao (ByteDance)",
		BaseURL:    "https://api.doubao.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "baichuan",
		Name:       "Baichuan",
		BaseURL:    "https://api.baichuan-ai.com",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "yi",
		Name:       "Yi (01.AI)",
		BaseURL:    "https://api.01.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "glmt",
		Name:       "GLM-T",
		BaseURL:    "https://api.glmt.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:         "wandb",
		Name:       "Weights & Biases",
		BaseURL:    "https://api.wandb.ai",
		Format:     FormatOpenAI,
		AuthType:   AuthTypeAPIKey,
		PassthroughModels: true,
	})
}
