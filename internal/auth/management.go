package auth

import (
	"crypto/hmac"
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

// GetJWTSecret returns configured JWT signing secret. Empty means authentication is disabled.
func GetJWTSecret() string {
	return strings.TrimSpace(os.Getenv(JWTSecretEnv))
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
		"INSERT INTO key_value (namespace, key, value) VALUES ('settings', 'setupComplete', 'true') " +
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

// GenerateSessionToken creates main-branch-compatible HS256 JWT valid for 30 days.
func GenerateSessionToken() (string, error) {
	secret := GetJWTSecret()
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET not set")
	}
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, _ := json.Marshal(map[string]interface{}{
		"authenticated": true,
		"exp":           time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func validateSessionToken(token string) bool {
	secret := GetJWTSecret()
	parts := strings.Split(token, ".")
	if secret == "" || len(parts) != 3 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return false
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var payload struct {
		Authenticated bool  `json:"authenticated"`
		ExpiresAt     int64 `json:"exp"`
	}
	return json.Unmarshal(payloadJSON, &payload) == nil && payload.Authenticated && payload.ExpiresAt >= time.Now().Unix()
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
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Unauthorized — password already set","type":"authentication_error"}}`))
		return
	}

	var body struct {
		Password     string `json:"password"`
		RequireLogin *bool  `json:"requireLogin,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"Invalid JSON"}}`))
		return
	}

	if body.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"password is required"}}`))
		return
	}

	if len(body.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"password must be at least 6 characters"}}`))
		return
	}

	if err := SetManagementPassword(h.DB, body.Password); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Failed to set password"}}`))
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Server misconfigured: JWT_SECRET not set"}}`))
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"Invalid JSON"}}`))
		return
	}

	if body.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"password is required"}}`))
		return
	}

	if !VerifyManagementPassword(h.DB, body.Password) {
		log.Println("[AUTH] Login failed: invalid password")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"Invalid password"}}`))
		return
	}

	token, err := GenerateSessionToken()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":{"message":"Failed to generate session"}}`))
		return
	}

	log.Println("[AUTH] Login successful")

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
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
		Name:     "auth_token",
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

func secureCookie(r *http.Request) bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("AUTH_COOKIE_SECURE")), "true") {
		return true
	}
	proto := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]))
	return proto == "https" || r.TLS != nil
}
