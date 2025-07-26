package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"forum/internal/auth"
	"forum/internal/database"
	"forum/internal/handlers"
	"forum/internal/websocket"
)

// Valid frontend routes for SPA
// var validFrontendRoutes = map[string]bool{
// 	"/":         true,
// 	"/login":    true,
// 	"/register": true,
// 	"/profile":  true,
// 	"/post":     true,
// 	"/posts":    true,
// 	"/messages": true,

// 	"/create-post": true,
// 	"/settings":    true,
// }

func main() {
	// Initialize database
	if err := database.Initialize(); err != nil {
		log.Fatal(" Failed to initialize database:", err)
	}
	defer database.Close()

	// Initialize OAuth providers
	auth.InitializeOAuthProviders()

	// Initialize WebSocket hub
	websocket.InitializeHub()

	// Setup routes
	setupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Display startup information
	fmt.Println("Database initialized successfully")
	fmt.Println("✅WebSocket hub initialized")
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("🌐 Access the forum at: http://localhost:%s\n", port)
	fmt.Println("📊 Database: forum.db")
	fmt.Println("🌐 Frontend: Modern SPA with real-time features")
	fmt.Println("💬 Real-time messaging enabled")
	fmt.Println("👥 Online user tracking active")
	fmt.Println("")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func setupRoutes() {
	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static/"))))

	// API routes
	http.HandleFunc("/api/register", handlers.RegisterHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)
	http.HandleFunc("/api/logout", handlers.LogoutHandler)
	http.HandleFunc("/api/user", handlers.CurrentUserHandler)
	http.HandleFunc("/api/posts", handlers.PostsHandler)
	http.HandleFunc("/api/posts/", handlers.PostHandler)
	// Comment routes - use a custom handler to catch all comment requests
	http.HandleFunc("/api/comment", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔀 Comment router - Method: %s, URL: %s", r.Method, r.URL.Path)
		handlers.CommentHandler(w, r)
	})
	http.HandleFunc("/api/comment/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔀 Comment router with slash - Method: %s, URL: %s", r.Method, r.URL.Path)
		handlers.CommentHandler(w, r)
	})
	http.HandleFunc("/api/comments/", handlers.CommentsHandler)
	// Also handle without trailing slash for better compatibility
	http.HandleFunc("/api/comments", handlers.CommentsHandler)

	// Test endpoint
	http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🧪 Test endpoint called")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "Test endpoint working"}`))
	})
	http.HandleFunc("/api/like", handlers.LikeHandler)

	http.HandleFunc("/api/categories", handlers.CategoriesHandler)
	http.HandleFunc("/api/profile", handlers.ProfileHandler)

	http.HandleFunc("/api/online-users", handlers.OnlineUsersHandler)

	// Messaging endpoints
	http.HandleFunc("/api/conversations", handlers.ConversationsHandler)
	http.HandleFunc("/api/messages", handlers.MessagesHandler)
	http.HandleFunc("/api/messages/send", handlers.SendMessageHandler)
	http.HandleFunc("/api/messages/read", handlers.MarkMessageReadHandler)

	// Avatar upload routes
	http.HandleFunc("/api/upload/avatar", handlers.AvatarUploadHandler)
	http.HandleFunc("/api/profile/avatar", handlers.AvatarUpdateHandler)

	// OAuth routes
	http.HandleFunc("/auth/google/login", handlers.GoogleLoginHandler)
	http.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)
	http.HandleFunc("/auth/github/login", handlers.GithubLoginHandler)
	http.HandleFunc("/auth/github/callback", handlers.GithubCallbackHandler)

	// WebSocket endpoint
	http.HandleFunc("/ws", websocket.HandleWebSocket)

	// SPA handler - serve index.html for all frontend routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip API routes - let them handle their own 404s
		if strings.HasPrefix(path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// For all non-API routes, serve index.html and let frontend router handle routing
		// This allows the frontend to show custom 404 pages
		http.ServeFile(w, r, "frontend/static/index.html")
	})

}
