package mcp

func (s *MCPServer) registerSkillTools() {
	s.simpleTool("skills_list", "List all available skills", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"skills": []interface{}{}}, nil
		})

	s.simpleTool("skills_enable", "Enable or disable a skill",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"skillId": map[string]interface{}{"type": "string"}, "enabled": map[string]interface{}{"type": "boolean"}}, "required": []string{"skillId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true}, nil
		})

	s.simpleTool("skills_execute", "Execute a skill by ID",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"skillId": map[string]interface{}{"type": "string"}, "input": map[string]interface{}{"type": "object"}}, "required": []string{"skillId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": false, "message": "skill execution not yet implemented"}, nil
		})

	s.simpleTool("skills_executions", "List skill execution history",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"limit": map[string]interface{}{"type": "number"}}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"executions": []interface{}{}}, nil
		})
}
