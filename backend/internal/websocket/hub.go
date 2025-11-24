package websocket

import (
	"encoding/json"
	"log"
)

// Hub maintains active client connections and broadcasts messages
type Hub struct {
	// Map of user ID to client connection
	clients map[int]*Client

	// Channel for incoming messages
	broadcast chan Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

// Message structure for WebSocket communication
type Message struct {
	Type      string `json:"type"`      // "private_message", "user_online", "user_offline"
	From      int    `json:"from"`      // Sender user ID
	To        int    `json:"to"`        // Receiver user ID (for private messages)
	Content   string `json:"content"`   // Message content
	Timestamp string `json:"timestamp"` // Message timestamp
	Username  string `json:"username"`  // Sender's username
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[int]*Client),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Register new client
			h.clients[client.userID] = client
			log.Printf("User %d connected. Total clients: %d", client.userID, len(h.clients))
			
			// Notify all clients that user is online
			h.broadcastUserStatus(client.userID, client.username, true)

		case client := <-h.unregister:
			// Unregister client
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("User %d disconnected. Total clients: %d", client.userID, len(h.clients))
				
				// Notify all clients that user is offline
				h.broadcastUserStatus(client.userID, client.username, false)
			}

		case message := <-h.broadcast:
			// Handle different message types
			switch message.Type {
			case "private_message":
				// Send to specific user
				if recipient, ok := h.clients[message.To]; ok {
					select {
					case recipient.send <- message:
					default:
						// Client's send channel is full, close it
						close(recipient.send)
						delete(h.clients, message.To)
					}
				}
				
				// Also send back to sender for confirmation
				if sender, ok := h.clients[message.From]; ok {
					messageCopy := message
					messageCopy.Type = "message_sent"
					select {
					case sender.send <- messageCopy:
					default:
						close(sender.send)
						delete(h.clients, message.From)
					}
				}

			case "broadcast":
				// Send to all connected clients
				for userID, client := range h.clients {
					if userID != message.From { // Don't send back to sender
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, userID)
						}
					}
				}
			}
		}
	}
}

// broadcastUserStatus notifies all clients about user online/offline status
func (h *Hub) broadcastUserStatus(userID int, username string, isOnline bool) {
	statusType := "user_offline"
	if isOnline {
		statusType = "user_online"
	}

	statusMessage := Message{
		Type:     statusType,
		From:     userID,
		Username: username,
	}

	// Send to all connected clients
	for _, client := range h.clients {
		select {
		case client.send <- statusMessage:
		default:
			close(client.send)
			delete(h.clients, client.userID)
		}
	}
}

// GetOnlineUsers returns list of online user IDs
func (h *Hub) GetOnlineUsers() []int {
	users := make([]int, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}
