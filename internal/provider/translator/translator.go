package translator

import "strings"

// Format constants matching the TypeScript FORMATS.
const (
	FormatOpenAI          = "openai"
	FormatOpenAIResponses = "openai-responses"
	FormatClaude          = "claude"
	FormatGemini          = "gemini"
)

// TranslateRequest translates a request body from one format to another.
func TranslateRequest(body map[string]interface{}, sourceFormat, targetFormat string) map[string]interface{} {
	if sourceFormat == targetFormat {
		return body
	}

	switch {
	case sourceFormat == FormatOpenAI && targetFormat == FormatClaude:
		return openAIToClaude(body)
	case sourceFormat == FormatClaude && targetFormat == FormatOpenAI:
		return claudeToOpenAI(body)
	case sourceFormat == FormatOpenAI && targetFormat == FormatGemini:
		return openAIToGemini(body)
	case sourceFormat == FormatGemini && targetFormat == FormatOpenAI:
		return geminiToOpenAI(body)
	default:
		return body
	}
}

// NeedsTranslation returns true if source and target formats differ.
func NeedsTranslation(sourceFormat, targetFormat string) bool {
	return sourceFormat != targetFormat && sourceFormat != "" && targetFormat != ""
}

// openAIToClaude converts an OpenAI Chat Completions request to Anthropic Messages format.
func openAIToClaude(body map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"model": body["model"],
	}

	// Map max_tokens (Anthropic requires it)
	if v, ok := body["max_tokens"]; ok {
		result["max_tokens"] = v
	} else if v, ok := body["max_completion_tokens"]; ok {
		result["max_tokens"] = v
	} else {
		result["max_tokens"] = 4096
	}

	// Map stream
	if v, ok := body["stream"]; ok {
		result["stream"] = v
	}

	// Extract system from messages and convert messages
	var systemParts []interface{}
	var messages []interface{}
	if msgs, ok := body["messages"].([]interface{}); ok {
		for _, msg := range msgs {
			m, ok := msg.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			switch role {
			case "system", "developer":
				systemParts = append(systemParts, map[string]interface{}{
					"type": "text",
					"text": extractTextContent(m["content"]),
				})
			case "assistant":
				// Handle tool_calls in assistant messages
				if toolCalls, ok := m["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
					var content []interface{}
					// Add text content if present
					if textContent := extractTextContent(m["content"]); textContent != "" {
						content = append(content, map[string]interface{}{
							"type": "text",
							"text": textContent,
						})
					}
					// Add tool_use blocks
					for _, tc := range toolCalls {
						tcm, ok := tc.(map[string]interface{})
						if !ok {
							continue
						}
						content = append(content, map[string]interface{}{
							"type":  "tool_use",
							"id":    tcm["id"],
							"name":  funcName(tcm),
							"input": funcArgs(tcm),
						})
					}
					messages = append(messages, map[string]interface{}{
						"role":    "assistant",
						"content": content,
					})
				} else {
					messages = append(messages, convertToClaudeMessage(m))
				}
			case "tool":
				// Convert tool response to tool_result
				messages = append(messages, map[string]interface{}{
					"role": "user",
					"content": []interface{}{
						map[string]interface{}{
							"type":       "tool_result",
							"tool_use_id": m["tool_call_id"],
							"content":    m["content"],
						},
					},
				})
			default:
				messages = append(messages, convertToClaudeMessage(m))
			}
		}
	}
	result["messages"] = messages
	if len(systemParts) > 0 {
		result["system"] = systemParts
	}

	// Map tools
	if tools, ok := body["tools"].([]interface{}); ok {
		var claudeTools []interface{}
		for _, tool := range tools {
			t, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			fn, _ := t["function"].(map[string]interface{})
			if fn == nil {
				continue
			}
			claudeTool := map[string]interface{}{
				"name":         fn["name"],
				"description":  fn["description"],
				"input_schema": fn["parameters"],
			}
			claudeTools = append(claudeTools, claudeTool)
		}
		result["tools"] = claudeTools
	}

	// Map tool_choice
	if tc, ok := body["tool_choice"]; ok {
		result["tool_choice"] = mapToolChoiceToClaude(tc)
	}

	// Map reasoning/thinking
	if re, ok := body["reasoning_effort"]; ok {
		result["thinking"] = map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": mapReasoningBudget(re),
		}
	}

	// Map temperature, top_p, stop_sequences
	if v, ok := body["temperature"]; ok {
		result["temperature"] = v
	}
	if v, ok := body["top_p"]; ok {
		result["top_p"] = v
	}
	if v, ok := body["stop"]; ok {
		result["stop_sequences"] = v
	}

	// Anthropic version header
	result["anthropic_version"] = "2023-06-01"

	return result
}

