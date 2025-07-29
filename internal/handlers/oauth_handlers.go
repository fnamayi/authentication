package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"forum/internal/auth"
	"golang.org/x/oauth2"
)

// ====== GOOGLE ======
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.GoogleConfig.AuthCodeURL("state-google", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in URL", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := auth.GoogleConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Google OAuth error - Failed to exchange token: %v", err)
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := auth.GoogleConfig.Client(context.Background(), token)

	// Get user profile information
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("Google OAuth error - Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("Google OAuth error - Failed to parse user info: %v", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Extract user information
	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		log.Printf("Google OAuth error - No email found in user info")
		http.Error(w, "No email found in Google account", http.StatusBadRequest)
		return
	}

	googleID, ok := userInfo["id"].(string)
	if !ok || googleID == "" {
		log.Printf("Google OAuth error - No Google ID found")
		http.Error(w, "No Google ID found", http.StatusBadRequest)
		return
	}

	// Extract name information
	firstName := ""
	lastName := ""
	nickname := ""

	if givenName, ok := userInfo["given_name"].(string); ok {
		firstName = givenName
	}

	if familyName, ok := userInfo["family_name"].(string); ok {
		lastName = familyName
	}

	// Use name as nickname, fallback to email prefix
	if name, ok := userInfo["name"].(string); ok && name != "" {
		// Remove spaces and special characters for nickname
		nickname = strings.ReplaceAll(strings.ToLower(name), " ", "")
		// Remove non-alphanumeric characters except underscores
		nickname = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
				return r
			}
			return -1
		}, nickname)
	}

	if nickname == "" {
		// Fallback to email prefix
		emailParts := strings.Split(email, "@")
		if len(emailParts) > 0 {
			nickname = emailParts[0]
		}
	}

	// Get avatar URL
	avatarURL := ""
	if picture, ok := userInfo["picture"].(string); ok {
		avatarURL = picture
	}

	log.Printf("Google OAuth - Processing user: %s (%s)", nickname, email)

	// Create or update user
	user, err := auth.CreateOrUpdateGoogleUser(googleID, email, nickname, firstName, lastName, avatarURL)
	if err != nil {
		log.Printf("Google OAuth error - Failed to create/update user: %v", err)
		http.Error(w, "Failed to create user account", http.StatusInternalServerError)
		return
	}

	log.Printf("Google OAuth - User created/updated successfully: %s", user.ID)

	// Create session
	session, err := auth.CreateSession(user.ID)
	if err != nil {
		log.Printf("Google OAuth error - Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	auth.SetSessionCookie(w, session.ID)

	log.Printf("Google OAuth - Login successful for user: %s", user.Nickname)

	// Redirect to main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ====== GITHUB ======
func GithubLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.GithubConfig.AuthCodeURL("state-github", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in URL", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := auth.GithubConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("GitHub OAuth error - Failed to exchange token: %v", err)
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := auth.GithubConfig.Client(context.Background(), token)

	// Get user profile information
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		log.Printf("GitHub OAuth error - Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		log.Printf("GitHub OAuth error - Failed to parse user info: %v", err)
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Get user emails
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		log.Printf("GitHub OAuth error - Failed to get emails: %v", err)
		http.Error(w, "Failed to get emails: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer emailResp.Body.Close()

	var emails []map[string]interface{}
	if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
		log.Printf("GitHub OAuth error - Failed to parse emails: %v", err)
		http.Error(w, "Failed to parse emails", http.StatusInternalServerError)
		return
	}

	// Find primary email
	var primaryEmail string
	for _, email := range emails {
		if primary, ok := email["primary"].(bool); ok && primary {
			if emailStr, ok := email["email"].(string); ok {
				primaryEmail = emailStr
				break
			}
		}
	}

	if primaryEmail == "" {
		log.Printf("GitHub OAuth error - No primary email found")
		http.Error(w, "No primary email found in GitHub account", http.StatusBadRequest)
		return
	}

	// Extract user information
	githubID := fmt.Sprintf("%.0f", userInfo["id"].(float64))
	nickname := userInfo["login"].(string)
	firstName := ""
	lastName := ""
	avatarURL := ""

	if name, ok := userInfo["name"].(string); ok && name != "" {
		// Split name into first and last name
		names := strings.Fields(name)
		if len(names) > 0 {
			firstName = names[0]
		}
		if len(names) > 1 {
			lastName = strings.Join(names[1:], " ")
		}
	}

	if avatar, ok := userInfo["avatar_url"].(string); ok {
		avatarURL = avatar
	}

	log.Printf("GitHub OAuth - Processing user: %s (%s)", nickname, primaryEmail)

	// Create or update user
	user, err := auth.CreateOrUpdateGithubUser(githubID, primaryEmail, nickname, firstName, lastName, avatarURL)
	if err != nil {
		log.Printf("GitHub OAuth error - Failed to create/update user: %v", err)
		http.Error(w, "Failed to create user account", http.StatusInternalServerError)
		return
	}

	log.Printf("GitHub OAuth - User created/updated successfully: %s", user.ID)

	// Create session
	session, err := auth.CreateSession(user.ID)
	if err != nil {
		log.Printf("GitHub OAuth error - Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	auth.SetSessionCookie(w, session.ID)

	log.Printf("GitHub OAuth - Login successful for user: %s", user.Nickname)

	// Redirect to main page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
