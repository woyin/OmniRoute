package registry

// RegisterOAuthProviders registers all OAuth-based providers.
// These providers use OAuth token exchange flows for authentication.
func RegisterOAuthProviders() {
	RegisterCursor()
	RegisterKiro()
	RegisterQoder()
	RegisterGitHubCopilot()
	RegisterWindsurf()
	RegisterClaudeCode()
	RegisterAntigravity()
	RegisterClaude()
	RegisterGitHub()
	RegisterQwen()
	RegisterGrokCLI()
	RegisterKimiCoding()
	RegisterKilocode()
	RegisterCodebuddyCN()
	RegisterAGY()
	RegisterGitLabDuo()
	RegisterDevinCLI()
	RegisterAmazonQ()
	RegisterZedHosted()
}

// RegisterCursor registers the Cursor OAuth provider.
func RegisterCursor() {
	Register(&RegistryEntry{
		ID:     "cursor",
		Name:   "Cursor",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api2.cursor.sh/aiserver.v1.AiService/StreamChat",
		AuthType: AuthTypeOAuth,
		DefaultContextLength: 200000,
		Models: []RegistryModel{
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7 (Cursor)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
			{ID: "gpt-5.5", Name: "GPT-5.5 (Cursor)", ContextLength: 128000, SupportsReasoning: true},
			{ID: "gpt-5.4", Name: "GPT-5.4 (Cursor)", ContextLength: 128000, SupportsReasoning: true},
		},
	})
}

// RegisterKiro registers the Kiro OAuth provider.
func RegisterKiro() {
	Register(&RegistryEntry{
		ID:     "kiro",
		Name:   "Kiro",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.kiro.dev/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeOAuth,
		DefaultContextLength: 200000,
		Models: []RegistryModel{
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (Kiro)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7 (Kiro)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
		},
	})
}

// RegisterQoder registers the Qoder OAuth provider.
func RegisterQoder() {
	Register(&RegistryEntry{
		ID:       "qoder",
		Alias:    "if",
		Name:     "Qoder AI",
		Format:   FormatOpenAI,
		Executor: "qoder",
		BaseURL:  "https://api.qoder.com/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
		OAuth: &OAuthConfig{
			ClientIDEnv:     "QODER_OAUTH_CLIENT_ID",
			ClientSecretEnv: "QODER_OAUTH_CLIENT_SECRET",
		},
		Models: []RegistryModel{
			{ID: "qoder-rome-30ba3b", Name: "Qoder ROME (Qoder)", ContextLength: 131072},
			{ID: "glm-5.2", Name: "GLM-5.2 (Qoder)", ContextLength: 131072},
			{ID: "minimax-m3", Name: "MiniMax M3 (Qoder)", ContextLength: 131072, SupportsVision: true},
			{ID: "qwen3-coder-plus", Name: "Qwen3 Coder Plus (Qoder)", ContextLength: 131072},
			{ID: "qwen3-max", Name: "Qwen3 Max (Qoder)", ContextLength: 131072},
			{ID: "qwen3-vl-plus", Name: "Qwen3 Vision Plus (Qoder)", ContextLength: 131072, SupportsVision: true},
			{ID: "kimi-k2-0905", Name: "Kimi K2 0905 (Qoder)", ContextLength: 131072},
			{ID: "qwen3-max-preview", Name: "Qwen3 Max Preview (Qoder)", ContextLength: 131072},
			{ID: "kimi-k2", Name: "Kimi K2 (Qoder)", ContextLength: 131072},
			{ID: "deepseek-v3.2", Name: "DeepSeek-V3.2-Exp (Qoder)", ContextLength: 131072},
			{ID: "deepseek-r1", Name: "DeepSeek R1 (Qoder)", ContextLength: 131072, SupportsReasoning: true},
			{ID: "deepseek-v3", Name: "DeepSeek V3 (Qoder)", ContextLength: 131072},
			{ID: "qwen3-32b", Name: "Qwen3 32B (Qoder)", ContextLength: 131072},
			{ID: "qwen3-235b-a22b-thinking-2507", Name: "Qwen3 235B A22B Thinking 2507 (Qoder)", ContextLength: 131072, SupportsReasoning: true},
			{ID: "qwen3-235b-a22b-instruct", Name: "Qwen3 235B A22B Instruct (Qoder)", ContextLength: 131072},
			{ID: "qwen3-235b", Name: "Qwen3 235B (Qoder)", ContextLength: 131072},
		},
	})
}

// RegisterGitHubCopilot registers the GitHub Copilot OAuth provider.
func RegisterGitHubCopilot() {
	Register(&RegistryEntry{
		ID:     "github-copilot",
		Name:   "GitHub Copilot",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.githubcopilot.com",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeOAuth,
		DefaultContextLength: 128000,
		Models: []RegistryModel{
			{ID: "gpt-5.5", Name: "GPT-5.5 (Copilot)", ContextLength: 128000, SupportsReasoning: true},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (Copilot)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
			{ID: "gpt-5.4", Name: "GPT-5.4 (Copilot)", ContextLength: 128000},
		},
	})
}

