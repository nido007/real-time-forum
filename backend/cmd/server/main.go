package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"real-time-forum/internal/database"
	"real-time-forum/internal/handlers"
	"real-time-forum/internal/middleware"
	"real-time-forum/internal/websocket"
)

func main() {
	fmt.Println("🚀 Starting Forum Server...")

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatal("❌ Failed to initialize database:", err)
	}
	defer db.Close()

	// Create handlers and middleware
	authHandler := handlers.NewAuthHandler(db)
	authMiddleware := middleware.NewAuthMiddleware(db)
	postsHandler := handlers.NewPostsHandler(db, authMiddleware)
	commentsHandler := handlers.NewCommentsHandler(db, authMiddleware)
	votesHandler := handlers.NewVotesHandler(db, authMiddleware)

	// Create WebSocket hub
	hub := websocket.NewHub()
	go hub.Run() // Start hub in a goroutine
	log.Println("🔌 WebSocket hub initialized")

	// Create messages handler
	messagesHandler := handlers.NewMessagesHandler(db, hub, authMiddleware)

	// Set up routes
	setupRoutes(authHandler, authMiddleware, postsHandler, commentsHandler, votesHandler, hub, messagesHandler)

	// Start cleanup routine
	go startSessionCleanup(authMiddleware)

	// Start server
	port := getPort()
	fmt.Printf("🌐 Server running on http://localhost%s\n", port)
	fmt.Println("📝 Available endpoints:")
	fmt.Println("   - GET  / (home page)")
	fmt.Println("   - GET  /register, POST /register")
	fmt.Println("   - GET  /login, POST /login")
	fmt.Println("   - GET  /logout")
	fmt.Println("   - GET  /posts (list posts)")
	fmt.Println("   - GET  /posts/create, POST /posts/create")
	fmt.Println("   - GET  /posts/view?id=X")
	fmt.Println("   - POST /comments/create")
	fmt.Println("   - POST /vote")
	fmt.Println("   - WS   /ws (WebSocket connection)")
	fmt.Println("   - POST /api/messages/send")
	fmt.Println("   - GET  /api/messages/history")
	fmt.Println("   - GET  /api/online-users")

	// Graceful shutdown
	setupGracefulShutdown(db)

	// Start HTTP server
	log.Fatal(http.ListenAndServe(port, nil))
}

func setupRoutes(authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware,
	postsHandler *handlers.PostsHandler, commentsHandler *handlers.CommentsHandler,
	votesHandler *handlers.VotesHandler, hub *websocket.Hub, messagesHandler *handlers.MessagesHandler) {

	// Home page
	http.HandleFunc("/", logRequest(homeHandler(authMiddleware)))

	// Authentication routes
	http.HandleFunc("/register", logRequest(authHandler.RegisterHandler))
	http.HandleFunc("/login", logRequest(authHandler.LoginHandler))
	http.HandleFunc("/logout", logRequest(authHandler.LogoutHandler))

	// Posts routes
	http.HandleFunc("/posts", logRequest(postsHandler.ListPostsHandler))
	http.HandleFunc("/posts/create", logRequest(authMiddleware.RequireAuth(postsHandler.CreatePostHandler)))
	http.HandleFunc("/posts/view", logRequest(postsHandler.ViewPostHandler))

	// Comments routes
	http.HandleFunc("/comments/create", logRequest(authMiddleware.RequireAuth(commentsHandler.CreateCommentHandler)))

	// Voting routes
	http.HandleFunc("/vote", logRequest(authMiddleware.RequireAuth(votesHandler.VoteHandler)))

	// Message API routes
	http.HandleFunc("/api/messages/send", logRequest(authMiddleware.RequireAuth(messagesHandler.SendMessage)))
	http.HandleFunc("/api/messages/history", logRequest(authMiddleware.RequireAuth(messagesHandler.GetMessageHistory)))
	http.HandleFunc("/api/online-users", logRequest(authMiddleware.RequireAuth(messagesHandler.GetOnlineUsers)))
	log.Println("💬 Message API endpoints registered")

	// WebSocket endpoint
	http.HandleFunc("/ws", logRequest(func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(hub, func(req *http.Request) (int, error) {
			return getUserIDFromRequest(req, authMiddleware)
		})(w, r)
	}))
	log.Println("🔌 WebSocket endpoint registered: /ws")

	// Static file serving
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("✅ All routes configured successfully!")
}

func homeHandler(authMiddleware *middleware.AuthMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		currentUser := authMiddleware.GetCurrentUser(r)
		html := generateHomepage(currentUser)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	}
}

