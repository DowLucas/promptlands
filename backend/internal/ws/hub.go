package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID     uuid.UUID
	GameID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	hub    *Hub
}

// Hub manages all WebSocket connections
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	gameRooms  map[uuid.UUID]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
}

// BroadcastMessage contains a message to broadcast to a game room
type BroadcastMessage struct {
	GameID  uuid.UUID
	Message interface{}
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		gameRooms:  make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case msg := <-h.broadcast:
			h.broadcastToGame(msg)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Add to game room
	if client.GameID != uuid.Nil {
		if h.gameRooms[client.GameID] == nil {
			h.gameRooms[client.GameID] = make(map[*Client]bool)
		}
		h.gameRooms[client.GameID][client] = true
		log.Printf("Client %s joined game %s", client.ID, client.GameID)
	}
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)

		// Remove from game room
		if room, ok := h.gameRooms[client.GameID]; ok {
			delete(room, client)
			if len(room) == 0 {
				delete(h.gameRooms, client.GameID)
			}
		}
		log.Printf("Client %s disconnected", client.ID)
	}
}

// broadcastToGame sends a message to all clients in a game room
func (h *Hub) broadcastToGame(msg BroadcastMessage) {
	h.mu.RLock()
	room, ok := h.gameRooms[msg.GameID]
	if !ok {
		h.mu.RUnlock()
		return
	}

	// Make a copy of clients to avoid holding lock during send
	clients := make([]*Client, 0, len(room))
	for client := range room {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	data, err := json.Marshal(msg.Message)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- data:
		default:
			// Client buffer full, disconnect
			h.unregister <- client
		}
	}
}

// BroadcastToGame sends a message to all clients watching a game
// This implements the game.Broadcaster interface
func (h *Hub) BroadcastToGame(gameID uuid.UUID, message interface{}) {
	h.broadcast <- BroadcastMessage{
		GameID:  gameID,
		Message: message,
	}
}

// Register adds a new client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// GetClientCount returns the total number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetGameClientCount returns the number of clients watching a specific game
func (h *Hub) GetGameClientCount(gameID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.gameRooms[gameID]; ok {
		return len(room)
	}
	return 0
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(clientID uuid.UUID, message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	for client := range h.clients {
		if client.ID == clientID {
			select {
			case client.Send <- data:
			default:
				// Buffer full
			}
			return
		}
	}
}
