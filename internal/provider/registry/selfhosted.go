package registry

// RegisterSelfHostedProviders adds self-hosted/local providers.
func RegisterSelfHostedProviders() {
	Register(&RegistryEntry{
		ID:                "vllm",
		Name:              "vLLM",
		BaseURL:           "http://localhost:8000",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:                "llamafile",
		Name:              "Llamafile",
		BaseURL:           "http://localhost:8080",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:                "triton",
		Name:              "NVIDIA Triton",
		BaseURL:           "http://localhost:8000",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:                "xinference",
		Name:              "XInference",
		BaseURL:           "http://localhost:9997",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:                "oobabooga",
		Name:              "oobabooga",
		BaseURL:           "http://localhost:5000",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:                "lemonade",
		Name:              "Lemonade Server",
		BaseURL:           "http://localhost:8000",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:                "sdwebui",
		Name:              "SD WebUI",
		BaseURL:           "http://localhost:7860",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})

	Register(&RegistryEntry{
		ID:                "comfyui",
		Name:              "ComfyUI",
		BaseURL:           "http://localhost:8188",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeNoAuth,
		PassthroughModels: true,
	})
}
