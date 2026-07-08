package db

import (
	"database/sql"
	"encoding/json"
	
	"time"
)

// ComboTarget represents a single target in a combo configuration.
type ComboTarget struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Account  string `json:"account,omitempty"`
	Weight   int    `json:"weight,omitempty"`
}

// Combo represents a routing combo configuration.
type Combo struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Strategy  string        `json:"strategy"`
	Targets   []ComboTarget `json:"targets"`
	IsActive  bool          `json:"isActive"`
	Domain    string        `json:"domain"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// ListCombos returns all combos.
func ListCombos(db *sql.DB) ([]Combo, error) {
	rows, err := db.Query(
		"SELECT id, name, strategy, targets, is_active, domain, created_at, updated_at FROM combos ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Combo, 0)
	for rows.Next() {
		var c Combo
		var isActive int
		var targetsJSON string
		var createdAt, updatedAt string
		err := rows.Scan(&c.ID, &c.Name, &c.Strategy, &targetsJSON, &isActive, &c.Domain, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		c.IsActive = isActive == 1
		c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		c.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		json.Unmarshal([]byte(targetsJSON), &c.Targets)
		if c.Targets == nil {
			c.Targets = []ComboTarget{}
		}
		results = append(results, c)
	}
	return results, nil
}

// GetCombo returns a combo by ID.
func GetCombo(db *sql.DB, id string) (*Combo, error) {
	var c Combo
	var isActive int
	var targetsJSON string
	var createdAt, updatedAt string

	err := db.QueryRow(
		"SELECT id, name, strategy, targets, is_active, domain, created_at, updated_at FROM combos WHERE id = ?", id,
	).Scan(&c.ID, &c.Name, &c.Strategy, &targetsJSON, &isActive, &c.Domain, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.IsActive = isActive == 1
	json.Unmarshal([]byte(targetsJSON), &c.Targets)
	if c.Targets == nil {
		c.Targets = []ComboTarget{}
	}
	return &c, nil
}

// SaveCombo creates or updates a combo.
func SaveCombo(db *sql.DB, c Combo) error {
	targetsJSON, _ := json.Marshal(c.Targets)
	isActive := 0
	if c.IsActive {
		isActive = 1
	}
	_, err := db.Exec(`INSERT INTO combos (id, name, strategy, targets, is_active, domain)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			strategy = excluded.strategy,
			targets = excluded.targets,
			is_active = excluded.is_active,
			domain = excluded.domain,
			updated_at = CURRENT_TIMESTAMP`,
		c.ID, c.Name, c.Strategy, string(targetsJSON), isActive, c.Domain)
	return err
}

// DeleteCombo removes a combo by ID.
func DeleteCombo(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM combos WHERE id = ?", id)
	return err
}

// GetActiveCombos returns all active combos.
func GetActiveCombos(db *sql.DB) ([]Combo, error) {
	rows, err := db.Query(
		"SELECT id, name, strategy, targets, is_active, domain, created_at, updated_at FROM combos WHERE is_active = 1 ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Combo, 0)
	for rows.Next() {
		var c Combo
		var isActive int
		var targetsJSON string
		var createdAt, updatedAt string
		err := rows.Scan(&c.ID, &c.Name, &c.Strategy, &targetsJSON, &isActive, &c.Domain, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		c.IsActive = isActive == 1
		json.Unmarshal([]byte(targetsJSON), &c.Targets)
		if c.Targets == nil {
			c.Targets = []ComboTarget{}
		}
		results = append(results, c)
	}
	return results, nil
}
