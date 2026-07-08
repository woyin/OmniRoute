package registry

// RegisterInferenceProviders registers inference provider entries.
func RegisterInferenceProviders() {
	Register(&RegistryEntry{
		ID:       "dgrid",
		Name:     "DGrid Free Models Router",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.dgrid.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "dgridai/free", Name: "DGrid Free Models Router"},
		},
	})

	Register(&RegistryEntry{
		ID:       "hackclub",
		Name:     "Llama 3.3 70B",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://ai.hackclub.com/proxy/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "meta-llama/llama-3.3-70b-instruct", Name: "Llama 3.3 70B"},
			{ID: "mistralai/mistral-7b-instruct", Name: "Mistral 7B"},
			{ID: "deepseek-ai/deepseek-coder-33b", Name: "DeepSeek Coder 33B"},
		},
	})

	Register(&RegistryEntry{
		ID:       "lambda-ai",
		Name:     "lambda-ai",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.lambda.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	})

	Register(&RegistryEntry{
		ID:       "llm7",
		Name:     "GPT-4o mini (LLM7)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.llm7.io/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "gpt-4o-mini-2024-07-18", Name: "GPT-4o mini (LLM7)"},
			{ID: "gpt-4.1-nano-2025-04-14", Name: "GPT-4.1 nano (LLM7)"},
			{ID: "deepseek-r1-0528", Name: "DeepSeek R1 (LLM7)"},
			{ID: "qwen2.5-coder-32b-instruct", Name: "Qwen2.5 Coder 32B (LLM7)"},
		},
	})

	Register(&RegistryEntry{
		ID:       "monsterapi",
		Name:     "Llama 3.1 8B Instruct",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://api.monsterapi.ai/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "meta-llama/Meta-Llama-3.1-8B-Instruct", Name: "Llama 3.1 8B Instruct"},
			{ID: "meta-llama/Llama-3.3-70B-Instruct", Name: "Llama 3.3 70B Instruct"},
		},
	})

	Register(&RegistryEntry{
		ID:       "nous-research",
		Name:     "Hermes 4 7B (Nous Research)",
		Format:   FormatOpenAI,
		AuthType: AuthTypeAPIKey,
		BaseURL:  "https://inference-api.nousresearch.com/v1/chat/completions",
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		Models: []RegistryModel{
			{ID: "Hermes-4-405B", Name: "Hermes 4 7B (Nous Research)"},
			{ID: "Hermes-4-70B", Name: "Hermes 4 70B (Nous Research)"},
		},
	})

}