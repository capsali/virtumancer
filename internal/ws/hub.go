package ws

import (
	"encoding/json"
	log "github.com/capsali/virtumancer/internal/logging"
)

// MessagePayload defines the structure for data sent with a message.
type MessagePayload map[string]interface{}

// Message is the structured message sent over WebSocket.
type Message struct {
	Type    string         `json:"type"`
	Payload MessagePayload `json:"payload,omitempty"`
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Verbosef("WebSocket client connected: %p", client)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Verbosef("WebSocket client disconnected: %p", client)
			}
		case message := <-h.broadcast:
			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Verbosef("Error marshalling broadcast message: %v", err)
				continue
			}
			log.Debugf("Broadcasting message to %d clients: type=%s", len(h.clients), message.Type)
			for client := range h.clients {
				select {
				case client.send <- messageBytes:
					log.Debugf("Queued message for client %p", client)
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastMessage sends a message to all connected clients.
func (h *Hub) BroadcastMessage(message Message) {
	h.broadcast <- message
}


