package routing

import (
	"testing"
	"time"

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

// resetTrackingMaps clears all routing state between tests.
func resetTrackingMaps() {
	lkgpMu.Lock()
	lastGoodProvider = make(map[string]string)
	lkgpMu.Unlock()

	strictRandMu.Lock()
	lastRandomPick = make(map[string]string)
	strictRandMu.Unlock()

	metricsMu.Lock()
	providerMetrics = make(map[string]ProviderMetrics)
	metricsMu.Unlock()
}

// helper to build a standard 3-provider test combo with connections.
func makeThreeProviderCombo(strategy string) (db.Combo, []db.ProviderConnection) {
	combo := db.Combo{
		ID:       "test-" + strategy,
		Strategy: strategy,
		IsActive: true,
		Targets: []db.ComboTarget{
			{Provider: "alpha", Model: "model-a", Weight: 1},
			{Provider: "beta", Model: "model-b", Weight: 1},
			{Provider: "gamma", Model: "model-c", Weight: 1},
		},
	}
	connections := []db.ProviderConnection{
		{ID: "c1", Provider: "alpha", IsActive: true, Priority: 10, APIKey: "key-alpha"},
		{ID: "c2", Provider: "beta", IsActive: true, Priority: 5, APIKey: "key-beta"},
		{ID: "c3", Provider: "gamma", IsActive: true, Priority: 1, APIKey: "key-gamma"},
	}
	return combo, connections
}

// --- strict-random tests ---

func TestResolveStrictRandom(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("strict-random")

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	firstPick := targets[0].Provider

	// Resolve again — the first pick should NOT be the same as before.
	targets2 := ResolveTargets(combo, connections)
	if len(targets2) != 3 {
		t.Fatalf("expected 3 targets on second call, got %d", len(targets2))
	}
	if targets2[0].Provider == firstPick {
		t.Errorf("strict-random: second call first pick (%s) should differ from first call (%s)",
			targets2[0].Provider, firstPick)
	}
}

func TestResolveStrictRandom_SingleTarget(t *testing.T) {
	resetTrackingMaps()
	combo := db.Combo{
		ID:       "strict-single",
		Strategy: "strict-random",
		IsActive: true,
		Targets: []db.ComboTarget{
			{Provider: "only-one", Model: "model-x"},
		},
	}
	connections := []db.ProviderConnection{
		{ID: "c1", Provider: "only-one", IsActive: true, Priority: 5, APIKey: "k1"},
	}

	targets := ResolveTargets(combo, connections)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Provider != "only-one" {
		t.Errorf("expected provider=only-one, got %s", targets[0].Provider)
	}
}

// --- LKG tests ---

func TestResolveLKGP_NoHistory(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("lkgp")

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	// No history — should return priority order (alpha first).
	if targets[0].Provider != "alpha" {
		t.Errorf("expected priority fallback first=alpha, got %s", targets[0].Provider)
	}
}

func TestResolveLKGP_WithHistory(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("lkgp")

	// Record that "gamma" was the last good provider.
	RecordSuccess(combo.ID, "gamma")

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	if targets[0].Provider != "gamma" {
		t.Errorf("expected LKG provider=gamma first, got %s", targets[0].Provider)
	}
	// The other two should still be present.
	providers := map[string]bool{}
	for _, tgt := range targets {
		providers[tgt.Provider] = true
	}
	if !providers["alpha"] || !providers["beta"] || !providers["gamma"] {
		t.Errorf("expected all 3 providers present, got %v", providers)
	}
}

func TestResolveLKGP_SingleTarget(t *testing.T) {
	resetTrackingMaps()
	combo := db.Combo{
		ID:       "lkgp-single",
		Strategy: "lkgp",
		IsActive: true,
		Targets: []db.ComboTarget{
			{Provider: "solo", Model: "model-x"},
		},
	}
	connections := []db.ProviderConnection{
		{ID: "c1", Provider: "solo", IsActive: true, Priority: 5, APIKey: "k1"},
	}

	targets := ResolveTargets(combo, connections)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
}

// --- context-relay tests ---

func TestResolveContextRelay(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("context-relay")

	// alpha: 128k context, 100k used → 28k headroom
	// beta: 32k context, 1k used → 31k headroom
	// gamma: 64k context, 10k used → 54k headroom
	SetProviderMetrics("alpha", ProviderMetrics{ContextLength: 128000, UsedTokens: 100000})
	SetProviderMetrics("beta", ProviderMetrics{ContextLength: 32000, UsedTokens: 1000})
	SetProviderMetrics("gamma", ProviderMetrics{ContextLength: 64000, UsedTokens: 10000})

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	// Expected order: gamma (54k) > beta (31k) > alpha (28k)
	if targets[0].Provider != "gamma" {
		t.Errorf("expected first=gamma (most context headroom), got %s", targets[0].Provider)
	}
	if targets[1].Provider != "beta" {
		t.Errorf("expected second=beta, got %s", targets[1].Provider)
	}
	if targets[2].Provider != "alpha" {
		t.Errorf("expected third=alpha (least headroom), got %s", targets[2].Provider)
	}
}

func TestResolveContextRelay_NoMetrics(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("context-relay")
	// No metrics set — all headroom is 0, order should remain stable (priority order).
	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	if targets[0].Provider != "alpha" {
		t.Errorf("expected priority fallback first=alpha, got %s", targets[0].Provider)
	}
}

// --- headroom tests ---

func TestResolveHeadroom(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("headroom")

	// alpha: 1000 limit, 900 used → 100 headroom
	// beta: 500 limit, 100 used → 400 headroom
	// gamma: 200 limit, 50 used → 150 headroom
	SetProviderMetrics("alpha", ProviderMetrics{QuotaLimit: 1000, QuotaUsed: 900})
	SetProviderMetrics("beta", ProviderMetrics{QuotaLimit: 500, QuotaUsed: 100})
	SetProviderMetrics("gamma", ProviderMetrics{QuotaLimit: 200, QuotaUsed: 50})

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	// Expected order: beta (400) > gamma (150) > alpha (100)
	if targets[0].Provider != "beta" {
		t.Errorf("expected first=beta (most headroom), got %s", targets[0].Provider)
	}
	if targets[1].Provider != "gamma" {
		t.Errorf("expected second=gamma, got %s", targets[1].Provider)
	}
	if targets[2].Provider != "alpha" {
		t.Errorf("expected third=alpha (least headroom), got %s", targets[2].Provider)
	}
}

func TestResolveHeadroom_NoMetrics(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("headroom")
	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	// No metrics → all headroom is 0 → stable sort preserves priority order.
	if targets[0].Provider != "alpha" {
		t.Errorf("expected priority fallback first=alpha, got %s", targets[0].Provider)
	}
}

// --- fusion tests ---

func TestResolveFusion(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("fusion")

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	// Fusion returns all targets in priority order.
	if targets[0].Provider != "alpha" {
		t.Errorf("expected first=alpha (highest priority), got %s", targets[0].Provider)
	}
	if targets[1].Provider != "beta" {
		t.Errorf("expected second=beta, got %s", targets[1].Provider)
	}
	if targets[2].Provider != "gamma" {
		t.Errorf("expected third=gamma, got %s", targets[2].Provider)
	}
}

// --- reset-aware tests ---

func TestResolveResetAware(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("reset-aware")

	now := time.Now()
	// alpha: resets in 2 minutes (near reset → deprioritized)
	// beta: resets in 30 minutes (normal)
	// gamma: resets in 1 minute (near reset → deprioritized)
	SetProviderMetrics("alpha", ProviderMetrics{ResetAt: now.Add(2 * time.Minute)})
	SetProviderMetrics("beta", ProviderMetrics{ResetAt: now.Add(30 * time.Minute)})
	SetProviderMetrics("gamma", ProviderMetrics{ResetAt: now.Add(1 * time.Minute)})

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	// beta should be first (not near reset), then alpha and gamma (near reset).
	if targets[0].Provider != "beta" {
		t.Errorf("expected first=beta (not near reset), got %s", targets[0].Provider)
	}
	// alpha and gamma should be after beta.
	if targets[1].Provider == "beta" {
		t.Errorf("expected beta to be first only, but found it at position 2")
	}
}

func TestResolveResetAware_AllNormal(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("reset-aware")

	future := time.Now().Add(1 * time.Hour)
	SetProviderMetrics("alpha", ProviderMetrics{ResetAt: future})
	SetProviderMetrics("beta", ProviderMetrics{ResetAt: future})
	SetProviderMetrics("gamma", ProviderMetrics{ResetAt: future})

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	// All normal → priority order.
	if targets[0].Provider != "alpha" {
		t.Errorf("expected first=alpha (priority), got %s", targets[0].Provider)
	}
}

func TestResolveResetAware_AllNearReset(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("reset-aware")

	soon := time.Now().Add(1 * time.Minute)
	SetProviderMetrics("alpha", ProviderMetrics{ResetAt: soon})
	SetProviderMetrics("beta", ProviderMetrics{ResetAt: soon})
	SetProviderMetrics("gamma", ProviderMetrics{ResetAt: soon})

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	// All near reset → all demoted but still returned.
	providers := map[string]bool{}
	for _, tgt := range targets {
		providers[tgt.Provider] = true
	}
	if !providers["alpha"] || !providers["beta"] || !providers["gamma"] {
		t.Errorf("expected all providers present, got %v", providers)
	}
}

// --- reset-window tests ---

func TestResolveResetWindow(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("reset-window")

	now := time.Now()
	// alpha: reset 2 minutes ago (in reset window → excluded)
	// beta: reset 30 minutes ago (outside window → included)
	// gamma: reset 1 minute ago (in reset window → excluded)
	SetProviderMetrics("alpha", ProviderMetrics{ResetAt: now.Add(-2 * time.Minute)})
	SetProviderMetrics("beta", ProviderMetrics{ResetAt: now.Add(-30 * time.Minute)})
	SetProviderMetrics("gamma", ProviderMetrics{ResetAt: now.Add(-1 * time.Minute)})

	targets := ResolveTargets(combo, connections)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target (only beta outside reset window), got %d", len(targets))
	}
	if targets[0].Provider != "beta" {
		t.Errorf("expected beta (outside reset window), got %s", targets[0].Provider)
	}
}

