package routing

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/omniroute/omniroute/internal/db"
)

// Strategy represents a routing strategy.
type Strategy string

const (
	StrategyPriority        Strategy = "priority"
	StrategyWeighted        Strategy = "weighted"
	StrategyFillFirst       Strategy = "fill-first"
	StrategyRoundRobin      Strategy = "round-robin"
	StrategyP2C             Strategy = "p2c"
	StrategyRandom          Strategy = "random"
	StrategyLeastUsed       Strategy = "least-used"
	StrategyCostOptimized   Strategy = "cost-optimized"
	StrategyAuto            Strategy = "auto"
	StrategyContextOptimized Strategy = "context-optimized"
	StrategyStrictRandom    Strategy = "strict-random"
	StrategyLKGP            Strategy = "lkgp"
	StrategyContextRelay    Strategy = "context-relay"
	StrategyHeadroom        Strategy = "headroom"
	StrategyFusion          Strategy = "fusion"
	StrategyResetAware      Strategy = "reset-aware"
	StrategyResetWindow     Strategy = "reset-window"
)

// ResolvedTarget is a fully resolved routing target.
type ResolvedTarget struct {
	Provider      string
	Model         string
	Account       string
	APIKey        string
	AccessToken   string
	Weight        int
	Priority      int
	QuotaLimit    int64
	QuotaUsed     int64
	ContextLength int64
	UsedTokens    int64
	ResetAt       time.Time
}

// ProviderMetrics carries runtime metrics for a provider, used by advanced
// routing strategies. Populate before calling ResolveTargets when using
// headroom, context-relay, reset-aware, or reset-window strategies.
type ProviderMetrics struct {
	QuotaLimit    int64
	QuotaUsed     int64
	ContextLength int64
	UsedTokens    int64
	ResetAt       time.Time
}

// metricsByProvider stores ProviderMetrics keyed by provider name.
// Callers may set this before ResolveTargets to supply live metrics.
var (
	metricsMu       sync.RWMutex
	providerMetrics map[string]ProviderMetrics
)

func init() {
	providerMetrics = make(map[string]ProviderMetrics)
	lastGoodProvider = make(map[string]string)
	lastRandomPick = make(map[string]string)
}

// SetProviderMetrics stores runtime metrics for a provider. Thread-safe.
func SetProviderMetrics(provider string, m ProviderMetrics) {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	providerMetrics[provider] = m
}

// GetProviderMetrics retrieves runtime metrics for a provider. Thread-safe.
func GetProviderMetrics(provider string) (ProviderMetrics, bool) {
	metricsMu.RLock()
	defer metricsMu.RUnlock()
	m, ok := providerMetrics[provider]
	return m, ok
}

// ClearProviderMetrics removes metrics for a provider. Thread-safe.
func ClearProviderMetrics(provider string) {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	delete(providerMetrics, provider)
}

// lastGoodProvider tracks the last provider that succeeded for each combo (LKG strategy).
var (
	lkgpMu           sync.RWMutex
	lastGoodProvider map[string]string
)

// lastRandomPick tracks the last randomly picked provider per combo (strict-random strategy).
var (
	strictRandMu   sync.RWMutex
	lastRandomPick map[string]string
)

// RecordSuccess records that a provider succeeded for a given combo.
// This feeds the LKG (Last Known Good Provider) strategy.
func RecordSuccess(comboID, provider string) {
	lkgpMu.Lock()
	defer lkgpMu.Unlock()
	lastGoodProvider[comboID] = provider
}

// ResolveTargets expands a combo configuration into an ordered list of routing targets.
func ResolveTargets(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	switch Strategy(combo.Strategy) {
	case StrategyPriority:
		return resolvePriority(combo, connections)
	case StrategyWeighted:
		return resolveWeighted(combo, connections)
	case StrategyRoundRobin:
		return resolveRoundRobin(combo, connections)
	case StrategyRandom:
		return resolveRandom(combo, connections)
	case StrategyFillFirst:
		return resolveFillFirst(combo, connections)
	case StrategyStrictRandom:
		return resolveStrictRandom(combo, connections)
	case StrategyLKGP:
		return resolveLKGP(combo, connections)
	case StrategyContextRelay:
		return resolveContextRelay(combo, connections)
	case StrategyHeadroom:
		return resolveHeadroom(combo, connections)
	case StrategyFusion:
		return resolveFusion(combo, connections)
	case StrategyResetAware:
		return resolveResetAware(combo, connections)
	case StrategyResetWindow:
		return resolveResetWindow(combo, connections)
	default:
		return resolvePriority(combo, connections)
	}
}

// buildTargets is a shared helper that constructs ResolvedTargets from combo
// targets and provider connections, enriching each target with metrics from
// the providerMetrics store.
func buildTargets(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	connMap := map[string]db.ProviderConnection{}
	for _, c := range connections {
		connMap[c.ID] = c
	}

	var targets []ResolvedTarget
	for _, t := range combo.Targets {
		target := ResolvedTarget{
			Provider: t.Provider,
			Model:    t.Model,
			Account:  t.Account,
		}

		if t.Account != "" {
			if conn, ok := connMap[t.Account]; ok {
				target.APIKey = conn.APIKey
				target.AccessToken = conn.AccessToken
				target.Priority = conn.Priority
			}
		} else {
			for _, conn := range connections {
				if conn.Provider == t.Provider && conn.IsActive {
					target.APIKey = conn.APIKey
					target.AccessToken = conn.AccessToken
					target.Priority = conn.Priority
					break
				}
			}
		}

		// Enrich with runtime metrics if available.
		metricsMu.RLock()
		if m, ok := providerMetrics[t.Provider]; ok {
			target.QuotaLimit = m.QuotaLimit
			target.QuotaUsed = m.QuotaUsed
			target.ContextLength = m.ContextLength
			target.UsedTokens = m.UsedTokens
			target.ResetAt = m.ResetAt
		}
		metricsMu.RUnlock()

		targets = append(targets, target)
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Priority > targets[j].Priority
	})
	return targets
}

