package routing

import (
	"testing"

	"github.com/omniroute/omniroute/internal/db"
)

func TestResolvePriority(t *testing.T) {
	combo := db.Combo{
		ID:       "test-combo",
		Strategy: "priority",
		IsActive: true,
		Targets: []db.ComboTarget{
			{Provider: "opencode", Model: "deepseek-v4-flash-free", Weight: 1},
			{Provider: "ollama-cloud", Model: "deepseek-v4-pro", Weight: 2},
		},
	}

	connections := []db.ProviderConnection{
		{ID: "conn1", Provider: "opencode", IsActive: true, Priority: 10, APIKey: "key1"},
		{ID: "conn2", Provider: "ollama-cloud", IsActive: true, Priority: 5, APIKey: "key2"},
	}

	targets := ResolveTargets(combo, connections)
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	// Priority order: conn1 has higher priority (10 > 5)
	if targets[0].Provider != "opencode" {
		t.Errorf("expected first target provider=opencode, got %s", targets[0].Provider)
	}
	if targets[0].APIKey != "key1" {
		t.Errorf("expected first target APIKey=key1, got %s", targets[0].APIKey)
	}
}

func TestResolveWeighted(t *testing.T) {
	combo := db.Combo{
		ID:       "weighted-combo",
		Strategy: "weighted",
		IsActive: true,
		Targets: []db.ComboTarget{
			{Provider: "opencode", Model: "deepseek-v4-flash-free", Weight: 3},
			{Provider: "ollama-cloud", Model: "deepseek-v4-pro", Weight: 1},
		},
	}

	connections := []db.ProviderConnection{
		{ID: "conn1", Provider: "opencode", IsActive: true, APIKey: "key1"},
		{ID: "conn2", Provider: "ollama-cloud", IsActive: true, APIKey: "key2"},
	}

	targets := ResolveTargets(combo, connections)
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0].Weight != 3 {
		t.Errorf("expected first target weight=3, got %d", targets[0].Weight)
	}
}

func TestResolveEmpty(t *testing.T) {
	combo := db.Combo{
		ID:       "empty-combo",
		Strategy: "priority",
		IsActive: true,
		Targets:  []db.ComboTarget{},
	}

	targets := ResolveTargets(combo, nil)
	if len(targets) != 0 {
		t.Errorf("expected 0 targets for empty combo, got %d", len(targets))
	}
}

func TestResolveWithAccount(t *testing.T) {
	combo := db.Combo{
		ID:       "account-combo",
		Strategy: "priority",
		IsActive: true,
		Targets: []db.ComboTarget{
			{Provider: "opencode", Model: "deepseek-v4-flash-free", Account: "conn1"},
		},
	}

	connections := []db.ProviderConnection{
		{ID: "conn1", Provider: "opencode", IsActive: true, APIKey: "key1", Priority: 5},
	}

	targets := ResolveTargets(combo, connections)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].APIKey != "key1" {
		t.Errorf("expected APIKey=key1, got %s", targets[0].APIKey)
	}
}
