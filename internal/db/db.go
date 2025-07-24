package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var Globaldb *sql.DB

func InitializeDB() (*sql.DB, error) {
	var err error
	Globaldb, err = sql.Open("sqlite", "./forum.db")
	if err != nil {
		return nil, err
	}

	// Create all tables in a single transaction
	tx, err := Globaldb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// SQL statements for creating tables
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY NOT NULL,
			username TEXT UNIQUE,
			password TEXT,
			email TEXT UNIQUE,
			profile_pic TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,

		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT,
			title TEXT NOT NULL,
			content TEXT,
			image_path TEXT,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`,

		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS post_categories (
			post_id TEXT NOT NULL,
			category_id INTEGER NOT NULL,
			PRIMARY KEY (post_id, category_id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		);`,

		`CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY NOT NULL,
			post_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,

		`CREATE TABLE IF NOT EXISTS likes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			post_id TEXT,
			comment_id TEXT,
			comment_like INTEGER CHECK (comment_like IN (0, 1)), -- 1 for like, 0 for dislike,
			like INTEGER CHECK (like IN (0, 1)), -- 1 for like, 0 for dislike,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (comment_id) REFERENCES comments(id),

			UNIQUE(user_id, post_id, comment_id)
		);
		CREATE INDEX IF NOT EXISTS idx_likes_post_id ON likes(post_id);`,
	}

	// Execute each table creation statement
	for _, table := range tables {
		if _, err := tx.Exec(table); err != nil {
			return nil, fmt.Errorf("error creating table: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	err = InsertDefaultCategories()
	if err != nil {
		return nil, fmt.Errorf("failed to insert default categories: %v", err)
	}

	// Run database migrations
	err = MigrateDB(Globaldb)
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %v", err)
	}

	return Globaldb, nil
}

func InsertDefaultCategories() error {
	categories := []string{
		"General",
		"Entertainment",
		"Health",
		"Technology",
		"Business",
		"Lifestyle",
		"Politics",
	}

	for _, category := range categories {
		_, err := Globaldb.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
		if err != nil {
			return fmt.Errorf("failed to insert category %s: %v", category, err)
		}
	}
	return nil
}
