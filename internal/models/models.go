package models

import "time"

// User represents a user in the system with comprehensive profile information
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Nickname  string    `json:"nickname" db:"nickname"`
	Password  string    `json:"-" db:"password"` // Never expose password in JSON
	FirstName string    `json:"firstName" db:"first_name"`
	LastName  string    `json:"lastName" db:"last_name"`
	Age       int       `json:"age" db:"age"`
	Gender    string    `json:"gender" db:"gender"`
	GoogleID  *string   `json:"googleId,omitempty" db:"google_id"`
	GithubID  *string   `json:"githubId,omitempty" db:"github_id"`
	AvatarURL *string   `json:"avatarUrl,omitempty" db:"avatar_url"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Post represents a forum post with enhanced features
type Post struct {
	ID           int       `json:"id" db:"id"`
	UserID       string    `json:"userId" db:"user_id"`
	Title        string    `json:"title" db:"title"`
	Content      string    `json:"content" db:"content"`
	ImagePath    *string   `json:"imagePath,omitempty" db:"image_path"`
	Categories   []string  `json:"categories" db:"-"` // Loaded separately
	Author       string    `json:"author" db:"nickname"`
	AuthorAvatar *string   `json:"authorAvatar,omitempty" db:"avatar_url"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	LikeCount    int       `json:"likeCount" db:"like_count"`
	DislikeCount int       `json:"dislikeCount" db:"dislike_count"`
	CommentCount int       `json:"commentCount" db:"comment_count"`
	UserLiked    bool      `json:"userLiked" db:"user_liked"`
	UserDisliked bool      `json:"userDisliked" db:"user_disliked"`
	Comments     []Comment `json:"comments,omitempty" db:"-"` // Loaded on demand
}

// Comment represents a comment on a post with threading support
type Comment struct {
	ID           int       `json:"id" db:"id"`
	PostID       int       `json:"postId" db:"post_id"`
	UserID       string    `json:"userId" db:"user_id"`
	ParentID     *int      `json:"parentId,omitempty" db:"parent_id"`
	Content      string    `json:"content" db:"content"`
	Author       string    `json:"author" db:"nickname"`
	AuthorAvatar *string   `json:"authorAvatar,omitempty" db:"avatar_url"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	LikeCount    int       `json:"likeCount" db:"like_count"`
	DislikeCount int       `json:"dislikeCount" db:"dislike_count"`
	UserLiked    bool      `json:"userLiked" db:"user_liked"`
	UserDisliked bool      `json:"userDisliked" db:"user_disliked"`
}

// Like represents a like/dislike on a post or comment
type Like struct {
	ID        int       `json:"id" db:"id"`
	UserID    string    `json:"userId" db:"user_id"`
	PostID    *int      `json:"postId,omitempty" db:"post_id"`
	CommentID *int      `json:"commentId,omitempty" db:"comment_id"`
	IsLike    bool      `json:"isLike" db:"is_like"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"userId" db:"user_id"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// OnlineUser represents an online user
type OnlineUser struct {
	UserID    string    `json:"userId" db:"user_id"`
	Nickname  string    `json:"nickname" db:"nickname"`
	FirstName string    `json:"firstName" db:"first_name"`
	LastName  string    `json:"lastName" db:"last_name"`
	AvatarURL *string   `json:"avatarUrl,omitempty" db:"avatar_url"`
	LastSeen  time.Time `json:"lastSeen" db:"last_seen"`
}

// Message represents a private message between users
type Message struct {
	ID         string    `json:"id" db:"id"`
	SenderID   string    `json:"senderId" db:"sender_id"`
	ReceiverID string    `json:"receiverId" db:"receiver_id"`
	Content    string    `json:"content" db:"content"`
	IsRead     bool      `json:"isRead" db:"is_read"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
	// Additional fields for frontend display
	SenderNickname   string  `json:"senderNickname,omitempty" db:"sender_nickname"`
	SenderAvatarURL  *string `json:"senderAvatarUrl,omitempty" db:"sender_avatar_url"`
	ReceiverNickname string  `json:"receiverNickname,omitempty" db:"receiver_nickname"`
}

// Conversation represents a conversation between two users
type Conversation struct {
	ID              string    `json:"id" db:"id"`
	User1ID         string    `json:"user1Id" db:"user1_id"`
	User2ID         string    `json:"user2Id" db:"user2_id"`
	LastMessageID   *string   `json:"lastMessageId,omitempty" db:"last_message_id"`
	LastMessageTime time.Time `json:"lastMessageTime" db:"last_message_time"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" db:"updated_at"`
	// Additional fields for frontend display
	OtherUserID       string  `json:"otherUserId,omitempty"`
	OtherUserNickname string  `json:"otherUserNickname,omitempty"`
	OtherUserAvatar   *string `json:"otherUserAvatar,omitempty"`
	LastMessage       string  `json:"lastMessage,omitempty"`
	UnreadCount       int     `json:"unreadCount,omitempty"`
	IsOnline          bool    `json:"isOnline,omitempty"`
}

// Category represents a post category
type Category struct {
	Name      string `json:"name" db:"name"`
	PostCount int    `json:"postCount" db:"post_count"`
}

// Request/Response models for API endpoints

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Identifier string `json:"identifier"` // Can be email or nickname
	Password   string `json:"password"`
}

// PostRequest represents the post creation request payload
type PostRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"`
	ImagePath  *string  `json:"imagePath,omitempty"`
}

// CommentRequest represents the comment creation request payload
type CommentRequest struct {
	PostID   int    `json:"postId"`
	ParentID *int   `json:"parentId,omitempty"`
	Content  string `json:"content"`
}

// LikeRequest represents the like/dislike request payload
type LikeRequest struct {
	PostID    *int `json:"postId,omitempty"`
	CommentID *int `json:"commentId,omitempty"`
	IsLike    bool `json:"isLike"`
}

// MessageRequest represents the message sending request payload
type MessageRequest struct {
	ReceiverID string `json:"receiverId"`
	Content    string `json:"content"`
}

// ConversationRequest represents the conversation creation request payload
type ConversationRequest struct {
	User1ID string `json:"user1Id"`
	User2ID string `json:"user2Id"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NotificationData represents notification data for WebSocket
type NotificationData struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	UserID  string `json:"userId,omitempty"`
	PostID  *int   `json:"postId,omitempty"`
}

// UserStatusData represents user online/offline status for WebSocket
type UserStatusData struct {
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"` // "online" or "offline"
}

// MessageData represents message data for WebSocket
type MessageData struct {
	Message *Message `json:"message"`
	Type    string   `json:"type"` // "new_message", "message_read", etc.
}

// TypingIndicatorData represents typing indicator data
type TypingIndicatorData struct {
	UserID     string `json:"userId"`
	ReceiverID string `json:"receiverId"`
	IsTyping   bool   `json:"isTyping"`
}
