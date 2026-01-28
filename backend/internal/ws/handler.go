package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Add proper origin checking in production
		return true
	},
}

// GameStateProvider provides game state for new connections
type GameStateProvider interface {
	GetFullState(gameID uuid.UUID) (interface{}, error)
}

// Handler handles WebSocket connections
type Handler struct {
	hub           *Hub
	stateProvider GameStateProvider
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, stateProvider GameStateProvider) *Handler {
	return &Handler{
		hub:           hub,
		stateProvider: stateProvider,
	}
}

// ServeWS handles WebSocket requests from clients
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request, gameID uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		ID:     uuid.New(),
		GameID: gameID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		hub:    h.hub,
	}

	h.hub.Register(client)

	// Send initial game state
	if h.stateProvider != nil {
		state, err := h.stateProvider.GetFullState(gameID)
		if err == nil {
			data, _ := json.Marshal(state)
			client.Send <- data
		}
	}

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from clients
func (c *Client) handleMessage(message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse client message: %v", err)
		return
	}

	switch msg.Type {
	case "ping":
		// Respond with pong
		response, _ := json.Marshal(map[string]string{"type": "pong"})
		c.Send <- response

	case "subscribe":
		// Client wants to watch a different game
		if msg.GameID != uuid.Nil && msg.GameID != c.GameID {
			c.hub.Unregister(c)
			c.GameID = msg.GameID
			c.hub.Register(c)
		}

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// ClientMessage represents a message from a WebSocket client
type ClientMessage struct {
	Type   string    `json:"type"`
	GameID uuid.UUID `json:"game_id,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}
