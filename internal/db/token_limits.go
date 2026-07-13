package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TokenLimit struct {
	ID            string `json:"id"`
	APIKeyID      string `json:"apiKeyId"`
	ScopeType     string `json:"scopeType"`
	ScopeValue    string `json:"scopeValue"`
	TokenLimit    int64  `json:"tokenLimit"`
	ResetInterval string `json:"resetInterval"`
	ResetTime     string `json:"resetTime"`
	Enabled       bool   `json:"enabled"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func UpsertTokenLimit(db *sql.DB, limit TokenLimit) (TokenLimit, error) {
	if limit.ID == "" {
		limit.ID = uuid.NewString()
	}
	if limit.ScopeType == "global" {
		limit.ScopeValue = ""
	}
	if limit.ResetInterval == "" {
		limit.ResetInterval = "monthly"
	}
	if limit.ResetTime == "" {
		limit.ResetTime = "00:00"
	}
	enabled := 0
	if limit.Enabled {
		enabled = 1
	}
	_, err := db.Exec(`INSERT INTO api_key_token_limits
		(id, api_key_id, scope_type, scope_value, token_limit, reset_interval, reset_time, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(api_key_id, scope_type, scope_value) DO UPDATE SET
		 token_limit=excluded.token_limit, reset_interval=excluded.reset_interval,
		 reset_time=excluded.reset_time, enabled=excluded.enabled, updated_at=datetime('now')`,
		limit.ID, limit.APIKeyID, limit.ScopeType, limit.ScopeValue, limit.TokenLimit,
		limit.ResetInterval, limit.ResetTime, enabled)
	if err != nil {
		return TokenLimit{}, err
	}
	return GetTokenLimit(db, limit.APIKeyID, limit.ScopeType, limit.ScopeValue)
}

func GetTokenLimit(db *sql.DB, apiKeyID, scopeType, scopeValue string) (TokenLimit, error) {
	var limit TokenLimit
	var enabled int
	err := db.QueryRow(`SELECT id, api_key_id, scope_type, scope_value, token_limit,
		reset_interval, reset_time, enabled, created_at, updated_at
		FROM api_key_token_limits WHERE api_key_id=? AND scope_type=? AND scope_value=?`,
		apiKeyID, scopeType, scopeValue).Scan(&limit.ID, &limit.APIKeyID, &limit.ScopeType,
		&limit.ScopeValue, &limit.TokenLimit, &limit.ResetInterval, &limit.ResetTime,
		&enabled, &limit.CreatedAt, &limit.UpdatedAt)
	limit.Enabled = enabled != 0
	return limit, err
}

func ListTokenLimits(db *sql.DB, apiKeyID string) ([]TokenLimit, error) {
	rows, err := db.Query(`SELECT id, api_key_id, scope_type, scope_value, token_limit,
		reset_interval, reset_time, enabled, created_at, updated_at
		FROM api_key_token_limits WHERE api_key_id=?
		ORDER BY CASE scope_type WHEN 'model' THEN 0 WHEN 'provider' THEN 1 ELSE 2 END, scope_value`, apiKeyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	limits := make([]TokenLimit, 0)
	for rows.Next() {
		var limit TokenLimit
		var enabled int
		if err := rows.Scan(&limit.ID, &limit.APIKeyID, &limit.ScopeType, &limit.ScopeValue,
			&limit.TokenLimit, &limit.ResetInterval, &limit.ResetTime, &enabled,
			&limit.CreatedAt, &limit.UpdatedAt); err != nil {
			return nil, err
		}
		limit.Enabled = enabled != 0
		limits = append(limits, limit)
	}
	return limits, rows.Err()
}

func DeleteTokenLimit(db *sql.DB, id string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM api_key_token_counters WHERE limit_id=?", id); err != nil {
		return false, err
	}
	if _, err := tx.Exec("DELETE FROM api_key_token_limit_reset_logs WHERE limit_id=?", id); err != nil {
		return false, err
	}
	result, err := tx.Exec("DELETE FROM api_key_token_limits WHERE id=?", id)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return count > 0, tx.Commit()
}

func TokenLimitWindow(limit TokenLimit, now time.Time) (start, next time.Time, err error) {
	hour, minute := 0, 0
	if _, err = fmt.Sscanf(limit.ResetTime, "%02d:%02d", &hour, &minute); err != nil || hour > 23 || minute > 59 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid resetTime")
	}
	loc := now.Location()
	boundary := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	switch limit.ResetInterval {
	case "daily":
		if now.Before(boundary) {
			boundary = boundary.AddDate(0, 0, -1)
		}
		return boundary, boundary.AddDate(0, 0, 1), nil
	case "weekly":
		days := (int(boundary.Weekday()) + 6) % 7
		boundary = boundary.AddDate(0, 0, -days)
		if now.Before(boundary) {
			boundary = boundary.AddDate(0, 0, -7)
		}
		return boundary, boundary.AddDate(0, 0, 7), nil
	case "monthly":
		boundary = time.Date(now.Year(), now.Month(), 1, hour, minute, 0, 0, loc)
		if now.Before(boundary) {
			boundary = boundary.AddDate(0, -1, 0)
		}
		return boundary, boundary.AddDate(0, 1, 0), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid resetInterval")
	}
}
