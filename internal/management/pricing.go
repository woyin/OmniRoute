package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// PricingHandler provides real DB-backed pricing management endpoints.
// Since there is no model_pricing table, we aggregate from usage_history and key_value.
type PricingHandler struct {
	DB *sql.DB
}

// List returns pricing data aggregated from usage history.
func (h *PricingHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type pricingRow struct {
		Provider    string  `json:"provider"`
		Model       string  `json:"model"`
		TotalCost   float64 `json:"totalCost"`
		Requests    int     `json:"requests"`
		InputTokens int64   `json:"inputTokens"`
		OutputTokens int64  `json:"outputTokens"`
	}

	var rows []pricingRow
	if h.DB != nil {
		result, err := h.DB.Query(`
			SELECT provider, model,
			       COALESCE(SUM(cost), 0) as total_cost,
			       COUNT(*) as requests,
			       COALESCE(SUM(input_tokens), 0) as input_tokens,
			       COALESCE(SUM(output_tokens), 0) as output_tokens
			FROM usage_history
			GROUP BY provider, model
			ORDER BY total_cost DESC
			LIMIT 100
		`)
		if err == nil {
			defer result.Close()
			for result.Next() {
				var pr pricingRow
				if err := result.Scan(&pr.Provider, &pr.Model, &pr.TotalCost,
					&pr.Requests, &pr.InputTokens, &pr.OutputTokens); err == nil {
					rows = append(rows, pr)
				}
			}
		}
	}

	if rows == nil {
		rows = []pricingRow{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   rows,
		"total":  len(rows),
	})
}

// Defaults returns default pricing configuration from key_value store.
func (h *PricingHandler) Defaults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defaults := map[string]interface{}{
		"inputPricePer1k":  0.01,
		"outputPricePer1k": 0.03,
	}

	if h.DB != nil {
		var val string
		if err := h.DB.QueryRow("SELECT value FROM key_value WHERE namespace = 'pricing' AND key = 'defaults'").Scan(&val); err == nil {
			var stored map[string]interface{}
			if json.Unmarshal([]byte(val), &stored) == nil {
				defaults = stored
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"defaults": defaults,
	})
}

// Models returns per-model cost data from usage history.
func (h *PricingHandler) Models(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type modelPricing struct {
		Provider     string  `json:"provider"`
		Model        string  `json:"model"`
		AvgCostPerReq float64 `json:"avgCostPerRequest"`
		TotalCost    float64 `json:"totalCost"`
		TotalRequests int    `json:"totalRequests"`
		AvgInputTokens  float64 `json:"avgInputTokens"`
		AvgOutputTokens float64 `json:"avgOutputTokens"`
	}

	var models []modelPricing
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT provider, model,
			       CASE WHEN COUNT(*) > 0 THEN SUM(cost) / COUNT(*) ELSE 0 END as avg_cost,
			       COALESCE(SUM(cost), 0) as total_cost,
			       COUNT(*) as total_requests,
			       CASE WHEN COUNT(*) > 0 THEN CAST(SUM(input_tokens) AS REAL) / COUNT(*) ELSE 0 END,
			       CASE WHEN COUNT(*) > 0 THEN CAST(SUM(output_tokens) AS REAL) / COUNT(*) ELSE 0 END
			FROM usage_history
			GROUP BY provider, model
			ORDER BY total_cost DESC
			LIMIT 50
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var mp modelPricing
				if err := rows.Scan(&mp.Provider, &mp.Model, &mp.AvgCostPerReq,
					&mp.TotalCost, &mp.TotalRequests,
					&mp.AvgInputTokens, &mp.AvgOutputTokens); err == nil {
					models = append(models, mp)
				}
			}
		}
	}

	if models == nil {
		models = []modelPricing{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   models,
		"total":  len(models),
	})
}

// Sync triggers a pricing data sync from usage history into key_value store.
func (h *PricingHandler) Sync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.DB == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not available")
		return
	}

	// Aggregate current pricing from usage history
	var totalCost float64
	var totalRequests int
	h.DB.QueryRow("SELECT COALESCE(SUM(cost), 0), COUNT(*) FROM usage_history").
		Scan(&totalCost, &totalRequests)

	// Store sync metadata
	syncData := map[string]interface{}{
		"lastSync":      time.Now().UTC().Format(time.RFC3339),
		"totalCost":     totalCost,
		"totalRequests": totalRequests,
	}
	data, _ := json.Marshal(syncData)

	_, err := h.DB.Exec(`
		INSERT INTO key_value (namespace, key, value)
		VALUES ('pricing', 'last_sync', ?)
		ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value
	`, string(data))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save sync metadata")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"synced":        true,
		"modelsUpdated": totalRequests,
		"totalCost":     totalCost,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	})
}
