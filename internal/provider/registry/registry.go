// Package registry provides a global provider registry for OmniRoute.
// Each provider entry describes an AI service endpoint including its
// authentication method, base URL, supported models, and wire format.
//
// Providers are registered at startup via RegisterBuiltinProviders() and
// can be queried by ID or alias.
package registry

import (
	"sync"
)

// AuthType enumerates provider authentication methods.
type AuthType string

const (
	AuthTypeAPIKey AuthType = "apikey"
	AuthTypeOAuth  AuthType = "oauth"
	AuthTypeNoAuth AuthType = "noauth"
)

// Format enumerates API wire formats.
type Format string

const (
	FormatOpenAI          Format = "openai"
	FormatOpenAIResponses Format = "openai-responses"
	FormatClaude          Format = "claude"
	FormatGemini          Format = "gemini"
)

// RegistryModel describes a model available through a provider.
type RegistryModel struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	ContextLength     int     `json:"contextLength,omitempty"`
	MaxInputTokens    int     `json:"maxInputTokens,omitempty"`
	MaxOutputTokens   int     `json:"maxOutputTokens,omitempty"`
	SupportsReasoning bool    `json:"supportsReasoning,omitempty"`
	SupportsVision    bool    `json:"supportsVision,omitempty"`
	SupportsXHighEffort bool  `json:"supportsXHighEffort,omitempty"`
	TargetFormat      Format  `json:"targetFormat,omitempty"`
}

// OAuthConfig describes OAuth parameters for a provider.
type OAuthConfig struct {
	ClientIDEnv       string `json:"clientIdEnv,omitempty"`
	ClientIDDefault   string `json:"clientIdDefault,omitempty"`
	ClientSecretEnv   string `json:"clientSecretEnv,omitempty"`
	ClientSecretDefault string `json:"clientSecretDefault,omitempty"`
	TokenURL          string `json:"tokenUrl,omitempty"`
	RefreshURL        string `json:"refreshUrl,omitempty"`
	AuthURL           string `json:"authUrl,omitempty"`
}

// RegistryEntry is the single source of truth for a provider's configuration.
type RegistryEntry struct {
	ID                   string          `json:"id"`
	Alias                string          `json:"alias,omitempty"`
	Name                 string          `json:"name,omitempty"`
	Format               Format          `json:"format"`
	Executor             string          `json:"executor,omitempty"`
	BaseURL              string          `json:"baseUrl,omitempty"`
	ChatPath             string          `json:"chatPath,omitempty"`
	ModelsURL            string          `json:"modelsUrl,omitempty"`
	AuthType             AuthType        `json:"authType"`
	AuthHeader           string          `json:"authHeader,omitempty"`
	AuthPrefix           string          `json:"authPrefix,omitempty"`
	DefaultContextLength int             `json:"defaultContextLength,omitempty"`
	Models               []RegistryModel `json:"models,omitempty"`
	OAuth                *OAuthConfig    `json:"oauth,omitempty"`
	Headers              map[string]string `json:"headers,omitempty"`
	PassthroughModels    bool            `json:"passthroughModels,omitempty"`
	HasFree              bool            `json:"hasFree,omitempty"`
	Deprecated           bool            `json:"deprecated,omitempty"`
	SystemOnly           bool            `json:"systemOnly,omitempty"`
}

// globalRegistry holds all provider entries.
var (
	globalRegistry = make(map[string]*RegistryEntry)
	registryMu     sync.RWMutex
)

// Register adds a provider entry to the global registry.
func Register(entry *RegistryEntry) {
	registryMu.Lock()
	defer registryMu.Unlock()
	globalRegistry[entry.ID] = entry
	if entry.Alias != "" {
		globalRegistry[entry.Alias] = entry
	}
}

// Get retrieves a provider entry by ID or alias.
func Get(id string) *RegistryEntry {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return globalRegistry[id]
}

// List returns all registered provider entries (by ID, not alias).
func List() []*RegistryEntry {
	registryMu.RLock()
	defer registryMu.RUnlock()
	seen := make(map[string]bool)
	var result []*RegistryEntry
	for _, entry := range globalRegistry {
		if seen[entry.ID] {
			continue
		}
		seen[entry.ID] = true
		result = append(result, entry)
	}
	return result
}

// GetModel finds a model by ID within a provider's model list.
func (e *RegistryEntry) GetModel(modelID string) *RegistryModel {
	for i := range e.Models {
		if e.Models[i].ID == modelID {
			return &e.Models[i]
		}
	}
	return nil
}
