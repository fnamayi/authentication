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

	token, err := auth.GoogleConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := auth.GoogleConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	// Example: Print email to log
	log.Printf("Google User: %v", userInfo["email"])
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
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
