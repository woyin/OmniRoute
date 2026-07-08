package db

import (
	"database/sql"
	
	"time"
)

type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	IsEnabled   bool   `json:"isEnabled"`
	IsBuiltin   bool   `json:"isBuiltin"`
	SkillType   string `json:"skillType"`
	Config      string `json:"config"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type SkillExecution struct {
	ID         int    `json:"id"`
	SkillID    string `json:"skillId"`
	Input      string `json:"input"`
	Output     string `json:"output"`
	Success    bool   `json:"success"`
	DurationMs int    `json:"durationMs"`
	CreatedAt  string `json:"createdAt"`
}

func ListSkills(db *sql.DB) ([]Skill, error) {
	rows, err := db.Query("SELECT id, name, description, version, is_enabled, is_builtin, skill_type, config, created_at, updated_at FROM skills ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]Skill, 0)
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Version, &s.IsEnabled, &s.IsBuiltin, &s.SkillType, &s.Config, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		results = append(results, s)
	}
	return results, nil
}

func SaveSkill(db *sql.DB, s Skill) error {
	_, err := db.Exec(
		"INSERT INTO skills (id, name, description, version, is_enabled, is_builtin, skill_type, config, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) "+
			"ON CONFLICT(id) DO UPDATE SET name=excluded.name, description=excluded.description, version=excluded.version, is_enabled=excluded.is_enabled, skill_type=excluded.skill_type, config=excluded.config, updated_at=excluded.updated_at",
		s.ID, s.Name, s.Description, s.Version, s.IsEnabled, s.IsBuiltin, s.SkillType, s.Config, time.Now().UTC().Format(time.RFC3339))
	return err
}

func DeleteSkill(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM skills WHERE id = ?", id)
	return err
}

func RecordSkillExecution(db *sql.DB, e SkillExecution) error {
	_, err := db.Exec(
		"INSERT INTO skill_executions (skill_id, input, output, success, duration_ms) VALUES (?, ?, ?, ?, ?)",
		e.SkillID, e.Input, e.Output, e.Success, e.DurationMs)
	return err
}
