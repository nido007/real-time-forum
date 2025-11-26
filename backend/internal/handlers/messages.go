package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"real-time-forum/internal/middleware"
	"real-time-forum/internal/websocket"
)

type MessagesHandler struct {
	db             *sql.DB
	hub            *websocket.Hub
	authMiddleware *middleware.AuthMiddleware
}

type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	IsRead     bool      `json:"is_read"`
}

type SendMessageRequest struct {
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
}

func NewMessagesHandler(db *sql.DB, hub *websocket.Hub, authMiddleware *middleware.AuthMiddleware) *MessagesHandler {
	return &MessagesHandler{
		db:             db,
		hub:            hub,
		authMiddleware: authMiddleware,
	}
}

// SendMessage handles sending a private message
func (h *MessagesHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.Content == "" {
		http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}

	if req.ReceiverID == currentUser.ID {
		http.Error(w, "Cannot send message to yourself", http.StatusBadRequest)
		return
	}

	// Save message to database
	query := `
		INSERT INTO messages (sender_id, receiver_id, content, created_at, is_read)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := h.db.Exec(query, currentUser.ID, req.ReceiverID, req.Content, time.Now(), false)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	messageID, _ := result.LastInsertId()

	// Create message object
	message := Message{
		ID:         int(messageID),
		SenderID:   currentUser.ID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		CreatedAt:  time.Now(),
		IsRead:     false,
	}

	// Send via WebSocket if receiver is online
	err = h.hub.SendToUser(req.ReceiverID, map[string]interface{}{
		"type":    "new_message",
		"message": message,
	})
	if err != nil {
		log.Printf("Error sending WebSocket message: %v", err)
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": message,
	})
}

// GetMessageHistory retrieves message history between two users
func (h *MessagesHandler) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get other user ID from query params
	otherUserIDStr := r.URL.Query().Get("user_id")
	if otherUserIDStr == "" {
		http.Error(w, "user_id parameter required", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	// Get limit (default 50)
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Fetch messages from database
	query := `
		SELECT id, sender_id, receiver_id, content, created_at, is_read
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := h.db.Query(query, currentUser.ID, otherUserID, otherUserID, currentUser.ID, limit)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.CreatedAt, &msg.IsRead)
		if err != nil {
			log.Printf("Error scanning message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// Mark messages as read
	updateQuery := `
		UPDATE messages 
		SET is_read = 1 
		WHERE sender_id = ? AND receiver_id = ? AND is_read = 0
	`
	_, err = h.db.Exec(updateQuery, otherUserID, currentUser.ID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
	}

	// Return messages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

// GetOnlineUsers returns a list of currently online users
func (h *MessagesHandler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get online user IDs from hub
	onlineUserIDs := h.hub.GetOnlineUserIDs()

	// Fetch user details from database
	if len(onlineUserIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"users":   []interface{}{},
		})
		return
	}

	// Build query with placeholders
	placeholders := ""
	args := []interface{}{}
	for i, id := range onlineUserIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}

	query := `
		SELECT id, username, email, created_at
		FROM users
		WHERE id IN (` + placeholders + `)
	`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("Error fetching online users: %v", err)
		http.Error(w, "Failed to fetch online users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var username, email string
		var createdAt time.Time

		err := rows.Scan(&id, &username, &email, &createdAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}

		// Don't include current user in the list
		if id == currentUser.ID {
			continue
		}

		users = append(users, map[string]interface{}{
			"id":         id,
			"username":   username,
			"email":      email,
			"created_at": createdAt,
		})
	}

	// Return users
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
	})
}
