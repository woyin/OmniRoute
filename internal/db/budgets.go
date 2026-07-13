package db

import (
	"database/sql"
	"math"
	"time"
)

type Budget struct {
	APIKeyID          string  `json:"-"`
	DailyLimitUSD     float64 `json:"dailyLimitUsd"`
	WeeklyLimitUSD    float64 `json:"weeklyLimitUsd"`
	MonthlyLimitUSD   float64 `json:"monthlyLimitUsd"`
	WarningThreshold  float64 `json:"warningThreshold"`
	ResetInterval     string  `json:"resetInterval"`
	ResetTime         string  `json:"resetTime"`
	BudgetResetAt     *int64  `json:"budgetResetAt"`
	LastBudgetResetAt *int64  `json:"lastBudgetResetAt"`
}

func UpsertBudget(db *sql.DB, budget Budget) (Budget, error) {
	_, err := db.Exec(`INSERT INTO domain_budgets
		(api_key_id, daily_limit_usd, weekly_limit_usd, monthly_limit_usd, warning_threshold, reset_interval, reset_time)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(api_key_id) DO UPDATE SET daily_limit_usd=excluded.daily_limit_usd,
		weekly_limit_usd=excluded.weekly_limit_usd, monthly_limit_usd=excluded.monthly_limit_usd,
		warning_threshold=excluded.warning_threshold, reset_interval=excluded.reset_interval,
		reset_time=excluded.reset_time`, budget.APIKeyID, budget.DailyLimitUSD, budget.WeeklyLimitUSD,
		budget.MonthlyLimitUSD, budget.WarningThreshold, budget.ResetInterval, budget.ResetTime)
	if err != nil {
		return Budget{}, err
	}
	return GetBudget(db, budget.APIKeyID)
}

func GetBudget(db *sql.DB, apiKeyID string) (Budget, error) {
	var budget Budget
	var resetAt, lastResetAt sql.NullInt64
	err := db.QueryRow(`SELECT api_key_id, daily_limit_usd, weekly_limit_usd, monthly_limit_usd,
		warning_threshold, reset_interval, reset_time, budget_reset_at, last_budget_reset_at
		FROM domain_budgets WHERE api_key_id=?`, apiKeyID).Scan(&budget.APIKeyID, &budget.DailyLimitUSD,
		&budget.WeeklyLimitUSD, &budget.MonthlyLimitUSD, &budget.WarningThreshold, &budget.ResetInterval,
		&budget.ResetTime, &resetAt, &lastResetAt)
	if resetAt.Valid {
		budget.BudgetResetAt = &resetAt.Int64
	}
	if lastResetAt.Valid {
		budget.LastBudgetResetAt = &lastResetAt.Int64
	}
	return budget, err
}

