package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"real-time-forum/internal/middleware"
)

type VotesHandler struct {
	db             *sql.DB
	authMiddleware *middleware.AuthMiddleware
}

func NewVotesHandler(db *sql.DB, authMiddleware *middleware.AuthMiddleware) *VotesHandler {
	return &VotesHandler{
		db:             db,
		authMiddleware: authMiddleware,
	}
}

func (h *VotesHandler) VoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUser := h.authMiddleware.GetCurrentUser(r)
	if currentUser == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	voteType := r.FormValue("type")     // "like" or "dislike"
	targetType := r.FormValue("target") // "post" or "comment"
	targetIDStr := r.FormValue("target_id")
	redirectURL := r.FormValue("redirect")

	if voteType == "" || targetType == "" || targetIDStr == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	if voteType != "like" && voteType != "dislike" {
		http.Error(w, "Invalid vote type", http.StatusBadRequest)
		return
	}

	if targetType != "post" && targetType != "comment" {
		http.Error(w, "Invalid target type", http.StatusBadRequest)
		return
	}

	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		http.Error(w, "Invalid target ID", http.StatusBadRequest)
		return
	}

	// Process vote
	err = h.processVote(currentUser.ID, voteType, targetType, targetID)
	if err != nil {
		fmt.Printf("❌ Vote error: %v\n", err)
		http.Error(w, "Error processing vote", http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ Vote processed: user=%d, %s %s:%d\n",
		currentUser.ID, voteType, targetType, targetID)

	// Redirect back to where user came from
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *VotesHandler) processVote(userID int, voteType, targetType string, targetID int) error {
	isLike := voteType == "like"

	// Check if user has already voted
	var existingVote sql.NullBool
	var query string

	if targetType == "post" {
		query = "SELECT is_like FROM likes WHERE user_id = ? AND post_id = ?"
	} else {
		query = "SELECT is_like FROM likes WHERE user_id = ? AND comment_id = ?"
	}

	err := h.db.QueryRow(query, userID, targetID).Scan(&existingVote)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking existing vote: %w", err)
	}

	// If no existing vote, insert new vote
	if err == sql.ErrNoRows {
		return h.insertVote(userID, targetType, targetID, isLike)
	}

	// If existing vote is the same, remove it (toggle off)
	if existingVote.Valid && existingVote.Bool == isLike {
		return h.deleteVote(userID, targetType, targetID)
	}

	// If existing vote is different, update it
	return h.updateVote(userID, targetType, targetID, isLike)
}

func (h *VotesHandler) insertVote(userID int, targetType string, targetID int, isLike bool) error {
	var query string

	if targetType == "post" {
		query = "INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, ?)"
	} else {
		query = "INSERT INTO likes (user_id, comment_id, is_like) VALUES (?, ?, ?)"
	}

	_, err := h.db.Exec(query, userID, targetID, isLike)
	return err
}

func (h *VotesHandler) updateVote(userID int, targetType string, targetID int, isLike bool) error {
	var query string

	if targetType == "post" {
		query = "UPDATE likes SET is_like = ? WHERE user_id = ? AND post_id = ?"
	} else {
		query = "UPDATE likes SET is_like = ? WHERE user_id = ? AND comment_id = ?"
	}

	_, err := h.db.Exec(query, isLike, userID, targetID)
	return err
}

func (h *VotesHandler) deleteVote(userID int, targetType string, targetID int) error {
	var query string

	if targetType == "post" {
		query = "DELETE FROM likes WHERE user_id = ? AND post_id = ?"
	} else {
		query = "DELETE FROM likes WHERE user_id = ? AND comment_id = ?"
	}

	_, err := h.db.Exec(query, userID, targetID)
	return err
}
