package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin (in production, restrict this)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket upgrades HTTP connection to WebSocket and manages the client
func HandleWebSocket(hub *Hub, getUserID func(*http.Request) (int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from session
		userID, err := getUserID(r)
		if err != nil {
			log.Println("❌ Unauthorized WebSocket connection attempt")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("❌ Failed to upgrade to WebSocket: %v", err)
			return
		}

		// Create new client
		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan *Message, 256),
			UserID: userID,
		}

		// Register client with hub
		client.hub.register <- client

		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump()

		log.Printf("🔌 WebSocket connection established for user %d", userID)
	}
}