// RegisterWindsurf registers the Windsurf OAuth provider.
func RegisterWindsurf() {
	Register(&RegistryEntry{
		ID:     "windsurf",
		Name:   "Windsurf",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://codeium.com/windsurf/chat",
		ChatPath: "/completions",
		AuthType: AuthTypeOAuth,
		DefaultContextLength: 200000,
		Models: []RegistryModel{
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (Windsurf)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
			{ID: "gpt-5.5", Name: "GPT-5.5 (Windsurf)", ContextLength: 128000, SupportsReasoning: true},
		},
	})
}

// RegisterClaudeCode registers the Claude Code OAuth provider.
func RegisterClaudeCode() {
	Register(&RegistryEntry{
		ID:     "claude-code",
		Name:   "Claude Code",
		Format: FormatClaude,
		Executor: "default",
		BaseURL: "https://api.anthropic.com/v1",
		ChatPath: "/messages",
		AuthType: AuthTypeOAuth,
		DefaultContextLength: 200000,
		Headers: map[string]string{
			"anthropic-version": "2023-06-01",
		},
		Models: []RegistryModel{
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7 (CC)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (CC)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
		},
	})
}

// RegisterAntigravity registers the Antigravity OAuth provider.
func RegisterAntigravity() {
	Register(&RegistryEntry{
		ID:     "antigravity",
		Name:   "Antigravity",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.antigravity.com/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeOAuth,
		DefaultContextLength: 200000,
		Models: []RegistryModel{
			{ID: "gpt-5.5", Name: "GPT-5.5 (Antigravity)", ContextLength: 128000, SupportsReasoning: true},
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7 (Antigravity)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
		},
	})
}

