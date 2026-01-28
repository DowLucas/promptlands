package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
	"github.com/lucas/promptlands/internal/game"
	"github.com/lucas/promptlands/internal/ws"
)

// Handler contains HTTP handler methods
type Handler struct {
	gameManager *game.Manager
	hub         *ws.Hub
	wsHandler   *ws.Handler
	cfg         *config.Config
}

// NewHandler creates a new API handler
func NewHandler(gameManager *game.Manager, hub *ws.Hub, cfg *config.Config) *Handler {
	h := &Handler{
		gameManager: gameManager,
		hub:         hub,
		cfg:         cfg,
	}
	h.wsHandler = ws.NewHandler(hub, &gameStateAdapter{gameManager})
	return h
}

// gameStateAdapter adapts game.Manager to ws.GameStateProvider
type gameStateAdapter struct {
	manager *game.Manager
}

func (a *gameStateAdapter) GetFullState(gameID uuid.UUID) (interface{}, error) {
	engine, err := a.manager.GetGame(gameID)
	if err != nil {
		return nil, err
	}
	return engine.GetFullState(), nil
}

// Health returns server health status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// ListGames returns all active games
func (h *Handler) ListGames(w http.ResponseWriter, r *http.Request) {
	games := h.gameManager.ListGames()
	writeJSON(w, http.StatusOK, games)
}

// CreateGame creates a new multiplayer game
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	engine, err := h.gameManager.CreateGame()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":     engine.ID,
		"status": engine.GetStatus(),
	})
}

// CreateSingleplayerGame creates a game with AI adversaries
func (h *Handler) CreateSingleplayerGame(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerPrompt string   `json:"player_prompt"`
		PlayerName   string   `json:"player_name"`
		Adversaries  []string `json:"adversaries"`
		Seed         int64    `json:"seed,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerPrompt == "" {
		writeError(w, http.StatusBadRequest, "player_prompt is required")
		return
	}

	if len(req.Adversaries) == 0 {
		req.Adversaries = []string{"aggressive"} // Default adversary
	}

	if req.PlayerName == "" {
		req.PlayerName = "Player"
	}

	engine, playerAgentID, err := h.gameManager.CreateSingleplayerGameWithSeed(req.PlayerPrompt, req.Adversaries, req.Seed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Auto-start singleplayer games
	if err := engine.Start(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get the generated seed from the world
	state := engine.GetFullState()

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"game_id":         engine.ID,
		"player_agent_id": playerAgentID,
		"status":          engine.GetStatus(),
		"seed":            state.World.Seed,
	})
}

// GetGame returns game details
func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return
	}

	engine, err := h.gameManager.GetGame(gameID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	state := engine.GetFullState()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           engine.ID,
		"status":       engine.GetStatus(),
		"tick":         engine.GetTick(),
		"seed":         state.World.Seed,
		"viewer_count": h.hub.GetGameClientCount(gameID),
	})
}

// JoinGame adds a player to a game
func (h *Handler) JoinGame(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return
	}

	var req struct {
		PlayerName   string `json:"player_name"`
		SystemPrompt string `json:"system_prompt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PlayerName == "" {
		req.PlayerName = "Anonymous"
	}

	if req.SystemPrompt == "" {
		writeError(w, http.StatusBadRequest, "system_prompt is required")
		return
	}

	agent, err := h.gameManager.JoinGame(gameID, req.PlayerName, req.SystemPrompt)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id": agent.ID,
		"name":     agent.Name,
		"position": agent.Position,
	})
}

// StartGame starts a waiting game
func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return
	}

	if err := h.gameManager.StartGame(gameID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "started",
	})
}

// GetGameState returns the full current game state
func (h *Handler) GetGameState(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return
	}

	engine, err := h.gameManager.GetGame(gameID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, engine.GetFullState())
}

// WebSocket handles WebSocket connections
func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return
	}

	// Verify game exists
	_, err = h.gameManager.GetGame(gameID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.wsHandler.ServeWS(w, r, gameID)
}

// ForceTick manually triggers a tick (dev only)
func (h *Handler) ForceTick(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return
	}

	if err := h.gameManager.ForceTick(gameID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "tick processed",
	})
}

// DebugState returns full state for debugging (dev only)
func (h *Handler) DebugState(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return
	}

	engine, err := h.gameManager.GetGame(gameID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, engine.GetFullState())
}

// ListAdversaries returns available AI adversary types
func (h *Handler) ListAdversaries(w http.ResponseWriter, r *http.Request) {
	types := game.GetAdversaryTypes()
	adversaries := make([]map[string]string, len(types))

	for i, t := range types {
		config, _ := game.GetAdversary(t)
		adversaries[i] = map[string]string{
			"type": t,
			"name": config.Name,
		}
	}

	writeJSON(w, http.StatusOK, adversaries)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
