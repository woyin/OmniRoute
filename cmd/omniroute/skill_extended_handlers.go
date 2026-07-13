package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/omniroute/omniroute/internal/db"
)

func skillUpdateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var skill db.Skill
		err := dbConn.QueryRow(`SELECT id,name,description,version,is_enabled,is_builtin,skill_type,config,created_at,updated_at FROM skills WHERE id=?`, id).Scan(&skill.ID, &skill.Name, &skill.Description, &skill.Version, &skill.IsEnabled, &skill.IsBuiltin, &skill.SkillType, &skill.Config, &skill.CreatedAt, &skill.UpdatedAt)
		if err == sql.ErrNoRows {
			jsonError(w, 404, "Skill not found")
			return
		}
		if err != nil {
			jsonError(w, 500, "Failed to load skill")
			return
		}
		var body struct {
			Enabled *bool  `json:"enabled"`
			Mode    string `json:"mode"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			jsonError(w, 400, "Invalid JSON body")
			return
		}
		if body.Enabled == nil && body.Mode == "" {
			jsonError(w, 400, "No update payload provided")
			return
		}
		if body.Mode != "" && body.Mode != "on" && body.Mode != "off" && body.Mode != "auto" {
			jsonError(w, 400, "Invalid mode")
			return
		}
		if body.Enabled != nil {
			skill.IsEnabled = *body.Enabled
		}
		if body.Mode != "" {
			skill.IsEnabled = body.Mode != "off"
		}
		if db.SaveSkill(dbConn, skill) != nil {
			jsonError(w, 500, "Failed to update skill")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"success": true, "enabled": skill.IsEnabled, "mode": body.Mode})
	}
}
func skillExecutionsHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			limit := queryInt(r, "limit", 20, 1, 100)
			page := queryInt(r, "page", 1, 1, 100000)
			var total int
			_ = dbConn.QueryRow(`SELECT COUNT(*) FROM skill_executions`).Scan(&total)
			rows, err := dbConn.Query(`SELECT id,skill_id,input,output,success,duration_ms,created_at FROM skill_executions ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, (page-1)*limit)
			if err != nil {
				jsonError(w, 500, "Failed to list executions")
				return
			}
			defer rows.Close()
			items := make([]db.SkillExecution, 0)
			for rows.Next() {
				var e db.SkillExecution
				if rows.Scan(&e.ID, &e.SkillID, &e.Input, &e.Output, &e.Success, &e.DurationMs, &e.CreatedAt) == nil {
					items = append(items, e)
				}
			}
			writeJSONResponse(w, map[string]interface{}{"data": items, "total": total, "page": page, "limit": limit, "totalPages": (total + limit - 1) / limit})
			return
		}
		var body struct {
			SkillName string                 `json:"skillName"`
			APIKeyID  string                 `json:"apiKeyId"`
			Input     map[string]interface{} `json:"input"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.SkillName == "" || body.APIKeyID == "" {
			jsonError(w, 400, "Invalid request")
			return
		}
		var skill db.Skill
		err := dbConn.QueryRow(`SELECT id,name,description,version,is_enabled,is_builtin,skill_type,config,created_at,updated_at FROM skills WHERE name=?`, body.SkillName).Scan(&skill.ID, &skill.Name, &skill.Description, &skill.Version, &skill.IsEnabled, &skill.IsBuiltin, &skill.SkillType, &skill.Config, &skill.CreatedAt, &skill.UpdatedAt)
		if err == sql.ErrNoRows {
			jsonError(w, 404, "Skill not found")
			return
		}
		if err != nil {
			jsonError(w, 500, "Failed to execute skill")
			return
		}
		if !skill.IsEnabled {
			jsonError(w, 503, "Skill is disabled")
			return
		}
		raw, _ := json.Marshal(body.Input)
		execution := db.SkillExecution{SkillID: skill.ID, Input: string(raw), Output: string(raw), Success: true}
		if db.RecordSkillExecution(dbConn, execution) != nil {
			jsonError(w, 500, "Failed to execute skill")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"execution": execution})
	}
}
func sessionCreateHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SessionID    string `json:"sessionId"`
			Provider     string `json:"provider"`
			ConnectionID string `json:"connectionId"`
			Model        string `json:"model"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.SessionID == "" || body.Provider == "" {
			jsonError(w, 400, "Invalid request")
			return
		}
		_, err := dbConn.Exec(`INSERT INTO session_account_affinity(session_id,provider,connection_id,model) VALUES(?,?,?,?) ON CONFLICT(session_id,provider) DO UPDATE SET connection_id=excluded.connection_id,model=excluded.model,last_used_at=CURRENT_TIMESTAMP`, body.SessionID, body.Provider, body.ConnectionID, body.Model)
		if err != nil {
			jsonError(w, 500, "Failed to create session")
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSONResponse(w, map[string]interface{}{"success": true, "sessionId": body.SessionID})
	}
}
