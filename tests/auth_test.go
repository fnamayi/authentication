package main

import (
	"strings"
	"testing"

	"forum/internal/auth"
)

// Test password hashing functions
func TestPasswordHashing(t *testing.T) {
	t.Run("Hash Password", func(t *testing.T) {
		password := "testpassword123"
		hash, err := auth.HashPassword(password)

		if err != nil {
			t.Errorf("HashPassword should not return error, got: %v", err)
		}

		if hash == "" {
			t.Error("Hash should not be empty")
		}

		if hash == password {
			t.Error("Hash should not equal original password")
		}

		if len(hash) < 50 {
			t.Error("Hash should be at least 50 characters long")
		}
	})

	t.Run("Check Password Hash - Valid", func(t *testing.T) {
		password := "testpassword123"
		hash, err := auth.HashPassword(password)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		isValid := auth.CheckPasswordHash(password, hash)
		if !isValid {
			t.Error("CheckPasswordHash should return true for correct password")
		}
	})

	t.Run("Check Password Hash - Invalid", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hash, err := auth.HashPassword(password)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		isValid := auth.CheckPasswordHash(wrongPassword, hash)
		if isValid {
			t.Error("CheckPasswordHash should return false for incorrect password")
		}
	})

	t.Run("Empty Password", func(t *testing.T) {
		hash, err := auth.HashPassword("")
		if err != nil {
			t.Errorf("HashPassword should handle empty password, got error: %v", err)
		}
		if hash == "" {
			t.Error("Hash should not be empty even for empty password")
		}
	})
}

// Test session ID generation
func TestSessionIDGeneration(t *testing.T) {
	t.Run("Generate Session ID", func(t *testing.T) {
		sessionID := auth.GenerateSessionID()

		if sessionID == "" {
			t.Error("Session ID should not be empty")
		}

		if len(sessionID) != 64 { // 32 bytes = 64 hex characters
			t.Errorf("Session ID should be 64 characters long, got %d", len(sessionID))
		}

		// Check if it's valid hex
		for _, char := range sessionID {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("Session ID should only contain hex characters, found: %c", char)
			}
		}
	})

	t.Run("Generate Unique Session IDs", func(t *testing.T) {
		id1 := auth.GenerateSessionID()
		id2 := auth.GenerateSessionID()

		if id1 == id2 {
			t.Error("Generated session IDs should be unique")
		}
	})

	t.Run("Multiple Session IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := auth.GenerateSessionID()
			if ids[id] {
				t.Error("Session IDs should be unique")
			}
			ids[id] = true
		}
	})
}

// Test string validation functions
func TestStringValidation(t *testing.T) {
	t.Run("Email Validation", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"user+tag@example.org",
		}

		invalidEmails := []string{
			"invalid-email",
			"@example.com",
			"test@",
			"",
		}

		for _, email := range validEmails {
			if !isValidEmailFormat(email) {
				t.Errorf("Email '%s' should be valid", email)
			}
		}

		for _, email := range invalidEmails {
			if isValidEmailFormat(email) {
				t.Errorf("Email '%s' should be invalid", email)
			}
		}
	})

	t.Run("String Trimming", func(t *testing.T) {
		input := "  test string  "
		trimmed := strings.TrimSpace(input)
		if trimmed != "test string" {
			t.Errorf("Expected 'test string', got '%s'", trimmed)
		}
	})
}

// Helper function for email validation (proper version)
func isValidEmailFormat(email string) bool {
	if email == "" {
		return false
	}

	// Must contain exactly one @
	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}

	// Split by @
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	username := parts[0]
	domain := parts[1]

	// Username must not be empty
	if len(username) == 0 {
		return false
	}

	// Domain must contain at least one dot and not be empty
	if len(domain) == 0 || !strings.Contains(domain, ".") {
		return false
	}

	// Domain must not start or end with dot
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	// Basic length check
	return len(email) >= 5
}
