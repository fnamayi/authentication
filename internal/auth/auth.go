package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"forum/internal/database"
	"forum/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionCookieName = "forum_session"
	SessionDuration   = 24 * time.Hour * 7 // 7 days
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateSessionID generates a random session ID
func GenerateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateSession creates a new session for a user
func CreateSession(userID string) (*models.Session, error) {
	sessionID := GenerateSessionID()
	expiresAt := time.Now().Add(SessionDuration)

	session := &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	_, err := database.DB.Exec(`
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`, session.ID, session.UserID, session.ExpiresAt, session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return session, nil
}

// GetSessionByID retrieves a session by its ID
func GetSessionByID(sessionID string) (*models.Session, error) {
	session := &models.Session{}
	err := database.DB.QueryRow(`
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ? AND expires_at > ?
	`, sessionID, time.Now()).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %v", err)
	}

	return session, nil
}

// DeleteSession deletes a session
func DeleteSession(sessionID string) error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %v", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func CleanupExpiredSessions() error {
	_, err := database.DB.Exec("DELETE FROM sessions WHERE expires_at <= ?", time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %v", err)
	}
	return nil
}

// SetSessionCookie sets the session cookie on the response
func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetSessionFromRequest extracts session from request cookie
func GetSessionFromRequest(r *http.Request) (*models.Session, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, nil // No session cookie - this is normal
		}
		return nil, fmt.Errorf("failed to read session cookie: %v", err)
	}

	if cookie.Value == "" {
		return nil, nil // Empty session cookie
	}

	session, err := GetSessionByID(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session: %v", err)
	}

	return session, nil
}

// GetUserFromSession gets the user associated with the session in the request
func GetUserFromSession(r *http.Request) *models.User {
	session, err := GetSessionFromRequest(r)
	if err != nil || session == nil {
		return nil
	}

	user, err := GetUserByID(session.UserID)
	if err != nil {
		return nil
	}

	return user
}

// CreateUser creates a new user in the database
func CreateUser(req *models.RegisterRequest) (*models.User, error) {
	// Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Generate UUID for user
	userID := uuid.New().String()

	user := &models.User{
		ID:        userID,
		Email:     req.Email,
		Nickname:  req.Nickname,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		Gender:    req.Gender,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = database.DB.Exec(`
		INSERT INTO users (id, email, nickname, password, first_name, last_name, age, gender, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.Email, user.Nickname, user.Password, user.FirstName, user.LastName, user.Age, user.Gender, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Don't return the password
	user.Password = ""
	return user, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(userID string) (*models.User, error) {
	user := &models.User{}
	err := database.DB.QueryRow(`
		SELECT id, email, nickname, first_name, last_name, age, gender, google_id, github_id, avatar_url, created_at, updated_at
		FROM users
		WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Nickname,
		&user.FirstName,
		&user.LastName,
		&user.Age,
		&user.Gender,
		&user.GoogleID,
		&user.GithubID,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetUserByEmailOrNickname retrieves a user by email or nickname
func GetUserByEmailOrNickname(identifier string) (*models.User, error) {
	user := &models.User{}
	err := database.DB.QueryRow(`
		SELECT id, email, nickname, password, first_name, last_name, age, gender, google_id, github_id, avatar_url, created_at, updated_at
		FROM users
		WHERE email = ? OR nickname = ?
	`, identifier, identifier).Scan(
		&user.ID,
		&user.Email,
		&user.Nickname,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Age,
		&user.Gender,
		&user.GoogleID,
		&user.GithubID,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email/nickname and password
func AuthenticateUser(identifier, password string) (*models.User, error) {
	user, err := GetUserByEmailOrNickname(identifier)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if !CheckPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	// Don't return the password
	user.Password = ""
	return user, nil
}

// RequireAuth middleware to require authentication
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromSession(r)
		if user == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// InitializeOAuthProviders initializes OAuth providers (placeholder)
func InitializeOAuthProviders() {
	// OAuth initialization would go here
	// For now, this is a placeholder to satisfy the main.go import
}
