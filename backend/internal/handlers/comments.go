package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"real-time-forum/internal/middleware"
)

// CommentsHandler handles all comment-related HTTP requests
type CommentsHandler struct {
	db             *sql.DB
	authMiddleware *middleware.AuthMiddleware
}

// NewCommentsHandler creates a new comments handler
func NewCommentsHandler(db *sql.DB, authMiddleware *middleware.AuthMiddleware) *CommentsHandler {
	return &CommentsHandler{
		db:             db,
		authMiddleware: authMiddleware,
	}
}

// CreateCommentRequest represents the JSON payload for creating a comment
type CreateCommentRequest struct {
	PostID  int    `json:"post_id"`
	Content string `json:"content"`
}

// CreateCommentHandler handles comment creation via JSON
func (h *CommentsHandler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		h.respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate input
	if req.PostID == 0 || req.Content == "" {
		h.respondWithError(w, http.StatusBadRequest, "Post ID and content are required")
		return
	}

	if len(req.Content) < 1 {
		h.respondWithError(w, http.StatusBadRequest, "Comment content cannot be empty")
		return
	}

	// Verify post exists
	if !h.postExists(req.PostID) {
		h.respondWithError(w, http.StatusNotFound, "Post not found")
		return
	}

	// Create comment
	commentID, err := h.createComment(req.PostID, currentUser.ID, req.Content)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error creating comment")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "Comment created successfully",
		"comment_id": commentID,
	})
}

// createComment creates a new comment in the database
func (h *CommentsHandler) createComment(postID, userID int, content string) (int64, error) {
	result, err := h.db.Exec(`
		INSERT INTO comments (post_id, user_id, content) 
		VALUES (?, ?, ?)
	`, postID, userID, content)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// postExists checks if a post with the given ID exists
func (h *CommentsHandler) postExists(postID int) bool {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", postID).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func (h *CommentsHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *CommentsHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
