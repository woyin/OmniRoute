package management

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// GuardrailsHandler provides real DB-backed guardrails management endpoints.
type GuardrailsHandler struct {
	DB *sql.DB
}

// builtinGuardrails returns the 3 built-in guardrails that always exist.
func builtinGuardrails() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "builtin-pii-masker",
			"name":        "pii-masker",
			"type":        "pii",
			"enabled":     false,
			"description": "PII masking guardrail (opt-in via PII_REDACTION_ENABLED)",
			"builtin":     true,
		},
		{
			"id":          "builtin-prompt-injection",
			"name":        "prompt-injection",
			"type":        "security",
			"enabled":     true,
			"description": "Prompt injection detection",
			"builtin":     true,
		},
		{
			"id":          "builtin-vision-bridge",
			"name":        "vision-bridge",
			"type":        "content",
			"enabled":     false,
			"description": "Vision content bridge",
			"builtin":     true,
		},
	}
}

// List returns all guardrails, merging built-in defaults with DB-configured ones.
func (h *GuardrailsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Start with built-in guardrails
	guardrails := builtinGuardrails()

	// Merge with DB-configured guardrails
	if h.DB != nil {
		rows, err := h.DB.Query(`
			SELECT id, name, guardrail_type, is_enabled, config, created_at
			FROM guardrails_config
			ORDER BY created_at DESC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, name, gtype, config, createdAt string
				var enabled int
				if err := rows.Scan(&id, &name, &gtype, &enabled, &config, &createdAt); err == nil {
					// Check if this overrides a builtin
					isBuiltin := false
					for i, bg := range guardrails {
						if bg["name"] == name {
							guardrails[i]["enabled"] = enabled == 1
							guardrails[i]["id"] = id
							guardrails[i]["config"] = config
							isBuiltin = true
							break
						}
					}
					if !isBuiltin {
						guardrails = append(guardrails, map[string]interface{}{
							"id":        id,
							"name":      name,
							"type":      gtype,
							"enabled":   enabled == 1,
							"config":    config,
							"builtin":   false,
							"createdAt": createdAt,
						})
					}
				}
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"guardrails": guardrails,
		"total":      len(guardrails),
	})
}

// Create creates a new custom guardrail configuration.
func (h *GuardrailsHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Config string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.Config == "" {
		body.Config = "{}"
	}

	if h.DB != nil {
		id := generateID()
		_, err := h.DB.Exec(`
			INSERT INTO guardrails_config (id, name, guardrail_type, is_enabled, config)
			VALUES (?, ?, ?, 0, ?)
		`, id, body.Name, body.Type, body.Config)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create guardrail")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      id,
			"name":    body.Name,
			"success": true,
		})
		return
	}

	writeJSONError(w, http.StatusServiceUnavailable, "database not available")
}

// Test runs a guardrail test against the provided input.
func (h *GuardrailsHandler) Test(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Input      string `json:"input"`
		Guardrails []string `json:"guardrails"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	// Run basic built-in checks
	var results []map[string]interface{}
	blocked := false

	if body.Input != "" {
		// Prompt injection detection (simple heuristic)
		injectionPatterns := []string{"ignore previous", "disregard all", "forget everything", "system prompt"}
		for _, pattern := range injectionPatterns {
			if containsIgnoreCase(body.Input, pattern) {
				blocked = true
				results = append(results, map[string]interface{}{
					"guardrail": "prompt-injection",
					"blocked":   true,
					"reason":    "Suspicious prompt injection pattern detected: " + pattern,
				})
				break
			}
		}
	}

	if !blocked {
		results = append(results, map[string]interface{}{
			"guardrail": "prompt-injection",
			"blocked":   false,
			"reason":    "No injection patterns detected",
		})
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"blocked":   blocked,
		"results":   results,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return indexOf(sLower, substrLower) >= 0
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
