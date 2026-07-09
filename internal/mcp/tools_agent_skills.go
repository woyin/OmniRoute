package mcp

func (s *MCPServer) registerAgentSkillsTools() {
	s.simpleTool("omniroute_agent_skills_list", "List agent skills with filtering",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"filter":  map[string]interface{}{"type": "string"},
			"enabled": map[string]interface{}{"type": "boolean"},
		}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"skills": []interface{}{}, "total": 0}, nil
		})

	s.simpleTool("omniroute_agent_skills_get", "Get skill metadata and SKILL.md content",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"skillId": map[string]interface{}{"type": "string"},
		}, "required": []string{"skillId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"skillId":  args["skillId"],
				"metadata": map[string]interface{}{},
				"content":  "",
			}, nil
		})

	s.simpleTool("omniroute_agent_skills_coverage", "Compute skill coverage stats",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"agentId": map[string]interface{}{"type": "string"},
		}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"totalSkills":  0,
				"coveredSkills": 0,
				"coveragePct":  0,
			}, nil
		})
}
