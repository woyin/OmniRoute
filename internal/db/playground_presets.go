package db

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
)

type PlaygroundPreset struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Endpoint  string                 `json:"endpoint"`
	Model     string                 `json:"model"`
	System    *string                `json:"system"`
	Params    map[string]interface{} `json:"params"`
	CreatedAt string                 `json:"created_at"`
}

func ListPlaygroundPresets(db *sql.DB) ([]PlaygroundPreset, error) {
	rows, err := db.Query("SELECT id, name, endpoint, model, system, params_json, created_at FROM playground_presets ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	presets := make([]PlaygroundPreset, 0)
	for rows.Next() {
		preset, err := scanPlaygroundPreset(rows.Scan)
		if err != nil {
			return nil, err
		}
		presets = append(presets, preset)
	}
	return presets, rows.Err()
}

func GetPlaygroundPreset(db *sql.DB, id string) (*PlaygroundPreset, error) {
	preset, err := scanPlaygroundPreset(db.QueryRow("SELECT id, name, endpoint, model, system, params_json, created_at FROM playground_presets WHERE id=?", id).Scan)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &preset, err
}

func CreatePlaygroundPreset(db *sql.DB, preset PlaygroundPreset) (PlaygroundPreset, error) {
	preset.ID = uuid.NewString()
	params, err := json.Marshal(preset.Params)
	if err != nil {
		return PlaygroundPreset{}, err
	}
	_, err = db.Exec("INSERT INTO playground_presets(id,name,endpoint,model,system,params_json) VALUES(?,?,?,?,?,?)", preset.ID, preset.Name, preset.Endpoint, preset.Model, preset.System, string(params))
	if err != nil {
		return PlaygroundPreset{}, err
	}
	stored, err := GetPlaygroundPreset(db, preset.ID)
	if err != nil {
		return PlaygroundPreset{}, err
	}
	return *stored, nil
}

func UpdatePlaygroundPreset(db *sql.DB, id string, patch map[string]interface{}) (*PlaygroundPreset, error) {
	preset, err := GetPlaygroundPreset(db, id)
	if err != nil || preset == nil {
		return preset, err
	}
	if value, ok := patch["name"].(string); ok {
		preset.Name = value
	}
	if value, ok := patch["endpoint"].(string); ok {
		preset.Endpoint = value
	}
	if value, ok := patch["model"].(string); ok {
		preset.Model = value
	}
	if value, exists := patch["system"]; exists {
		if value == nil {
			preset.System = nil
		} else if text, ok := value.(string); ok {
			preset.System = &text
		}
	}
	if value, ok := patch["params"].(map[string]interface{}); ok {
		preset.Params = value
	}
	params, err := json.Marshal(preset.Params)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("UPDATE playground_presets SET name=?,endpoint=?,model=?,system=?,params_json=? WHERE id=?", preset.Name, preset.Endpoint, preset.Model, preset.System, string(params), id)
	if err != nil {
		return nil, err
	}
	return GetPlaygroundPreset(db, id)
}

func DeletePlaygroundPreset(db *sql.DB, id string) (bool, error) {
	result, err := db.Exec("DELETE FROM playground_presets WHERE id=?", id)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

type scanPreset func(...interface{}) error

func scanPlaygroundPreset(scan scanPreset) (PlaygroundPreset, error) {
	var preset PlaygroundPreset
	var system sql.NullString
	var params string
	err := scan(&preset.ID, &preset.Name, &preset.Endpoint, &preset.Model, &system, &params, &preset.CreatedAt)
	if err != nil {
		return PlaygroundPreset{}, err
	}
	if system.Valid {
		preset.System = &system.String
	}
	if err := json.Unmarshal([]byte(params), &preset.Params); err != nil {
		preset.Params = map[string]interface{}{}
	}
	return preset, nil
}
