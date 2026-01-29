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

func (a *gameStateAdapter) GetFullState(gameID uuid.UUID, playerAgentID *uuid.UUID) (interface{}, error) {
	engine, err := a.manager.GetGame(gameID)
	if err != nil {
		return nil, err
	}
	// If player agent ID is provided, return state with visible tiles
	if playerAgentID != nil {
		log.Printf("gameStateAdapter.GetFullState: calling GetFullStateForPlayer with playerAgentID=%s", *playerAgentID)
		return engine.GetFullStateForPlayer(*playerAgentID), nil
	}
	log.Printf("gameStateAdapter.GetFullState: no playerAgentID, returning basic state")
	return engine.GetFullState(), nil
}

// parseGameID parses the game UUID from the request path.
// Returns the parsed ID and true, or writes an error and returns false.
func (h *Handler) parseGameID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid game ID")
		return uuid.Nil, false
	}
	return gameID, true
}

// getGameEngine parses the game ID and looks up the engine.
// Returns the engine and true, or writes an error and returns false.
func (h *Handler) getGameEngine(w http.ResponseWriter, r *http.Request) (*game.Engine, uuid.UUID, bool) {
	gameID, ok := h.parseGameID(w, r)
	if !ok {
		return nil, uuid.Nil, false
	}
	engine, err := h.gameManager.GetGame(gameID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return nil, uuid.Nil, false
	}
	return engine, gameID, true
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
		MapConfig    *struct {
			Preset     string `json:"preset"`
			Size       string `json:"size"`
			CustomSize int    `json:"custom_size"`
			Seed       int64  `json:"seed"`
		} `json:"map_config,omitempty"`
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

	// Build map size override from request
	var mapSizeOverride string
	if req.MapConfig != nil && req.MapConfig.Size != "" {
		mapSizeOverride = req.MapConfig.Size
	}

	engine, playerAgentID, err := h.gameManager.CreateSingleplayerGameWithSeed(req.PlayerPrompt, req.Adversaries, req.Seed, mapSizeOverride)
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
	engine, gameID, ok := h.getGameEngine(w, r)
	if !ok {
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
	gameID, ok := h.parseGameID(w, r)
	if !ok {
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
	gameID, ok := h.parseGameID(w, r)
	if !ok {
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
	engine, _, ok := h.getGameEngine(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, engine.GetFullState())
}

// WebSocket handles WebSocket connections
func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	_, gameID, ok := h.getGameEngine(w, r)
	if !ok {
		return
	}

	// Parse optional player_agent_id for fog of war
	var playerAgentID *uuid.UUID
	if playerIDStr := r.URL.Query().Get("player_agent_id"); playerIDStr != "" {
		if parsed, err := uuid.Parse(playerIDStr); err == nil {
			playerAgentID = &parsed
			log.Printf("WebSocket connection with player_agent_id: %s", parsed)
		}
	} else {
		log.Printf("WebSocket connection without player_agent_id")
	}

	h.wsHandler.ServeWS(w, r, gameID, playerAgentID)
}

// ForceTick manually triggers a tick (dev only)
func (h *Handler) ForceTick(w http.ResponseWriter, r *http.Request) {
	gameID, ok := h.parseGameID(w, r)
	if !ok {
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

// PauseGame pauses a running game (dev only)
func (h *Handler) PauseGame(w http.ResponseWriter, r *http.Request) {
	gameID, ok := h.parseGameID(w, r)
	if !ok {
		return
	}

	if err := h.gameManager.PauseGame(gameID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "game paused",
	})
}

// ResumeGame resumes a paused game (dev only)
func (h *Handler) ResumeGame(w http.ResponseWriter, r *http.Request) {
	gameID, ok := h.parseGameID(w, r)
	if !ok {
		return
	}

	if err := h.gameManager.ResumeGame(gameID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "game resumed",
	})
}

// DebugState returns full state for debugging (dev only)
func (h *Handler) DebugState(w http.ResponseWriter, r *http.Request) {
	engine, _, ok := h.getGameEngine(w, r)
	if !ok {
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
