package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// CreateCommentHandler handles comment creation (POST only)
func (h *CommentsHandler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication
	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		h.showError(w, "Invalid form data")
		return
	}

	postIDStr := r.FormValue("post_id")
	content := strings.TrimSpace(r.FormValue("content"))

	// Validate input
	if postIDStr == "" || content == "" {
		h.showError(w, "Post ID and comment content are required")
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		h.showError(w, "Invalid post ID")
		return
	}

	if len(content) < 1 {
		h.showError(w, "Comment content cannot be empty")
		return
	}

	// Verify post exists
	if !h.postExists(postID) {
		h.showError(w, "Post not found")
		return
	}

	// Create comment
	commentID, err := h.createComment(postID, currentUser.ID, content)
	if err != nil {
		h.showError(w, "Error creating comment: "+err.Error())
		return
	}

	fmt.Printf("✅ Comment created successfully: ID=%d, PostID=%d, UserID=%d\n",
		commentID, postID, currentUser.ID)

	// Redirect back to post view
	http.Redirect(w, r, fmt.Sprintf("/posts/view?id=%d", postID), http.StatusSeeOther)
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

// showError displays an error page for comment operations
func (h *CommentsHandler) showError(w http.ResponseWriter, message string) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Comment Error</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
			.error-container { background: #f8d7da; padding: 30px; border-radius: 8px; 
							   border: 1px solid #f5c6cb; margin: 20px 0; text-align: center; }
			.btn { background: #6c757d; color: white; padding: 12px 20px; border: none; 
				   border-radius: 4px; cursor: pointer; font-size: 16px; text-decoration: none; 
				   display: inline-block; margin: 5px; }
			.btn:hover { background: #545b62; }
			h1 { color: #721c24; }
		</style>
	</head>
	<body>
		<div class="error-container">
			<h1>❌ Comment Error</h1>
			<p>%s</p>
			<a href="javascript:history.back()" class="btn">← Go Back</a>
			<a href="/posts" class="btn">📝 All Posts</a>
		</div>
	</body>
	</html>`, message)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, html)
}
