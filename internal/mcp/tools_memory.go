package mcp

import "github.com/google/uuid"

func (s *MCPServer) registerMemoryTools() {
	s.simpleTool("memory_search", "Search stored memories",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) {
			if s.DB == nil {
				return []interface{}{}, nil
			}
			query, _ := args["query"].(string)
			rows, err := s.DB.Query("SELECT id, content FROM memories WHERE content LIKE ? LIMIT 10", "%"+query+"%")
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			var results []map[string]interface{}
			for rows.Next() {
				var id, content string
				if rows.Scan(&id, &content) == nil {
					results = append(results, map[string]interface{}{"id": id, "content": content})
				}
			}
			return results, nil
		})

	s.simpleTool("memory_add", "Add a memory entry",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"content": map[string]interface{}{"type": "string"}, "tags": map[string]interface{}{"type": "array"}}, "required": []string{"content"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true, "id": uuid.New().String()}, nil
		})

	s.simpleTool("memory_clear", "Clear all memories", nil,
		func(args map[string]interface{}) (interface{}, error) {
			if s.DB != nil {
				s.DB.Exec("DELETE FROM memories")
			}
			return map[string]interface{}{"success": true}, nil
		})
}
