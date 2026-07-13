package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// SettingValue represents a stored setting value.
type SettingValue struct {
	Namespace string                 `json:"namespace"`
	Key       string                 `json:"key"`
	Value     map[string]interface{} `json:"value"`
}

// GetSetting retrieves a single setting from the key_value store.
func GetSetting(dbConn *sql.DB, namespace, key string) (map[string]interface{}, error) {
	var raw string
	err := dbConn.QueryRow("SELECT value FROM key_value WHERE namespace = ? AND key = ?", namespace, key).Scan(&raw)
	if err == sql.ErrNoRows {
		return map[string]interface{}{}, nil
	}
	if err != nil {
		return nil, err
	}
	var val map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &val); err != nil {
		return nil, fmt.Errorf("unmarshal setting: %w", err)
	}
	return val, nil
}

// SetSetting stores a setting in the key_value store.
func SetSetting(dbConn *sql.DB, namespace, key string, value map[string]interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal setting: %w", err)
	}
	_, err = dbConn.Exec(`
		INSERT INTO key_value (namespace, key, value)
		VALUES (?, ?, ?)
		ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value`,
		namespace, key, string(raw))
	return err
}

// DeleteSetting removes a setting from the key_value store.
func DeleteSetting(dbConn *sql.DB, namespace, key string) error {
	_, err := dbConn.Exec("DELETE FROM key_value WHERE namespace = ? AND key = ?", namespace, key)
	return err
}

// ListSettingsByNamespace retrieves all settings under a namespace.
func ListSettingsByNamespace(dbConn *sql.DB, namespace string) (map[string]map[string]interface{}, error) {
	rows, err := dbConn.Query("SELECT key, value FROM key_value WHERE namespace = ?", namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]interface{})
	for rows.Next() {
		var key, raw string
		if err := rows.Scan(&key, &raw); err != nil {
			continue
		}
		var val map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &val); err != nil {
			continue
		}
		result[key] = val
	}
	return result, rows.Err()
}
