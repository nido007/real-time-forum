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
	fmt.Println("🚀 Starting Real-Time Forum Server...")

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
	// Serve JS files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../frontend/js"))))

	// Serve CSS file
	http.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/styles.css")
	})

	fmt.Println("✅ All routes configured successfully!")
}

func homeHandler(authMiddleware *middleware.AuthMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "../frontend/index.html")
	}
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
