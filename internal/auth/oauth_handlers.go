package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// OAuthHandler handles OAuth authentication requests
type OAuthHandler struct {
	db        *sql.DB
	authHandler *AuthHandler
}

// NewOAuthHandler creates a new OAuthHandler
func NewOAuthHandler(db *sql.DB, authHandler *AuthHandler) *OAuthHandler {
	return &OAuthHandler{
		db:        db,
		authHandler: authHandler,
	}
}

// GoogleLoginHandler initiates the Google OAuth flow
func (h *OAuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GoogleLoginHandler called with method: %s", r.Method)

	// Generate a state parameter for CSRF protection
	state, err := GenerateState()
	if err != nil {
		log.Printf("Error generating state: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Store the state in a cookie
	stateCookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
	}
	http.SetCookie(w, stateCookie)

	// Get the authorization URL
	authURL := Config.Google.GetAuthURL(state)
	log.Printf("Redirecting to Google auth URL: %s", authURL)

	// Redirect to the authorization URL
	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

// GoogleCallbackHandler handles the callback from Google OAuth
func (h *OAuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GoogleCallbackHandler called with method: %s", r.Method)

	// Get the state and code from the request
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	errorParam := r.URL.Query().Get("error")

	// Check if there was an error
	if errorParam != "" {
		log.Printf("Error from Google: %s", errorParam)
		http.Redirect(w, r, "/login?error=oauth_error", http.StatusSeeOther)
		return
	}

	// Verify the state parameter
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		log.Printf("Invalid state parameter: %v", err)
		http.Redirect(w, r, "/login?error=invalid_state", http.StatusSeeOther)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Exchange the code for an access token
	accessToken, err := Config.Google.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Redirect(w, r, "/login?error=token_exchange", http.StatusSeeOther)
		return
	}

	// Get the user information
	userInfo, err := Config.Google.GetUserInfo(accessToken)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Redirect(w, r, "/login?error=user_info", http.StatusSeeOther)
		return
	}

	// Process the user information
	err = h.processOAuthUser(w, "google", userInfo)
	if err != nil {
		log.Printf("Error processing OAuth user: %v", err)
		http.Redirect(w, r, "/login?error=user_processing", http.StatusSeeOther)
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GitHubLoginHandler initiates the GitHub OAuth flow
func (h *OAuthHandler) GitHubLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GitHubLoginHandler called with method: %s", r.Method)

	// Generate a state parameter for CSRF protection
	state, err := GenerateState()
	if err != nil {
		log.Printf("Error generating state: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Store the state in a cookie
	stateCookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
	}
	http.SetCookie(w, stateCookie)

	// Get the authorization URL
	authURL := Config.GitHub.GetAuthURL(state)
	log.Printf("Redirecting to GitHub auth URL: %s", authURL)

	// Redirect to the authorization URL
	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

// GitHubCallbackHandler handles the callback from GitHub OAuth
func (h *OAuthHandler) GitHubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GitHubCallbackHandler called with method: %s", r.Method)

	// Get the state and code from the request
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	errorParam := r.URL.Query().Get("error")

	// Check if there was an error
	if errorParam != "" {
		log.Printf("Error from GitHub: %s", errorParam)
		http.Redirect(w, r, "/login?error=oauth_error", http.StatusSeeOther)
		return
	}

	// Verify the state parameter
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		log.Printf("Invalid state parameter: %v", err)
		http.Redirect(w, r, "/login?error=invalid_state", http.StatusSeeOther)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Exchange the code for an access token
	accessToken, err := Config.GitHub.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Redirect(w, r, "/login?error=token_exchange", http.StatusSeeOther)
		return
	}

	// Get the user information
	userInfo, err := Config.GitHub.GetUserInfo(accessToken)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Redirect(w, r, "/login?error=user_info", http.StatusSeeOther)
		return
	}

	// Process the user information
	err = h.processOAuthUser(w, "github", userInfo)
	if err != nil {
		log.Printf("Error processing OAuth user: %v", err)
		http.Redirect(w, r, "/login?error=user_processing", http.StatusSeeOther)
		return
	}

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// processOAuthUser processes the OAuth user information
func (h *OAuthHandler) processOAuthUser(w http.ResponseWriter, provider string, userInfo map[string]interface{}) error {
	// Get the provider configuration
	var providerConfig OAuthProvider
	if provider == "google" {
		providerConfig = Config.Google
	} else if provider == "github" {
		providerConfig = Config.GitHub
	} else {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	// Extract user information
	oauthID, ok := userInfo[providerConfig.UserIDKey].(string)
	if !ok {
		if id, ok := userInfo[providerConfig.UserIDKey].(float64); ok {
			oauthID = fmt.Sprintf("%.0f", id)
		} else {
			return fmt.Errorf("invalid user ID: %v", userInfo[providerConfig.UserIDKey])
		}
	}

	username, _ := userInfo[providerConfig.UserNameKey].(string)
	email, _ := userInfo[providerConfig.UserEmailKey].(string)
	profilePic, _ := userInfo[providerConfig.UserProfilePicKey].(string)

	log.Printf("Processing OAuth user: provider=%s, id=%s, username=%s, email=%s", provider, oauthID, username, email)

	// Check if the user already exists
	var userID string
	err := h.db.QueryRow(
		"SELECT id FROM users WHERE oauth_provider = ? AND oauth_id = ?",
		provider, oauthID,
	).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			// User doesn't exist, create a new one
			return h.createOAuthUser(w, provider, oauthID, username, email, profilePic)
		}
		return err
	}

	// User exists, create a session
	CreateSession(w, userID)
	return nil
}

// createOAuthUser creates a new user from OAuth information
func (h *OAuthHandler) createOAuthUser(w http.ResponseWriter, provider, oauthID, username, email, profilePic string) error {
	// Generate a unique username if needed
	if username == "" {
		username = fmt.Sprintf("%s_user_%s", provider, strings.Split(oauthID, "-")[0])
	}

	// Check if the username already exists
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return err
	}

	// If the username exists, append a random string
	if count > 0 {
		username = fmt.Sprintf("%s_%s", username, strings.Split(uuid.New().String(), "-")[0])
	}

	// Generate a user ID
	userID := uuid.New().String()

	// Insert the new user
	_, err = h.db.Exec(
		"INSERT INTO users (id, username, email, profile_pic, oauth_provider, oauth_id) VALUES (?, ?, ?, ?, ?, ?)",
		userID, username, email, profilePic, provider, oauthID,
	)
	if err != nil {
		return err
	}

	// Create a session for the new user
	CreateSession(w, userID)
	return nil
}