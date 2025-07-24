package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"forum/internal/auth"
	"forum/internal/comments"
	"forum/internal/db"
	"forum/internal/filters"
	"forum/internal/likes"
	"forum/internal/posts"
	"forum/internal/xerrors"
)

// Global variables
var (
	templates *template.Template
	database  *sql.DB
)

// serveHome handles requests to "/"
func ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		IsLoggedIn bool
		User       *auth.User
		Posts      []posts.Post
	}{}

	// Check if user is logged in
	userID, isLoggedIn := auth.GetUserIDFromSession(r)
	data.IsLoggedIn = isLoggedIn

	if isLoggedIn {
		// Get user data
		user, err := auth.GetUserByID(database, userID)
		if err != nil {
			log.Printf("Error fetching user: %v", err)
		} else if user == nil {
			// Handle "not found" case
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else {
			data.User = user
		}
	}

	// Get posts
	postService := posts.NewPostService(database)
	posts, err := postService.GetAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		xerrors.RenderErrorPage(w, http.StatusNotFound, xerrors.ErrPageNotFound)

		return
	}
	if len(posts) == 0 {
		log.Println("No posts found.")
	} else {
		data.Posts = posts
	}

	log.Printf("Template data: IsLoggedIn=%v, User=%+v", data.IsLoggedIn, data.User)

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// middleware for logging requests
func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		log.Printf("Cookies: %v", r.Cookies())
		handler(w, r)
	}
}

func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, isLoggedIn := auth.GetUserIDFromSession(r)

		// Check if user is logged in
		if !isLoggedIn || userID == "" {
			// Prevent redirect loop by checking if we're already on the login page
			if r.URL.Path == "/login" {
				// Already on login page, just process it normally
				handler(w, r)
				return
			}

			// Set a cookie to indicate where to redirect after login
			returnCookie := &http.Cookie{
				Name:     "return_to",
				Value:    r.URL.Path,
				Path:     "/",
				MaxAge:   300, // 5 minutes expiration
				HttpOnly: true,
			}
			http.SetCookie(w, returnCookie)

			// Redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// User is authenticated, proceed with the handler
		handler(w, r)
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting the Forum App...")

	// Initialize templates
	var err error
	templatePath := filepath.Join("internal", "web", "templates", "*.html")
	templates, err = template.ParseGlob(templatePath)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Ensure the correct number of arguments
	if len(os.Args) != 1 {
		fmt.Println("Invalid number of arguments.")
		fmt.Println("Usage: go run main.go")
		return
	}

	// Initialize the database
	database, err = db.InitializeDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize services
	postService := posts.NewPostService(database)
	postHandler := posts.NewPostHandler(postService, templates)
	authHandler := auth.NewAuthHandler(database, templates)
	commentHandler := comments.NewCommentHandler(database)
	
	// Initialize OAuth handler
	oauthHandler := auth.NewOAuthHandler(database, authHandler)
	
	// Register OAuth routes with middleware
	http.HandleFunc("/auth/google/login", logRequest(oauthHandler.GoogleLoginHandler))
	http.HandleFunc("/auth/google/callback", logRequest(oauthHandler.GoogleCallbackHandler))
	http.HandleFunc("/auth/github/login", logRequest(oauthHandler.GitHubLoginHandler))
	http.HandleFunc("/auth/github/callback", logRequest(oauthHandler.GitHubCallbackHandler))
	
	// Initialize OAuth configuration
	auth.InitOAuthConfig("http://localhost:3000")

	// Register routes
	http.HandleFunc("/", logRequest(ServeHome))
	http.HandleFunc("/login", logRequest(authHandler.LoginHandler))
	http.HandleFunc("/register", logRequest(authHandler.RegisterHandler))
	http.HandleFunc("/logout", logRequest(authHandler.LogoutHandler))
	http.HandleFunc("/create", requireAuth(logRequest(postHandler.CreatePostHandler)))
	http.HandleFunc("/posts", requireAuth(logRequest(postHandler.GetAllPostsHandler)))
	http.HandleFunc("/profile/", requireAuth(logRequest(authHandler.ProfileHandler))) // Note the trailing slash
	http.HandleFunc("/comments", requireAuth(commentHandler.CreateComment))
	http.HandleFunc("/comments/post", requireAuth(commentHandler.GetCommentsForPost))
	http.HandleFunc("/editcomment", requireAuth(comments.HandleEditComment))
	http.HandleFunc("/deletecomment", requireAuth(comments.HandleDeleteComment))
	http.HandleFunc("/commentreact", requireAuth(comments.HandleCommentReactions))

	http.HandleFunc("/react", requireAuth(likes.HandleReactions))
	http.HandleFunc("/created", requireAuth(filters.CreatedPosts))
	http.HandleFunc("/liked", requireAuth(filters.LikedPosts))

	categoryHandler := filters.NewCategoryHandler()
	http.Handle("/categories", categoryHandler)
	http.Handle("/category", categoryHandler)

	// Serve static files (CSS, JS, images)
	http.Handle("/static/", http.StripPrefix("/static/", auth.SecureStaticHandler()))

	// Create uploads directory
	if err := os.MkdirAll("internal/web/static/uploads", 0o755); err != nil {
		log.Printf("Warning: Failed to create uploads directory: %v", err)
	}

	// Start the server
	log.Printf("Server starting on http://localhost:3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
