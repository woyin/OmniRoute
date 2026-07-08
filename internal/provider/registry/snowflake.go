package registry

func RegisterSnowflake() {
	Register(&RegistryEntry{
		ID:     "snowflake",
		Name:   "Snowflake Cortex",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://{account}.snowflakecomputing.com/api/v2/databases/{db}/schemas/{schema}/models",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "llama-4-maverick-17b-128e-instruct", Name: "Llama 4 Maverick (Cortex)", ContextLength: 131072},
			{ID: "snowflake-arctic", Name: "Snowflake Arctic", ContextLength: 131072},
		},
	})
}
