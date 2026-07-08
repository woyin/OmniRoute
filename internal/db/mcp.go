package db

import (
	"database/sql"
	"encoding/json"
)

type MCPAuditEntry struct {
	ID           int    `json:"id"`
	ToolName     string `json:"toolName"`
	Args         string `json:"args"`
	ResultSummary string `json:"resultSummary"`
	Success      bool   `json:"success"`
	APIKeyID     string `json:"apiKeyId"`
	DurationMs   int    `json:"durationMs"`
	CreatedAt    string `json:"createdAt"`
}

func RecordMCPAudit(db *sql.DB, entry MCPAuditEntry) error {
	argsJSON, _ := json.Marshal(entry.Args)
	_, err := db.Exec(
		"INSERT INTO mcp_tool_audit (tool_name, args, result_summary, success, api_key_id, duration_ms) VALUES (?, ?, ?, ?, ?, ?)",
		entry.ToolName, string(argsJSON), entry.ResultSummary, entry.Success, entry.APIKeyID, entry.DurationMs)
	return err
}

func ListMCPAudit(db *sql.DB, limit int) ([]MCPAuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.Query(
		"SELECT id, tool_name, args, result_summary, success, COALESCE(api_key_id,''), duration_ms, created_at FROM mcp_tool_audit ORDER BY id DESC LIMIT ?",
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MCPAuditEntry
	for rows.Next() {
		var e MCPAuditEntry
		if err := rows.Scan(&e.ID, &e.ToolName, &e.Args, &e.ResultSummary, &e.Success, &e.APIKeyID, &e.DurationMs, &e.CreatedAt); err != nil {
			continue
		}
		results = append(results, e)
	}
	return results, nil
}
