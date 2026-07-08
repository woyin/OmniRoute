package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server
	Port            int
	DataDir         string
	RequireApiKey   bool

	// Timeouts
	FetchTimeoutMs       int
	StreamIdleTimeoutMs  int
	StreamReadinessMs    int
	StreamReadinessMaxMs int
	SSEHeartbeatMs       int
	FetchBodyTimeoutMs   int

	// DB
	SQLiteFile string

	// Auth
	CodexOAuthClientID     string
	CodexOAuthClientSecret string
	CodexOAuthTokenURL     string

	// Feature flags
	OpenCodeSynthesizeCliHeaders bool
}

var defaultConfig = Config{
	Port:                         3456,
	DataDir:                      "",
	RequireApiKey:                false,
	FetchTimeoutMs:               120000,
	StreamIdleTimeoutMs:          60000,
	StreamReadinessMs:            60000,
	StreamReadinessMaxMs:         180000,
	SSEHeartbeatMs:               15000,
	FetchBodyTimeoutMs:           120000,
	CodexOAuthTokenURL:           "https://auth.openai.com/oauth/token",
	OpenCodeSynthesizeCliHeaders: false,
}

// Load reads configuration from environment variables and defaults.
func Load() *Config {
	c := &Config{}
	*c = defaultConfig

	c.Port = envInt("PORT", c.Port)
	c.DataDir = envStr("DATA_DIR", c.DataDir)
	c.RequireApiKey = envBool("REQUIRE_API_KEY", c.RequireApiKey)
	c.FetchTimeoutMs = envInt("FETCH_TIMEOUT_MS", c.FetchTimeoutMs)
	c.StreamIdleTimeoutMs = envInt("STREAM_IDLE_TIMEOUT_MS", c.StreamIdleTimeoutMs)
	c.StreamReadinessMs = envInt("STREAM_READINESS_TIMEOUT_MS", c.StreamReadinessMs)
	c.StreamReadinessMaxMs = envInt("STREAM_READINESS_MAX_TIMEOUT_MS", c.StreamReadinessMaxMs)
	c.SSEHeartbeatMs = envInt("SSE_HEARTBEAT_INTERVAL_MS", c.SSEHeartbeatMs)
	c.FetchBodyTimeoutMs = envInt("FETCH_BODY_TIMEOUT_MS", c.FetchBodyTimeoutMs)
	c.CodexOAuthClientID = envStr("CODEX_OAUTH_CLIENT_ID", c.CodexOAuthClientID)
	c.CodexOAuthClientSecret = envStr("CODEX_OAUTH_CLIENT_SECRET", c.CodexOAuthClientSecret)
	c.CodexOAuthTokenURL = envStr("CODEX_OAUTH_TOKEN_URL", c.CodexOAuthTokenURL)
	c.OpenCodeSynthesizeCliHeaders = envBool("OPENCODE_SYNTHESIZE_CLI_HEADERS", c.OpenCodeSynthesizeCliHeaders)

	if c.DataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/tmp"
		}
		c.DataDir = homeDir + "/.omniroute"
	}
	c.SQLiteFile = c.DataDir + "/storage.sqlite"

	return c
}

// ResolveDataDir returns the configured data directory, creating it if needed.
func (c *Config) ResolveDataDir() string {
	os.MkdirAll(c.DataDir, 0o755)
	return c.DataDir
}

func envStr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	lower := strings.ToLower(v)
	return lower == "true" || lower == "1" || lower == "yes" || lower == "on"
}

// Duration helpers
func (c *Config) FetchTimeout() time.Duration {
	return time.Duration(c.FetchTimeoutMs) * time.Millisecond
}

func (c *Config) StreamIdleTimeout() time.Duration {
	return time.Duration(c.StreamIdleTimeoutMs) * time.Millisecond
}

func (c *Config) StreamReadinessTimeout() time.Duration {
	return time.Duration(c.StreamReadinessMs) * time.Millisecond
}

func (c *Config) SSEHeartbeatInterval() time.Duration {
	return time.Duration(c.SSEHeartbeatMs) * time.Millisecond
}

func (c *Config) FetchBodyTimeout() time.Duration {
	return time.Duration(c.FetchBodyTimeoutMs) * time.Millisecond
}
