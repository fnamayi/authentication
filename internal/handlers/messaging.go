package handlers

import (
	"database/sql"
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
	"forum/internal/websocket"

	"github.com/google/uuid"
)

// ConversationsHandler handles fetching user conversations
func ConversationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	conversations, err := getUserConversations(user.ID)
	if err != nil {
		log.Printf("Error fetching conversations for user %s: %v", user.ID, err)
		RenderError(w, "Failed to fetch conversations", http.StatusInternalServerError)
		return
	}

	RenderSuccess(w, "Conversations retrieved successfully", conversations)
}

// MessagesHandler handles fetching messages for a conversation
func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get the other user ID from query parameters
	otherUserID := r.URL.Query().Get("user")
	if otherUserID == "" {
		RenderError(w, "User parameter is required", http.StatusBadRequest)
		return
	}

	// Get limit and offset for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	log.Printf("📥 Fetching messages: user=%s, otherUser=%s, limit=%d, offset=%d", user.ID, otherUserID, limit, offset)

	messages, err := getConversationMessages(user.ID, otherUserID, limit, offset)
	if err != nil {
		log.Printf("Error fetching messages between %s and %s: %v", user.ID, otherUserID, err)
		RenderError(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	log.Printf("📤 Returning %d messages for conversation", len(messages))

	// Mark messages as read
	err = markMessagesAsRead(user.ID, otherUserID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
		// Don't fail the request, just log the error
	}

	RenderSuccess(w, "Messages retrieved successfully", messages)
}

// SendMessageHandler handles sending a new message
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.ReceiverID == "" || req.Content == "" {
		RenderError(w, "Receiver ID and content are required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		RenderError(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if receiver exists
	var receiverExists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", req.ReceiverID).Scan(&receiverExists)
	if err != nil {
		log.Printf("Error checking receiver existence: %v", err)
		RenderError(w, "Failed to validate receiver", http.StatusInternalServerError)
		return
	}

	if !receiverExists {
		RenderError(w, "Receiver not found", http.StatusNotFound)
		return
	}

	// Create message
	message := &models.Message{
		ID:         uuid.New().String(),
		SenderID:   user.ID,
		ReceiverID: req.ReceiverID,
		Content:    strings.TrimSpace(req.Content),
		IsRead:     false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = createMessage(message)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		RenderError(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	// Create or update conversation
	err = createOrUpdateConversation(user.ID, req.ReceiverID, message.ID)
	if err != nil {
		log.Printf("Error updating conversation: %v", err)
		// Don't fail the request, just log the error
	}

	// Get the complete message with sender info for response
	completeMessage, err := getMessageWithUserInfo(message.ID)
	if err != nil {
		log.Printf("Error fetching complete message: %v", err)
		// Return the basic message if we can't get the complete one
		completeMessage = message
	}

	// Broadcast the new message to the receiver via WebSocket
	websocket.BroadcastNewMessage(completeMessage)

	RenderSuccess(w, "Message sent successfully", completeMessage)
}

// MarkMessageReadHandler handles marking messages as read
func MarkMessageReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUserFromSession(r)
	if user == nil {
		RenderError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get the sender ID from request body
	var req struct {
		SenderID string `json:"senderId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SenderID == "" {
		RenderError(w, "Sender ID is required", http.StatusBadRequest)
		return
	}

	err := markMessagesAsRead(user.ID, req.SenderID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
		RenderError(w, "Failed to mark messages as read", http.StatusInternalServerError)
		return
	}

	RenderSuccess(w, "Messages marked as read", nil)
}

// Helper functions

// getUserConversations retrieves all conversations for a user
func getUserConversations(userID string) ([]models.Conversation, error) {
	query := `
		SELECT DISTINCT
			c.id,
			c.user1_id,
			c.user2_id,
			c.last_message_id,
			c.last_message_time,
			c.created_at,
			c.updated_at,
			CASE 
				WHEN c.user1_id = ? THEN u2.id
				ELSE u1.id
			END as other_user_id,
			CASE 
				WHEN c.user1_id = ? THEN u2.nickname
				ELSE u1.nickname
			END as other_user_nickname,
			CASE 
				WHEN c.user1_id = ? THEN u2.avatar_url
				ELSE u1.avatar_url
			END as other_user_avatar,
			COALESCE(m.content, '') as last_message,
			COALESCE(unread.count, 0) as unread_count,
			CASE WHEN ou.user_id IS NOT NULL THEN 1 ELSE 0 END as is_online
		FROM conversations c
		JOIN users u1 ON c.user1_id = u1.id
		JOIN users u2 ON c.user2_id = u2.id
		LEFT JOIN messages m ON c.last_message_id = m.id
		LEFT JOIN (
			SELECT receiver_id, COUNT(*) as count
			FROM messages
			WHERE receiver_id = ? AND is_read = 0
			GROUP BY receiver_id, sender_id
		) unread ON (
			(c.user1_id = ? AND unread.receiver_id = c.user1_id) OR
			(c.user2_id = ? AND unread.receiver_id = c.user2_id)
		)
		LEFT JOIN online_users ou ON (
			(c.user1_id = ? AND ou.user_id = c.user2_id) OR
			(c.user2_id = ? AND ou.user_id = c.user1_id)
		)
		WHERE c.user1_id = ? OR c.user2_id = ?
		ORDER BY c.last_message_time DESC
	`

	rows, err := database.DB.Query(query, userID, userID, userID, userID, userID, userID, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %v", err)
	}
	defer rows.Close()

	var conversations []models.Conversation
	for rows.Next() {
		var conv models.Conversation
		var lastMessageID sql.NullString
		var otherUserAvatar sql.NullString

		err := rows.Scan(
			&conv.ID,
			&conv.User1ID,
			&conv.User2ID,
			&lastMessageID,
			&conv.LastMessageTime,
			&conv.CreatedAt,
			&conv.UpdatedAt,
			&conv.OtherUserID,
			&conv.OtherUserNickname,
			&otherUserAvatar,
			&conv.LastMessage,
			&conv.UnreadCount,
			&conv.IsOnline,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %v", err)
		}

		if lastMessageID.Valid {
			conv.LastMessageID = &lastMessageID.String
		}
		if otherUserAvatar.Valid {
			conv.OtherUserAvatar = &otherUserAvatar.String
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// getConversationMessages retrieves messages between two users
func getConversationMessages(userID, otherUserID string, limit, offset int) ([]models.Message, error) {
	query := `
		SELECT
			m.id,
			m.sender_id,
			m.receiver_id,
			m.content,
			m.is_read,
			m.created_at,
			m.updated_at,
			sender.nickname as sender_nickname,
			sender.avatar_url as sender_avatar_url,
			receiver.nickname as receiver_nickname
		FROM messages m
		JOIN users sender ON m.sender_id = sender.id
		JOIN users receiver ON m.receiver_id = receiver.id
		WHERE
			(m.sender_id = ? AND m.receiver_id = ?) OR
			(m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := database.DB.Query(query, userID, otherUserID, otherUserID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %v", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var senderAvatarURL, receiverNickname sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.IsRead,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&msg.SenderNickname,
			&senderAvatarURL,
			&receiverNickname,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %v", err)
		}

		if senderAvatarURL.Valid {
			msg.SenderAvatarURL = &senderAvatarURL.String
		}
		if receiverNickname.Valid {
			msg.ReceiverNickname = receiverNickname.String
		}

		messages = append(messages, msg)
	}

	// Reverse the slice to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// createMessage inserts a new message into the database
func createMessage(message *models.Message) error {
	query := `
		INSERT INTO messages (id, sender_id, receiver_id, content, is_read, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := database.DB.Exec(query,
		message.ID,
		message.SenderID,
		message.ReceiverID,
		message.Content,
		message.IsRead,
		message.CreatedAt,
		message.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert message: %v", err)
	}

	return nil
}

// createOrUpdateConversation creates a new conversation or updates existing one
func createOrUpdateConversation(user1ID, user2ID, messageID string) error {
	// Ensure consistent ordering of user IDs for conversation lookup
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	// Check if conversation exists
	var conversationID string
	err := database.DB.QueryRow(
		"SELECT id FROM conversations WHERE user1_id = ? AND user2_id = ?",
		user1ID, user2ID,
	).Scan(&conversationID)

	if err == sql.ErrNoRows {
		// Create new conversation
		conversationID = uuid.New().String()
		_, err = database.DB.Exec(`
			INSERT INTO conversations (id, user1_id, user2_id, last_message_id, last_message_time, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, conversationID, user1ID, user2ID, messageID, time.Now(), time.Now(), time.Now())

		if err != nil {
			return fmt.Errorf("failed to create conversation: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check conversation existence: %v", err)
	} else {
		// Update existing conversation
		_, err = database.DB.Exec(`
			UPDATE conversations
			SET last_message_id = ?, last_message_time = ?, updated_at = ?
			WHERE id = ?
		`, messageID, time.Now(), time.Now(), conversationID)

		if err != nil {
			return fmt.Errorf("failed to update conversation: %v", err)
		}
	}

	return nil
}

// markMessagesAsRead marks all unread messages from a sender as read
func markMessagesAsRead(receiverID, senderID string) error {
	query := `
		UPDATE messages
		SET is_read = 1, updated_at = ?
		WHERE receiver_id = ? AND sender_id = ? AND is_read = 0
	`

	_, err := database.DB.Exec(query, time.Now(), receiverID, senderID)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %v", err)
	}

	return nil
}

// getMessageWithUserInfo retrieves a message with complete user information
func getMessageWithUserInfo(messageID string) (*models.Message, error) {
	query := `
		SELECT
			m.id,
			m.sender_id,
			m.receiver_id,
			m.content,
			m.is_read,
			m.created_at,
			m.updated_at,
			sender.nickname as sender_nickname,
			sender.avatar_url as sender_avatar_url,
			receiver.nickname as receiver_nickname
		FROM messages m
		JOIN users sender ON m.sender_id = sender.id
		JOIN users receiver ON m.receiver_id = receiver.id
		WHERE m.id = ?
	`

	var msg models.Message
	var senderAvatarURL, receiverNickname sql.NullString

	err := database.DB.QueryRow(query, messageID).Scan(
		&msg.ID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Content,
		&msg.IsRead,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&msg.SenderNickname,
		&senderAvatarURL,
		&receiverNickname,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get message with user info: %v", err)
	}

	if senderAvatarURL.Valid {
		msg.SenderAvatarURL = &senderAvatarURL.String
	}
	if receiverNickname.Valid {
		msg.ReceiverNickname = receiverNickname.String
	}

	return &msg, nil
}
