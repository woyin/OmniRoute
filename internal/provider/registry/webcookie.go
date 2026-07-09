package registry

// RegisterWebCookieProviders adds web/cookie-based providers.
// These providers use browser session cookies for authentication.
func RegisterWebCookieProviders() {
	RegisterChatGPTWeb()
	RegisterGrokWeb()
	RegisterGeminiWeb()
	RegisterPerplexityWeb()
	RegisterBlackboxWeb()
	RegisterMuseSparkWeb()
	RegisterClaudeWeb()
	RegisterDeepSeekWeb()
	RegisterCopilotWeb()
	RegisterCopilotM365Web()
	RegisterT3Web()
	RegisterInnerAI()
	RegisterAdaptaWeb()
	RegisterLMArena()
	RegisterYuanbaoWeb()
	RegisterHuggingChat()
	RegisterPoeWeb()
	RegisterVeniceWeb()
	RegisterV0VercelWeb()
	RegisterKimiWeb()
	RegisterDoubaoWeb()
	RegisterQwenWeb()
	RegisterGeminiBusiness()
	RegisterZenmuxFree()
}

// RegisterChatGPTWeb registers the ChatGPT Web (Plus/Pro) provider.
func RegisterChatGPTWeb() {
	Register(&RegistryEntry{
		ID:                "chatgpt-web",
		Alias:             "cgpt-web",
		Name:              "ChatGPT Web (Plus/Pro)",
		BaseURL:           "https://chatgpt.com/backend-api/conversation",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterGrokWeb registers the Grok Web (Subscription) provider.
func RegisterGrokWeb() {
	Register(&RegistryEntry{
		ID:                "grok-web",
		Alias:             "gw",
		Name:              "Grok Web (Subscription)",
		BaseURL:           "https://grok.com/rest/app/chat",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterGeminiWeb registers the Gemini Web (Free) provider.
func RegisterGeminiWeb() {
	Register(&RegistryEntry{
		ID:                "gemini-web",
		Alias:             "gweb",
		Name:              "Gemini Web (Free)",
		BaseURL:           "https://gemini.google.com",
		Format:            FormatGemini,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterPerplexityWeb registers the Perplexity Web (Pro/Max) provider.
func RegisterPerplexityWeb() {
	Register(&RegistryEntry{
		ID:                "perplexity-web",
		Alias:             "pplx-web",
		Name:              "Perplexity Web (Pro/Max)",
		BaseURL:           "https://www.perplexity.ai",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterBlackboxWeb registers the Blackbox Web (Subscription) provider.
func RegisterBlackboxWeb() {
	Register(&RegistryEntry{
		ID:                "blackbox-web",
		Alias:             "bb-web",
		Name:              "Blackbox Web (Subscription)",
		BaseURL:           "https://www.blackbox.ai",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterMuseSparkWeb registers the Muse Spark Web (Meta AI) provider.
func RegisterMuseSparkWeb() {
	Register(&RegistryEntry{
		ID:                "muse-spark-web",
		Alias:             "ms-web",
		Name:              "Muse Spark Web (Meta AI)",
		BaseURL:           "https://www.meta.ai/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterClaudeWeb registers the Claude Web provider.
// Distinct from the claude OAuth provider (claude-code / claude).
func RegisterClaudeWeb() {
	Register(&RegistryEntry{
		ID:                "claude-web",
		Alias:             "cw",
		Name:              "Claude Web",
		BaseURL:           "https://claude.ai/api",
		Format:            FormatClaude,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterDeepSeekWeb registers the DeepSeek Web provider.
func RegisterDeepSeekWeb() {
	Register(&RegistryEntry{
		ID:                "deepseek-web",
		Alias:             "ds-web",
		Name:              "DeepSeek Web",
		BaseURL:           "https://chat.deepseek.com/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterCopilotWeb registers the Microsoft Copilot Web provider.
func RegisterCopilotWeb() {
	Register(&RegistryEntry{
		ID:                "copilot-web",
		Alias:             "copilot",
		Name:              "Microsoft Copilot Web",
		BaseURL:           "https://copilot.microsoft.com",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterCopilotM365Web registers the Microsoft 365 Copilot (BizChat) provider.
func RegisterCopilotM365Web() {
	Register(&RegistryEntry{
		ID:                "copilot-m365-web",
		Alias:             "m365copilot",
		Name:              "Microsoft 365 Copilot (BizChat)",
		BaseURL:           "https://m365.cloud.microsoft/chat",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterT3Web registers the t3.chat (Pro/Free) provider.
func RegisterT3Web() {
	Register(&RegistryEntry{
		ID:                "t3-web",
		Alias:             "t3chat",
		Name:              "t3.chat (Pro/Free)",
		BaseURL:           "https://t3.chat/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterInnerAI registers the Inner.ai (Subscription) provider.
func RegisterInnerAI() {
	Register(&RegistryEntry{
		ID:                "inner-ai",
		Alias:             "in-ai",
		Name:              "Inner.ai (Subscription)",
		BaseURL:           "https://app.innerai.com/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterAdaptaWeb registers the Adapta.org (Adapta One Web) provider.
func RegisterAdaptaWeb() {
	Register(&RegistryEntry{
		ID:                "adapta-web",
		Alias:             "adp-web",
		Name:              "Adapta.org (Adapta One Web)",
		BaseURL:           "https://agent.adapta.one",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterLMArena registers the LMArena (Free) provider.
func RegisterLMArena() {
	Register(&RegistryEntry{
		ID:                "lmarena",
		Alias:             "lma",
		Name:              "LMArena (Free)",
		BaseURL:           "https://lmarena.ai/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterYuanbaoWeb registers the Tencent Yuanbao (Free) provider.
func RegisterYuanbaoWeb() {
	Register(&RegistryEntry{
		ID:                "yuanbao-web",
		Alias:             "ybw",
		Name:              "Tencent Yuanbao (Free)",
		BaseURL:           "https://yuanbao.tencent.com/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterHuggingChat registers the HuggingChat (Free) provider.
func RegisterHuggingChat() {
	Register(&RegistryEntry{
		ID:                "huggingchat",
		Name:              "HuggingChat (Free)",
		BaseURL:           "https://huggingface.co/chat/conversation",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterPoeWeb registers the Poe Web (Subscription) provider.
func RegisterPoeWeb() {
	Register(&RegistryEntry{
		ID:                "poe-web",
		Alias:             "poew",
		Name:              "Poe Web (Subscription)",
		BaseURL:           "https://poe.com/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterVeniceWeb registers the Venice Web (Privacy) provider.
func RegisterVeniceWeb() {
	Register(&RegistryEntry{
		ID:                "venice-web",
		Alias:             "ven",
		Name:              "Venice Web (Privacy)",
		BaseURL:           "https://venice.ai/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterV0VercelWeb registers the v0 Vercel Web (Code Gen) provider.
func RegisterV0VercelWeb() {
	Register(&RegistryEntry{
		ID:                "v0-vercel-web",
		Alias:             "v0",
		Name:              "v0 Vercel Web (Code Gen)",
		BaseURL:           "https://v0.dev/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterKimiWeb registers the Kimi Web (Moonshot AI) provider.
func RegisterKimiWeb() {
	Register(&RegistryEntry{
		ID:                "kimi-web",
		Alias:             "kimi-web",
		Name:              "Kimi Web (Moonshot AI)",
		BaseURL:           "https://www.kimi.com/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterDoubaoWeb registers the Doubao Web (ByteDance) provider.
func RegisterDoubaoWeb() {
	Register(&RegistryEntry{
		ID:                "doubao-web",
		Alias:             "db",
		Name:              "Doubao Web (ByteDance)",
		BaseURL:           "https://www.doubao.com/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
	})
}

// RegisterQwenWeb registers the Qwen Web (Free) provider.
func RegisterQwenWeb() {
	Register(&RegistryEntry{
		ID:                "qwen-web",
		Alias:             "qwen-web",
		Name:              "Qwen Web (Free)",
		BaseURL:           "https://chat.qwen.ai/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterGeminiBusiness registers the Gemini Business (Enterprise) provider.
func RegisterGeminiBusiness() {
	Register(&RegistryEntry{
		ID:                "gemini-business",
		Alias:             "gembiz",
		Name:              "Gemini Business (Enterprise)",
		BaseURL:           "https://business.gemini.google",
		Format:            FormatGemini,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}

// RegisterZenmuxFree registers the ZenMux Free (Web) provider.
func RegisterZenmuxFree() {
	Register(&RegistryEntry{
		ID:                "zenmux-free",
		Alias:             "zmf",
		Name:              "ZenMux Free (Web)",
		BaseURL:           "https://zenmux.ai/api",
		Format:            FormatOpenAI,
		AuthType:          AuthTypeOAuth,
		PassthroughModels: true,
		HasFree:           true,
	})
}