// claudeToOpenAI converts an Anthropic Messages request to OpenAI Chat Completions format.
func claudeToOpenAI(body map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"model": body["model"],
	}

	// Map max_tokens
	if v, ok := body["max_tokens"]; ok {
		result["max_tokens"] = v
	}

	// Map stream
	if v, ok := body["stream"]; ok {
		result["stream"] = v
	}

	// Map system to messages
	var messages []interface{}
	if system, ok := body["system"].([]interface{}); ok {
		for _, s := range system {
			if sm, ok := s.(map[string]interface{}); ok {
				messages = append(messages, map[string]interface{}{
					"role":    "system",
					"content": sm["text"],
				})
			}
		}
	}
	if system, ok := body["system"].(string); ok && system != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": system,
		})
	}

	// Map Claude messages to OpenAI format
	if msgs, ok := body["messages"].([]interface{}); ok {
		for _, msg := range msgs {
			m, ok := msg.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			content := m["content"]

			// Handle content blocks (tool_use, tool_result, text, image)
			switch role {
			case "user":
				if contentBlocks, ok := content.([]interface{}); ok {
					// Check for tool_result blocks
					var toolResults []interface{}
					var textParts []string
					hasToolResult := false
					for _, cb := range contentBlocks {
						block, ok := cb.(map[string]interface{})
						if !ok {
							continue
						}
						blockType, _ := block["type"].(string)
						switch blockType {
						case "tool_result":
							hasToolResult = true
							toolResults = append(toolResults, map[string]interface{}{
								"role":         "tool",
								"tool_call_id": block["tool_use_id"],
								"content":      block["content"],
							})
						case "text":
							if text, ok := block["text"].(string); ok {
								textParts = append(textParts, text)
							}
						}
					}
					if hasToolResult {
						if len(textParts) > 0 {
							messages = append(messages, map[string]interface{}{
								"role":    "user",
								"content": strings.Join(textParts, "\n"),
							})
						}
						messages = append(messages, toolResults...)
					} else {
						messages = append(messages, convertFromClaudeMessage(m))
					}
				} else {
					messages = append(messages, map[string]interface{}{
						"role":    "user",
						"content": content,
					})
				}
			case "assistant":
				if contentBlocks, ok := content.([]interface{}); ok {
					var toolCalls []interface{}
					var textParts []string
					hasToolUse := false
					for _, cb := range contentBlocks {
						block, ok := cb.(map[string]interface{})
						if !ok {
							continue
						}
						blockType, _ := block["type"].(string)
						switch blockType {
						case "tool_use":
							hasToolUse = true
							toolCalls = append(toolCalls, map[string]interface{}{
								"id":   block["id"],
								"type": "function",
								"function": map[string]interface{}{
									"name":      block["name"],
									"arguments": serializeArgs(block["input"]),
								},
							})
						case "text":
							if text, ok := block["text"].(string); ok {
								textParts = append(textParts, text)
							}
						}
					}
					if hasToolUse {
						openaiMsg := map[string]interface{}{
							"role": "assistant",
						}
						if len(textParts) > 0 {
							openaiMsg["content"] = strings.Join(textParts, "\n")
						}
						openaiMsg["tool_calls"] = toolCalls
						messages = append(messages, openaiMsg)
					} else if len(textParts) > 0 {
						messages = append(messages, map[string]interface{}{
							"role":    "assistant",
							"content": strings.Join(textParts, "\n"),
						})
					}
				} else {
					messages = append(messages, map[string]interface{}{
						"role":    "assistant",
						"content": content,
					})
				}
			default:
				messages = append(messages, map[string]interface{}{
					"role":    role,
					"content": content,
				})
			}
		}
	}
	result["messages"] = messages

	// Map tools
	if tools, ok := body["tools"].([]interface{}); ok {
		var openaiTools []interface{}
		for _, tool := range tools {
			t, ok := tool.(map[string]interface{})
			if !ok {
				continue
			}
			openaiTools = append(openaiTools, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        t["name"],
					"description": t["description"],
					"parameters":  t["input_schema"],
				},
			})
		}
		result["tools"] = openaiTools
	}

	// Map temperature, top_p, stop_sequences → stop
	if v, ok := body["temperature"]; ok {
		result["temperature"] = v
	}
	if v, ok := body["top_p"]; ok {
		result["top_p"] = v
	}
	if v, ok := body["stop_sequences"]; ok {
		result["stop"] = v
	}

	return result
}

