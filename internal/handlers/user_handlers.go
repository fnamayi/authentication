package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/models"
)

// ProfileHandler handles profile operations
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getProfileHandler(w, r)
	case http.MethodPut:
		updateProfileHandler(w, r)
	default:
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getProfileHandler gets user profile
func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	RenderSuccess(w, "Profile retrieved successfully", user)
}

// updateProfileHandler updates user profile
func updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Age       int    `json:"age"`
		Gender    string `json:"gender"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.FirstName == "" || req.LastName == "" {
		RenderError(w, "First name and last name are required", http.StatusBadRequest)
		return
	}

	if req.Age < 13 {
		RenderError(w, "Age must be at least 13", http.StatusBadRequest)
		return
	}

	if req.Gender != "male" && req.Gender != "female" {
		RenderError(w, "Gender must be 'male' or 'female'", http.StatusBadRequest)
		return
	}

	// Update user profile
	_, err := database.DB.Exec(`
		UPDATE users
		SET first_name = ?, last_name = ?, age = ?, gender = ?, updated_at = ?
		WHERE id = ?
	`, req.FirstName, req.LastName, req.Age, req.Gender, time.Now(), user.ID)

	if err != nil {
		RenderError(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Get updated user data
	updatedUser, err := auth.GetUserByID(user.ID)
	if err != nil {
		RenderError(w, "Failed to retrieve updated profile", http.StatusInternalServerError)
		return
	}

	RenderSuccess(w, "Profile updated successfully", updatedUser)
}

// OnlineUsersHandler handles online users listing
func OnlineUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	onlineUsers, err := getOnlineUsers()
	if err != nil {
		log.Printf("❌ OnlineUsersHandler: Error fetching online users: %v", err)
		RenderError(w, "Failed to fetch online users", http.StatusInternalServerError)
		return
	}

	log.Printf("👥 OnlineUsersHandler: Found %d online users for %s", len(onlineUsers), user.Nickname)

	RenderSuccess(w, "Online users retrieved successfully", onlineUsers)
}

// Helper functions for user handlers

// getOnlineUsers gets all currently online users
func getOnlineUsers() ([]models.OnlineUser, error) {
	// Get users from the database who have been active recently (group by user_id to avoid duplicates)
	rows, err := database.DB.Query(`
		SELECT ou.user_id, u.nickname, u.first_name, u.last_name, u.avatar_url, MAX(ou.last_seen) as last_seen
		FROM online_users ou
		JOIN users u ON ou.user_id = u.id
		WHERE ou.last_seen > datetime('now', '-10 minutes')
		GROUP BY ou.user_id, u.nickname, u.first_name, u.last_name, u.avatar_url
		ORDER BY u.nickname ASC
	`)

	if err != nil {
		log.Printf("❌ getOnlineUsers: Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var onlineUsers []models.OnlineUser
	for rows.Next() {
		var user models.OnlineUser
		var lastSeenStr string

		err := rows.Scan(
			&user.UserID, &user.Nickname, &user.FirstName, &user.LastName, &user.AvatarURL, &lastSeenStr,
		)
		if err != nil {
			log.Printf("❌ getOnlineUsers: Row scan error: %v", err)
			return nil, err
		}

		// Parse the timestamp string into time.Time
		if lastSeenStr != "" {
			parsedTime, err := time.Parse("2006-01-02T15:04:05Z", lastSeenStr)
			if err != nil {
				// Try alternative format
				parsedTime, err = time.Parse("2006-01-02 15:04:05", lastSeenStr)
				if err != nil {
					log.Printf("⚠️ getOnlineUsers: Could not parse timestamp %s: %v", lastSeenStr, err)
					parsedTime = time.Now() // Use current time as fallback
				}
			}
			user.LastSeen = parsedTime
		} else {
			user.LastSeen = time.Now()
		}

		onlineUsers = append(onlineUsers, user)
	}

	return onlineUsers, nil
}

// AvatarUploadHandler handles avatar file uploads
func AvatarUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(5 << 20) // 5MB limit
	if err != nil {
		RenderError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		RenderError(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		RenderError(w, "File must be an image", http.StatusBadRequest)
		return
	}

	// Validate file size (5MB limit)
	if header.Size > 5<<20 {
		RenderError(w, "File size must be less than 5MB", http.StatusBadRequest)
		return
	}

	// Get current working directory for debugging
	wd, _ := os.Getwd()
	log.Printf("📂 Current working directory: %s", wd)

	// Create uploads directory if it doesn't exist
	uploadsDir := "frontend/static/uploads/avatars"
	log.Printf("📁 Creating uploads directory: %s", uploadsDir)
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("❌ Failed to create upload directory: %v", err)
		RenderError(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := user.ID + "_" + time.Now().Format("20060102_150405") + ext
	filePath := filepath.Join(uploadsDir, filename)
	log.Printf("📄 Saving file to: %s", filePath)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("❌ Failed to create file: %v", err)
		RenderError(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		log.Printf("❌ Failed to copy file: %v", err)
		RenderError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("❌ File was not created: %s", filePath)
		RenderError(w, "File was not saved properly", http.StatusInternalServerError)
		return
	}
	log.Printf("✅ File created successfully: %s", filePath)

	// Update user avatar URL in database (use forward slashes for web URLs)
	avatarURL := "/static/uploads/avatars/" + filename
	log.Printf("🔗 Avatar URL: %s", avatarURL)
	_, err = database.DB.Exec(`
		UPDATE users SET avatar_url = ?, updated_at = ? WHERE id = ?
	`, avatarURL, time.Now(), user.ID)

	if err != nil {
		log.Printf("❌ Failed to update database: %v", err)
		RenderError(w, "Failed to update avatar", http.StatusInternalServerError)
		return
	}

	// Test if file is accessible (optional debug step)
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("✅ File exists and is accessible: %s", filePath)
	} else {
		log.Printf("⚠️ File may not be accessible: %v", err)
	}

	log.Printf("✅ Avatar uploaded successfully for user %s: %s", user.ID, avatarURL)
	RenderSuccess(w, "Avatar uploaded successfully", map[string]string{
		"avatarURL": avatarURL,
	})
}

// AvatarUpdateHandler handles avatar URL updates (for default avatars)
func AvatarUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		AvatarURL string `json:"avatarUrl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate avatar URL (must be a default avatar or uploaded file)
	if req.AvatarURL == "" {
		RenderError(w, "Avatar URL is required", http.StatusBadRequest)
		return
	}

	// Update user avatar URL in database
	_, err := database.DB.Exec(`
		UPDATE users SET avatar_url = ?, updated_at = ? WHERE id = ?
	`, req.AvatarURL, time.Now(), user.ID)

	if err != nil {
		RenderError(w, "Failed to update avatar", http.StatusInternalServerError)
		return
	}

	// Get updated user data
	updatedUser, err := auth.GetUserByID(user.ID)
	if err != nil {
		RenderError(w, "Failed to retrieve updated user", http.StatusInternalServerError)
		return
	}

	RenderSuccess(w, "Avatar updated successfully", updatedUser)
}
