package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Memory struct {
	ID        string   `json:"id"`
	Content   string   `json:"content"`
	Tags      []string `json:"tags"`
	Provider  string   `json:"provider"`
	SessionID string   `json:"sessionId"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

func ListMemories(db *sql.DB, limit int) ([]Memory, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.Query("SELECT id, content, tags, provider, session_id, created_at, updated_at FROM memories ORDER BY created_at DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]Memory, 0)
	for rows.Next() {
		var m Memory
		var tagsJSON string
		if err := rows.Scan(&m.ID, &m.Content, &tagsJSON, &m.Provider, &m.SessionID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &m.Tags)
		results = append(results, m)
	}
	return results, nil
}

func SaveMemory(db *sql.DB, m Memory) error {
	tagsJSON, _ := json.Marshal(m.Tags)
	_, err := db.Exec(
		"INSERT INTO memories (id, content, tags, provider, session_id, updated_at) VALUES (?, ?, ?, ?, ?, ?) "+
			"ON CONFLICT(id) DO UPDATE SET content=excluded.content, tags=excluded.tags, updated_at=excluded.updated_at",
		m.ID, m.Content, string(tagsJSON), m.Provider, m.SessionID, time.Now().UTC().Format(time.RFC3339))
	return err
}

func DeleteMemory(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM memories WHERE id = ?", id)
	return err
}

func SearchMemories(db *sql.DB, query string, limit int) ([]Memory, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(
		"SELECT id, content, tags, provider, session_id, created_at, updated_at FROM memories WHERE content LIKE ? ORDER BY created_at DESC LIMIT ?",
		"%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]Memory, 0)
	for rows.Next() {
		var m Memory
		var tagsJSON string
		if err := rows.Scan(&m.ID, &m.Content, &tagsJSON, &m.Provider, &m.SessionID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &m.Tags)
		results = append(results, m)
	}
	return results, nil
}
