package db

import (
	"database/sql"
)

type ReasoningCache struct {
	ID               string `json:"id"`
	SessionID        string `json:"sessionId"`
	Model            string `json:"model"`
	ReasoningContent string `json:"reasoningContent"`
	TurnIndex        int    `json:"turnIndex"`
	CreatedAt        string `json:"createdAt"`
}

func SaveReasoning(db *sql.DB, r ReasoningCache) error {
	_, err := db.Exec(
		"INSERT INTO reasoning_cache (id, session_id, model, reasoning_content, turn_index) VALUES (?, ?, ?, ?, ?) "+
			"ON CONFLICT(id) DO UPDATE SET reasoning_content=excluded.reasoning_content",
		r.ID, r.SessionID, r.Model, r.ReasoningContent, r.TurnIndex)
	return err
}

func GetReasoningForSession(db *sql.DB, sessionID string) ([]ReasoningCache, error) {
	rows, err := db.Query(
		"SELECT id, session_id, model, reasoning_content, turn_index, created_at FROM reasoning_cache WHERE session_id = ? ORDER BY turn_index",
		sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []ReasoningCache
	for rows.Next() {
		var r ReasoningCache
		if err := rows.Scan(&r.ID, &r.SessionID, &r.Model, &r.ReasoningContent, &r.TurnIndex, &r.CreatedAt); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}