// RegisterClaude registers the Claude (OAuth) provider.
// Distinct from claude-code: this is the Claude.ai OAuth flow for direct API access.
func RegisterClaude() {
	Register(&RegistryEntry{
		ID:        "claude",
		Alias:     "cc",
		Name:      "Claude (OAuth)",
		Format:    FormatClaude,
		Executor:  "default",
		BaseURL:   "https://api.anthropic.com/v1",
		ChatPath:  "/messages",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "x-api-key",
		DefaultContextLength: 200000,
		Headers: map[string]string{
			"anthropic-version": "2023-06-01",
		},
		Models: []RegistryModel{
			{ID: "claude-opus-4-7", Name: "Claude Opus 4.7", SupportsReasoning: true, ContextLength: 200000},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6", SupportsReasoning: true, ContextLength: 200000},
			{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", SupportsReasoning: true, ContextLength: 200000},
			{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5", ContextLength: 200000},
		},
	})
}

// RegisterGitHub registers the GitHub Copilot (OAuth) provider.
// Distinct from github-copilot: uses the GitHub OAuth app flow directly.
func RegisterGitHub() {
	Register(&RegistryEntry{
		ID:        "github",
		Alias:     "gh",
		Name:      "GitHub Copilot (OAuth)",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://api.githubcopilot.com",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		Models: []RegistryModel{
			{ID: "gpt-5.5", Name: "GPT-5.5 (GitHub)", ContextLength: 128000, SupportsReasoning: true},
			{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6 (GitHub)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
		},
	})
}

// RegisterQwen registers the Qwen Code OAuth provider.
func RegisterQwen() {
	Register(&RegistryEntry{
		ID:        "qwen",
		Alias:     "qw",
		Name:      "Qwen Code",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://chat.qwen.ai/api/v1/services/aigc/text-generation/generation",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
	})
}

// RegisterGrokCLI registers the Grok Build (CLI) OAuth provider.
func RegisterGrokCLI() {
	Register(&RegistryEntry{
		ID:        "grok-cli",
		Alias:     "gc",
		Name:      "Grok Build",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://api.x.ai/v1",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		Models: []RegistryModel{
			{ID: "grok-4", Name: "Grok 4 (CLI)", ContextLength: 131072, SupportsReasoning: true},
			{ID: "grok-3", Name: "Grok 3 (CLI)", ContextLength: 131072, SupportsReasoning: true},
		},
	})
}

// RegisterKimiCoding registers the Kimi Coding OAuth provider.
func RegisterKimiCoding() {
	Register(&RegistryEntry{
		ID:        "kimi-coding",
		Alias:     "kmc",
		Name:      "Kimi Coding",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://api.kimi.com/coding/v1",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
	})
}

// RegisterKilocode registers the Kilocode OAuth provider.
// Kilo's gateway serves free models anonymously when no account is connected;
// HasFree flags this capability for the dashboard's anonymous-fallback logic.
func RegisterKilocode() {
	Register(&RegistryEntry{
		ID:        "kilocode",
		Alias:     "kc",
		Name:      "Kilocode",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://api.kilo.ai/api/openrouter/chat/completions",
		ModelsURL: "https://api.kilo.ai/api/openrouter/models",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		HasFree:   true,
		PassthroughModels: true,
		Headers: map[string]string{
			"X-KILOCODE-EDITORNAME": "OmniRoute",
		},
		Models: []RegistryModel{
			{ID: "openrouter/free", Name: "Free Models Router (Kilo)", ContextLength: 131072},
			{ID: "qwen/qwen3.6-plus", Name: "Qwen3.6 Plus (Kilo)", ContextLength: 131072},
			{ID: "qwen/qwen3.5-397b-a17b", Name: "Qwen3.5 397B A17B (Kilo)", ContextLength: 131072},
			{ID: "openai/gpt-5.5", Name: "GPT-5.5 (Kilo)", ContextLength: 128000, SupportsReasoning: true},
			{ID: "openai/gpt-5.4-mini", Name: "GPT-5.4 Mini (Kilo)", ContextLength: 128000},
			{ID: "anthropic/claude-opus-4.7", Name: "Claude Opus 4.7 (Kilo)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
			{ID: "anthropic/claude-sonnet-4.6", Name: "Claude Sonnet 4.6 (Kilo)", ContextLength: 200000, TargetFormat: FormatClaude, SupportsReasoning: true},
			{ID: "anthropic/claude-haiku-4.5", Name: "Claude Haiku 4.5 (Kilo)", ContextLength: 200000, TargetFormat: FormatClaude},
			{ID: "google/gemini-3.1-pro-preview", Name: "Gemini 3.1 Pro (Kilo)", ContextLength: 200000},
			{ID: "google/gemini-3-flash-preview", Name: "Gemini 3 Flash (Kilo)", ContextLength: 131072},
			{ID: "google/gemini-3.1-flash-lite", Name: "Gemini 3.1 Flash Lite (Kilo)", ContextLength: 131072},
			{ID: "deepseek/deepseek-v4-pro", Name: "DeepSeek V4 Pro (Kilo)", ContextLength: 131072, SupportsReasoning: true},
			{ID: "deepseek/deepseek-v4-flash", Name: "DeepSeek V4 Flash (Kilo)", ContextLength: 131072, SupportsReasoning: true},
			{ID: "moonshotai/kimi-k2.6", Name: "Kimi K2.6 (Kilo)", ContextLength: 131072},
		},
	})
}

// RegisterCodebuddyCN registers the CodeBuddy CN (Tencent) OAuth provider.
func RegisterCodebuddyCN() {
	Register(&RegistryEntry{
		ID:        "codebuddy-cn",
		Alias:     "cbcn",
		Name:      "CodeBuddy CN",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://copilot.tencent.com/v2",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 131072,
		PassthroughModels: true,
	})
}

// RegisterAGY registers the Antigravity CLI (AGY) OAuth provider.
func RegisterAGY() {
	Register(&RegistryEntry{
		ID:        "agy",
		Name:      "Antigravity CLI",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://api.antigravity.com/v1",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 200000,
		HasFree:   true,
		PassthroughModels: true,
	})
}

// RegisterGitLabDuo registers the GitLab Duo OAuth provider.
func RegisterGitLabDuo() {
	Register(&RegistryEntry{
		ID:        "gitlab-duo",
		Alias:     "gld",
		Name:      "GitLab Duo",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://gitlab.com/api/v4",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		PassthroughModels: true,
	})
}

// RegisterDevinCLI registers the Devin CLI OAuth provider.
func RegisterDevinCLI() {
	Register(&RegistryEntry{
		ID:        "devin-cli",
		Alias:     "dv",
		Name:      "Devin CLI",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://api.devin.ai/v1",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		PassthroughModels: true,
	})
}

// RegisterAmazonQ registers the Amazon Q OAuth provider.
func RegisterAmazonQ() {
	Register(&RegistryEntry{
		ID:        "amazon-q",
		Alias:     "aq",
		Name:      "Amazon Q",
		Format:    FormatOpenAI,
		Executor:  "default",
		BaseURL:   "https://q.us-east-1.amazonaws.com",
		ChatPath:  "/chat/completions",
		AuthType:  AuthTypeOAuth,
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		DefaultContextLength: 128000,
		HasFree:   true,
		PassthroughModels: true,
	})
}

// RegisterZedHosted registers the Zed Hosted Models OAuth provider.
// Zed's hosted catalog changes frequently and is fetched live per-connection,
// so models are not hardcoded here; modelsUrl feeds live discovery.
func RegisterZedHosted() {
	Register(&RegistryEntry{
		ID:                   "zed-hosted",
		Name:                 "Zed Hosted Models",
		Format:               FormatOpenAI,
		Executor:             "zed-hosted",
		BaseURL:              "https://cloud.zed.dev/completions",
		ModelsURL:            "https://cloud.zed.dev/models",
		AuthType:             AuthTypeOAuth,
		AuthHeader:           "Authorization",
		AuthPrefix:           "Bearer ",
		DefaultContextLength: 200000,
		PassthroughModels:    true,
	})
}
