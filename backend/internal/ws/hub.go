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
	ID            uuid.UUID
	GameID        uuid.UUID
	PlayerAgentID *uuid.UUID // Player agent ID for fog of war calculation
	Conn          *websocket.Conn
	Send          chan []byte
	hub           *Hub
}

// Hub manages all WebSocket connections
type Hub struct {
	mu                 sync.RWMutex
	clients            map[*Client]bool
	gameRooms          map[uuid.UUID]map[*Client]bool
	register           chan *Client
	unregister         chan *Client
	broadcast          chan BroadcastMessage
	perPlayerBroadcast chan PerPlayerBroadcastMessage
}

// BroadcastMessage contains a message to broadcast to a game room
type BroadcastMessage struct {
	GameID  uuid.UUID
	Message interface{}
}

// PerPlayerBroadcastMessage contains a message to broadcast with per-player customization
type PerPlayerBroadcastMessage struct {
	GameID             uuid.UUID
	BaseMessage        interface{}
	VisibilityProvider func(playerAgentID uuid.UUID) []string
	InventoryProvider  func(playerAgentID uuid.UUID) interface{}
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:            make(map[*Client]bool),
		gameRooms:          make(map[uuid.UUID]map[*Client]bool),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		broadcast:          make(chan BroadcastMessage, 256),
		perPlayerBroadcast: make(chan PerPlayerBroadcastMessage, 256),
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
		case msg := <-h.perPlayerBroadcast:
			h.broadcastToGamePerPlayer(msg)
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

// TickUpdateMessage matches game.TickUpdate structure for JSON marshaling
type TickUpdateMessage struct {
	Type    string                 `json:"type"`
	Tick    int                    `json:"tick"`
	GameID  uuid.UUID              `json:"game_id"`
	Changes TickChangesMessage     `json:"changes"`
}

// TickChangesMessage matches game.TickChanges structure
type TickChangesMessage struct {
	Tiles          json.RawMessage `json:"tiles"`
	Agents         json.RawMessage `json:"agents"`
	Messages       json.RawMessage `json:"messages"`
	Results        json.RawMessage `json:"results"`
	ObjectsAdded   json.RawMessage `json:"objects_added,omitempty"`
	ObjectsRemoved json.RawMessage `json:"objects_removed,omitempty"`
	Respawned      json.RawMessage `json:"respawned,omitempty"`
	VisibleTiles   []string        `json:"visible_tiles,omitempty"`
}

// BroadcastToGameWithVisibility sends a tick update with per-player visible tiles and inventory
// This implements the game.Broadcaster interface
func (h *Hub) BroadcastToGameWithVisibility(gameID uuid.UUID, baseUpdate interface{}, visibilityProvider func(playerAgentID uuid.UUID) []string, inventoryProvider func(playerAgentID uuid.UUID) interface{}) {
	h.perPlayerBroadcast <- PerPlayerBroadcastMessage{
		GameID:             gameID,
		BaseMessage:        baseUpdate,
		VisibilityProvider: visibilityProvider,
		InventoryProvider:  inventoryProvider,
	}
}

// broadcastToGamePerPlayer sends customized tick updates to each player
func (h *Hub) broadcastToGamePerPlayer(msg PerPlayerBroadcastMessage) {
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

	// Marshal the base message once to extract components
	baseData, err := json.Marshal(msg.BaseMessage)
	if err != nil {
		log.Printf("Failed to marshal base broadcast message: %v", err)
		return
	}

	// Parse the base message to extract components
	var baseUpdate map[string]json.RawMessage
	if err := json.Unmarshal(baseData, &baseUpdate); err != nil {
		log.Printf("Failed to parse base broadcast message: %v", err)
		return
	}

	for _, client := range clients {
		var data []byte

		if client.PlayerAgentID != nil {
			// Get visible tiles for this player
			visibleTiles := msg.VisibilityProvider(*client.PlayerAgentID)
			log.Printf("broadcastToGamePerPlayer: client %s has PlayerAgentID %s, visibleTiles=%d", client.ID, *client.PlayerAgentID, len(visibleTiles))

			// Create customized message with visible tiles and inventory
			customUpdate := make(map[string]interface{})
			for k, v := range baseUpdate {
				if k == "changes" {
					// Parse changes and add per-player fields
					var changes map[string]json.RawMessage
					if err := json.Unmarshal(v, &changes); err == nil {
						customChanges := make(map[string]interface{})
						for ck, cv := range changes {
							customChanges[ck] = cv
						}
						customChanges["visible_tiles"] = visibleTiles
						if msg.InventoryProvider != nil {
							inv := msg.InventoryProvider(*client.PlayerAgentID)
							if inv != nil {
								customChanges["player_inventory"] = inv
							}
						}
						customUpdate["changes"] = customChanges
					} else {
						customUpdate[k] = v
					}
				} else {
					customUpdate[k] = v
				}
			}
			data, err = json.Marshal(customUpdate)
			if err != nil {
				log.Printf("Failed to marshal customized message: %v", err)
				data = baseData // Fallback to base
			}
		} else {
			// Non-player clients get the base message
			data = baseData
		}

		select {
		case client.Send <- data:
		default:
			// Client buffer full, disconnect
			h.unregister <- client
		}
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
