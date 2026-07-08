package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ProviderConnection represents a stored provider connection.
type ProviderConnection struct {
	ID                   string                 `json:"id"`
	Provider             string                 `json:"provider"`
	Name                 string                 `json:"name"`
	APIKey               string                 `json:"apiKey,omitempty"`
	AccessToken          string                 `json:"accessToken,omitempty"`
	RefreshToken         string                 `json:"refreshToken,omitempty"`
	ProjectID            string                 `json:"projectId,omitempty"`
	ExpiresAt            string                 `json:"expiresAt,omitempty"`
	IsActive             bool                   `json:"isActive"`
	TestStatus           string                 `json:"testStatus"`
	Priority             int                    `json:"priority"`
	ProviderSpecificData map[string]interface{} `json:"providerSpecificData"`
	CreatedAt            time.Time              `json:"createdAt"`
	UpdatedAt            time.Time              `json:"updatedAt"`
}

// ListProviderConnections returns all provider connections, optionally filtered by provider.
func ListProviderConnections(db *sql.DB, provider string) ([]ProviderConnection, error) {
	var rows *sql.Rows
	var err error

	if provider != "" {
		rows, err = db.Query("SELECT id, provider, name, api_key, access_token, refresh_token, project_id, expires_at, is_active, test_status, priority, provider_specific_data, created_at, updated_at FROM provider_connections WHERE provider = ? ORDER BY priority DESC", provider)
	} else {
		rows, err = db.Query("SELECT id, provider, name, api_key, access_token, refresh_token, project_id, expires_at, is_active, test_status, priority, provider_specific_data, created_at, updated_at FROM provider_connections ORDER BY priority DESC")
	}
	if err != nil {
		return nil, fmt.Errorf("query provider_connections: %w", err)
	}
	defer rows.Close()

	results := make([]ProviderConnection, 0)
	for rows.Next() {
		var pc ProviderConnection
		var isActive int
		var psdJSON string
		var createdAt, updatedAt string

		err := rows.Scan(&pc.ID, &pc.Provider, &pc.Name, &pc.APIKey, &pc.AccessToken,
			&pc.RefreshToken, &pc.ProjectID, &pc.ExpiresAt, &isActive, &pc.TestStatus,
			&pc.Priority, &psdJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan provider_connection: %w", err)
		}
		pc.IsActive = isActive == 1
		pc.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		pc.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		if psdJSON != "" {
			json.Unmarshal([]byte(psdJSON), &pc.ProviderSpecificData)
		}
		if pc.ProviderSpecificData == nil {
			pc.ProviderSpecificData = make(map[string]interface{})
		}
		results = append(results, pc)
	}
	return results, nil
}

// GetProviderConnection returns a single provider connection by ID.
func GetProviderConnection(db *sql.DB, id string) (*ProviderConnection, error) {
	var pc ProviderConnection
	var isActive int
	var psdJSON string
	var createdAt, updatedAt string

	err := db.QueryRow(
		"SELECT id, provider, name, api_key, access_token, refresh_token, project_id, expires_at, is_active, test_status, priority, provider_specific_data, created_at, updated_at FROM provider_connections WHERE id = ?",
		id,
	).Scan(&pc.ID, &pc.Provider, &pc.Name, &pc.APIKey, &pc.AccessToken,
		&pc.RefreshToken, &pc.ProjectID, &pc.ExpiresAt, &isActive, &pc.TestStatus,
		&pc.Priority, &psdJSON, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get provider_connection %s: %w", id, err)
	}

	pc.IsActive = isActive == 1
	pc.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	pc.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if psdJSON != "" {
		json.Unmarshal([]byte(psdJSON), &pc.ProviderSpecificData)
	}
	if pc.ProviderSpecificData == nil {
		pc.ProviderSpecificData = make(map[string]interface{})
	}
	return &pc, nil
}

// SaveProviderConnection creates or updates a provider connection.
func SaveProviderConnection(db *sql.DB, pc ProviderConnection) error {
	psdJSON, _ := json.Marshal(pc.ProviderSpecificData)
	isActive := 0
	if pc.IsActive {
		isActive = 1
	}

	_, err := db.Exec(`INSERT INTO provider_connections (id, provider, name, api_key, access_token, refresh_token, project_id, expires_at, is_active, test_status, priority, provider_specific_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider = excluded.provider,
			name = excluded.name,
			api_key = excluded.api_key,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			project_id = excluded.project_id,
			expires_at = excluded.expires_at,
			is_active = excluded.is_active,
			test_status = excluded.test_status,
			priority = excluded.priority,
			provider_specific_data = excluded.provider_specific_data,
			updated_at = CURRENT_TIMESTAMP`,
		pc.ID, pc.Provider, pc.Name, pc.APIKey, pc.AccessToken, pc.RefreshToken,
		pc.ProjectID, pc.ExpiresAt, isActive, pc.TestStatus, pc.Priority, string(psdJSON))
	return err
}

// DeleteProviderConnection removes a provider connection.
func DeleteProviderConnection(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM provider_connections WHERE id = ?", id)
	return err
}

// GetActiveProviderConnections returns active connections for a provider.
func GetActiveProviderConnections(db *sql.DB, provider string) ([]ProviderConnection, error) {
	rows, err := db.Query(
		"SELECT id, provider, name, api_key, access_token, refresh_token, project_id, expires_at, is_active, test_status, priority, provider_specific_data, created_at, updated_at FROM provider_connections WHERE provider = ? AND is_active = 1 ORDER BY priority DESC",
		provider)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ProviderConnection, 0)
	for rows.Next() {
		var pc ProviderConnection
		var isActive int
		var psdJSON string
		var createdAt, updatedAt string
		err := rows.Scan(&pc.ID, &pc.Provider, &pc.Name, &pc.APIKey, &pc.AccessToken,
			&pc.RefreshToken, &pc.ProjectID, &pc.ExpiresAt, &isActive, &pc.TestStatus,
			&pc.Priority, &psdJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		pc.IsActive = isActive == 1
		if psdJSON != "" {
			json.Unmarshal([]byte(psdJSON), &pc.ProviderSpecificData)
		}
		if pc.ProviderSpecificData == nil {
			pc.ProviderSpecificData = make(map[string]interface{})
		}
		results = append(results, pc)
	}
	return results, nil
}
