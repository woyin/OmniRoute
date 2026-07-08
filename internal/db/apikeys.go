package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// APIKey represents a stored API key.
type APIKey struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	IsActive  bool     `json:"isActive"`
	Scopes    []string `json:"scopes"`
	CreatedAt time.Time `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

// ValidateAPIKey checks if an API key is valid and active.
func ValidateAPIKey(db *sql.DB, key string) (*APIKey, error) {
	var ak APIKey
	var isActive int
	var scopesJSON string
	var createdAt string
	var lastUsedAt sql.NullString

	err := db.QueryRow(
		"SELECT key, name, is_active, scopes, created_at, last_used_at FROM api_keys WHERE key = ? AND is_active = 1",
		key,
	).Scan(&ak.Key, &ak.Name, &isActive, &scopesJSON, &createdAt, &lastUsedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("validate api key: %w", err)
	}

	ak.IsActive = isActive == 1
	json.Unmarshal([]byte(scopesJSON), &ak.Scopes)
	ak.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	if lastUsedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lastUsedAt.String)
		ak.LastUsedAt = &t
	}

	// Update last used timestamp
	go db.Exec("UPDATE api_keys SET last_used_at = CURRENT_TIMESTAMP WHERE key = ?", key)

	return &ak, nil
}

// ListAPIKeys returns all API keys.
func ListAPIKeys(db *sql.DB) ([]APIKey, error) {
	rows, err := db.Query("SELECT key, name, is_active, scopes, created_at, last_used_at FROM api_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]APIKey, 0)
	for rows.Next() {
		var ak APIKey
		var isActive int
		var scopesJSON string
		var createdAt string
		var lastUsedAt sql.NullString
		err := rows.Scan(&ak.Key, &ak.Name, &isActive, &scopesJSON, &createdAt, &lastUsedAt)
		if err != nil {
			return nil, err
		}
		ak.IsActive = isActive == 1
		json.Unmarshal([]byte(scopesJSON), &ak.Scopes)
		ak.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		if lastUsedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", lastUsedAt.String)
			ak.LastUsedAt = &t
		}
		results = append(results, ak)
	}
	return results, nil
}

// CreateAPIKey creates a new API key.
func CreateAPIKey(db *sql.DB, ak APIKey) error {
	scopesJSON, _ := json.Marshal(ak.Scopes)
	isActive := 0
	if ak.IsActive {
		isActive = 1
	}
	_, err := db.Exec("INSERT INTO api_keys (key, name, is_active, scopes) VALUES (?, ?, ?, ?)",
		ak.Key, ak.Name, isActive, string(scopesJSON))
	return err
}

// DeleteAPIKey removes an API key.
func DeleteAPIKey(db *sql.DB, key string) error {
	_, err := db.Exec("DELETE FROM api_keys WHERE key = ?", key)
	return err
}