func ListBudgets(db *sql.DB) ([]Budget, error) {
	rows, err := db.Query(`SELECT api_key_id, daily_limit_usd, weekly_limit_usd, monthly_limit_usd,
		warning_threshold, reset_interval, reset_time, budget_reset_at, last_budget_reset_at
		FROM domain_budgets ORDER BY api_key_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	budgets := make([]Budget, 0)
	for rows.Next() {
		var budget Budget
		var resetAt, lastResetAt sql.NullInt64
		if err := rows.Scan(&budget.APIKeyID, &budget.DailyLimitUSD, &budget.WeeklyLimitUSD,
			&budget.MonthlyLimitUSD, &budget.WarningThreshold, &budget.ResetInterval, &budget.ResetTime,
			&resetAt, &lastResetAt); err != nil {
			return nil, err
		}
		if resetAt.Valid {
			budget.BudgetResetAt = &resetAt.Int64
		}
		if lastResetAt.Valid {
			budget.LastBudgetResetAt = &lastResetAt.Int64
		}
		budgets = append(budgets, budget)
	}
	return budgets, rows.Err()
}

func BudgetSummary(db *sql.DB, apiKeyID string, now time.Time) (map[string]interface{}, error) {
	budget, err := GetBudget(db, apiKeyID)
	hasBudget := err == nil
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	daily, dailyCount, err := costSince(db, apiKeyID, dayStart)
	if err != nil {
		return nil, err
	}
	monthly, totalEntries, err := costSince(db, apiKeyID, monthStart)
	if err != nil {
		return nil, err
	}
	var period, active float64
	var periodStart, nextReset *int64
	if hasBudget {
		start, next, err := TokenLimitWindow(TokenLimit{ResetInterval: budget.ResetInterval, ResetTime: budget.ResetTime}, now)
		if err != nil {
			return nil, err
		}
		period, _, err = costSince(db, apiKeyID, start)
		if err != nil {
			return nil, err
		}
		startMillis, nextMillis := start.UnixMilli(), next.UnixMilli()
		periodStart, nextReset = &startMillis, &nextMillis
		active = activeBudgetLimit(budget)
		budget.BudgetResetAt, budget.LastBudgetResetAt = nextReset, periodStart
	}
	_ = dailyCount
	var budgetValue interface{}
	if hasBudget {
		budgetValue = budget
	}
	return map[string]interface{}{
		"dailyTotal": daily, "monthlyTotal": monthly, "totalEntries": totalEntries, "budget": budgetValue,
		"totalCostToday": daily, "totalCostMonth": monthly, "totalCostPeriod": period,
		"activeLimitUsd": active, "resetInterval": nullableBudgetString(hasBudget, budget.ResetInterval),
		"resetTime": nullableBudgetString(hasBudget, budget.ResetTime), "budgetResetAt": nextReset,
		"lastBudgetResetAt": periodStart, "periodStartAt": periodStart, "nextResetAt": nextReset,
		"dailyLimitUsd": budget.DailyLimitUSD, "weeklyLimitUsd": budget.WeeklyLimitUSD,
		"monthlyLimitUsd": budget.MonthlyLimitUSD, "warningThreshold": nullableBudgetFloat(hasBudget, budget.WarningThreshold),
	}, nil
}

func BudgetCheck(summary map[string]interface{}) map[string]interface{} {
	used := summary["totalCostPeriod"].(float64)
	limit := summary["activeLimitUsd"].(float64)
	threshold, _ := summary["warningThreshold"].(*float64)
	warning := threshold != nil && limit > 0 && used >= limit**threshold
	remaining := math.Max(limit-used, 0)
	check := map[string]interface{}{
		"allowed": limit <= 0 || used <= limit, "dailyUsed": used, "dailyLimit": limit,
		"warningReached": warning, "remaining": remaining, "periodUsed": used, "activeLimitUsd": limit,
		"resetInterval": summary["resetInterval"], "resetTime": summary["resetTime"],
		"budgetResetAt": summary["budgetResetAt"], "lastBudgetResetAt": summary["lastBudgetResetAt"],
		"periodStartAt": summary["periodStartAt"],
	}
	return check
}

func costSince(db *sql.DB, apiKeyID string, since time.Time) (float64, int, error) {
	var total float64
	var count int
	err := db.QueryRow(`SELECT COALESCE(SUM(cost),0), COUNT(*) FROM usage_history
		WHERE api_key=? AND created_at>=?`, apiKeyID, since.UTC().Format("2006-01-02 15:04:05")).Scan(&total, &count)
	return total, count, err
}

func activeBudgetLimit(b Budget) float64 {
	switch b.ResetInterval {
	case "daily":
		return b.DailyLimitUSD
	case "weekly":
		return b.WeeklyLimitUSD
	case "monthly":
		return b.MonthlyLimitUSD
	default:
		return 0
	}
}

func nullableBudgetString(ok bool, value string) interface{} {
	if !ok {
		return nil
	}
	return value
}

func nullableBudgetFloat(ok bool, value float64) interface{} {
	if !ok {
		return nil
	}
	return &value
}
