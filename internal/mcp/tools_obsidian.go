package mcp

func (s *MCPServer) registerObsidianTools() {
	s.simpleTool("obsidian_search", "Search Obsidian vault",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"results": []interface{}{}}, nil })

	s.simpleTool("obsidian_read_note", "Read an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"content": ""}, nil })

	s.simpleTool("obsidian_create_note", "Create an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path", "content"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	s.simpleTool("obsidian_update_note", "Update an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	s.simpleTool("obsidian_delete_note", "Delete an Obsidian note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	s.simpleTool("obsidian_list_notes", "List notes in a folder",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"folder": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"notes": []interface{}{}}, nil })

	s.simpleTool("obsidian_list_tags", "List all tags in vault", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"tags": []interface{}{}}, nil })

	s.simpleTool("obsidian_list_links", "List all links in vault", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"links": []interface{}{}}, nil })

	s.simpleTool("obsidian_get_backlinks", "Get backlinks for a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"backlinks": []interface{}{}}, nil })

	s.simpleTool("obsidian_get_metadata", "Get note metadata (frontmatter)",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"metadata": map[string]interface{}{}}, nil })

	s.simpleTool("obsidian_search_tags", "Search notes by tag",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"tags": map[string]interface{}{"type": "array"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"results": []interface{}{}}, nil })

	s.simpleTool("obsidian_append_note", "Append content to a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path", "content"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	s.simpleTool("obsidian_move_note", "Move/rename a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"fromPath": map[string]interface{}{"type": "string"}, "toPath": map[string]interface{}{"type": "string"}}, "required": []string{"fromPath", "toPath"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	s.simpleTool("obsidian_get_graph", "Get vault graph data", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}}, nil })

	s.simpleTool("obsidian_get_daily_note", "Get or create daily note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"date": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"content": ""}, nil })

	s.simpleTool("obsidian_list_templates", "List available templates", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"templates": []interface{}{}}, nil })

	s.simpleTool("obsidian_apply_template", "Apply a template to a note",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "template": map[string]interface{}{"type": "string"}}, "required": []string{"path", "template"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil })

	s.simpleTool("obsidian_webdav_status", "Check WebDAV connection status", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"connected": false}, nil })

	s.simpleTool("obsidian_webdav_sync", "Trigger WebDAV sync", nil,
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"synced": false, "message": "WebDAV not configured"}, nil })

	s.simpleTool("obsidian_webdav_list", "List files via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"files": []interface{}{}}, nil })

	s.simpleTool("obsidian_webdav_read", "Read a file via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"content": ""}, nil })

	s.simpleTool("obsidian_webdav_write", "Write a file via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}, "content": map[string]interface{}{"type": "string"}}, "required": []string{"path", "content"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "WebDAV not configured"}, nil })

	s.simpleTool("obsidian_webdav_delete", "Delete a file via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "WebDAV not configured"}, nil })

	s.simpleTool("obsidian_webdav_mkdir", "Create directory via WebDAV",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{"path": map[string]interface{}{"type": "string"}}, "required": []string{"path"}},
		func(args map[string]interface{}) (interface{}, error) { return map[string]interface{}{"success": false, "message": "WebDAV not configured"}, nil })

	s.simpleTool("obsidian_check_status", "Check if the Obsidian local API is reachable", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"reachable": false, "version": ""}, nil
		})

	s.simpleTool("obsidian_search_structured", "Search Obsidian vault using JSON Logic queries",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"query": map[string]interface{}{"type": "object", "description": "JSON Logic query"},
			"limit": map[string]interface{}{"type": "number"},
		}, "required": []string{"query"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"results": []interface{}{}}, nil
		})

	s.simpleTool("obsidian_get_document_map", "Get the document relationship map for the vault", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"nodes": []interface{}{}, "edges": []interface{}{}}, nil
		})

	s.simpleTool("obsidian_get_active_file", "Get the currently open file in Obsidian", nil,
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"path": "", "content": "", "error": "Obsidian integration not configured"}, nil
		})

	s.simpleTool("obsidian_execute_command", "Execute a registered Obsidian command",
		map[string]interface{}{"type": "object", "properties": map[string]interface{}{
			"commandId": map[string]interface{}{"type": "string"},
		}, "required": []string{"commandId"}},
		func(args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": false, "message": "Obsidian integration not configured"}, nil
		})
}