func TestResolveResetWindow_AllInWindow(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("reset-window")

	recent := time.Now().Add(-1 * time.Minute)
	SetProviderMetrics("alpha", ProviderMetrics{ResetAt: recent})
	SetProviderMetrics("beta", ProviderMetrics{ResetAt: recent})
	SetProviderMetrics("gamma", ProviderMetrics{ResetAt: recent})

	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets (fallback to all), got %d", len(targets))
	}
	// All in reset window → fall back to returning all in priority order.
	if targets[0].Provider != "alpha" {
		t.Errorf("expected priority fallback first=alpha, got %s", targets[0].Provider)
	}
}

func TestResolveResetWindow_NoMetrics(t *testing.T) {
	resetTrackingMaps()
	combo, connections := makeThreeProviderCombo("reset-window")

	// No metrics → ResetAt is zero → not in reset window → all included.
	targets := ResolveTargets(combo, connections)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
}

// --- default / unknown strategy test ---

func TestResolveUnknownStrategy(t *testing.T) {
	resetTrackingMaps()
	combo := db.Combo{
		ID:       "unknown-strategy-combo",
		Strategy: "nonexistent-strategy",
		IsActive: true,
		Targets: []db.ComboTarget{
			{Provider: "alpha", Model: "model-a"},
			{Provider: "beta", Model: "model-b"},
		},
	}
	connections := []db.ProviderConnection{
		{ID: "c1", Provider: "alpha", IsActive: true, Priority: 10, APIKey: "k1"},
		{ID: "c2", Provider: "beta", IsActive: true, Priority: 5, APIKey: "k2"},
	}

	targets := ResolveTargets(combo, connections)
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	// Unknown strategy falls back to priority.
	if targets[0].Provider != "alpha" {
		t.Errorf("expected first=alpha (priority fallback), got %s", targets[0].Provider)
	}
}

