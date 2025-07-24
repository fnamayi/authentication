package auth

import (
	"database/sql"
	"time"
)

// User represents a forum user
type User struct {
	ID           string
	Username     string
	Email        string
	Password     string
	ProfilePic   sql.NullString
	OAuthProvider sql.NullString
	OAuthID      sql.NullString
}

// Post represents a user's post for profile viewing
type Post struct {
	ID        string
	Title     string
	Content   string
	ImagePath string
	CreatedAt time.Time
	Likes     int
	Dislikes  int
	Comments  int
}
