package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/models"
	"forum/internal/websocket"
)

// RenderSuccess renders a successful JSON response
func RenderSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

// RenderError renders an error JSON response
func RenderError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := models.APIResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(response)
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Nickname == "" || req.Password == "" {
		RenderError(w, "Email, nickname, and password are required", http.StatusBadRequest)
		return
	}

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

	// Create user
	user, err := auth.CreateUser(&req)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "email") {
				RenderError(w, "Email already exists", http.StatusConflict)
			} else if strings.Contains(err.Error(), "nickname") {
				RenderError(w, "Nickname already exists", http.StatusConflict)
			} else {
				RenderError(w, "User already exists", http.StatusConflict)
			}
		} else {
			RenderError(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	// Create session
	session, err := auth.CreateSession(user.ID)
	if err != nil {
		RenderError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	auth.SetSessionCookie(w, session.ID)

	RenderSuccess(w, "User registered successfully", user)
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Login error - Invalid request body: %v", err)
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Identifier == "" || req.Password == "" {
		log.Printf("Login error - Missing credentials")
		RenderError(w, "Email/nickname and password are required", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for identifier: %s", req.Identifier)

	// Authenticate user
	user, err := auth.AuthenticateUser(req.Identifier, req.Password)
	if err != nil {
		log.Printf("Login error - Authentication failed: %v", err)
		RenderError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("User authenticated successfully: %s", user.ID)

	// Create session
	session, err := auth.CreateSession(user.ID)
	if err != nil {
		log.Printf("Login error - Failed to create session: %v", err)
		RenderError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	log.Printf("Session created successfully: %s", session.ID)

	// Set session cookie
	auth.SetSessionCookie(w, session.ID)

	log.Printf("Login successful for user: %s", user.Nickname)
	RenderSuccess(w, "Login successful", user)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("🔒 Logout attempt started")

	// Get session from request
	session, err := auth.GetSessionFromRequest(r)
	if err != nil {
		log.Printf("❌ Logout error - Failed to get session from request: %v", err)
		// Still clear the cookie even if session retrieval fails
		auth.ClearSessionCookie(w)
		RenderSuccess(w, "Logout successful", nil)
		return
	}

	if session == nil {
		log.Printf("⚠️ Logout warning - No active session found, clearing cookie anyway")
		// Still clear the cookie even if no session exists
		auth.ClearSessionCookie(w)
		RenderSuccess(w, "Logout successful", nil)
		return
	}

	log.Printf("🔒 Deleting session: %s for user: %s", session.ID, session.UserID)

	// Delete session from database
	if err := auth.DeleteSession(session.ID); err != nil {
		log.Printf("❌ Logout error - Failed to delete session: %v", err)
		// Still clear the cookie even if database deletion fails
		auth.ClearSessionCookie(w)
		RenderError(w, "Failed to logout completely, but session cleared", http.StatusInternalServerError)
		return
	}

	// Remove user from online_users table for this specific session
	_, err = database.DB.Exec(`
		DELETE FROM online_users WHERE user_id = ? AND session_id = ?
	`, session.UserID, session.ID)
	if err != nil {
		log.Printf("⚠️ Logout warning - Failed to remove user from online_users: %v", err)
		// Don't fail the logout, just log the warning
	} else {
		log.Printf("✅ User %s session %s removed from online_users", session.UserID, session.ID)
	}

	// Check if user has any remaining sessions online
	var remainingSessions int
	err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM online_users WHERE user_id = ?
	`, session.UserID).Scan(&remainingSessions)

	if err != nil {
		log.Printf("⚠️ Logout warning - Failed to check remaining sessions: %v", err)
	} else if remainingSessions == 0 {
		// User has no more online sessions, broadcast offline status
		log.Printf("📡 Broadcasting offline status for user %s (no remaining sessions)", session.UserID)
		websocket.BroadcastUserOffline(session.UserID)
	} else {
		log.Printf("📡 User %s still has %d remaining online sessions", session.UserID, remainingSessions)
	}

	// Clear session cookie
	auth.ClearSessionCookie(w)

	log.Printf("✅ Logout successful for session: %s", session.ID)
	RenderSuccess(w, "Logout successful", nil)
}

// CurrentUserHandler returns the current authenticated user
func CurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	RenderSuccess(w, "Current user retrieved", user)
}

// PostsHandler handles posts listing and creation
func PostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getPostsHandler(w, r)
	case http.MethodPost:
		createPostHandler(w, r)
	default:
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getPostsHandler handles getting posts
func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	category := r.URL.Query().Get("category")

	// Build query
	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.image_path, p.created_at,
		       u.nickname, u.avatar_url,
		       COALESCE(like_counts.like_count, 0) as like_count,
		       COALESCE(like_counts.dislike_count, 0) as dislike_count,
		       COALESCE(comment_counts.comment_count, 0) as comment_count
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT post_id,
			       SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END) as like_count,
			       SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END) as dislike_count
			FROM likes
			WHERE post_id IS NOT NULL
			GROUP BY post_id
		) like_counts ON p.id = like_counts.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as comment_count
			FROM comments
			GROUP BY post_id
		) comment_counts ON p.id = comment_counts.post_id
	`

	args := []interface{}{}

	if category != "" {
		query += ` WHERE p.id IN (
			SELECT post_id FROM post_categories WHERE category = ?
		)`
		args = append(args, category)
	}

	query += ` ORDER BY p.created_at DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		RenderError(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImagePath, &post.CreatedAt,
			&post.Author, &post.AuthorAvatar,
			&post.LikeCount, &post.DislikeCount, &post.CommentCount,
		)
		if err != nil {
			RenderError(w, "Failed to scan post", http.StatusInternalServerError)
			return
		}

		// Load categories for this post
		categories, err := getPostCategories(post.ID)
		if err != nil {
			RenderError(w, "Failed to load post categories", http.StatusInternalServerError)
			return
		}
		post.Categories = categories

		// Check if current user liked/disliked this post
		user := auth.GetUserFromSession(r)
		if user != nil {
			userLike, err := getUserLikeStatus(user.ID, &post.ID, nil)
			if err == nil && userLike != nil {
				post.UserLiked = userLike.IsLike
				post.UserDisliked = !userLike.IsLike
			}
		}

		posts = append(posts, post)
	}

	if posts == nil {
		posts = []models.Post{}
	}

	RenderSuccess(w, "Posts retrieved successfully", posts)
}

// createPostHandler handles post creation
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromSession(r)
	if user == nil {
		log.Printf("Post creation error - Authentication required")
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	log.Printf("Creating post for user: %s", user.ID)

	var req models.PostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Post creation error - Invalid request body: %v", err)
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Post request: Title=%s, Content length=%d, Categories=%v", req.Title, len(req.Content), req.Categories)

	// Validate input
	if req.Title == "" || req.Content == "" {
		log.Printf("Post creation error - Missing title or content")
		RenderError(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	// Insert post
	result, err := database.DB.Exec(`
		INSERT INTO posts (user_id, title, content, image_path, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, user.ID, req.Title, req.Content, req.ImagePath, time.Now())

	if err != nil {
		log.Printf("Post creation error - Failed to insert post: %v", err)
		RenderError(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	postID, _ := result.LastInsertId()
	log.Printf("Post created with ID: %d", postID)

	// Insert categories
	for _, category := range req.Categories {
		if category != "" {
			_, err := database.DB.Exec(`
				INSERT INTO post_categories (post_id, category)
				VALUES (?, ?)
			`, postID, category)
			if err != nil {
				log.Printf("Warning - Failed to insert category %s: %v", category, err)
			}
		}
	}

	// Get the created post
	post, err := getPostByID(int(postID))
	if err != nil {
		log.Printf("Post creation error - Failed to retrieve created post: %v", err)
		RenderError(w, "Failed to retrieve created post", http.StatusInternalServerError)
		return
	}

	log.Printf("Post retrieved successfully: %s", post.Title)

	// Broadcast new post to all users
	websocket.BroadcastNewPost(post)

	log.Printf("Post creation successful for user: %s", user.Nickname)
	RenderSuccess(w, "Post created successfully", post)
}

// Helper functions

// getPostCategories gets categories for a post
func getPostCategories(postID int) ([]string, error) {
	rows, err := database.DB.Query("SELECT category FROM post_categories WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// getUserLikeStatus gets user's like status for a post or comment
func getUserLikeStatus(userID string, postID *int, commentID *int) (*models.Like, error) {
	var like models.Like
	var err error

	if postID != nil {
		err = database.DB.QueryRow(`
			SELECT id, user_id, post_id, comment_id, is_like, created_at
			FROM likes
			WHERE user_id = ? AND post_id = ?
		`, userID, *postID).Scan(
			&like.ID, &like.UserID, &like.PostID, &like.CommentID, &like.IsLike, &like.CreatedAt,
		)
	} else if commentID != nil {
		err = database.DB.QueryRow(`
			SELECT id, user_id, post_id, comment_id, is_like, created_at
			FROM likes
			WHERE user_id = ? AND comment_id = ?
		`, userID, *commentID).Scan(
			&like.ID, &like.UserID, &like.PostID, &like.CommentID, &like.IsLike, &like.CreatedAt,
		)
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &like, nil
}

// getPostByID gets a post by ID
func getPostByID(postID int) (*models.Post, error) {
	var post models.Post
	err := database.DB.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.content, p.image_path, p.created_at,
		       u.nickname, u.avatar_url,
		       COALESCE(like_counts.like_count, 0) as like_count,
		       COALESCE(like_counts.dislike_count, 0) as dislike_count,
		       COALESCE(comment_counts.comment_count, 0) as comment_count
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT post_id,
			       SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END) as like_count,
			       SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END) as dislike_count
			FROM likes
			WHERE post_id IS NOT NULL
			GROUP BY post_id
		) like_counts ON p.id = like_counts.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as comment_count
			FROM comments
			GROUP BY post_id
		) comment_counts ON p.id = comment_counts.post_id
		WHERE p.id = ?
	`, postID).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImagePath, &post.CreatedAt,
		&post.Author, &post.AuthorAvatar,
		&post.LikeCount, &post.DislikeCount, &post.CommentCount,
	)

	if err != nil {
		return nil, err
	}

	// Load categories
	categories, err := getPostCategories(post.ID)
	if err != nil {
		return nil, err
	}
	post.Categories = categories

	return &post, nil
}
