package auth

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "my-secure-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	// Correct password
	if !VerifyPassword(password, hash) {
		t.Error("expected password to verify")
	}

	// Wrong password
	if VerifyPassword("wrong-password", hash) {
		t.Error("expected wrong password to fail verification")
	}

	// Empty password
	if VerifyPassword("", hash) {
		t.Error("expected empty password to fail verification")
	}

	// Invalid hash format
	if VerifyPassword(password, "invalid-hash") {
		t.Error("expected invalid hash to fail verification")
	}
}

func TestConstantTimeEqual(t *testing.T) {
	if !constantTimeEqual("abc", "abc") {
		t.Error("expected equal strings to match")
	}
	if constantTimeEqual("abc", "abd") {
		t.Error("expected different strings to not match")
	}
	if constantTimeEqual("abc", "ab") {
		t.Error("expected different length strings to not match")
	}
	if !constantTimeEqual("", "") {
		t.Error("expected empty strings to be equal")
	}
}

func TestGenerateRandomSecret(t *testing.T) {
	secret1 := generateRandomSecret(32)
	secret2 := generateRandomSecret(32)
	if secret1 == secret2 {
		t.Error("expected different random secrets")
	}
	if len(secret1) < 32 {
		t.Errorf("expected secret length >= 32, got %d", len(secret1))
	}
}

func TestGenerateSessionToken(t *testing.T) {
	t.Setenv(JWTSecretEnv, "test-secret-at-least-32-bytes-long")
	token, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken failed: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if len(token) < 20 {
		t.Errorf("expected token length >= 20, got %d", len(token))
	}
	if !validateSessionToken(token) {
		t.Error("expected generated JWT to validate")
	}
	if validateSessionToken(token + "tampered") {
		t.Error("expected tampered JWT to fail validation")
	}
}
