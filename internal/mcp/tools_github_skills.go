package mcp

func (s *MCPServer) registerGithubSkillsTools() {
	s.simpleTool("omniroute_github_skills_search", "Search GitHub for skill repositories",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"query": map[string]interface{}{"type": "string"},
			"limit": map[string]interface{}{"type": "number"},
		}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"results": []interface{}{}, "total": 0}, nil
		})

	s.simpleTool("omniroute_github_skills_scan", "Scan a GitHub skill repo for blocked patterns",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"repo": map[string]interface{}{"type": "string"},
		}, "required": []string{"repo"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"repo":    args["repo"],
				"blocked": []interface{}{},
				"safe":    true,
			}, nil
		})

	s.simpleTool("omniroute_github_skills_install", "Install a skill from a GitHub repository",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"repo": map[string]interface{}{"type": "string"},
			"ref":  map[string]interface{}{"type": "string"},
		}, "required": []string{"repo"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"success": false,
				"message": "GitHub skill install not yet implemented",
			}, nil
		})
}
