package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// TelemetryHandler provides real DB-backed telemetry aggregation endpoints.
type TelemetryHandler struct {
	DB *sql.DB
}

// Summary returns aggregated telemetry data from usage_history and daily_usage_summary.
func (h *TelemetryHandler) Summary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Aggregate from usage_history
	totalRequests := 0
	totalInputTokens := int64(0)
	totalOutputTokens := int64(0)
	totalCost := 0.0
	totalErrors := 0
	avgLatency := 0.0

	if h.DB != nil {
		h.DB.QueryRow(`
			SELECT COUNT(*),
			       COALESCE(SUM(input_tokens), 0),
			       COALESCE(SUM(output_tokens), 0),
			       COALESCE(SUM(cost), 0),
			       SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END),
			       COALESCE(AVG(latency_ms), 0)
			FROM usage_history
		`).Scan(&totalRequests, &totalInputTokens, &totalOutputTokens, &totalCost, &totalErrors, &avgLatency)
	}

	// Get unique provider count
	uniqueProviders := 0
	if h.DB != nil {
		h.DB.QueryRow("SELECT COUNT(DISTINCT provider) FROM usage_history").Scan(&uniqueProviders)
	}

	// Get unique model count
	uniqueModels := 0
	if h.DB != nil {
		h.DB.QueryRow("SELECT COUNT(DISTINCT model) FROM usage_history").Scan(&uniqueModels)
	}

	// Recent daily summary (last 7 days)
	type daySummary struct {
		Date         string  `json:"date"`
		Requests     int     `json:"requests"`
		InputTokens  int64   `json:"inputTokens"`
		OutputTokens int64   `json:"outputTokens"`
		Cost         float64 `json:"cost"`
		Errors       int     `json:"errors"`
	}
	var dailySummary []daySummary
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT date, SUM(request_count), SUM(input_tokens), SUM(output_tokens),
			       SUM(cost), SUM(error_count)
			FROM daily_usage_summary
			GROUP BY date
			ORDER BY date DESC
			LIMIT 7
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var ds daySummary
				if err := rows.Scan(&ds.Date, &ds.Requests, &ds.InputTokens,
					&ds.OutputTokens, &ds.Cost, &ds.Errors); err == nil {
					dailySummary = append(dailySummary, ds)
				}
			}
		}
	}

	if dailySummary == nil {
		dailySummary = []daySummary{}
	}

	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests":       totalRequests,
		"tokens":         totalInputTokens + totalOutputTokens,
		"inputTokens":    totalInputTokens,
		"outputTokens":   totalOutputTokens,
		"cost":           totalCost,
		"errors":         totalErrors,
		"errorRate":      errorRate,
		"avgLatency":     avgLatency,
		"uniqueProviders":    uniqueProviders,
		"uniqueModels":   uniqueModels,
		"dailySummary":   dailySummary,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}