func resolvePriority(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	return buildTargets(combo, connections)
}

func resolveWeighted(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	connMap := map[string]db.ProviderConnection{}
	for _, c := range connections {
		connMap[c.ID] = c
	}

	var targets []ResolvedTarget
	for _, t := range combo.Targets {
		target := ResolvedTarget{
			Provider: t.Provider,
			Model:    t.Model,
			Account:  t.Account,
			Weight:   t.Weight,
		}
		if target.Weight <= 0 {
			target.Weight = 1
		}
		if t.Account != "" {
			if conn, ok := connMap[t.Account]; ok {
				target.APIKey = conn.APIKey
				target.AccessToken = conn.AccessToken
			}
		}
		targets = append(targets, target)
	}
	return targets
}

func resolveRoundRobin(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	if len(targets) > 1 {
		targets = append(targets[1:], targets[0])
	}
	return targets
}

func resolveRandom(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})
	return targets
}

func resolveFillFirst(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	return buildTargets(combo, connections)
}

// resolveStrictRandom shuffles targets randomly but ensures the first target
// is not the same as the last pick for this combo. The last pick is tracked
// in the lastRandomPick map.
func resolveStrictRandom(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	if len(targets) <= 1 {
		return targets
	}

	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})

	// Check if the first pick was the last pick; if so, swap it with the second.
	strictRandMu.RLock()
	lastPick := lastRandomPick[combo.ID]
	strictRandMu.RUnlock()

	if targets[0].Provider == lastPick && len(targets) > 1 {
		targets[0], targets[1] = targets[1], targets[0]
	}

	// Record the new pick.
	strictRandMu.Lock()
	lastRandomPick[combo.ID] = targets[0].Provider
	strictRandMu.Unlock()

	return targets
}

// resolveLKGP routes to the provider that last succeeded for this combo.
// Falls back to priority ordering if no history exists.
func resolveLKGP(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	if len(targets) <= 1 {
		return targets
	}

	lkgpMu.RLock()
	lastGood := lastGoodProvider[combo.ID]
	lkgpMu.RUnlock()

	if lastGood == "" {
		// No history — return priority order.
		return targets
	}

	// Move the last-known-good provider to the front.
	for i, t := range targets {
		if t.Provider == lastGood {
			if i > 0 {
				copy(targets[1:i+1], targets[0:i])
				targets[0] = t
			}
			break
		}
	}
	return targets
}

// resolveContextRelay sorts targets by remaining context headroom descending
// (contextLength - usedTokens). Providers with more available context window
// are preferred.
func resolveContextRelay(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	if len(targets) <= 1 {
		return targets
	}

	sort.SliceStable(targets, func(i, j int) bool {
		headroomI := targets[i].ContextLength - targets[i].UsedTokens
		headroomJ := targets[j].ContextLength - targets[j].UsedTokens
		return headroomI > headroomJ
	})
	return targets
}

// resolveHeadroom sorts targets by remaining quota headroom descending
// (quotaLimit - quotaUsed). Providers with more quota remaining are preferred.
func resolveHeadroom(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	if len(targets) <= 1 {
		return targets
	}

	sort.SliceStable(targets, func(i, j int) bool {
		headroomI := targets[i].QuotaLimit - targets[i].QuotaUsed
		headroomJ := targets[j].QuotaLimit - targets[j].QuotaUsed
		return headroomI > headroomJ
	})
	return targets
}

// resolveFusion returns all targets ordered by priority. The caller is
// responsible for fan-out (sending to all concurrently and returning the
// first successful response).
func resolveFusion(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	return buildTargets(combo, connections)
}

// resolveResetAware is like priority but deprioritizes providers whose quota
// resets within the next 5 minutes. Those providers are moved to the end
// while preserving their relative order.
func resolveResetAware(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	if len(targets) <= 1 {
		return targets
	}

	resetThreshold := time.Now().Add(5 * time.Minute)

	var normal []ResolvedTarget
	var nearReset []ResolvedTarget
	for _, t := range targets {
		if !t.ResetAt.IsZero() && t.ResetAt.Before(resetThreshold) {
			nearReset = append(nearReset, t)
		} else {
			normal = append(normal, t)
		}
	}

	return append(normal, nearReset...)
}

// resolveResetWindow routes only to providers that are NOT currently in a
// reset window (quota has not reset in the last 5 minutes). If all providers
// are in a reset window, falls back to priority order.
func resolveResetWindow(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := buildTargets(combo, connections)
	if len(targets) <= 1 {
		return targets
	}

	resetWindowStart := time.Now().Add(-5 * time.Minute)

	var outsideWindow []ResolvedTarget
	for _, t := range targets {
		// A provider is "in a reset window" if its reset time is very recent
		// (within the last 5 minutes).
		if !t.ResetAt.IsZero() && t.ResetAt.After(resetWindowStart) && !t.ResetAt.After(time.Now()) {
			continue
		}
		outsideWindow = append(outsideWindow, t)
	}

	if len(outsideWindow) == 0 {
		// All providers are in the reset window — fall back to priority.
		return targets
	}
	return outsideWindow
}
