package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"forum/internal/models"
)

// Test JSON serialization of models
func TestModelSerialization(t *testing.T) {
	t.Run("User JSON Serialization", func(t *testing.T) {
		user := &models.User{
			ID:        "test-id",
			Email:     "test@example.com",
			Nickname:  "testuser",
			Password:  "hashedpassword",
			FirstName: "Test",
			LastName:  "User",
			Age:       25,
			Gender:    "male",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		jsonData, err := json.Marshal(user)
		if err != nil {
			t.Errorf("Failed to marshal user to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "test@example.com") {
			t.Error("Email should be included in JSON serialization")
		}
		if !strings.Contains(jsonString, "testuser") {
			t.Error("Nickname should be included in JSON serialization")
		}
	})

	t.Run("Post JSON Serialization", func(t *testing.T) {
		post := &models.Post{
			ID:         1,
			UserID:     "user-id",
			Title:      "Test Post",
			Content:    "This is a test post content",
			Categories: []string{"general", "test"},
			Author:     "testuser",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		jsonData, err := json.Marshal(post)
		if err != nil {
			t.Errorf("Failed to marshal post to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "Test Post") {
			t.Error("Title should be included in JSON serialization")
		}
		if !strings.Contains(jsonString, "general") {
			t.Error("Categories should be included in JSON serialization")
		}
	})

	t.Run("Comment JSON Serialization", func(t *testing.T) {
		comment := &models.Comment{
			ID:        1,
			PostID:    1,
			UserID:    "user-id",
			Content:   "This is a test comment",
			Author:    "testuser",
			CreatedAt: time.Now(),
		}

		jsonData, err := json.Marshal(comment)
		if err != nil {
			t.Errorf("Failed to marshal comment to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "This is a test comment") {
			t.Error("Content should be included in JSON serialization")
		}
		if !strings.Contains(jsonString, "testuser") {
			t.Error("Author should be included in JSON serialization")
		}
	})

	t.Run("Message JSON Serialization", func(t *testing.T) {
		message := &models.Message{
			ID:         "msg_1",
			SenderID:   "sender-id",
			ReceiverID: "receiver-id",
			Content:    "Hello, this is a test message",
			CreatedAt:  time.Now(),
		}

		jsonData, err := json.Marshal(message)
		if err != nil {
			t.Errorf("Failed to marshal message to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "Hello, this is a test message") {
			t.Error("Content should be included in JSON serialization")
		}
		if !strings.Contains(jsonString, "sender-id") {
			t.Error("SenderID should be included in JSON serialization")
		}
	})
}

// Test API response models
func TestAPIResponseModels(t *testing.T) {
	t.Run("APIResponse Success", func(t *testing.T) {
		response := models.APIResponse{
			Success: true,
			Message: "Operation successful",
			Data:    map[string]string{"key": "value"},
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Errorf("Failed to marshal APIResponse to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "true") {
			t.Error("Success should be true in JSON")
		}
		if !strings.Contains(jsonString, "Operation successful") {
			t.Error("Message should be included in JSON")
		}
	})

	t.Run("APIResponse Error", func(t *testing.T) {
		response := models.APIResponse{
			Success: false,
			Error:   "Something went wrong",
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Errorf("Failed to marshal APIResponse to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "false") {
			t.Error("Success should be false in JSON")
		}
		if !strings.Contains(jsonString, "Something went wrong") {
			t.Error("Error should be included in JSON")
		}
	})
}

// Test request model validation
func TestRequestValidation(t *testing.T) {
	t.Run("RegisterRequest JSON Deserialization", func(t *testing.T) {
		jsonData := `{
			"email": "test@example.com",
			"nickname": "testuser",
			"password": "password123",
			"firstName": "Test",
			"lastName": "User",
			"age": 25,
			"gender": "male"
		}`

		var req models.RegisterRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		if err != nil {
			t.Errorf("Failed to unmarshal RegisterRequest from JSON: %v", err)
		}

		if req.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", req.Email)
		}
		if req.Age != 25 {
			t.Errorf("Expected age 25, got %d", req.Age)
		}
		if req.Gender != "male" {
			t.Errorf("Expected gender 'male', got '%s'", req.Gender)
		}
	})

	t.Run("LoginRequest JSON Deserialization", func(t *testing.T) {
		jsonData := `{
			"identifier": "test@example.com",
			"password": "password123"
		}`

		var req models.LoginRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		if err != nil {
			t.Errorf("Failed to unmarshal LoginRequest from JSON: %v", err)
		}

		if req.Identifier != "test@example.com" {
			t.Errorf("Expected identifier 'test@example.com', got '%s'", req.Identifier)
		}
		if req.Password != "password123" {
			t.Errorf("Expected password 'password123', got '%s'", req.Password)
		}
	})

	t.Run("PostRequest JSON Deserialization", func(t *testing.T) {
		jsonData := `{
			"title": "Test Post",
			"content": "This is test content",
			"categories": ["general", "test"]
		}`

		var req models.PostRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		if err != nil {
			t.Errorf("Failed to unmarshal PostRequest from JSON: %v", err)
		}

		if req.Title != "Test Post" {
			t.Errorf("Expected title 'Test Post', got '%s'", req.Title)
		}
		if len(req.Categories) != 2 {
			t.Errorf("Expected 2 categories, got %d", len(req.Categories))
		}
		if req.Categories[0] != "general" {
			t.Errorf("Expected first category 'general', got '%s'", req.Categories[0])
		}
	})
}

// Test conversation and online user models
func TestCommunicationModels(t *testing.T) {
	t.Run("Conversation JSON Serialization", func(t *testing.T) {
		conversation := &models.Conversation{
			ID:                "conv-id",
			User1ID:           "user1-id",
			User2ID:           "user2-id",
			LastMessageTime:   time.Now(),
			OtherUserNickname: "testuser",
			LastMessage:       "Hello there!",
			UnreadCount:       3,
			IsOnline:          true,
		}

		jsonData, err := json.Marshal(conversation)
		if err != nil {
			t.Errorf("Failed to marshal conversation to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "Hello there!") {
			t.Error("LastMessage should be included in JSON")
		}
		if !strings.Contains(jsonString, "testuser") {
			t.Error("Nickname should be included in JSON")
		}
	})

	t.Run("OnlineUser JSON Serialization", func(t *testing.T) {
		onlineUser := &models.OnlineUser{
			UserID:    "user-id",
			Nickname:  "testuser",
			FirstName: "Test",
			LastName:  "User",
			LastSeen:  time.Now(),
		}

		jsonData, err := json.Marshal(onlineUser)
		if err != nil {
			t.Errorf("Failed to marshal OnlineUser to JSON: %v", err)
		}

		jsonString := string(jsonData)
		if !strings.Contains(jsonString, "testuser") {
			t.Error("Nickname should be included in JSON")
		}
		if !strings.Contains(jsonString, "user-id") {
			t.Error("UserID should be included in JSON")
		}
	})
}
