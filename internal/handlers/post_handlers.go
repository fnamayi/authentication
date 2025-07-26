package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/models"
)

// PostHandler handles individual post operations
func PostHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🚀 PostHandler START - URL: %s", r.URL.Path)

	if r.Method != http.MethodGet {
		log.Printf("❌ Method not allowed: %s", r.Method)
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("📝 PostHandler called with URL: %s", r.URL.Path)

	// Extract path after /api/posts/
	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	log.Printf("📝 Extracted path: %s", path)

	// Check if this is a comments request
	if strings.Contains(path, "/comments") {
		log.Printf("📝 Detected comments request, delegating to handleCommentsRequest")
		handleCommentsRequest(w, r, path)
		return
	}

	// Handle regular post request
	postID, err := strconv.Atoi(path)
	if err != nil {
		RenderError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Get post with comments
	post, err := getPostWithComments(postID)
	if err != nil {
		RenderError(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check if current user liked/disliked this post
	user := auth.GetUserFromSession(r)
	if user != nil {
		userLike, err := getUserLikeStatus(user.ID, &post.ID, nil)
		if err == nil && userLike != nil {
			post.UserLiked = userLike.IsLike
			post.UserDisliked = !userLike.IsLike
		}

		// Check like status for comments
		for i := range post.Comments {
			commentLike, err := getUserLikeStatus(user.ID, nil, &post.Comments[i].ID)
			if err == nil && commentLike != nil {
				post.Comments[i].UserLiked = commentLike.IsLike
				post.Comments[i].UserDisliked = !commentLike.IsLike
			}
		}
	}

	RenderSuccess(w, "Post retrieved successfully", post)
}

// CommentHandler handles comment operations (create, update, delete)
func CommentHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🚀 CommentHandler - Method: %s, URL: %s", r.Method, r.URL.Path)

	// Check if this is a specific comment operation (has ID in path)
	path := strings.TrimPrefix(r.URL.Path, "/api/comment")
	log.Printf("📝 Extracted path after /api/comment: '%s'", path)

	if path == "" || path == "/" {
		// POST /api/comment - create comment
		if r.Method == http.MethodPost {
			handleCreateComment(w, r)
		} else {
			RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// Operations on specific comments: /api/comment/{id}
	switch r.Method {
	case http.MethodPut:
		// PUT /api/comment/{id} - update comment
		handleUpdateComment(w, r)
	case http.MethodDelete:
		// DELETE /api/comment/{id} - delete comment
		handleDeleteComment(w, r)
	default:
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateComment handles comment creation
func handleCreateComment(w http.ResponseWriter, r *http.Request) {

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.CommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.PostID == 0 || req.Content == "" {
		RenderError(w, "Post ID and content are required", http.StatusBadRequest)
		return
	}

	// Insert comment
	result, err := database.DB.Exec(`
		INSERT INTO comments (post_id, user_id, parent_id, content, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, req.PostID, user.ID, req.ParentID, req.Content, time.Now())

	if err != nil {
		RenderError(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	commentID, _ := result.LastInsertId()

	// Get the created comment
	comment, err := getCommentByID(int(commentID))
	if err != nil {
		RenderError(w, "Failed to retrieve created comment", http.StatusInternalServerError)
		return
	}

	RenderSuccess(w, "Comment created successfully", comment)
}

// handleUpdateComment handles comment updates
func handleUpdateComment(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔄 handleUpdateComment - URL: %s", r.URL.Path)

	// Get user from session with better error handling
	user := auth.GetUserFromSession(r)
	if user == nil {
		log.Printf("❌ No user in session for comment update")
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	log.Printf("✅ User %s (%s) updating comment", user.Nickname, user.ID)

	// Extract comment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/comment/")
	if path == "" {
		path = strings.TrimPrefix(r.URL.Path, "/api/comment")
		path = strings.TrimPrefix(path, "/")
	}
	log.Printf("📝 Extracted path: %s", path)

	commentID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("❌ Invalid comment ID: %s, error: %v", path, err)
		RenderError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	log.Printf("📝 Comment ID: %d", commentID)

	// Get existing comment to verify ownership
	existingComment, err := getCommentByID(commentID)
	if err != nil {
		RenderError(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Check if user owns the comment
	if existingComment.UserID != user.ID {
		RenderError(w, "You can only edit your own comments", http.StatusForbidden)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Content == "" {
		RenderError(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Update comment
	_, err = database.DB.Exec(`
		UPDATE comments SET content = ?, updated_at = ? WHERE id = ?
	`, req.Content, time.Now(), commentID)

	if err != nil {
		RenderError(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	// Get updated comment
	updatedComment, err := getCommentByID(commentID)
	if err != nil {
		RenderError(w, "Failed to retrieve updated comment", http.StatusInternalServerError)
		return
	}

	RenderSuccess(w, "Comment updated successfully", updatedComment)
}

// handleDeleteComment handles comment deletion
func handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	log.Printf("🗑️ handleDeleteComment - URL: %s", r.URL.Path)

	user := auth.GetUserFromSession(r)
	if user == nil {
		log.Printf("❌ No user in session")
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract comment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/comment/")
	if path == "" {
		path = strings.TrimPrefix(r.URL.Path, "/api/comment")
		path = strings.TrimPrefix(path, "/")
	}
	log.Printf("📝 Extracted path: %s", path)

	commentID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("❌ Invalid comment ID: %s, error: %v", path, err)
		RenderError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	log.Printf("📝 Comment ID: %d", commentID)

	// Get existing comment to verify ownership
	existingComment, err := getCommentByID(commentID)
	if err != nil {
		RenderError(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Check if user owns the comment
	if existingComment.UserID != user.ID {
		RenderError(w, "You can only delete your own comments", http.StatusForbidden)
		return
	}

	// Delete comment likes first (foreign key constraint)
	_, err = database.DB.Exec(`DELETE FROM likes WHERE comment_id = ?`, commentID)
	if err != nil {
		RenderError(w, "Failed to delete comment likes", http.StatusInternalServerError)
		return
	}

	// Delete comment
	_, err = database.DB.Exec(`DELETE FROM comments WHERE id = ?`, commentID)
	if err != nil {
		RenderError(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	RenderSuccess(w, "Comment deleted successfully", nil)
}

// handleCommentsRequest handles requests for post comments
func handleCommentsRequest(w http.ResponseWriter, r *http.Request, path string) {
	log.Printf("📝 handleCommentsRequest called with path: %s", path)

	// Extract post ID from path like "123/comments"
	parts := strings.Split(path, "/")
	log.Printf("📝 Path parts: %v", parts)

	if len(parts) < 2 || parts[1] != "comments" {
		log.Printf("❌ Invalid comments request - parts: %v", parts)
		RenderError(w, "Invalid comments request", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Printf("❌ Invalid post ID: %s, error: %v", parts[0], err)
		RenderError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	log.Printf("📝 Getting comments for post ID: %d", postID)

	// Get comments for the post
	comments, err := getPostComments(postID)
	if err != nil {
		log.Printf("❌ Failed to get comments for post %d: %v", postID, err)
		RenderError(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Found %d comments for post %d", len(comments), postID)

	// Check if current user liked/disliked comments
	user := auth.GetUserFromSession(r)
	if user != nil {
		for i := range comments {
			commentLike, err := getUserLikeStatus(user.ID, nil, &comments[i].ID)
			if err == nil && commentLike != nil {
				comments[i].UserLiked = commentLike.IsLike
				comments[i].UserDisliked = !commentLike.IsLike
			}
		}
	}

	log.Printf("✅ Returning %d comments for post %d", len(comments), postID)
	RenderSuccess(w, "Comments retrieved successfully", comments)
}

// CommentsHandler handles comments requests with cleaner URL structure
func CommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("🚀 CommentsHandler START - URL: %s", r.URL.Path)
	fmt.Printf("🚀 CommentsHandler START - URL: %s\n", r.URL.Path)

	// Extract post ID from URL like /api/comments/123
	path := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	log.Printf("📝 Extracted path: %s", path)

	postID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("❌ Invalid post ID: %s, error: %v", path, err)
		RenderError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	log.Printf("📝 Getting comments for post ID: %d", postID)

	// Get comments for the post
	comments, err := getPostComments(postID)
	if err != nil {
		log.Printf("❌ Failed to get comments for post %d: %v", postID, err)
		RenderError(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Found %d comments for post %d", len(comments), postID)

	// Check if current user liked/disliked comments
	user := auth.GetUserFromSession(r)
	if user != nil {
		for i := range comments {
			commentLike, err := getUserLikeStatus(user.ID, nil, &comments[i].ID)
			if err == nil && commentLike != nil {
				comments[i].UserLiked = commentLike.IsLike
				comments[i].UserDisliked = !commentLike.IsLike
			}
		}
	}

	log.Printf("✅ Returning %d comments for post %d", len(comments), postID)
	RenderSuccess(w, "Comments retrieved successfully", comments)
}

// LikeHandler handles like/dislike operations
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("👍 LikeHandler - Method: %s, URL: %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		log.Printf("❌ No user in session for like request")
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Invalid request body: %v", err)
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("📝 Like request - PostID: %v, CommentID: %v, IsLike: %v", req.PostID, req.CommentID, req.IsLike)

	// Validate that either postID or commentID is provided
	if req.PostID == nil && req.CommentID == nil {
		RenderError(w, "Either postId or commentId is required", http.StatusBadRequest)
		return
	}

	// Check if user already liked/disliked this item
	var existingLike *models.Like
	var err error

	if req.PostID != nil {
		existingLike, err = getUserLikeStatus(user.ID, req.PostID, nil)
	} else {
		existingLike, err = getUserLikeStatus(user.ID, nil, req.CommentID)
	}

	if err != nil {
		RenderError(w, "Failed to check existing like status", http.StatusInternalServerError)
		return
	}

	if existingLike != nil {
		// If clicking the same action (like->like or dislike->dislike), remove the like
		if existingLike.IsLike == req.IsLike {
			_, err = database.DB.Exec(`DELETE FROM likes WHERE id = ?`, existingLike.ID)
		} else {
			// Update existing like to opposite action
			_, err = database.DB.Exec(`
				UPDATE likes SET is_like = ? WHERE id = ?
			`, req.IsLike, existingLike.ID)
		}
	} else {
		// Create new like
		_, err = database.DB.Exec(`
			INSERT INTO likes (user_id, post_id, comment_id, is_like, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, user.ID, req.PostID, req.CommentID, req.IsLike, time.Now())
	}

	if err != nil {
		RenderError(w, "Failed to update like status", http.StatusInternalServerError)
		return
	}

	// Return updated counts
	var likeCount, dislikeCount int
	if req.PostID != nil {
		// Get updated post counts
		database.DB.QueryRow(`
			SELECT
				COALESCE(SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END), 0) as like_count,
				COALESCE(SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END), 0) as dislike_count
			FROM likes WHERE post_id = ?
		`, *req.PostID).Scan(&likeCount, &dislikeCount)
	} else {
		// Get updated comment counts
		database.DB.QueryRow(`
			SELECT
				COALESCE(SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END), 0) as like_count,
				COALESCE(SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END), 0) as dislike_count
			FROM likes WHERE comment_id = ?
		`, *req.CommentID).Scan(&likeCount, &dislikeCount)
	}

	response := map[string]interface{}{
		"likeCount":    likeCount,
		"dislikeCount": dislikeCount,
	}

	RenderSuccess(w, "Like status updated successfully", response)
}

// CategoriesHandler handles category operations
func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := database.DB.Query(`
		SELECT category, COUNT(*) as post_count
		FROM post_categories
		GROUP BY category
		ORDER BY post_count DESC, category ASC
	`)
	if err != nil {
		RenderError(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.Name, &category.PostCount); err != nil {
			RenderError(w, "Failed to scan category", http.StatusInternalServerError)
			return
		}
		categories = append(categories, category)
	}

	if categories == nil {
		categories = []models.Category{}
	}

	RenderSuccess(w, "Categories retrieved successfully", categories)
}

// Helper functions for post handlers

// getPostWithComments gets a post with its comments
func getPostWithComments(postID int) (*models.Post, error) {
	// Get the post
	post, err := getPostByID(postID)
	if err != nil {
		return nil, err
	}

	// Get comments for the post
	comments, err := getPostComments(postID)
	if err != nil {
		return nil, err
	}

	post.Comments = comments
	return post, nil
}

// getPostComments gets all comments for a post
func getPostComments(postID int) ([]models.Comment, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
		       u.nickname, u.avatar_url,
		       COALESCE(like_counts.like_count, 0) as like_count,
		       COALESCE(like_counts.dislike_count, 0) as dislike_count
		FROM comments c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN (
			SELECT comment_id,
			       SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END) as like_count,
			       SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END) as dislike_count
			FROM likes
			WHERE comment_id IS NOT NULL
			GROUP BY comment_id
		) like_counts ON c.id = like_counts.comment_id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &comment.ParentID, &comment.Content, &comment.CreatedAt,
			&comment.Author, &comment.AuthorAvatar,
			&comment.LikeCount, &comment.DislikeCount,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// getCommentByID gets a comment by ID
func getCommentByID(commentID int) (*models.Comment, error) {
	var comment models.Comment
	err := database.DB.QueryRow(`
		SELECT c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
		       u.nickname, u.avatar_url,
		       COALESCE(like_counts.like_count, 0) as like_count,
		       COALESCE(like_counts.dislike_count, 0) as dislike_count
		FROM comments c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN (
			SELECT comment_id,
			       SUM(CASE WHEN is_like = 1 THEN 1 ELSE 0 END) as like_count,
			       SUM(CASE WHEN is_like = 0 THEN 1 ELSE 0 END) as dislike_count
			FROM likes
			WHERE comment_id IS NOT NULL
			GROUP BY comment_id
		) like_counts ON c.id = like_counts.comment_id
		WHERE c.id = ?
	`, commentID).Scan(
		&comment.ID, &comment.PostID, &comment.UserID, &comment.ParentID, &comment.Content, &comment.CreatedAt,
		&comment.Author, &comment.AuthorAvatar,
		&comment.LikeCount, &comment.DislikeCount,
	)

	if err != nil {
		return nil, err
	}

	return &comment, nil
}
