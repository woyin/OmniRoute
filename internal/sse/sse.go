package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Event represents a Server-Sent Event.
type Event struct {
	ID    string
	Event string
	Data  string
}

// Writer writes SSE events to an HTTP response writer.
type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
	closed  bool
}

// NewWriter creates a new SSE writer wrapping an HTTP response writer.
func NewWriter(w http.ResponseWriter) *Writer {
	flusher, ok := w.(http.Flusher)
	if !ok {
		flusher = http.Flusher(nil)
	}
	return &Writer{w: w, flusher: flusher}
}

// WriteHeader writes the SSE response headers.
func (s *Writer) WriteHeader() {
	s.w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
	s.w.Header().Set("X-Accel-Buffering", "no")
	s.w.WriteHeader(200)
	if s.flusher != nil {
		s.flusher.Flush()
	}
}

// WriteEvent writes a single SSE event.
func (s *Writer) WriteEvent(event string, data interface{}) error {
	if s.closed {
		return fmt.Errorf("SSE writer is closed")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal SSE data: %w", err)
	}

	if event != "" {
		fmt.Fprintf(s.w, "event: %s\n", event)
	}
	fmt.Fprintf(s.w, "data: %s\n\n", jsonData)

	if s.flusher != nil {
		s.flusher.Flush()
	}
	return nil
}

// WriteData writes a data-only SSE event (no event type).
func (s *Writer) WriteData(data interface{}) error {
	return s.WriteEvent("", data)
}

// WriteDone writes the [DONE] sentinel.
func (s *Writer) WriteDone() error {
	if s.closed {
		return nil
	}
	fmt.Fprintf(s.w, "data: [DONE]\n\n")
	if s.flusher != nil {
		s.flusher.Flush()
	}
	s.closed = true
	return nil
}

// WriteComment writes an SSE comment (keepalive).
func (s *Writer) WriteComment(comment string) {
	if s.closed {
		return
	}
	fmt.Fprintf(s.w, ": %s\n\n", comment)
	if s.flusher != nil {
		s.flusher.Flush()
	}
}

// Close marks the writer as closed.
func (s *Writer) Close() {
	s.closed = true
}

// IsClosed returns whether the writer has been closed.
func (s *Writer) IsClosed() bool {
	return s.closed
}

// StreamFromUpstream reads an upstream SSE response and forwards it to the client.
func StreamFromUpstream(client *Writer, upstream *http.Response, stopCh <-chan struct{}) error {
	defer upstream.Body.Close()

	buf := make([]byte, 4096)
	for {
		select {
		case <-stopCh:
			return nil
		default:
		}

		n, err := upstream.Body.Read(buf)
		if n > 0 {
			if _, err := client.w.Write(buf[:n]); err != nil {
				return err
			}
			if client.flusher != nil {
				client.flusher.Flush()
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read upstream: %w", err)
		}
	}
}

// StartHeartbeat starts a goroutine that sends SSE comments at the given interval.
func StartHeartbeat(writer *Writer, interval time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writer.WriteComment("heartbeat")
			case <-stopCh:
				return
			}
		}
	}()
}

// ChatCompletionChunk represents an OpenAI-format streaming chunk.
type ChatCompletionChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                    `json:"index"`
		Delta        map[string]interface{} `json:"delta"`
		FinishReason interface{}            `json:"finish_reason"`
	} `json:"choices"`
}

// NewChatCompletionChunk creates a streaming chunk.
func NewChatCompletionChunk(id, model string, delta map[string]interface{}, finishReason interface{}) ChatCompletionChunk {
	return ChatCompletionChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []struct {
			Index        int                    `json:"index"`
			Delta        map[string]interface{} `json:"delta"`
			FinishReason interface{}            `json:"finish_reason"`
		}{
			{Index: 0, Delta: delta, FinishReason: finishReason},
		},
	}
}
