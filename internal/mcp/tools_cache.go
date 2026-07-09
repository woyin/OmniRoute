package mcp

func (s *MCPServer) registerCacheTools() {
	s.simpleTool("cache_stats", "Get semantic cache statistics", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"entries": 0, "hitRate": 0}, nil
		})

	s.simpleTool("cache_flush", "Flush the semantic cache", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true, "flushed": 0}, nil
		})
}
