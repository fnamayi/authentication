package db

import (
	"database/sql"
	"log"
)

// MigrateDB applies all necessary database migrations
func MigrateDB(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Check if oauth_provider column already exists
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='oauth_provider'`).Scan(&count)
	if err != nil {
		log.Printf("Error checking if oauth_provider column exists: %v", err)
		return err
	}

	// Only add OAuth fields if they don't exist
	if count == 0 {
		_, err := db.Exec(`
			ALTER TABLE users ADD COLUMN oauth_provider TEXT DEFAULT NULL;
			ALTER TABLE users ADD COLUMN oauth_id TEXT DEFAULT NULL;
		`)
		if err != nil {
			log.Printf("Error adding OAuth fields to users table: %v", err)
			return err
		}
	}

	// Create index (IF NOT EXISTS ensures it won't error if already exists)
	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id)
		WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL;
	`)
	if err != nil {
		log.Printf("Error creating OAuth index: %v", err)
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}