// openAIToGemini converts an OpenAI request to Gemini format.
func openAIToGemini(body map[string]interface{}) map[string]interface{} {
	// Simplified Gemini translation
	var contents []interface{}
	if msgs, ok := body["messages"].([]interface{}); ok {
		for _, msg := range msgs {
			m, ok := msg.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			geminiRole := "user"
			if role == "assistant" {
				geminiRole = "model"
			}
			contents = append(contents, map[string]interface{}{
				"role": geminiRole,
				"parts": []interface{}{
					map[string]interface{}{
						"text": extractTextContent(m["content"]),
					},
				},
			})
		}
	}

	result := map[string]interface{}{
		"contents": contents,
	}

	// Map generation config
	genConfig := map[string]interface{}{}
	if v, ok := body["temperature"]; ok {
		genConfig["temperature"] = v
	}
	if v, ok := body["top_p"]; ok {
		genConfig["topP"] = v
	}
	if v, ok := body["max_tokens"]; ok {
		genConfig["maxOutputTokens"] = v
	}
	if len(genConfig) > 0 {
		result["generationConfig"] = genConfig
	}

	return result
}

// geminiToOpenAI converts a Gemini request to OpenAI format.
func geminiToOpenAI(body map[string]interface{}) map[string]interface{} {
	var messages []interface{}
	if contents, ok := body["contents"].([]interface{}); ok {
		for _, content := range contents {
			c, ok := content.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := c["role"].(string)
			openaiRole := "user"
			if role == "model" {
				openaiRole = "assistant"
			}
			messages = append(messages, map[string]interface{}{
				"role":    openaiRole,
				"content": c["parts"],
			})
		}
	}

	return map[string]interface{}{
		"model":    body["model"],
		"messages": messages,
	}
}

// --- Helper functions ---

// extractTextContent extracts text from various content formats.
func extractTextContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, item := range v {
			if block, ok := item.(map[string]interface{}); ok {
				if t, ok := block["text"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

// convertToClaudeMessage converts an OpenAI message to Claude format.
func convertToClaudeMessage(m map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"role": m["role"],
	}
	// Claude uses content as string or array of content blocks
	content := m["content"]
	switch v := content.(type) {
	case string:
		result["content"] = v
	case []interface{}:
		// Already content blocks
		result["content"] = v
	default:
		result["content"] = content
	}
	return result
}

// convertFromClaudeMessage converts a Claude message to OpenAI format.
func convertFromClaudeMessage(m map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"role":    m["role"],
		"content": extractTextContent(m["content"]),
	}
}

// funcName extracts the function name from a tool_call object.
func funcName(tc map[string]interface{}) interface{} {
	if fn, ok := tc["function"].(map[string]interface{}); ok {
		return fn["name"]
	}
	return tc["name"]
}

// funcArgs extracts the function arguments from a tool_call object.
func funcArgs(tc map[string]interface{}) interface{} {
	if fn, ok := tc["function"].(map[string]interface{}); ok {
		return fn["arguments"]
	}
	return tc["arguments"]
}

// serializeArgs converts arguments to a JSON string if needed.
func serializeArgs(args interface{}) interface{} {
	switch v := args.(type) {
	case string:
		return v
	case map[string]interface{}, []interface{}:
		// For simplicity, return as-is; the executor will serialize
		return v
	default:
		return args
	}
}

// mapToolChoiceToClaude maps OpenAI tool_choice to Anthropic format.
func mapToolChoiceToClaude(tc interface{}) interface{} {
	if tc == nil {
		return nil
	}

	switch v := tc.(type) {
	case string:
		switch v {
		case "auto":
			return map[string]interface{}{"type": "auto"}
		case "none":
			return map[string]interface{}{"type": "none"}
		case "required":
			return map[string]interface{}{"type": "any"}
		default:
			return map[string]interface{}{"type": "auto"}
		}
	case map[string]interface{}:
		if fn, ok := v["function"].(map[string]interface{}); ok {
			return map[string]interface{}{
				"type": "tool",
				"name": fn["name"],
			}
		}
		return map[string]interface{}{"type": "auto"}
	default:
		return map[string]interface{}{"type": "auto"}
	}
}

// mapReasoningBudget maps OpenAI reasoning_effort to Anthropic thinking budget_tokens.
func mapReasoningBudget(effort interface{}) int {
	switch v := effort.(type) {
	case string:
		switch v {
		case "low":
			return 1024
		case "medium":
			return 8192
		case "high":
			return 32768
		case "xhigh":
			return 65536
		default:
			return 8192
		}
	case float64:
		return int(v)
	default:
		return 8192
	}
}
