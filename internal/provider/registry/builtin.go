package registry

// RegisterBuiltinProviders registers all built-in provider entries.
func RegisterBuiltinProviders() {
	// Priority executors (custom logic)
	RegisterOpenCode()
	RegisterOpenCodeGo()
	RegisterOpenCodeZen()
	RegisterCodex()
	RegisterCommandCode()

	// Free / no-auth providers
	RegisterPollinations()
	RegisterPuter()

	// Tier-1 API key providers
	RegisterOpenAI()
	RegisterAnthropic()
	RegisterGemini()
	RegisterDeepSeek()
	RegisterGroq()
	RegisterXAI()
	RegisterMistral()

	// Tier-2 API key providers
	RegisterTogether()
	RegisterFireworks()
	RegisterCohere()
	RegisterCerebras()
	RegisterNVIDIA()
	RegisterPerplexity()
	RegisterOpenRouter()
	RegisterSiliconFlow()
	RegisterHuggingFace()
	RegisterDeepInfra()
	RegisterSambaNova()
	RegisterNebius()
	RegisterHyperbolic()

	// Enterprise / cloud
	RegisterVertex()
	RegisterVertexPartner()
	RegisterCloudflareAI()
	RegisterDatabricks()
	RegisterSnowflake()
	RegisterAzureOpenAI()
	RegisterBedrock()

	// Regional / Chinese providers
	RegisterAlibaba()
	RegisterAlibabaCN()
	RegisterVolcengine()
	RegisterGLM()
	RegisterGLMCN()
	RegisterMiniMax()
	RegisterMiniMaxCN()
	RegisterKimi()
	RegisterKimiCodingAPIKey()
	RegisterMoonshot()
	RegisterXiaomiMiMo()

	// Meta / Llama
	RegisterMetaLlama()

	// Specialty
	RegisterAI21()
	RegisterVenice()
	RegisterOllamaCloud()

	// Cloud/infrastructure API providers
	RegisterScaleway()
	RegisterAIMLAPI()
	RegisterLambdaAI()
	RegisterRunPod()
	RegisterNScale()
	RegisterOVHcloud()
	RegisterBaseten()
	RegisterBlackbox()

	// Self-hosted / local providers
	RegisterSelfHostedProviders()

	// Audio / media providers
	RegisterAudioProviders()

	// Web cookie / OAuth-like providers
	RegisterWebCookieProviders()

	// No-auth / free providers (DuckDuckGo, Chipotle, TheOldLLM, VeoAIFree)
	RegisterNoAuthProviders()

	// Regional / enterprise providers
	RegisterRegionalProviders()

	// Gateway / router / specialty providers

	// Extended: IDE/CLI, search, cloud-agents, misc
	RegisterExtendedProviders()

	// OAuth providers
	RegisterOAuthProviders()


	// Search providers
	RegisterSearchProviders()

	// Local/self-hosted providers
	RegisterLocalProviders()

	// Missing providers ported from main branch
	RegisterGatewayProviders()
	RegisterAggregatorProviders()
	RegisterRegionalCnProviders()
	RegisterSpecialtyProviders()
	RegisterCodingProviders()
	RegisterInferenceProviders()
	RegisterMiscProviders()

	// Custom-compatible (user-defined)
	RegisterOpenAICompatible()
	RegisterAnthropicCompatible()
}
