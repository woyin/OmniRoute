package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// EvalsHandler provides eval management endpoints.
// Uses skill_executions table for eval tracking since there's no dedicated evals table.
type EvalsHandler struct {
	DB *sql.DB
}

// List returns eval runs from skill_executions (filtered by eval-type skills).
func (h *EvalsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type evalRun struct {
		ID         int    `json:"id"`
		SkillID    string `json:"skillId"`
		SkillName  string `json:"skillName"`
		Input      string `json:"input"`
		Output     string `json:"output"`
		Success    bool   `json:"success"`
		DurationMs int    `json:"durationMs"`
		CreatedAt  string `json:"createdAt"`
	}

	var evals []evalRun
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT se.id, se.skill_id, COALESCE(s.name, ''),
			       se.input, se.output, se.success, se.duration_ms, se.created_at
			FROM skill_executions se
			LEFT JOIN skills s ON se.skill_id = s.id
			WHERE s.skill_type = 'eval' OR s.name LIKE '%eval%'
			ORDER BY se.created_at DESC
			LIMIT 100
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var e evalRun
				var success int
				if err := rows.Scan(&e.ID, &e.SkillID, &e.SkillName,
					&e.Input, &e.Output, &success, &e.DurationMs, &e.CreatedAt); err == nil {
					e.Success = success == 1
					evals = append(evals, e)
				}
			}
		}
	}

	// If no eval-specific records, return all skill executions as a fallback
	if len(evals) == 0 && h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT se.id, se.skill_id, COALESCE(s.name, ''),
			       se.input, se.output, se.success, se.duration_ms, se.created_at
			FROM skill_executions se
			LEFT JOIN skills s ON se.skill_id = s.id
			ORDER BY se.created_at DESC
			LIMIT 50
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var e evalRun
				var success int
				if err := rows.Scan(&e.ID, &e.SkillID, &e.SkillName,
					&e.Input, &e.Output, &success, &e.DurationMs, &e.CreatedAt); err == nil {
					e.Success = success == 1
					evals = append(evals, e)
				}
			}
		}
	}

	if evals == nil {
		evals = []evalRun{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   evals,
		"total":  len(evals),
	})
}

// Suites returns eval suites (eval-type skills).
func (h *EvalsHandler) Suites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type suite struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     string `json:"version"`
		IsEnabled   bool   `json:"isEnabled"`
		Config      string `json:"config"`
		CreatedAt   string `json:"createdAt"`
	}

	var suites []suite
	if h.DB != nil {
		// Query eval-type skills as suites
		rows, err := h.DB.Query(`
			SELECT id, name, description, version, is_enabled, config, created_at
			FROM skills
			WHERE skill_type = 'eval' OR name LIKE '%eval%'
			ORDER BY created_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var s suite
				var enabled int
				if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Version,
					&enabled, &s.Config, &s.CreatedAt); err == nil {
					s.IsEnabled = enabled == 1
					suites = append(suites, s)
				}
			}
		}

		// If no eval skills found, return all skills as potential suites
		if len(suites) == 0 {
			rows2, err := h.DB.Query(`
				SELECT id, name, description, version, is_enabled, config, created_at
				FROM skills
				ORDER BY created_at DESC
			`)
			if err == nil {
				defer rows2.Close()
				for rows2.Next() {
					var s suite
					var enabled int
					if err := rows2.Scan(&s.ID, &s.Name, &s.Description, &s.Version,
						&enabled, &s.Config, &s.CreatedAt); err == nil {
						s.IsEnabled = enabled == 1
						suites = append(suites, s)
					}
				}
			}
		}
	}

	if suites == nil {
		suites = []suite{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"object":    "list",
		"data":      suites,
		"total":     len(suites),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
