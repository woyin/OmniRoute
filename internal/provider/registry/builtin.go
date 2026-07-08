package registry

// RegisterBuiltinProviders registers all built-in provider entries.
func RegisterBuiltinProviders() {
	// Priority executors (custom logic)
	RegisterOpenCode()
	RegisterOpenCodeGo()
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
	RegisterCloudflareAI()
	RegisterDatabricks()
	RegisterSnowflake()
	RegisterAzureOpenAI()
	RegisterBedrock()

	// Regional / Chinese providers
	RegisterAlibaba()
	RegisterVolcengine()
	RegisterGLM()
	RegisterMiniMax()
	RegisterKimi()
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

	// Regional / enterprise providers
	RegisterRegionalProviders()

	// Gateway / router / specialty providers
	RegisterGatewayProviders()

	// Extended: IDE/CLI, search, cloud-agents, misc
	RegisterExtendedProviders()

	// OAuth providers
	RegisterOAuthProviders()

	// Custom-compatible (user-defined)
	RegisterOpenAICompatible()
	RegisterAnthropicCompatible()
}