// --- provider metrics API tests ---

func TestProviderMetricsCRUD(t *testing.T) {
	resetTrackingMaps()

	m := ProviderMetrics{QuotaLimit: 1000, QuotaUsed: 500, ContextLength: 32000}
	SetProviderMetrics("test-provider", m)

	got, ok := GetProviderMetrics("test-provider")
	if !ok {
		t.Fatal("expected metrics to exist")
	}
	if got.QuotaLimit != 1000 {
		t.Errorf("expected QuotaLimit=1000, got %d", got.QuotaLimit)
	}
	if got.QuotaUsed != 500 {
		t.Errorf("expected QuotaUsed=500, got %d", got.QuotaUsed)
	}

	ClearProviderMetrics("test-provider")
	_, ok = GetProviderMetrics("test-provider")
	if ok {
		t.Error("expected metrics to be cleared")
	}
}

// --- strategy constant validation ---

func TestAllStrategyConstants(t *testing.T) {
	strategies := []Strategy{
		StrategyPriority,
		StrategyWeighted,
		StrategyFillFirst,
		StrategyRoundRobin,
		StrategyP2C,
		StrategyRandom,
		StrategyLeastUsed,
		StrategyCostOptimized,
		StrategyAuto,
		StrategyContextOptimized,
		StrategyStrictRandom,
		StrategyLKGP,
		StrategyContextRelay,
		StrategyHeadroom,
		StrategyFusion,
		StrategyResetAware,
		StrategyResetWindow,
	}
	if len(strategies) != 17 {
		t.Errorf("expected 17 strategy constants, got %d", len(strategies))
	}

	// Verify string values for the new strategies.
	expected := map[Strategy]string{
		StrategyStrictRandom: "strict-random",
		StrategyLKGP:        "lkgp",
		StrategyContextRelay: "context-relay",
		StrategyHeadroom:    "headroom",
		StrategyFusion:      "fusion",
		StrategyResetAware:  "reset-aware",
		StrategyResetWindow: "reset-window",
	}
	for s, want := range expected {
		if string(s) != want {
			t.Errorf("expected strategy %q, got %q", want, string(s))
		}
	}
}
