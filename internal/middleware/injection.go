package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// injectionPatterns are common prompt injection markers.
var injectionPatterns = []string{
	"ignore previous instructions",
	"ignore all previous",
	"disregard all previous",
	"forget your instructions",
	"new instructions:",
	"system override",
	"jailbreak",
	"DAN mode",
	"you are now",
	"act as if you are",
	"pretend you are",
	"simulate being",
	"bypass",
	"override safety",
	"ignore the above",
}

// InjectionCheckResult holds the result of a prompt injection check.
type InjectionCheckResult struct {
	Blocked  bool   `json:"blocked"`
	Reason   string `json:"reason,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
}

// CheckPromptInjection scans a request body for common prompt injection patterns.
// Returns a result indicating whether the request should be blocked.
func CheckPromptInjection(body map[string]interface{}) *InjectionCheckResult {
	messages, ok := body["messages"].([]interface{})
	if !ok {
		return nil
	}

	for _, msg := range messages {
		m, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := m["content"].(string)
		if content == "" {
			continue
		}
		lower := strings.ToLower(content)
		for _, pattern := range injectionPatterns {
			if strings.Contains(lower, pattern) {
				return &InjectionCheckResult{
					Blocked: true,
					Reason:  "Potential prompt injection detected",
					Pattern: pattern,
				}
			}
		}
	}

	// Also check system/developer messages for unusual content
	if sysPrompt, ok := body["system"].(string); ok {
		lower := strings.ToLower(sysPrompt)
		for _, pattern := range injectionPatterns {
			if strings.Contains(lower, pattern) {
				return &InjectionCheckResult{
					Blocked: true,
					Reason:  "Potential prompt injection in system prompt",
					Pattern: pattern,
				}
			}
		}
	}

	return &InjectionCheckResult{Blocked: false}
}

// PromptInjectionGuard is middleware that checks for prompt injection.
// If injection is detected, it returns 400 with an error message.
func PromptInjectionGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check POST requests with JSON bodies
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Read the body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			// Can't parse — let the next handler deal with it
			next.ServeHTTP(w, r)
			return
		}

		// Check for injection
		result := CheckPromptInjection(body)
		if result != nil && result.Blocked {
			log.Printf("[GUARD] Prompt injection blocked: %s (pattern: %s)", result.Reason, result.Pattern)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Request blocked: potential prompt injection detected",
					"type":    "invalid_request_error",
					"code":    "prompt_injection_detected",
				},
			})
			return
		}

		// Re-encode the body and replace the request body
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		next.ServeHTTP(w, r)
	})
}
