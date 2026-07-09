package management

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

// writeJSONError writes a JSON-formatted error response.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"message": message},
	})
}

// generateID creates a random hex ID for new records.
func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "fallback-id"
	}
	return hex.EncodeToString(b)
}
