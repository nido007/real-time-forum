package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"real-time-forum/internal/database"

	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles all authentication-related HTTP requests
// This includes registration, login, logout, and session management
type AuthHandler struct {
	db *sql.DB
}

// NewAuthHandler creates a new authentication handler with database connection
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
	}
}

// RegisterHandler handles user registration (GET: show form, POST: process registration)
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showRegistrationForm(w, r)
	case http.MethodPost:
		h.processRegistration(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// showRegistrationForm displays the user registration form
func (h *AuthHandler) showRegistrationForm(w http.ResponseWriter, r *http.Request) {
	// Check if user is already logged in
	if h.isUserLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Simple HTML form for registration
	// In a production app, you'd use proper templates
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Register - Forum</title>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
			.form-container { background: #f9f9f9; padding: 30px; border-radius: 8px; border: 1px solid #ddd; }
			.form-group { margin-bottom: 20px; }
			label { display: block; margin-bottom: 5px; font-weight: bold; }
			input[type="text"], input[type="email"], input[type="password"] { 
				width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 4px; 
				font-size: 16px; box-sizing: border-box;
			}
			.btn { background: #007bff; color: white; padding: 12px 20px; border: none; 
				   border-radius: 4px; cursor: pointer; font-size: 16px; width: 100%; }
			.btn:hover { background: #0056b3; }
			.error { color: #dc3545; margin-bottom: 10px; padding: 10px; background: #f8d7da; 
					  border: 1px solid #f5c6cb; border-radius: 4px; }
			.links { text-align: center; margin-top: 20px; }
			.links a { color: #007bff; text-decoration: none; }
			h1 { text-align: center; color: #333; }
		</style>
	</head>
	<body>
		<h1>🔐 Create Account</h1>
		<div class="form-container">
			<form method="POST" action="/register">
				<div class="form-group">
					<label for="username">Username:</label>
					<input type="text" id="username" name="username" required 
						   minlength="3" maxlength="50" placeholder="Choose a unique username">
				</div>
				
				<div class="form-group">
					<label for="email">Email:</label>
					<input type="email" id="email" name="email" required 
						   placeholder="your@email.com">
				</div>
				
				<div class="form-group">
					<label for="password">Password:</label>
					<input type="password" id="password" name="password" required 
						   minlength="6" placeholder="Minimum 6 characters">
				</div>
				
				<div class="form-group">
					<label for="confirm_password">Confirm Password:</label>
					<input type="password" id="confirm_password" name="confirm_password" required 
						   placeholder="Re-enter your password">
				</div>
				
				<button type="submit" class="btn">Create Account</button>
			</form>
		</div>
		
		<div class="links">
			<p><a href="/login">Already have an account? Login here</a></p>
			<p><a href="/">← Back to Home</a></p>
		</div>
	</body>
	</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// processRegistration handles the registration form submission
func (h *AuthHandler) processRegistration(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		h.showError(w, "Invalid form data")
		return
	}

	// Extract form values
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Validate input
	if err := h.validateRegistrationInput(username, email, password, confirmPassword); err != nil {
		h.showError(w, err.Error())
		return
	}

	// Check if user already exists
	if h.userExists(username, email) {
		h.showError(w, "Username or email already exists. Please choose different credentials.")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.showError(w, "Error processing password. Please try again.")
		return
	}

	// Create user in database
	userID, err := h.createUser(username, email, string(hashedPassword))
	if err != nil {
		h.showError(w, "Error creating account. Please try again.")
		return
	}

	// Show success message
	h.showRegistrationSuccess(w, username, userID)
}

// LoginHandler handles user login (GET: show form, POST: process login)
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.showLoginForm(w, r)
	case http.MethodPost:
		h.processLogin(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// showLoginForm displays the login form
func (h *AuthHandler) showLoginForm(w http.ResponseWriter, r *http.Request) {
	// Check if user is already logged in
	if h.isUserLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Login - Forum</title>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body { font-family: Arial, sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; }
			.form-container { background: #f9f9f9; padding: 30px; border-radius: 8px; border: 1px solid #ddd; }
			.form-group { margin-bottom: 20px; }
			label { display: block; margin-bottom: 5px; font-weight: bold; }
			input[type="text"], input[type="email"], input[type="password"] { 
				width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 4px; 
				font-size: 16px; box-sizing: border-box;
			}
			.btn { background: #28a745; color: white; padding: 12px 20px; border: none; 
				   border-radius: 4px; cursor: pointer; font-size: 16px; width: 100%; }
			.btn:hover { background: #218838; }
			.error { color: #dc3545; margin-bottom: 10px; padding: 10px; background: #f8d7da; 
					  border: 1px solid #f5c6cb; border-radius: 4px; }
			.links { text-align: center; margin-top: 20px; }
			.links a { color: #007bff; text-decoration: none; }
			h1 { text-align: center; color: #333; }
		</style>
	</head>
	<body>
		<h1>🔑 Login</h1>
		<div class="form-container">
			<form method="POST" action="/login">
				<div class="form-group">
					<label for="login">Username or Email:</label>
					<input type="text" id="login" name="login" required 
						   placeholder="Enter your username or email">
				</div>
				
				<div class="form-group">
					<label for="password">Password:</label>
					<input type="password" id="password" name="password" required 
						   placeholder="Enter your password">
				</div>
				
				<button type="submit" class="btn">Login</button>
			</form>
		</div>
		
		<div class="links">
			<p><a href="/register">Don't have an account? Create one here</a></p>
			<p><a href="/">← Back to Home</a></p>
		</div>
	</body>
	</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// processLogin handles the login form submission
func (h *AuthHandler) processLogin(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		h.showError(w, "Invalid form data")
		return
	}

	// Extract form values
	login := strings.TrimSpace(r.FormValue("login"))
	password := r.FormValue("password")

	// Validate input
	if login == "" || password == "" {
		h.showError(w, "Please enter both username/email and password")
		return
	}

	// Authenticate user
	user, err := h.authenticateUser(login, password)
	if err != nil {
		h.showError(w, "Invalid username/email or password")
		return
	}

	// Create session
	err = h.createSession(w, user)
	if err != nil {
		h.showError(w, "Error creating session. Please try again.")
		return
	}

	// Show success and redirect
	h.showLoginSuccess(w, user)
}

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session
	h.clearSession(w, r)

	// Show logout success page
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Logged Out - Forum</title>
		<meta charset="UTF-8">
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; 
				   padding: 20px; text-align: center; }
			.success-container { background: #d4edda; padding: 30px; border-radius: 8px; 
								 border: 1px solid #c3e6cb; margin: 20px 0; }
			.btn { background: #007bff; color: white; padding: 12px 20px; border: none; 
				   border-radius: 4px; cursor: pointer; font-size: 16px; text-decoration: none; 
				   display: inline-block; margin: 5px; }
			.btn:hover { background: #0056b3; }
			h1 { color: #333; }
		</style>
	</head>
	<body>
		<h1>👋 Logged Out Successfully</h1>
		<div class="success-container">
			<p>You have been safely logged out of your account.</p>
			<p>Thank you for using our forum!</p>
		</div>
		<a href="/" class="btn">🏠 Back to Home</a>
		<a href="/login" class="btn">🔑 Login Again</a>
	</body>
	</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// HELPER METHODS

// validateRegistrationInput validates user registration input
func (h *AuthHandler) validateRegistrationInput(username, email, password, confirmPassword string) error {
	// Username validation
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	// Email validation (basic)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("please enter a valid email address")
	}

	// Password validation
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	// Confirm password
	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	return nil
}

// userExists checks if a user with the given username or email already exists
func (h *AuthHandler) userExists(username, email string) bool {
	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE username = ? OR email = ?
	`, username, email).Scan(&count)

	if err != nil {
		// If there's an error, assume user exists to be safe
		return true
	}

	return count > 0
}

// createUser creates a new user in the database
func (h *AuthHandler) createUser(username, email, hashedPassword string) (int64, error) {
	result, err := h.db.Exec(`
		INSERT INTO users (username, email, password_hash) 
		VALUES (?, ?, ?)
	`, username, email, hashedPassword)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// authenticateUser verifies user credentials and returns user info
func (h *AuthHandler) authenticateUser(login, password string) (*database.User, error) {
	var user database.User

	// Try to find user by username or email
	err := h.db.QueryRow(`
		SELECT id, username, email, password_hash, created_at 
		FROM users 
		WHERE username = ? OR email = ?
	`, login, login).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// generateSessionToken creates a secure random session token
func (h *AuthHandler) generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateSessionToken creates a secure UUID-based session token
func (h *AuthHandler) createSession(w http.ResponseWriter, user *database.User) error {
	// Generate session token
	token, err := h.generateSessionToken()
	if err != nil {
		return err
	}

	// Session expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour) // ← Should be inside function

	// Save session to database
	_, err = h.db.Exec(`
		INSERT INTO sessions (user_id, token, expires_at) 
		VALUES (?, ?, ?)
	`, user.ID, token, expiresAt)

	if err != nil {
		return err
	}

	// Set HTTP cookie
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	return nil
}

// clearSession removes the user's session (logout)
func (h *AuthHandler) clearSession(w http.ResponseWriter, r *http.Request) {
	// Get session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return // No session to clear
	}

	// Delete from database
	h.db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)

	// Clear cookie
	clearCookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	}

	http.SetCookie(w, clearCookie)
}

// isUserLoggedIn checks if the current request has a valid session
func (h *AuthHandler) isUserLoggedIn(r *http.Request) bool {
	user := h.GetCurrentUser(r)
	return user != nil
}

// GetCurrentUser returns the current logged-in user or nil
func (h *AuthHandler) GetCurrentUser(r *http.Request) *database.User {
	// Get session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}

	// Look up session in database
	var userID int
	var expiresAt time.Time
	err = h.db.QueryRow(`
		SELECT user_id, expires_at FROM sessions 
		WHERE token = ? AND expires_at > ?
	`, cookie.Value, time.Now()).Scan(&userID, &expiresAt)

	if err != nil {
		return nil
	}

	// Get user details
	var user database.User
	err = h.db.QueryRow(`
		SELECT id, username, email, created_at 
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		return nil
	}

	return &user
}

// showError displays an error page
func (h *AuthHandler) showError(w http.ResponseWriter, message string) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Error - Forum</title>
		<meta charset="UTF-8">
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; 
				   padding: 20px; text-align: center; }
			.error-container { background: #f8d7da; padding: 30px; border-radius: 8px; 
							   border: 1px solid #f5c6cb; margin: 20px 0; }
			.btn { background: #6c757d; color: white; padding: 12px 20px; border: none; 
				   border-radius: 4px; cursor: pointer; font-size: 16px; text-decoration: none; 
				   display: inline-block; margin: 5px; }
			.btn:hover { background: #545b62; }
			h1 { color: #721c24; }
		</style>
	</head>
	<body>
		<h1>❌ Error</h1>
		<div class="error-container">
			<p>%s</p>
		</div>
		<a href="javascript:history.back()" class="btn">← Go Back</a>
		<a href="/" class="btn">🏠 Home</a>
	</body>
	</html>`, template.HTMLEscapeString(message))

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, html)
}

// showRegistrationSuccess displays successful registration page
func (h *AuthHandler) showRegistrationSuccess(w http.ResponseWriter, username string, userID int64) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Registration Successful - Forum</title>
		<meta charset="UTF-8">
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; 
				   padding: 20px; text-align: center; }
			.success-container { background: #d4edda; padding: 30px; border-radius: 8px; 
								 border: 1px solid #c3e6cb; margin: 20px 0; }
			.btn { background: #28a745; color: white; padding: 12px 20px; border: none; 
				   border-radius: 4px; cursor: pointer; font-size: 16px; text-decoration: none; 
				   display: inline-block; margin: 5px; }
			.btn:hover { background: #218838; }
			h1 { color: #155724; }
		</style>
	</head>
	<body>
		<h1>🎉 Account Created Successfully!</h1>
		<div class="success-container">
			<p><strong>Welcome to our forum, %s!</strong></p>
			<p>Your account has been created successfully.</p>
			<p>User ID: #%d</p>
			<p>You can now login and start participating in discussions!</p>
		</div>
		<a href="/login" class="btn">🔑 Login Now</a>
		<a href="/" class="btn">🏠 Back to Home</a>
	</body>
	</html>`, template.HTMLEscapeString(username), userID)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// showLoginSuccess displays successful login page and redirects
func (h *AuthHandler) showLoginSuccess(w http.ResponseWriter, user *database.User) {
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Login Successful - Forum</title>
		<meta charset="UTF-8">
		<meta http-equiv="refresh" content="3;url=/">
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; 
				   padding: 20px; text-align: center; }
			.success-container { background: #d4edda; padding: 30px; border-radius: 8px; 
								 border: 1px solid #c3e6cb; margin: 20px 0; }
			.btn { background: #007bff; color: white; padding: 12px 20px; border: none; 
				   border-radius: 4px; cursor: pointer; font-size: 16px; text-decoration: none; 
				   display: inline-block; margin: 5px; }
			.btn:hover { background: #0056b3; }
			h1 { color: #155724; }
		</style>
	</head>
	<body>
		<h1>✅ Welcome back, %s!</h1>
		<div class="success-container">
			<p>You have successfully logged in to your account.</p>
			<p>Redirecting to home page in 3 seconds...</p>
		</div>
		<a href="/" class="btn">🏠 Go to Home</a>
	</body>
	</html>`, template.HTMLEscapeString(user.Username))

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}
