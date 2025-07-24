package auth_test

import (
	"database/sql"
	"forum/internal/auth"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestGoogleLoginHandler(t *testing.T) {
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
	defer os.Unsetenv("GOOGLE_CLIENT_ID")
	defer os.Unsetenv("GOOGLE_CLIENT_SECRET")

	// Initialize OAuth configuration
	auth.InitOAuthConfig("http://localhost:3000")

	// Create a request to the Google login handler
	req, err := http.NewRequest("GET", "/auth/google/login", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	oauthHandler.GoogleLoginHandler(rr, req)

	// Check the response
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Check that we got a cookie
	cookies := rr.Result().Cookies()
	var stateCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "oauth_state" {
			stateCookie = cookie
			break
		}
	}
	if stateCookie == nil {
		t.Errorf("No state cookie found")
	}

	// Check the redirect URL
	location := rr.Header().Get("Location")
	if location == "" {
		t.Errorf("No redirect location found")
	}

	// Parse the URL to check the parameters
	redirectURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("Failed to parse redirect URL: %v", err)
	}

	// Check that the URL is for Google
	if !containsSubstring(redirectURL.Host, "google") {
		t.Errorf("Expected redirect to Google, got %s", redirectURL.Host)
	}

	// Check that the client ID is included
	if redirectURL.Query().Get("client_id") != "test-client-id" {
		t.Errorf("Expected client_id=test-client-id, got %s", redirectURL.Query().Get("client_id"))
	}

	// Check that the state parameter matches the cookie
	if redirectURL.Query().Get("state") != stateCookie.Value {
		t.Errorf("State parameter doesn't match cookie: %s != %s", redirectURL.Query().Get("state"), stateCookie.Value)
	}
}

func TestGitHubLoginHandler(t *testing.T) {
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
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	defer os.Unsetenv("GITHUB_CLIENT_ID")
	defer os.Unsetenv("GITHUB_CLIENT_SECRET")

	// Initialize OAuth configuration
	auth.InitOAuthConfig("http://localhost:3000")

	// Create a request to the GitHub login handler
	req, err := http.NewRequest("GET", "/auth/github/login", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	oauthHandler.GitHubLoginHandler(rr, req)

	// Check the response
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Check that we got a cookie
	cookies := rr.Result().Cookies()
	var stateCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "oauth_state" {
			stateCookie = cookie
			break
		}
	}
	if stateCookie == nil {
		t.Errorf("No state cookie found")
	}

	// Check the redirect URL
	location := rr.Header().Get("Location")
	if location == "" {
		t.Errorf("No redirect location found")
	}

	// Parse the URL to check the parameters
	redirectURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("Failed to parse redirect URL: %v", err)
	}

	// Check that the URL is for GitHub
	if !containsSubstring(redirectURL.Host, "github") {
		t.Errorf("Expected redirect to GitHub, got %s", redirectURL.Host)
	}

	// Check that the client ID is included
	if redirectURL.Query().Get("client_id") != "test-client-id" {
		t.Errorf("Expected client_id=test-client-id, got %s", redirectURL.Query().Get("client_id"))
	}

	// Check that the state parameter matches the cookie
	if redirectURL.Query().Get("state") != stateCookie.Value {
		t.Errorf("State parameter doesn't match cookie: %s != %s", redirectURL.Query().Get("state"), stateCookie.Value)
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}