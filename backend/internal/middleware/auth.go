package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"real-time-forum/internal/database"
)

// AuthMiddleware provides authentication middleware for protecting routes
type AuthMiddleware struct {
	db *sql.DB
}

// NewAuthMiddleware creates a new authentication middleware instance
func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{
		db: db,
	}
}

// RequireAuth is a middleware that requires user authentication
// It checks if the user has a valid session and redirects to login if not
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is authenticated
		user := m.GetCurrentUser(r)
		if user == nil {
			// User is not authenticated, return 401
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized"}`))
			return
		}

		// User is authenticated, continue to next handler
		next(w, r)
	}
}

// RequireGuest is a middleware that requires user to NOT be authenticated
// It redirects authenticated users away from login/register pages
func (m *AuthMiddleware) RequireGuest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is authenticated
		user := m.GetCurrentUser(r)
		if user != nil {
			// User is already authenticated, redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// User is not authenticated, continue to next handler
		next(w, r)
	}
}

// GetCurrentUser extracts the current user from the request session
// Returns nil if user is not authenticated or session is invalid
func (m *AuthMiddleware) GetCurrentUser(r *http.Request) *database.User {
	// Get session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil // No session cookie found
	}

	// Look up session in database
	var userID int
	var expiresAt time.Time
	err = m.db.QueryRow(`
		SELECT user_id, expires_at FROM sessions 
		WHERE token = ?
	`, cookie.Value).Scan(&userID, &expiresAt)

	f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	if err != nil {
		fmt.Fprintf(f, "Lookup failed. Token: %s, Error: %v\n", cookie.Value, err)
		return nil // Session not found
	}

	if time.Now().UTC().After(expiresAt) {
		fmt.Fprintf(f, "Expired. Token: %s, Expires: %v, Now: %v\n", cookie.Value, expiresAt, time.Now().UTC())
		return nil // Session expired
	}
	fmt.Fprintf(f, "Success. Token: %s, UserID: %d\n", cookie.Value, userID)

	// Get user details
	var user database.User
	err = m.db.QueryRow(`
		SELECT id, username, email, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		fmt.Fprintf(f, "User lookup failed for ID %d: %v\n", userID, err)
		return nil // User not found
	}
	fmt.Fprintf(f, "User found: %s\n", user.Username)

	return &user
}

// CleanupExpiredSessions removes expired sessions from the database
// This should be called periodically to keep the sessions table clean
func (m *AuthMiddleware) CleanupExpiredSessions() error {
	_, err := m.db.Exec(`
		DELETE FROM sessions 
		WHERE expires_at <= ?
	`, time.Now())

	return err
}

// RevokeUserSessions revokes all sessions for a specific user
// Useful for logout from all devices functionality
func (m *AuthMiddleware) RevokeUserSessions(userID int) error {
	_, err := m.db.Exec(`
		DELETE FROM sessions 
		WHERE user_id = ?
	`, userID)

	return err
}

// ExtendSession extends the expiration time of a session
// Can be used to implement "remember me" functionality
func (m *AuthMiddleware) ExtendSession(token string, duration time.Duration) error {
	newExpiresAt := time.Now().Add(duration)

	_, err := m.db.Exec(`
		UPDATE sessions 
		SET expires_at = ? 
		WHERE token = ?
	`, newExpiresAt, token)

	return err
}

// SessionStats provides statistics about active sessions
// Useful for admin dashboards and monitoring
type SessionStats struct {
	TotalSessions   int `json:"total_sessions"`
	ActiveSessions  int `json:"active_sessions"`
	ExpiredSessions int `json:"expired_sessions"`
	UniqueUsers     int `json:"unique_users"`
}

// GetSessionStats returns statistics about sessions in the database
func (m *AuthMiddleware) GetSessionStats() (*SessionStats, error) {
	stats := &SessionStats{}

	// Get total sessions
	err := m.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&stats.TotalSessions)
	if err != nil {
		return nil, err
	}

	// Get active sessions (not expired)
	err = m.db.QueryRow(`
		SELECT COUNT(*) FROM sessions 
		WHERE expires_at > ?
	`, time.Now()).Scan(&stats.ActiveSessions)
	if err != nil {
		return nil, err
	}

	// Calculate expired sessions
	stats.ExpiredSessions = stats.TotalSessions - stats.ActiveSessions

	// Get unique users with active sessions
	err = m.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) FROM sessions 
		WHERE expires_at > ?
	`, time.Now()).Scan(&stats.UniqueUsers)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// AddUserToContext is a middleware that adds the current user to the request context
// This allows handlers to access user information without database queries
func (m *AuthMiddleware) AddUserToContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user
		_ = m.GetCurrentUser(r)

		// Add user to request context (if needed in the future)
		// For now, handlers can use GetCurrentUser directly

		// Continue to next handler
		next(w, r)
	}
}

// LogActivity logs user activity for security and analytics
// This is optional but useful for production applications
func (m *AuthMiddleware) LogActivity(userID int, action, ipAddress, userAgent string) error {
	// This would typically go to a separate activities/logs table
	// For now, we'll just log to console or implement later

	// In a production app, you might want to create an activities table:
	// CREATE TABLE activities (
	//     id INTEGER PRIMARY KEY AUTOINCREMENT,
	//     user_id INTEGER,
	//     action TEXT,
	//     ip_address TEXT,
	//     user_agent TEXT,
	//     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	//     FOREIGN KEY (user_id) REFERENCES users(id)
	// );

	return nil
}

// RateLimitByUser implements basic rate limiting per user
// This helps prevent spam and abuse
type RateLimiter struct {
	requests map[int][]time.Time // userID -> request timestamps
	limit    int                 // max requests
	window   time.Duration       // time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[int][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a user is allowed to make a request
func (rl *RateLimiter) Allow(userID int) bool {
	now := time.Now()

	// Clean old requests
	if timestamps, exists := rl.requests[userID]; exists {
		var validRequests []time.Time
		for _, timestamp := range timestamps {
			if now.Sub(timestamp) < rl.window {
				validRequests = append(validRequests, timestamp)
			}
		}
		rl.requests[userID] = validRequests
	}

	// Check if under limit
	if len(rl.requests[userID]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[userID] = append(rl.requests[userID], now)
	return true
}

// RateLimit is a middleware that implements rate limiting
func (m *AuthMiddleware) RateLimit(limiter *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user := m.GetCurrentUser(r)
			if user != nil {
				if !limiter.Allow(user.ID) {
					http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
					return
				}
			}

			next(w, r)
		}
	}
}
