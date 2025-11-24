package database

import (
	"database/sql"
	"fmt"
	"log"

	// SQLite driver - required for database/sql to work with SQLite
	_ "github.com/mattn/go-sqlite3"
)

// DB is the global database connection that other packages can use
// We make it a package-level variable so it's accessible throughout the application
var DB *sql.DB

// Initialize sets up the database connection and creates all required tables
// This function should be called once when the application starts
func Initialize() (*sql.DB, error) {
	log.Println("🗄️  Initializing database connection...")

	// Open database connection to forum.db file
	// If the file doesn't exist, SQLite will create it automatically
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test that we can actually connect to the database
	// Ping() verifies the connection is working
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Store in global variable so other packages can access it
	DB = db

	// Enable foreign key constraints in SQLite
	// SQLite has foreign keys disabled by default, so we need to enable them
	if err := enableForeignKeys(db); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create all the tables we need for our forum
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Insert default categories (General, Technology, etc.)
	if err := insertDefaultData(db); err != nil {
		return nil, fmt.Errorf("failed to insert default data: %w", err)
	}

	// Add real-time features (messages, user_status tables)
	if err := AddRealtimeFeatures(db); err != nil {
		log.Println("⚠️  Warning: Could not add real-time features:", err)
	}

	log.Println("✅ Database initialized successfully!")
	return db, nil
}

// enableForeignKeys turns on foreign key constraint checking
// This is essential for maintaining data integrity in our forum
func enableForeignKeys(db *sql.DB) error {
	log.Println("🔗 Enabling foreign key constraints...")

	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	log.Println("✅ Foreign key constraints enabled")
	return nil
}

// createTables creates all database tables with proper relationships and constraints
// Each table serves a specific purpose in our forum system
func createTables(db *sql.DB) error {
	log.Println("🗃️  Creating database tables...")

	// Define all table creation SQL statements
	// Each table has specific columns, constraints, and relationships
	tables := []struct {
		name string
		sql  string
	}{
		{
			name: "users",
			sql: `CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT UNIQUE NOT NULL CHECK(length(username) >= 3),
				email TEXT UNIQUE NOT NULL CHECK(email LIKE '%@%.%'),
				password_hash TEXT NOT NULL CHECK(length(password_hash) > 0),
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			name: "sessions",
			sql: `CREATE TABLE IF NOT EXISTS sessions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				token TEXT UNIQUE NOT NULL CHECK(length(token) >= 32),
				expires_at DATETIME NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`,
		},
		{
			name: "categories",
			sql: `CREATE TABLE IF NOT EXISTS categories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT UNIQUE NOT NULL CHECK(length(name) >= 2),
				description TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		{
			name: "posts",
			sql: `CREATE TABLE IF NOT EXISTS posts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				title TEXT NOT NULL CHECK(length(title) >= 3 AND length(title) <= 200),
				content TEXT NOT NULL CHECK(length(content) >= 10),
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`,
		},
		{
			name: "post_categories",
			sql: `CREATE TABLE IF NOT EXISTS post_categories (
				post_id INTEGER NOT NULL,
				category_id INTEGER NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (post_id, category_id),
				FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
				FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
			)`,
		},
		{
			name: "comments",
			sql: `CREATE TABLE IF NOT EXISTS comments (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				post_id INTEGER NOT NULL,
				user_id INTEGER NOT NULL,
				content TEXT NOT NULL CHECK(length(content) >= 1),
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			)`,
		},
		{
			name: "likes",
			sql: `CREATE TABLE IF NOT EXISTS likes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				post_id INTEGER,
				comment_id INTEGER,
				is_like BOOLEAN NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
				FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
				FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
				-- Ensure user can only vote once per post
				UNIQUE(user_id, post_id),
				-- Ensure user can only vote once per comment
				UNIQUE(user_id, comment_id),
				-- Ensure vote is either on post OR comment, not both
				CHECK ((post_id IS NOT NULL AND comment_id IS NULL) OR 
					   (post_id IS NULL AND comment_id IS NOT NULL))
			)`,
		},
		{
			name: "contact_messages",
			sql: `CREATE TABLE IF NOT EXISTS contact_messages (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL CHECK(length(name) >= 2),
				email TEXT CHECK(email LIKE '%@%.%'),
				message TEXT NOT NULL CHECK(length(message) >= 10),
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
		},
	}

	// Execute each table creation command
	for _, table := range tables {
		log.Printf("📋 Creating table: %s", table.name)

		if _, err := db.Exec(table.sql); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.name, err)
		}
	}

	// Create indexes for better query performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("✅ All tables created successfully!")
	return nil
}

// createIndexes creates database indexes for improved query performance
// Indexes speed up common queries like finding posts by user or sorting by date
func createIndexes(db *sql.DB) error {
	log.Println("📊 Creating database indexes...")

	indexes := []string{
		// Speed up session lookups by token
		"CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)",

		// Speed up post queries
		"CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC)",

		// Speed up comment queries
		"CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at)",

		// Speed up like/dislike queries
		"CREATE INDEX IF NOT EXISTS idx_likes_post_id ON likes(post_id)",
		"CREATE INDEX IF NOT EXISTS idx_likes_comment_id ON likes(comment_id)",
		"CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id)",

		// Speed up category queries
		"CREATE INDEX IF NOT EXISTS idx_post_categories_post_id ON post_categories(post_id)",
		"CREATE INDEX IF NOT EXISTS idx_post_categories_category_id ON post_categories(category_id)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("✅ Database indexes created successfully!")
	return nil
}

// insertDefaultData adds initial categories and any other default data
// This ensures the forum has some basic categories available from the start
func insertDefaultData(db *sql.DB) error {
	log.Println("📝 Inserting default categories...")

	// Check if categories already exist to avoid duplicates
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing categories: %w", err)
	}

	// If categories already exist, skip insertion
	if count > 0 {
		log.Printf("📂 Found %d existing categories, skipping default insertion", count)
		return nil
	}

	// Define default categories for the forum
	defaultCategories := []struct {
		name        string
		description string
	}{
		{"General", "General discussion topics and community chat"},
		{"Technology", "Programming, software, and tech discussions"},
		{"Gaming", "Video games, reviews, and gaming community"},
		{"Sports", "Sports discussions, news, and events"},
		{"Entertainment", "Movies, TV shows, music, and entertainment"},
		{"Science", "Scientific discussions, research, and discoveries"},
		{"Education", "Learning resources, tutorials, and academic topics"},
		{"Travel", "Travel experiences, destinations, and tips"},
		{"Food", "Cooking, recipes, restaurants, and food culture"},
		{"Books", "Book recommendations, reviews, and literary discussions"},
	}

	// Insert each default category
	for _, category := range defaultCategories {
		_, err := db.Exec(`
			INSERT INTO categories (name, description) 
			VALUES (?, ?)
		`, category.name, category.description)

		if err != nil {
			return fmt.Errorf("failed to insert category %s: %w", category.name, err)
		}

		log.Printf("📂 Added category: %s", category.name)
	}

	log.Printf("✅ Successfully inserted %d default categories!", len(defaultCategories))
	return nil
}

// Close safely closes the database connection
// This should be called when the application shuts down
func Close() error {
	if DB != nil {
		log.Println("🔒 Closing database connection...")
		return DB.Close()
	}
	return nil
}

// GetDB returns the global database connection
// Other packages can use this to access the database
func GetDB() *sql.DB {
	return DB
}
