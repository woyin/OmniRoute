package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Port != 3456 {
		t.Errorf("expected default port 3456, got %d", cfg.Port)
	}
	if cfg.RequireApiKey != false {
		t.Error("expected RequireApiKey=false by default")
	}
	if cfg.FetchTimeoutMs != 120000 {
		t.Errorf("expected default FetchTimeoutMs 120000, got %d", cfg.FetchTimeoutMs)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")

	os.Setenv("REQUIRE_API_KEY", "true")
	defer os.Unsetenv("REQUIRE_API_KEY")

	cfg := Load()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080 from env, got %d", cfg.Port)
	}
	if cfg.RequireApiKey != true {
		t.Error("expected RequireApiKey=true from env")
	}
}

func TestDurationHelpers(t *testing.T) {
	cfg := &Config{FetchTimeoutMs: 60000}
	if cfg.FetchTimeout() != 60*time.Second {
		t.Errorf("expected 60s, got %v", cfg.FetchTimeout())
	}
}

func TestResolveDataDir(t *testing.T) {
	cfg := &Config{DataDir: "/tmp/omniroute-test"}
	dir := cfg.ResolveDataDir()
	if dir != "/tmp/omniroute-test" {
		t.Errorf("expected /tmp/omniroute-test, got %s", dir)
	}
}

func TestEnvBool(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"false", false},
		{"0", false},
		{"", false},
		{"random", false},
	}
	for _, test := range tests {
		os.Setenv("TEST_BOOL", test.value)
		result := envBool("TEST_BOOL", false)
		if result != test.expected {
			t.Errorf("envBool(%q) = %v, want %v", test.value, result, test.expected)
		}
		os.Unsetenv("TEST_BOOL")
	}
}

func TestEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	result := envInt("TEST_INT", 0)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}

	// Invalid int
	os.Setenv("TEST_INT", "not-a-number")
	result = envInt("TEST_INT", 99)
	if result != 99 {
		t.Errorf("expected fallback 99, got %d", result)
	}
}
