package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// OAuthProvider represents an OAuth provider configuration
type OAuthProvider struct {
	Name             string
	ClientID         string
	ClientSecret     string
	AuthURL          string
	TokenURL         string
	UserInfoURL      string
	Scopes           []string
	RedirectURL      string
	UserIDKey        string
	UserNameKey      string
	UserEmailKey     string
	UserProfilePicKey string
}

// OAuthConfig holds the configuration for all OAuth providers
type OAuthConfig struct {
	Google OAuthProvider
	GitHub OAuthProvider
}

var Config OAuthConfig

// InitOAuthConfig initializes the OAuth configuration
func InitOAuthConfig(baseURL string) {
	// Google OAuth configuration
	Config.Google = OAuthProvider{
		Name:             "google",
		ClientID:         getEnv("GOOGLE_CLIENT_ID", "test_google_client_id"),
		ClientSecret:     getEnv("GOOGLE_CLIENT_SECRET", "test_google_client_secret"),
		AuthURL:          "https://accounts.google.com/o/oauth2/auth",
		TokenURL:         "https://oauth2.googleapis.com/token",
		UserInfoURL:      "https://www.googleapis.com/oauth2/v3/userinfo",
		Scopes:           []string{"profile", "email"},
		RedirectURL:      fmt.Sprintf("%s/auth/google/callback", baseURL),
		UserIDKey:        "sub",
		UserNameKey:      "name",
		UserEmailKey:     "email",
		UserProfilePicKey: "picture",
	}

	// GitHub OAuth configuration
	Config.GitHub = OAuthProvider{
		Name:             "github",
		ClientID:         getEnv("GITHUB_CLIENT_ID", "test_github_client_id"),
		ClientSecret:     getEnv("GITHUB_CLIENT_SECRET", "test_github_client_secret"),
		AuthURL:          "https://github.com/login/oauth/authorize",
		TokenURL:         "https://github.com/login/oauth/access_token",
		UserInfoURL:      "https://api.github.com/user",
		Scopes:           []string{"read:user", "user:email"},
		RedirectURL:      fmt.Sprintf("%s/auth/github/callback", baseURL),
		UserIDKey:        "id",
		UserNameKey:      "login",
		UserEmailKey:     "email",
		UserProfilePicKey: "avatar_url",
	}

	log.Printf("OAuth configuration initialized for providers: Google, GitHub")
}

// GetAuthURL generates the authorization URL for the specified provider
func (p *OAuthProvider) GetAuthURL(state string) string {
	u, err := url.Parse(p.AuthURL)
	if err != nil {
		log.Printf("Error parsing auth URL: %v", err)
		return ""
	}

	q := u.Query()
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURL)
	q.Set("response_type", "code")
	q.Set("state", state)
	q.Set("scope", strings.Join(p.Scopes, " "))
	u.RawQuery = q.Encode()

	return u.String()
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (p *OAuthProvider) ExchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", p.ClientID)
	data.Set("client_secret", p.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", p.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", p.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating token request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error exchanging code for token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading token response: %v", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		// GitHub might return non-JSON response, try parsing as query string
		if p.Name == "github" {
			values, parseErr := url.ParseQuery(string(body))
			if parseErr != nil {
				return "", fmt.Errorf("error parsing token response: %v", err)
			}
			tokenResp.AccessToken = values.Get("access_token")
			tokenResp.TokenType = values.Get("token_type")
			tokenResp.Error = values.Get("error")
		} else {
			return "", fmt.Errorf("error parsing token response: %v", err)
		}
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("error from provider: %s", tokenResp.Error)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return tokenResp.AccessToken, nil
}

// GetUserInfo fetches the user information using the access token
func (p *OAuthProvider) GetUserInfo(accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", p.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating user info request: %v", err)
	}

	if p.Name == "github" {
		req.Header.Set("Authorization", "token "+accessToken)
	} else {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching user info: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading user info response: %v", err)
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("error parsing user info: %v", err)
	}

	// For GitHub, we might need to make an additional request to get the email
	if p.Name == "github" && userInfo[p.UserEmailKey] == nil {
		email, err := p.getGitHubEmail(accessToken)
		if err != nil {
			log.Printf("Error fetching GitHub email: %v", err)
		} else {
			userInfo[p.UserEmailKey] = email
		}
	}

	return userInfo, nil
}

// getGitHubEmail fetches the user's email from GitHub's email API
func (p *OAuthProvider) getGitHubEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	// Find primary and verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// If no primary and verified email, return the first verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	// If no verified email, return the first email
	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}

// GenerateState generates a random state string for CSRF protection
func GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}