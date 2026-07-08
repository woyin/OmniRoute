package registry

func RegisterCloudflareAI() {
	Register(&RegistryEntry{
		ID:     "cloudflare-ai",
		Name:   "Cloudflare Workers AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.cloudflare.com/client/v4/accounts/{account_id}/ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 65536,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "@cf/meta/llama-4-scout-17b-16e-instruct", Name: "Llama 4 Scout", ContextLength: 131072},
			{ID: "@cf/deepseek-ai/deepseek-r1-distill-qwen-32b", Name: "DeepSeek R1 Distill", SupportsReasoning: true},
		},
	})
}
