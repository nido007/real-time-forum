package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"real-time-forum/internal/database"

	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles all authentication-related HTTP requests
type AuthHandler struct {
	db *sql.DB
}

// NewAuthHandler creates a new authentication handler with database connection
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

// RegisterRequest represents the JSON payload for registration
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest represents the JSON payload for login
type LoginRequest struct {
	Login    string `json:"login"` // Username or Email
	Password string `json:"password"`
}

// RegisterHandler handles user registration via JSON API
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate input
	if err := h.validateRegistrationInput(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user already exists
	if h.userExists(req.Username, req.Email) {
		h.respondWithError(w, http.StatusConflict, "Username or email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error processing password")
		return
	}

	// Create user in database
	userID, err := h.createUser(&req, string(hashedPassword))
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error creating account")
		return
	}

	h.respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user_id": userID,
	})
}

// LoginHandler handles user login via JSON API
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Login == "" || req.Password == "" {
		h.respondWithError(w, http.StatusBadRequest, "Login and password are required")
		return
	}

	// Authenticate user
	user, err := h.authenticateUser(req.Login, req.Password)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Create session
	err = h.createSession(w, user)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Error creating session")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"user":    user,
	})
}

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	h.clearSession(w, r)
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// HELPER METHODS

func (h *AuthHandler) validateRegistrationInput(req *RegisterRequest) error {
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return fmt.Errorf("invalid email address")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if req.Age <= 0 {
		return fmt.Errorf("invalid age")
	}
	if req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("first name and last name are required")
	}
	return nil
}

func (h *AuthHandler) userExists(username, email string) bool {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", username, email).Scan(&count)
	if err != nil {
		return true // Fail safe
	}
	return count > 0
}

func (h *AuthHandler) createUser(req *RegisterRequest, hashedPassword string) (int64, error) {
	result, err := h.db.Exec(`
		INSERT INTO users (username, email, password_hash, age, gender, first_name, last_name) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, req.Username, req.Email, hashedPassword, req.Age, req.Gender, req.FirstName, req.LastName)

	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (h *AuthHandler) authenticateUser(login, password string) (*database.User, error) {
	var user database.User
	err := h.db.QueryRow(`
		SELECT id, username, email, password_hash, age, gender, first_name, last_name, created_at 
		FROM users 
		WHERE username = ? OR email = ?
	`, login, login).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *AuthHandler) createSession(w http.ResponseWriter, user *database.User) error {
	token, err := h.generateSessionToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	_, err = h.db.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)", user.ID, token, expiresAt)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (h *AuthHandler) clearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		h.db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})
}

func (h *AuthHandler) generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *AuthHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
