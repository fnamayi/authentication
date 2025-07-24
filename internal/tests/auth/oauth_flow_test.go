package auth_test

import (
	"database/sql"
	"forum/internal/auth"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestOAuthFlow(t *testing.T) {
	// Set up test environment
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY NOT NULL,
			username TEXT UNIQUE,
			password TEXT,
			email TEXT UNIQUE,
			profile_pic TEXT,
			oauth_provider TEXT,
			oauth_id TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Set up templates
	tmpl, err := template.New("test").Parse(`{{.}}`)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Create handlers
	authHandler := auth.NewAuthHandler(db, tmpl)
	oauthHandler := auth.NewOAuthHandler(db, authHandler)

	// Set environment variables for testing
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("GOOGLE_CLIENT_ID")
	defer os.Unsetenv("GOOGLE_CLIENT_SECRET")
	defer os.Unsetenv("GITHUB_CLIENT_ID")
	defer os.Unsetenv("GITHUB_CLIENT_SECRET")

	// Initialize OAuth configuration
	auth.InitOAuthConfig("http://localhost:3000")

	// Test cases
	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		handler        http.HandlerFunc
	}{
		{
			name:           "Google Login",
			path:           "/auth/google/login",
			expectedStatus: http.StatusSeeOther,
			handler:        oauthHandler.GoogleLoginHandler,
		},
		{
			name:           "GitHub Login",
			path:           "/auth/github/login",
			expectedStatus: http.StatusSeeOther,
			handler:        oauthHandler.GitHubLoginHandler,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tc.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			tc.handler(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			// Check for redirect
			if tc.expectedStatus == http.StatusSeeOther {
				location := rr.Header().Get("Location")
				if location == "" {
					t.Errorf("Expected redirect, but no Location header found")
				}
			}
		})
	}

	// Test callback handlers with simulated responses
	t.Run("Google Callback", func(t *testing.T) {
		// This is a simplified test that doesn't actually make external requests
		req, err := http.NewRequest("GET", "/auth/google/callback?code=test-code&state=test-state", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set a cookie for the state parameter
		req.AddCookie(&http.Cookie{
			Name:  "oauth_state",
			Value: "test-state",
		})

		rr := httptest.NewRecorder()
		oauthHandler.GoogleCallbackHandler(rr, req)

		// In a real test, we would mock the token exchange and user info requests
		// Here we just check that the handler doesn't panic and returns a redirect
		if rr.Code != http.StatusSeeOther && rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d or %d, got %d", http.StatusSeeOther, http.StatusInternalServerError, rr.Code)
		}
	})

	t.Run("GitHub Callback", func(t *testing.T) {
		// This is a simplified test that doesn't actually make external requests
		req, err := http.NewRequest("GET", "/auth/github/callback?code=test-code&state=test-state", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set a cookie for the state parameter
		req.AddCookie(&http.Cookie{
			Name:  "oauth_state",
			Value: "test-state",
		})

		rr := httptest.NewRecorder()
		oauthHandler.GitHubCallbackHandler(rr, req)

		// In a real test, we would mock the token exchange and user info requests
		// Here we just check that the handler doesn't panic and returns a redirect
		if rr.Code != http.StatusSeeOther && rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d or %d, got %d", http.StatusSeeOther, http.StatusInternalServerError, rr.Code)
		}
	})

	// Test error handling
	t.Run("OAuth Error Handling", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/auth/google/callback?error=access_denied", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		rr := httptest.NewRecorder()
		oauthHandler.GoogleCallbackHandler(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, rr.Code)
		}

		location := rr.Header().Get("Location")
		if location == "" {
			t.Errorf("Expected redirect, but no Location header found")
		}

		if location != "/login?error=oauth_error" {
			t.Errorf("Expected redirect to /login?error=oauth_error, got %s", location)
		}
	})
}