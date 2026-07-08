package registry

func RegisterXiaomiMiMo() {
	Register(&RegistryEntry{
		ID:     "xiaomi-mimo",
		Name:   "Xiaomi MiMo",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://xiaomi-mimo.github.io/api/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "mimo-vl-2.1", Name: "MiMo VL 2.1", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}
