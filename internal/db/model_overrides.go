package db

import (
	"database/sql"
	"strconv"
)

type ModelCapabilityOverride struct {
	Provider    string `json:"provider"`
	ModelID     string `json:"modelId"`
	Target      string `json:"target"`
	Key         string `json:"key"`
	Value       int    `json:"value"`
	RefreshedAt string `json:"refreshedAt"`
}

func ListModelCapabilityOverrides(db *sql.DB) ([]ModelCapabilityOverride, error) {
	rows, err := db.Query("SELECT provider,model_id,override_key,override_value,refreshed_at FROM model_capability_overrides ORDER BY refreshed_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]ModelCapabilityOverride, 0)
	for rows.Next() {
		var item ModelCapabilityOverride
		var value string
		if err := rows.Scan(&item.Provider, &item.ModelID, &item.Key, &value, &item.RefreshedAt); err != nil {
			return nil, err
		}
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 || item.Key != "max_token" {
			continue
		}
		item.Value, item.Target = parsed, item.Provider+"/"+item.ModelID
		result = append(result, item)
	}
	return result, rows.Err()
}

func SetModelCapabilityOverride(db *sql.DB, provider, model, key string, value int) error {
	_, err := db.Exec(`INSERT INTO model_capability_overrides(provider,model_id,override_key,override_value,refreshed_at) VALUES(?,?,?,?,datetime('now'))
		ON CONFLICT(provider,model_id,override_key) DO UPDATE SET override_value=excluded.override_value,refreshed_at=excluded.refreshed_at`, provider, model, key, strconv.Itoa(value))
	return err
}

func DeleteModelCapabilityOverride(db *sql.DB, provider, model, key string) error {
	_, err := db.Exec("DELETE FROM model_capability_overrides WHERE provider=? AND model_id=? AND override_key=?", provider, model, key)
	return err
}
