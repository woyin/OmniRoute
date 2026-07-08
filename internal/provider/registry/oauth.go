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

// RegisterQoder registers the Qoder free/OAuth provider.
func RegisterQoder() {
	Register(&RegistryEntry{
		ID:     "qoder",
		Name:   "Qoder AI",
		Format: FormatOpenAI,
		Executor: "default",
		BaseURL: "https://api.qoder.ai/v1",
		ChatPath: "/chat/completions",
		AuthType: AuthTypeOAuth,
		DefaultContextLength: 131072,
		PassthroughModels: true,
		HasFree: true,
		Models: []RegistryModel{
			{ID: "deepseek-v4-pro", Name: "DeepSeek V4 Pro (Qoder)", ContextLength: 131072, SupportsReasoning: true},
			{ID: "gpt-5.4", Name: "GPT-5.4 (Qoder)", ContextLength: 128000},
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