func generateHomepage(user *database.User) string {
	userSection := ""

	if user != nil {
		// User is logged in - show user options and post links
		userSection = fmt.Sprintf(`
		<div class="user-info">
			<h3>👋 Welcome back, %s!</h3>
			<p><strong>Email:</strong> %s</p>
			<p><strong>Member since:</strong> %s</p>
			<div class="user-actions" style="margin-top: 20px;">
				<a href="/posts" class="btn btn-primary">📝 Browse Posts</a>
				<a href="/posts/create" class="btn btn-success">➕ Create New Post</a>
				<a href="/posts?filter=my-posts" class="btn btn-info">📋 My Posts</a>
				<a href="/posts?filter=liked-posts" class="btn btn-info">❤️ Liked Posts</a>
				<a href="/logout" class="btn btn-secondary">🚪 Logout</a>
			</div>
		</div>
		`, user.Username, user.Email, user.CreatedAt.Format("January 2, 2006"))
	} else {
		// User is not logged in - show posts link and registration
		userSection = `
		<div class="guest-info">
			<h3>🔐 Join Our Community</h3>
			<p>Create an account to participate in discussions, create posts, and interact with other members!</p>
			<div class="guest-actions" style="margin-top: 20px;">
				<a href="/posts" class="btn btn-primary">📝 Browse Posts</a>
				<a href="/register" class="btn btn-success">📝 Create Account</a>
				<a href="/login" class="btn btn-info">🔑 Login</a>
			</div>
		</div>
		`
	}

	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Forum Home</title>
		<style>
			* { margin: 0; padding: 0; box-sizing: border-box; }
			body { 
				font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
				line-height: 1.6; color: #333; background-color: #f8f9fa;
			}
			.container { max-width: 1200px; margin: 0 auto; padding: 20px; }
			.header { text-align: center; margin-bottom: 40px; }
			.header h1 { color: #2c3e50; margin-bottom: 10px; }
			.header p { color: #7f8c8d; font-size: 1.1em; }
			
			.card { 
				background: white; padding: 30px; margin: 20px 0; 
				border-radius: 12px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);
			}
			.user-info { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; }
			.guest-info { background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); color: white; }
			
			.btn { 
				display: inline-block; padding: 12px 24px; margin: 8px 8px 8px 0;
				text-decoration: none; border-radius: 6px; font-weight: 600;
				transition: transform 0.2s, box-shadow 0.2s;
			}
			.btn:hover { transform: translateY(-2px); box-shadow: 0 4px 15px rgba(0,0,0,0.2); }
			.btn-primary { background: #3498db; color: white; }
			.btn-success { background: #2ecc71; color: white; }
			.btn-secondary { background: #95a5a6; color: white; }
			.btn-info { background: #17a2b8; color: white; }
			
			.features { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
			.feature { text-align: center; padding: 20px; }
			.feature h3 { color: #2c3e50; margin-bottom: 15px; }
			.feature p { color: #7f8c8d; }
			
			.quick-nav { background: #e8f4fd; padding: 25px; border-radius: 8px; text-align: center; margin: 20px 0; }
			.quick-nav h3 { color: #2c3e50; margin-bottom: 15px; }
			
			.footer { text-align: center; margin-top: 40px; padding: 20px; 
					  background: #2c3e50; color: white; border-radius: 12px; }
			
			@media (max-width: 768px) {
				.container { padding: 10px; }
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1>🏠 Welcome to Our Forum!</h1>
				<p>A modern community platform built with Go - Connect, Share, Discuss</p>
			</div>

			<div class="card">
				%s
			</div>

			<div class="quick-nav">
				<h3>🚀 Quick Navigation</h3>
				<p>Explore our community and join the conversation!</p>
				<div style="margin-top: 15px;">
					<a href="/posts" class="btn btn-primary">📝 All Posts</a>
					<a href="/posts?category=1" class="btn btn-info">💻 Technology</a>
					<a href="/posts?category=2" class="btn btn-info">🎮 Gaming</a>
					<a href="/posts?category=3" class="btn btn-info">⚽ Sports</a>
				</div>
			</div>

			<div class="card">
				<h2 style="text-align: center; margin-bottom: 30px;">🌟 Platform Features</h2>
				<div class="features">
					<div class="feature">
						<h3>💬 Rich Discussions</h3>
						<p>Create posts and engage in conversations.</p>
					</div>
					<div class="feature">
						<h3>🏷️ Category System</h3>
						<p>Organize content with multiple categories.</p>
					</div>
					<div class="feature">
						<h3>👍 Voting System</h3>
						<p>Like and dislike posts and comments.</p>
					</div>
					<div class="feature">
						<h3>🔐 Secure Authentication</h3>
						<p>Safe user accounts with session management.</p>
					</div>
				</div>
			</div>

			<div class="footer">
				<p><strong>🚀 Forum v1.0</strong> - Built with Go & SQLite</p>
				<p>Server Time: %s</p>
			</div>
		</div>
	</body>
	</html>
	`, userSection, time.Now().Format("2006-01-02 15:04:05 MST"))
}

func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler(w, r)
		duration := time.Since(start)
		log.Printf("📥 %s %s from %s [%v]", r.Method, r.URL.Path, r.RemoteAddr, duration)
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if port[0] != ':' {
		port = ":" + port
	}
	return port
}

func startSessionCleanup(authMiddleware *middleware.AuthMiddleware) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		err := authMiddleware.CleanupExpiredSessions()
		if err != nil {
			log.Printf("⚠️ Error cleaning up expired sessions: %v", err)
		} else {
			log.Println("🧹 Cleaned up expired sessions")
		}
	}
}

func setupGracefulShutdown(db *sql.DB) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("\n🛑 Shutting down gracefully...")
		if db != nil {
			db.Close()
			log.Println("🗄️ Database connection closed")
		}
		log.Println("👋 Goodbye!")
		os.Exit(0)
	}()
}

// getUserIDFromRequest extracts user ID from session cookie
// This is needed for WebSocket authentication
func getUserIDFromRequest(r *http.Request, authMiddleware *middleware.AuthMiddleware) (int, error) {
	user := authMiddleware.GetCurrentUser(r)
	if user == nil {
		return 0, fmt.Errorf("user not authenticated")
	}
	return user.ID, nil
}

// my forum project
