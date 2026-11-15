package util

import (
	"testing"
	"time"
)

func TestTokenCreationAndVerification(t *testing.T) {
	validKey := "12345678901234567890123456789012" // 32 characters
	userID := int64(42)
	duration := 1 * time.Hour

	// Valid creation and verification
	token, err := CreateToken(userID, duration, validKey)
	if err != nil {
		t.Fatalf("expected no error creating token, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	extractedUserID, err := VerifyToken(token, validKey)
	if err != nil {
		t.Fatalf("expected no error verifying token, got: %v", err)
	}
	if extractedUserID != userID {
		t.Fatalf("expected user ID %d, got %d", userID, extractedUserID)
	}

	// Short key (< 32 chars) must be rejected
	shortKey := "too_short_secret_key"
	_, err = CreateToken(userID, duration, shortKey)
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey for short key, got: %v", err)
	}

	_, err = VerifyToken(token, shortKey)
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey for short key during verification, got: %v", err)
	}

	// Insecure default key must be rejected
	_, err = CreateToken(userID, duration, InsecureDefaultKey)
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey for insecure default key, got: %v", err)
	}
}
