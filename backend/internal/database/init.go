package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the global database connection that other packages can use
var DB *sql.DB

// Initialize sets up the database connection and creates all required tables
func Initialize() (*sql.DB, error) {
	log.Println("📊 Initializing database connection...")

	// Open database connection to forum.db file
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test that we can actually connect to the database
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Store in global variable so other packages can access it
	DB = db

	// Enable foreign key constraints in SQLite
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create all tables and indexes
	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Println("✅ Database initialized successfully")
	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			age INTEGER,
			gender TEXT,
			first_name TEXT,
			last_name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Sessions table
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Categories table
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Posts table
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Post categories (many-to-many)
		`CREATE TABLE IF NOT EXISTS post_categories (
			post_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (post_id, category_id),
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)`,

		// Comments table
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Votes table
		`CREATE TABLE IF NOT EXISTS votes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			post_id INTEGER,
			comment_id INTEGER,
			vote_type INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
			UNIQUE(user_id, post_id, comment_id)
		)`,

		// Messages table for private chat
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender_id INTEGER NOT NULL,
			receiver_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			is_read BOOLEAN DEFAULT 0,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Create indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_votes_user_id ON votes(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_votes_post_id ON votes(post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_votes_comment_id ON votes(comment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)`,
	}

	// Execute all queries
	for i, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("❌ Error executing query %d: %v", i+1, err)
			log.Printf("Query was: %s", query)
			return fmt.Errorf("query %d failed: %w", i+1, err)
		}
	}

	// Insert default categories if they don't exist
	categories := []string{"Technology", "Gaming", "Sports", "General"}
	for _, category := range categories {
		_, err := db.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
		if err != nil {
			log.Printf("⚠️ Error inserting category %s: %v", category, err)
		}
	}

	log.Println("✅ All tables and indexes created successfully")
	return nil
}
