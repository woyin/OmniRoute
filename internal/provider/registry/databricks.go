package registry

func RegisterDatabricks() {
	Register(&RegistryEntry{
		ID:     "databricks",
		Name:   "Databricks",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://adb-{workspace_id}.azuredatabricks.net/serving-endpoints",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "databricks-meta-llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick (Databricks)", ContextLength: 131072},
		},
	})
}
