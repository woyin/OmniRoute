package sse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// ResponsesAPIEvent represents a Responses API SSE event.
type ResponsesAPIEvent struct {
	Type       string      `json:"type"`
	ItemID     string      `json:"item_id,omitempty"`
	OutputIndex int        `json:"output_index,omitempty"`
	ContentIndex int       `json:"content_index,omitempty"`
	Delta      string      `json:"delta,omitempty"`
	Text       string      `json:"text,omitempty"`
	Status     string      `json:"status,omitempty"`
	Sequence   int         `json:"sequence,omitempty"`
}

// ResponsesAPIResponse represents the full response object.
type ResponsesAPIResponse struct {
	ID          string                 `json:"id"`
	Object      string                 `json:"object"`
	Model       string                 `json:"model"`
	CreatedAt   int64                  `json:"created_at"`
	Status      string                 `json:"status"`
	Output      []ResponsesAPIOutput   `json:"output"`
	Usage       map[string]interface{} `json:"usage,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ResponsesAPIOutput represents an output item in the Responses API.
type ResponsesAPIOutput struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id"`
	Status  string                 `json:"status"`
	Role    string                 `json:"role,omitempty"`
	Content []ResponsesAPIContent  `json:"content,omitempty"`
	Name    string                 `json:"name,omitempty"`
}

// ResponsesAPIContent represents content within an output item.
type ResponsesAPIContent struct {
	Type        string      `json:"type"`
	Text        string      `json:"text,omitempty"`
	Refusal     string      `json:"refusal,omitempty"`
	ToolCallID  string      `json:"tool_call_id,omitempty"`
	Function    interface{} `json:"function,omitempty"`
}

// TransformChatToResponsesStream converts a Chat Completions SSE stream
// into a Responses API SSE stream.
func TransformChatToResponsesStream(client *Writer, upstream io.Reader, responseID, model string, stopCh <-chan struct{}) error {
	itemID := fmt.Sprintf("item_%s", responseID)
	sequence := 0

	// Send response.created
	createdEvent := ResponsesAPIEvent{
		Type:     "response.created",
		Sequence: sequence,
	}
	client.WriteEvent("response.created", createdEvent)
	sequence++

	// Send response.output_item.added
	addedEvent := map[string]interface{}{
		"type":         "response.output_item.added",
		"output_index": 0,
		"item": map[string]interface{}{
			"type":   "message",
			"id":     itemID,
			"status": "in_progress",
			"role":   "assistant",
			"content": []map[string]interface{}{
				{"type": "output_text", "text": ""},
			},
		},
		"sequence": sequence,
	}
	client.WriteEvent("response.output_item.added", addedEvent)
	sequence++

	// Send content_part.added
	contentAdded := map[string]interface{}{
		"type":           "response.content_part.added",
		"output_index":   0,
		"content_index":  0,
		"part": map[string]interface{}{
			"type": "output_text",
			"text": "",
		},
		"sequence": sequence,
	}
	client.WriteEvent("response.content_part.added", contentAdded)
	sequence++

	// Scan the upstream SSE stream
	scanner := bufio.NewScanner(upstream)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-stopCh:
			return nil
		default:
		}

		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Handle data lines
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				// Send completed events
				completed := map[string]interface{}{
					"type":          "response.output_text.done",
					"output_index":  0,
					"content_index": 0,
					"text":          "",
					"sequence":      sequence,
				}
				client.WriteEvent("response.output_text.done", completed)
				sequence++

				itemDone := map[string]interface{}{
					"type":         "response.output_item.done",
					"output_index": 0,
					"item": map[string]interface{}{
						"type":   "message",
						"id":     itemID,
						"status": "completed",
						"role":   "assistant",
						"content": []map[string]interface{}{
							{"type": "output_text", "text": ""},
						},
					},
					"sequence": sequence,
				}
				client.WriteEvent("response.output_item.done", itemDone)
				sequence++

				respDone := map[string]interface{}{
					"type": "response.done",
					"response": map[string]interface{}{
						"id":     responseID,
						"object": "response",
						"model":  model,
						"status": "completed",
					},
					"sequence": sequence,
				}
				client.WriteEvent("response.done", respDone)
				return nil
			}

			// Parse the chunk
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			// Extract delta text from choices
			choices, ok := chunk["choices"].([]interface{})
			if !ok || len(choices) == 0 {
				continue
			}

			choice, ok := choices[0].(map[string]interface{})
			if !ok {
				continue
			}

			delta, _ := choice["delta"].(map[string]interface{})
			if delta == nil {
				continue
			}

			content, _ := delta["content"].(string)
			if content == "" {
				continue
			}

			// Send text delta event
			deltaEvent := map[string]interface{}{
				"type":           "response.output_text.delta",
				"output_index":   0,
				"content_index":  0,
				"delta":          content,
				"sequence":       sequence,
			}
			client.WriteEvent("response.output_text.delta", deltaEvent)
			sequence++
		}
	}

	return scanner.Err()
}

// BuildResponsesAPINonStream constructs a Responses API response from a Chat Completion.
func BuildResponsesAPINonStream(chatResponse map[string]interface{}, responseID string) map[string]interface{} {
	model, _ := chatResponse["model"].(string)

	var outputText string
	if choices, ok := chatResponse["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := msg["content"].(string); ok {
					outputText = content
				}
			}
		}
	}

	output := []map[string]interface{}{
		{
			"type":   "message",
			"id":     fmt.Sprintf("item_%s", responseID),
			"status": "completed",
			"role":   "assistant",
			"content": []map[string]interface{}{
				{"type": "output_text", "text": outputText},
			},
		},
	}

	result := map[string]interface{}{
		"id":         responseID,
		"object":     "response",
		"model":      model,
		"created_at": time.Now().Unix(),
		"status":     "completed",
		"output":     output,
	}

	// Copy usage if present
	if usage, ok := chatResponse["usage"]; ok {
		result["usage"] = usage
	}

	return result
}
