package registry

// RegisterBedrock registers the AWS Bedrock provider.
func RegisterBedrock() {
	Register(&RegistryEntry{
		ID:     "bedrock",
		Alias:  "aws",
		Name:   "AWS Bedrock",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://bedrock-runtime.{region}.amazonaws.com",
		ChatPath: "/model/{model}/converse",
		AuthType: AuthTypeAPIKey,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 200000,
		PassthroughModels: true,
		Models: []RegistryModel{
			{ID: "anthropic.claude-opus-4-7-v1:0", Name: "Claude Opus 4.7 (Bedrock)", ContextLength: 200000, SupportsReasoning: true},
			{ID: "anthropic.claude-sonnet-4-6-v1:0", Name: "Claude Sonnet 4.6 (Bedrock)", ContextLength: 200000, SupportsReasoning: true},
			{ID: "us.amazon.nova-pro-v1:0", Name: "Amazon Nova Pro", ContextLength: 300000},
			{ID: "us.amazon.nova-lite-v1:0", Name: "Amazon Nova Lite", ContextLength: 300000},
		},
	})
}
