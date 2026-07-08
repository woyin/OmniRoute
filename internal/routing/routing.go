package routing

import (
	"math/rand"
	"sort"

	"github.com/omniroute/omniroute/internal/db"
)

// Strategy represents a routing strategy.
type Strategy string

const (
	StrategyPriority       Strategy = "priority"
	StrategyWeighted       Strategy = "weighted"
	StrategyFillFirst      Strategy = "fill-first"
	StrategyRoundRobin     Strategy = "round-robin"
	StrategyP2C            Strategy = "p2c"
	StrategyRandom         Strategy = "random"
	StrategyLeastUsed      Strategy = "least-used"
	StrategyCostOptimized  Strategy = "cost-optimized"
	StrategyAuto           Strategy = "auto"
	StrategyContextOptimized Strategy = "context-optimized"
)

// ResolvedTarget is a fully resolved routing target.
type ResolvedTarget struct {
	Provider    string
	Model       string
	Account     string
	APIKey      string
	AccessToken string
	Weight      int
	Priority    int
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
	default:
		return resolvePriority(combo, connections)
	}
}

func resolvePriority(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
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
			// Find best connection for this provider
			for _, conn := range connections {
				if conn.Provider == t.Provider && conn.IsActive {
					target.APIKey = conn.APIKey
					target.AccessToken = conn.AccessToken
					target.Priority = conn.Priority
					break
				}
			}
		}
		targets = append(targets, target)
	}

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Priority > targets[j].Priority
	})
	return targets
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
		if t.Weight <= 0 {
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
	// Rotate through targets — simplified: shuffle then return
	targets := resolvePriority(combo, connections)
	if len(targets) > 1 {
		// Rotate by one
		targets = append(targets[1:], targets[0])
	}
	return targets
}

func resolveRandom(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	targets := resolvePriority(combo, connections)
	rand.Shuffle(len(targets), func(i, j int) {
		targets[i], targets[j] = targets[j], targets[i]
	})
	return targets
}

func resolveFillFirst(combo db.Combo, connections []db.ProviderConnection) []ResolvedTarget {
	// Fill-first: try the first target until it's at capacity, then move to the next
	return resolvePriority(combo, connections)
}
