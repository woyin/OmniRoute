package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/omniroute/omniroute/internal/db"
)

func evalRunHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			SuiteID string                 `json:"suiteId"`
			Outputs map[string]interface{} `json:"outputs"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.SuiteID == "" {
			jsonError(w, 400, "Invalid JSON body")
			return
		}
		var exists int
		if dbConn.QueryRow(`SELECT COUNT(*) FROM skills WHERE id=? AND skill_type='eval'`, body.SuiteID).Scan(&exists) != nil {
			jsonError(w, 500, "Failed to run eval suite")
			return
		}
		if exists == 0 {
			jsonError(w, 404, "Suite not found: "+body.SuiteID)
			return
		}
		raw, _ := json.Marshal(body.Outputs)
		start := time.Now()
		execution := db.SkillExecution{SkillID: body.SuiteID, Input: string(raw), Output: string(raw), Success: true, DurationMs: int(time.Since(start).Milliseconds())}
		if db.RecordSkillExecution(dbConn, execution) != nil {
			jsonError(w, 500, "Failed to run eval suite")
			return
		}
		writeJSONResponse(w, map[string]interface{}{"suiteId": body.SuiteID, "results": body.Outputs, "summary": map[string]interface{}{"total": len(body.Outputs), "passed": len(body.Outputs), "failed": 0}})
	}
}
func evalSuiteCollectionHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ID          string                 `json:"id"`
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			Version     string                 `json:"version"`
			Config      map[string]interface{} `json:"config"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.Name == "" {
			jsonError(w, 400, "Invalid JSON body")
			return
		}
		if body.ID == "" {
			body.ID = uuid.NewString()
		}
		if body.Version == "" {
			body.Version = "1.0.0"
		}
		raw, _ := json.Marshal(body.Config)
		skill := db.Skill{ID: body.ID, Name: body.Name, Description: body.Description, Version: body.Version, IsEnabled: true, SkillType: "eval", Config: string(raw)}
		if db.SaveSkill(dbConn, skill) != nil {
			jsonError(w, 500, "Failed to create eval suite")
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSONResponse(w, map[string]interface{}{"suite": skill})
	}
}
func evalSuiteDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "suiteId")
		var skill db.Skill
		err := dbConn.QueryRow(`SELECT id,name,description,version,is_enabled,is_builtin,skill_type,config,created_at,updated_at FROM skills WHERE id=? AND skill_type='eval'`, id).Scan(&skill.ID, &skill.Name, &skill.Description, &skill.Version, &skill.IsEnabled, &skill.IsBuiltin, &skill.SkillType, &skill.Config, &skill.CreatedAt, &skill.UpdatedAt)
		if err == sql.ErrNoRows {
			jsonError(w, 404, "Eval suite not found")
			return
		}
		if err != nil {
			jsonError(w, 500, "Failed to load eval suite")
			return
		}
		switch r.Method {
		case http.MethodGet:
			writeJSONResponse(w, map[string]interface{}{"suite": skill})
		case http.MethodPut:
			var body struct {
				Name        string                 `json:"name"`
				Description string                 `json:"description"`
				Version     string                 `json:"version"`
				Config      map[string]interface{} `json:"config"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || body.Name == "" {
				jsonError(w, 400, "Invalid JSON body")
				return
			}
			skill.Name = body.Name
			skill.Description = body.Description
			if body.Version != "" {
				skill.Version = body.Version
			}
			raw, _ := json.Marshal(body.Config)
			skill.Config = string(raw)
			if db.SaveSkill(dbConn, skill) != nil {
				jsonError(w, 500, "Failed to update eval suite")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"suite": skill})
		case http.MethodDelete:
			if db.DeleteSkill(dbConn, id) != nil {
				jsonError(w, 500, "Failed to delete eval suite")
				return
			}
			writeJSONResponse(w, map[string]interface{}{"success": true})
		}
	}
}
func evalSuitePublicDetailHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "suiteId")
		var skill db.Skill
		err := dbConn.QueryRow(`SELECT id,name,description,version,is_enabled,is_builtin,skill_type,config,created_at,updated_at FROM skills WHERE id=? AND skill_type='eval'`, id).Scan(&skill.ID, &skill.Name, &skill.Description, &skill.Version, &skill.IsEnabled, &skill.IsBuiltin, &skill.SkillType, &skill.Config, &skill.CreatedAt, &skill.UpdatedAt)
		if err == sql.ErrNoRows {
			jsonError(w, 404, "Suite not found: "+id)
			return
		}
		if err != nil {
			jsonError(w, 500, "Failed to load eval suite")
			return
		}
		writeJSONResponse(w, skill)
	}
}
