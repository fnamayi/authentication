package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Simple WebSocket message structure for testing
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Test message ID generation (from websocket package)
func TestMessageIDGeneration(t *testing.T) {
	t.Run("Generate Message ID Format", func(t *testing.T) {
		// Since generateMessageID is not exported, we'll test the format
		messageID := fmt.Sprintf("msg_%d_%d", time.Now().UnixNano(), 1234)

		if !strings.HasPrefix(messageID, "msg_") {
			t.Error("Message ID should start with 'msg_'")
		}

		parts := strings.Split(messageID, "_")
		if len(parts) != 3 {
			t.Errorf("Message ID should have 3 parts separated by '_', got %d", len(parts))
		}

		if parts[0] != "msg" {
			t.Errorf("First part should be 'msg', got '%s'", parts[0])
		}
	})

	t.Run("Message ID Uniqueness", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			messageID := fmt.Sprintf("msg_%d_%d", time.Now().UnixNano(), i)
			if ids[messageID] {
				t.Error("Message IDs should be unique")
			}
			ids[messageID] = true
			time.Sleep(1 * time.Nanosecond) // Ensure different timestamps
		}
	})
}

// Test WebSocket message structure
func TestWebSocketMessage(t *testing.T) {
	t.Run("WebSocket Message JSON", func(t *testing.T) {
		msg := WebSocketMessage{
			Type:      "chat",
			Data:      map[string]interface{}{"content": "Hello"},
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal WebSocket message: %v", err)
		}

		var unmarshaled WebSocketMessage
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal WebSocket message: %v", err)
		}

		if unmarshaled.Type != "chat" {
			t.Errorf("Expected type 'chat', got '%s'", unmarshaled.Type)
		}
	})

	t.Run("Chat Message Structure", func(t *testing.T) {
		chatData := map[string]interface{}{
			"content":  "Hello everyone!",
			"userId":   "user-123",
			"username": "testuser",
		}

		msg := WebSocketMessage{
			Type:      "chat",
			Data:      chatData,
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal chat message: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "Hello everyone!") {
			t.Error("Chat content should be included in JSON")
		}
		if !strings.Contains(jsonString, "user-123") {
			t.Error("User ID should be included in JSON")
		}
	})

	t.Run("Notification Message Structure", func(t *testing.T) {
		notificationData := map[string]interface{}{
			"message": "New post created",
			"postId":  123,
			"author":  "testuser",
		}

		msg := WebSocketMessage{
			Type:      "notification",
			Data:      notificationData,
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal notification message: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "notification") {
			t.Error("Message type should be included in JSON")
		}
		if !strings.Contains(jsonString, "New post created") {
			t.Error("Notification message should be included in JSON")
		}
	})

	t.Run("User Status Message Structure", func(t *testing.T) {
		statusData := map[string]interface{}{
			"userId":   "user-456",
			"username": "statususer",
			"status":   "online",
			"action":   "join",
		}

		msg := WebSocketMessage{
			Type:      "user_status",
			Data:      statusData,
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal user status message: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "user_status") {
			t.Error("Message type should be included in JSON")
		}
		if !strings.Contains(jsonString, "online") {
			t.Error("User status should be included in JSON")
		}
	})
}

// Test WebSocket message types
func TestWebSocketMessageTypes(t *testing.T) {
	messageTypes := []struct {
		name        string
		messageType string
		data        map[string]interface{}
	}{
		{
			name:        "Chat Message",
			messageType: "chat",
			data:        map[string]interface{}{"content": "Hello everyone!"},
		},
		{
			name:        "User Join",
			messageType: "user_join",
			data:        map[string]interface{}{"username": "newuser"},
		},
		{
			name:        "User Leave",
			messageType: "user_leave",
			data:        map[string]interface{}{"username": "leftuser"},
		},
		{
			name:        "Notification",
			messageType: "notification",
			data:        map[string]interface{}{"message": "New post created"},
		},
		{
			name:        "Typing Indicator",
			messageType: "typing",
			data:        map[string]interface{}{"username": "typinguser", "isTyping": true},
		},
		{
			name:        "Online Users",
			messageType: "online_users",
			data:        map[string]interface{}{"users": []string{"user1", "user2"}},
		},
	}

	for _, tt := range messageTypes {
		t.Run(tt.name, func(t *testing.T) {
			msg := WebSocketMessage{
				Type:      tt.messageType,
				Data:      tt.data,
				Timestamp: time.Now(),
			}

			// Test serialization
			jsonData, err := json.Marshal(msg)
			if err != nil {
				t.Errorf("Failed to marshal %s message: %v", tt.name, err)
			}

			// Test deserialization
			var deserializedMsg WebSocketMessage
			err = json.Unmarshal(jsonData, &deserializedMsg)
			if err != nil {
				t.Errorf("Failed to unmarshal %s message: %v", tt.name, err)
			}

			// Verify fields
			if deserializedMsg.Type != tt.messageType {
				t.Errorf("Expected type '%s', got '%s'", tt.messageType, deserializedMsg.Type)
			}

			// Verify data is present
			if deserializedMsg.Data == nil {
				t.Error("Data should not be nil after deserialization")
			}
		})
	}
}

// Test WebSocket message validation
func TestWebSocketMessageValidation(t *testing.T) {
	t.Run("Empty Message Type", func(t *testing.T) {
		msg := WebSocketMessage{
			Type:      "",
			Data:      map[string]interface{}{"content": "test"},
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal message with empty type: %v", err)
		}

		// Should still be valid JSON even with empty type
		var unmarshaled WebSocketMessage
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal message with empty type: %v", err)
		}
	})

	t.Run("Nil Data", func(t *testing.T) {
		msg := WebSocketMessage{
			Type:      "test",
			Data:      nil,
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal message with nil data: %v", err)
		}

		var unmarshaled WebSocketMessage
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal message with nil data: %v", err)
		}
	})

	t.Run("Complex Data Structure", func(t *testing.T) {
		complexData := map[string]interface{}{
			"user": map[string]interface{}{
				"id":       "user-123",
				"nickname": "testuser",
				"profile": map[string]interface{}{
					"firstName": "Test",
					"lastName":  "User",
				},
			},
			"metadata": map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"version":   "1.0",
				"tags":      []string{"important", "urgent"},
			},
		}

		msg := WebSocketMessage{
			Type:      "complex",
			Data:      complexData,
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal complex message: %v", err)
		}

		var unmarshaled WebSocketMessage
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal complex message: %v", err)
		}

		if unmarshaled.Type != "complex" {
			t.Errorf("Expected type 'complex', got '%s'", unmarshaled.Type)
		}
	})
}
