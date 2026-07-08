package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	JWTSecretEnv = "JWT_SECRET"
)

var jwtSecret string

// GetJWTSecret returns the JWT signing secret. Auto-generates one if not set.
func GetJWTSecret() string {
	if jwtSecret != "" {
		return jwtSecret
	}
	jwtSecret = strings.TrimSpace(os.Getenv(JWTSecretEnv))
	if jwtSecret == "" {
		secret := generateRandomSecret(32)
		jwtSecret = secret
		log.Println("[AUTH] JWT_SECRET not set — auto-generated a random secret for this session.")
		log.Println("[AUTH] Set JWT_SECRET env var for persistent sessions across restarts.")
	}
	return jwtSecret
}

func generateRandomSecret(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return uuid.New().String()
	}
	return base64.URLEncoding.EncodeToString(b)
}

// HashPassword hashes a password using SHA-256 with a random salt.
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	hash := sha256.Sum256([]byte(saltB64 + password))
	return saltB64 + ":" + base64.StdEncoding.EncodeToString(hash[:]), nil
}

// VerifyPassword verifies a password against a stored hash (salt:hash format).
func VerifyPassword(password, storedHash string) bool {
	parts := strings.SplitN(storedHash, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt := parts[0]
	hash := sha256.Sum256([]byte(salt + password))
	expected := base64.StdEncoding.EncodeToString(hash[:])
	return constantTimeEqual(parts[1], expected)
}

// HasManagementPassword checks if a management password is configured.
func HasManagementPassword(dbConn *sql.DB) bool {
	if dbConn == nil {
		return false
	}
	var count int
	err := dbConn.QueryRow(
		"SELECT COUNT(*) FROM key_value WHERE namespace = 'settings' AND key = 'password'",
	).Scan(&count)
	return err == nil && count > 0
}

// IsSetupComplete checks if initial setup has been completed.
func IsSetupComplete(dbConn *sql.DB) bool {
	if dbConn == nil {
		return false
	}
	var value string
	err := dbConn.QueryRow(
		"SELECT value FROM key_value WHERE namespace = 'settings' AND key = 'setupComplete'",
	).Scan(&value)
	if err != nil {
		return false
	}
	return value == "true" || value == "1"
}

// IsBootstrapWindow returns true if no password has been set yet (first-run state).
func IsBootstrapWindow(dbConn *sql.DB) bool {
	return !HasManagementPassword(dbConn)
}

// SetManagementPassword saves a hashed password to the DB.
func SetManagementPassword(dbConn *sql.DB, password string) error {
	hashed, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = dbConn.Exec(
		"INSERT INTO key_value (namespace, key, value) VALUES ('settings', 'password', ?) "+
			"ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value",
		hashed,
	)
	return err
}

// MarkSetupComplete marks the initial setup as done.
func MarkSetupComplete(dbConn *sql.DB) error {
	_, err := dbConn.Exec(
		"INSERT INTO key_value (namespace, key, value) VALUES ('settings', 'setupComplete', 'true') "+
			"ON CONFLICT(namespace, key) DO UPDATE SET value = excluded.value",
	)
	return err
}

// VerifyManagementPassword checks a password against the stored hash.
func VerifyManagementPassword(dbConn *sql.DB, password string) bool {
	var storedHash string
	err := dbConn.QueryRow(
		"SELECT value FROM key_value WHERE namespace = 'settings' AND key = 'password'",
	).Scan(&storedHash)
	if err != nil {
		return false
	}
	return VerifyPassword(password, storedHash)
}

// GenerateSessionToken creates a session token.
func GenerateSessionToken() (string, error) {
	payload := map[string]interface{}{
		"sub": "admin",
		"iat": time.Now().Unix(),
		"jti": uuid.New().String(),
	}
	payloadJSON, _ := json.Marshal(payload)
	token := base64.URLEncoding.EncodeToString(payloadJSON)
	return token, nil
}

// constantTimeEqual does a constant-time string comparison.
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// --- HTTP Handlers ---

// RequireLoginHandler handles GET/POST /api/settings/require-login
type RequireLoginHandler struct {
	DB *sql.DB
}

func (h *RequireLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *RequireLoginHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	hasPassword := HasManagementPassword(h.DB)
	setupComplete := IsSetupComplete(h.DB)
	requireLogin := true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requireLogin":  requireLogin,
		"hasPassword":   hasPassword,
		"setupComplete": setupComplete,
	})
}

func (h *RequireLoginHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	// Only allow setting password during bootstrap window (first run, no password set yet)
	// or if already authenticated (TODO: check session cookie)
	if !IsBootstrapWindow(h.DB) {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var body struct {
		Password     string `json:"password"`
		RequireLogin *bool  `json:"requireLogin,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON"}}`, http.StatusBadRequest)
		return
	}

	if body.Password == "" {
		http.Error(w, `{"error":{"message":"password is required"}}`, http.StatusBadRequest)
		return
	}

	if len(body.Password) < 6 {
		http.Error(w, `{"error":{"message":"password must be at least 6 characters"}}`, http.StatusBadRequest)
		return
	}

	if err := SetManagementPassword(h.DB, body.Password); err != nil {
		http.Error(w, `{"error":{"message":"Failed to set password"}}`, http.StatusInternalServerError)
		return
	}

	MarkSetupComplete(h.DB)
	log.Println("[AUTH] Management password set — bootstrap window closed.")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// LoginHandler handles POST /api/auth/login
type LoginHandler struct {
	DB *sql.DB
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	secret := GetJWTSecret()
	if secret == "" {
		http.Error(w, `{"error":"Server misconfigured: JWT_SECRET not set"}`, http.StatusInternalServerError)
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":{"message":"Invalid JSON"}}`, http.StatusBadRequest)
		return
	}

	if body.Password == "" {
		http.Error(w, `{"error":{"message":"password is required"}}`, http.StatusBadRequest)
		return
	}

	if !VerifyManagementPassword(h.DB, body.Password) {
		log.Println("[AUTH] Login failed: invalid password")
		http.Error(w, `{"error":{"message":"Invalid password"}}`, http.StatusUnauthorized)
		return
	}

	token, err := GenerateSessionToken()
	if err != nil {
		http.Error(w, `{"error":{"message":"Failed to generate session"}}`, http.StatusInternalServerError)
		return
	}

	log.Println("[AUTH] Login successful")

	http.SetCookie(w, &http.Cookie{
		Name:     "omniroute_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
	})
}

// LogoutHandler handles POST /api/auth/logout
type LogoutHandler struct{}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "omniroute_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
