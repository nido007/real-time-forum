package database

import (
	"database/sql"
	"log"
)

// AddRealtimeFeatures creates tables for real-time functionality
// This adds support for private messaging and online/offline status tracking
func AddRealtimeFeatures(db *sql.DB) error {
	log.Println("🔄 Adding real-time features to database...")

	// Table for private messages between users
	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER NOT NULL,
		receiver_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_read BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
	)`

	_, err := db.Exec(messagesTable)
	if err != nil {
		return err
	}
	log.Println("✅ Messages table created/verified")

	// Table for tracking online/offline status
	userStatusTable := `
	CREATE TABLE IF NOT EXISTS user_status (
		user_id INTEGER PRIMARY KEY,
		is_online BOOLEAN DEFAULT FALSE,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		websocket_id TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`

	_, err = db.Exec(userStatusTable)
	if err != nil {
		return err
	}
	log.Println("✅ User status table created/verified")

	// Create indexes for better query performance
	log.Println("📊 Creating indexes for real-time tables...")

	indexes := []string{
		// Speed up message queries
		"CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_receiver ON messages(receiver_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(receiver_id, is_read)",

		// Speed up user status queries
		"CREATE INDEX IF NOT EXISTS idx_user_status_online ON user_status(is_online)",
		"CREATE INDEX IF NOT EXISTS idx_user_status_websocket ON user_status(websocket_id)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			log.Printf("⚠️  Warning: Failed to create index: %v", err)
		}
	}

	log.Println("🎉 Real-time tables ready!")
	return nil
}
