package mcp

import "github.com/google/uuid"

func (s *MCPServer) registerAgentTools() {
	s.simpleTool("agent_skill_discover", "Discover available agent skills via A2A",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"agentId": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"skills": []interface{}{}, "agentId": args["agentId"]}, nil
		})

	s.simpleTool("agent_skill_invoke", "Invoke an agent skill",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"skillId": map[string]interface{}{"type": "string"}, "input": map[string]interface{}{"type": "object"}}, "required": []string{"skillId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"invocationId": uuid.New().String(), "status": "queued"}, nil
		})

	s.simpleTool("agent_skill_status", "Check agent skill execution status",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"invocationId": map[string]interface{}{"type": "string"}}, "required": []string{"invocationId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"invocationId": args["invocationId"], "status": "completed"}, nil
		})
}
