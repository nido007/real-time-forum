package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forum_git/internal/middleware"
)

type MessagesHandler struct {
	db             *sql.DB
	authMiddleware *middleware.AuthMiddleware
}

func NewMessagesHandler(db *sql.DB, authMiddleware *middleware.AuthMiddleware) *MessagesHandler {
	return &MessagesHandler{
		db:             db,
		authMiddleware: authMiddleware,
	}
}

// Message structure for API responses
type MessageResponse struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	IsRead     bool      `json:"is_read"`
	SenderName string    `json:"sender_name"`
}

// User structure for users list
type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsOnline bool   `json:"is_online"`
}

// GetUsersHandler returns list of all users with online status
func (h *MessagesHandler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query all users except current user
	rows, err := h.db.Query(`
		SELECT 
			u.id, 
			u.username, 
			u.email,
			CASE WHEN s.user_id IS NOT NULL AND s.expires_at > datetime('now') 
				THEN 1 ELSE 0 END as is_online
		FROM users u
		LEFT JOIN (
			SELECT user_id, MAX(expires_at) as expires_at 
			FROM sessions 
			WHERE expires_at > datetime('now')
			GROUP BY user_id
		) s ON u.id = s.user_id
		WHERE u.id != ?
		ORDER BY u.username
	`, currentUser.ID)

	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []UserResponse
	for rows.Next() {
		var user UserResponse
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsOnline)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
	})
}

// GetMessageHistoryHandler returns chat history between two users
func (h *MessagesHandler) GetMessageHistoryHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get other user ID from URL: /api/messages/history/123
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	otherUserID, err := strconv.Atoi(parts[4])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get limit and offset from query params
	limit := 10
	offset := 0
	
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}
	
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Query messages between the two users
	rows, err := h.db.Query(`
		SELECT 
			m.id,
			m.sender_id,
			m.receiver_id,
			m.content,
			m.created_at,
			m.is_read,
			u.username as sender_name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE (m.sender_id = ? AND m.receiver_id = ?) 
		   OR (m.sender_id = ? AND m.receiver_id = ?)
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, currentUser.ID, otherUserID, otherUserID, currentUser.ID, limit, offset)

	if err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []MessageResponse
	for rows.Next() {
		var msg MessageResponse
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, 
			&msg.Content, &msg.CreatedAt, &msg.IsRead, &msg.SenderName)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	// Mark messages as read
	h.markMessagesAsRead(currentUser.ID, otherUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"messages": messages,
	})
}

// SendMessageHandler stores a message in the database
func (h *MessagesHandler) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		ReceiverID int    `json:"receiver_id"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Content == "" {
		http.Error(w, "Message content cannot be empty", http.StatusBadRequest)
		return
	}

	// Insert message into database
	result, err := h.db.Exec(`
		INSERT INTO messages (sender_id, receiver_id, content, created_at, is_read)
		VALUES (?, ?, ?, datetime('now'), 0)
	`, currentUser.ID, request.ReceiverID, request.Content)

	if err != nil {
		http.Error(w, "Error sending message", http.StatusInternalServerError)
		return
	}

	messageID, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": map[string]interface{}{
			"id":          messageID,
			"sender_id":   currentUser.ID,
			"receiver_id": request.ReceiverID,
			"content":     request.Content,
			"created_at":  time.Now(),
		},
	})
}

// GetUnreadCountHandler returns count of unread messages
func (h *MessagesHandler) GetUnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM messages 
		WHERE receiver_id = ? AND is_read = 0
	`, currentUser.ID).Scan(&count)

	if err != nil {
		http.Error(w, "Error fetching unread count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   count,
	})
}

// markMessagesAsRead marks all messages between two users as read
func (h *MessagesHandler) markMessagesAsRead(currentUserID, otherUserID int) {
	h.db.Exec(`
		UPDATE messages SET is_read = 1 
		WHERE sender_id = ? AND receiver_id = ? AND is_read = 0
	`, otherUserID, currentUserID)
}
