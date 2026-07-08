package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// UsageEntry represents a usage history record.
type UsageEntry struct {
	ID          int64  `json:"id"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	APIKey      string `json:"apiKey,omitempty"`
	InputTokens int    `json:"inputTokens"`
	OutputTokens int   `json:"outputTokens"`
	Cost        float64 `json:"cost"`
	LatencyMs   int    `json:"latencyMs"`
	Success     bool   `json:"success"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CallLog represents a call log record.
type CallLog struct {
	ID           int64  `json:"id"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	StatusCode   int    `json:"statusCode"`
	LatencyMs    int    `json:"latencyMs"`
	RequestID    string `json:"requestId,omitempty"`
	APIKey       string `json:"apiKey,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// RecordUsage inserts a usage history entry.
func RecordUsage(db *sql.DB, entry UsageEntry) error {
	success := 0
	if entry.Success {
		success = 1
	}
	_, err := db.Exec(
		"INSERT INTO usage_history (provider, model, api_key, input_tokens, output_tokens, cost, latency_ms, success) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		entry.Provider, entry.Model, entry.APIKey, entry.InputTokens, entry.OutputTokens,
		entry.Cost, entry.LatencyMs, success,
	)
	return err
}

// RecordCallLog inserts a call log entry.
func RecordCallLog(db *sql.DB, log CallLog) error {
	_, err := db.Exec(
		"INSERT INTO call_logs (provider, model, status_code, latency_ms, request_id, api_key, error_message) VALUES (?, ?, ?, ?, ?, ?, ?)",
		log.Provider, log.Model, log.StatusCode, log.LatencyMs, log.RequestID,
		log.APIKey, log.ErrorMessage,
	)
	return err
}

// GetUsageHistory returns usage history entries.
func GetUsageHistory(db *sql.DB, provider string, limit int) ([]UsageEntry, error) {
	var rows *sql.Rows
	var err error

	if provider != "" {
		rows, err = db.Query(
			"SELECT id, provider, model, api_key, input_tokens, output_tokens, cost, latency_ms, success, created_at FROM usage_history WHERE provider = ? ORDER BY created_at DESC LIMIT ?",
			provider, limit,
		)
	} else {
		rows, err = db.Query(
			"SELECT id, provider, model, api_key, input_tokens, output_tokens, cost, latency_ms, success, created_at FROM usage_history ORDER BY created_at DESC LIMIT ?",
			limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("query usage_history: %w", err)
	}
	defer rows.Close()

	results := make([]UsageEntry, 0)
	for rows.Next() {
		var e UsageEntry
		var success int
		var createdAt string
		err := rows.Scan(&e.ID, &e.Provider, &e.Model, &e.APIKey, &e.InputTokens,
			&e.OutputTokens, &e.Cost, &e.LatencyMs, &success, &createdAt)
		if err != nil {
			return nil, err
		}
		e.Success = success == 1
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, e)
	}
	return results, nil
}

// GetCallLogs returns call log entries.
func GetCallLogs(db *sql.DB, provider string, limit int) ([]CallLog, error) {
	var rows *sql.Rows
	var err error

	if provider != "" {
		rows, err = db.Query(
			"SELECT id, provider, model, status_code, latency_ms, request_id, api_key, error_message, created_at FROM call_logs WHERE provider = ? ORDER BY created_at DESC LIMIT ?",
			provider, limit,
		)
	} else {
		rows, err = db.Query(
			"SELECT id, provider, model, status_code, latency_ms, request_id, api_key, error_message, created_at FROM call_logs ORDER BY created_at DESC LIMIT ?",
			limit,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("query call_logs: %w", err)
	}
	defer rows.Close()

	results := make([]CallLog, 0)
	for rows.Next() {
		var l CallLog
		var createdAt string
		err := rows.Scan(&l.ID, &l.Provider, &l.Model, &l.StatusCode, &l.LatencyMs,
			&l.RequestID, &l.APIKey, &l.ErrorMessage, &createdAt)
		if err != nil {
			return nil, err
		}
		l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		results = append(results, l)
	}
	return results, nil
}

// GetUsageSummary returns aggregated usage stats.
func GetUsageSummary(db *sql.DB) (map[string]interface{}, error) {
	var totalRequests int
	var totalInputTokens int
	var totalOutputTokens int
	var totalCost float64
	var successCount int

	err := db.QueryRow(
		"SELECT COUNT(*), COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COALESCE(SUM(cost),0), COALESCE(SUM(CASE WHEN success=1 THEN 1 ELSE 0 END),0) FROM usage_history",
	).Scan(&totalRequests, &totalInputTokens, &totalOutputTokens, &totalCost, &successCount)
	if err != nil {
		return nil, err
	}

	// Provider breakdown
	rows, err := db.Query(
		"SELECT provider, COUNT(*), COALESCE(SUM(input_tokens),0), COALESCE(SUM(output_tokens),0), COALESCE(SUM(cost),0) FROM usage_history GROUP BY provider ORDER BY SUM(cost) DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	providers := make(map[string]interface{})
	for rows.Next() {
		var provider string
		var count, inTok, outTok int
		var cost float64
		rows.Scan(&provider, &count, &inTok, &outTok, &cost)
		providers[provider] = map[string]interface{}{
			"requests":      count,
			"inputTokens":   inTok,
			"outputTokens":  outTok,
			"cost":          cost,
		}
	}

	return map[string]interface{}{
		"totalRequests":     totalRequests,
		"totalInputTokens":  totalInputTokens,
		"totalOutputTokens": totalOutputTokens,
		"totalCost":         totalCost,
		"successRate":       fmt.Sprintf("%.1f%%", float64(successCount)/float64(totalRequests)*100),
		"providers":         providers,
	}, nil
}

// PurgeUsageHistory deletes usage history entries older than the given number of days.
func PurgeUsageHistory(db *sql.DB, daysOld int) (int, error) {
	result, err := db.Exec(
		"DELETE FROM usage_history WHERE created_at < datetime('now', ?)",
		fmt.Sprintf("-%d days", daysOld),
	)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// PurgeCallLogs deletes call log entries older than the given number of days.
func PurgeCallLogs(db *sql.DB, daysOld int) (int, error) {
	result, err := db.Exec(
		"DELETE FROM call_logs WHERE created_at < datetime('now', ?)",
		fmt.Sprintf("-%d days", daysOld),
	)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// DBHealthCheck returns database health metrics.
func DBHealthCheck(db *sql.DB) (map[string]interface{}, error) {
	// Table row counts
	tables := []string{"provider_connections", "combos", "api_keys", "usage_history", "call_logs", "key_value", "semantic_cache"}
	counts := make(map[string]int)
	for _, t := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", t)).Scan(&count)
		if err != nil {
			counts[t] = -1
		} else {
			counts[t] = count
		}
	}

	// WAL check
	var walMode string
	db.QueryRow("PRAGMA journal_mode").Scan(&walMode)

	// DB size
	var pageCount, pageSize int
	db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	dbSizeMB := float64(pageCount * pageSize) / 1024 / 1024

	return map[string]interface{}{
		"tables":    counts,
		"walMode":   walMode,
		"dbSizeMB":  dbSizeMB,
		"pageCount": pageCount,
		"pageSize":  pageSize,
	}, nil
}

// ExportAll exports all data as JSON for backup.
func ExportAll(db *sql.DB) (map[string]interface{}, error) {
	export := make(map[string]interface{})

	providers, err := ListProviderConnections(db, "")
	if err == nil {
		export["providers"] = providers
	}

	combos, err := ListCombos(db)
	if err == nil {
		export["combos"] = combos
	}

	keys, err := ListAPIKeys(db)
	if err == nil {
		export["apiKeys"] = keys
	}

	// Settings
	rows, err := db.Query("SELECT key, value FROM key_value WHERE namespace = 'settings'")
	if err == nil {
		settings := make(map[string]string)
		defer rows.Close()
		for rows.Next() {
			var key, value string
			rows.Scan(&key, &value)
			settings[key] = value
		}
		export["settings"] = settings
	}

	export["exportedAt"] = time.Now().UTC().Format(time.RFC3339)
	return export, nil
}

// ImportAll imports data from a JSON backup.
func ImportAll(db *sql.DB, data map[string]interface{}) error {
	// Import providers
	if providersRaw, ok := data["providers"]; ok {
		providersJSON, _ := json.Marshal(providersRaw)
		var providers []ProviderConnection
		json.Unmarshal(providersJSON, &providers)
		for _, pc := range providers {
			SaveProviderConnection(db, pc)
		}
	}

	// Import combos
	if combosRaw, ok := data["combos"]; ok {
		combosJSON, _ := json.Marshal(combosRaw)
		var combos []Combo
		json.Unmarshal(combosJSON, &combos)
		for _, c := range combos {
			SaveCombo(db, c)
		}
	}

	// Import api keys
	if keysRaw, ok := data["apiKeys"]; ok {
		keysJSON, _ := json.Marshal(keysRaw)
		var keys []APIKey
		json.Unmarshal(keysJSON, &keys)
		for _, ak := range keys {
			CreateAPIKey(db, ak)
		}
	}

	// Import settings
	if settingsRaw, ok := data["settings"]; ok {
		if settings, ok := settingsRaw.(map[string]interface{}); ok {
			for key, value := range settings {
				if strVal, ok := value.(string); ok {
					db.Exec(
						"INSERT INTO key_value (namespace, key, value) VALUES ('settings', ?, ?) "+
							"ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value",
						key, strVal,
					)
				}
			}
		}
	}

	return nil
}
