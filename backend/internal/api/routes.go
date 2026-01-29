package api

import (
	"net/http"

	"github.com/lucas/promptlands/internal/config"
	"github.com/lucas/promptlands/internal/game"
	"github.com/lucas/promptlands/internal/ws"
)

// NewRouter creates the HTTP router with all routes
func NewRouter(gameManager *game.Manager, hub *ws.Hub, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	handler := NewHandler(gameManager, hub, cfg)

	// Health check
	mux.HandleFunc("GET /health", handler.Health)

	// Game routes
	mux.HandleFunc("GET /api/games", handler.ListGames)
	mux.HandleFunc("POST /api/games", handler.CreateGame)
	mux.HandleFunc("GET /api/games/{id}", handler.GetGame)
	mux.HandleFunc("POST /api/games/{id}/join", handler.JoinGame)
	mux.HandleFunc("POST /api/games/{id}/start", handler.StartGame)
	mux.HandleFunc("GET /api/games/{id}/state", handler.GetGameState)

	// Singleplayer
	mux.HandleFunc("POST /api/games/singleplayer", handler.CreateSingleplayerGame)

	// WebSocket
	mux.HandleFunc("GET /ws/game/{id}", handler.WebSocket)

	// Dev routes (only enabled in dev mode)
	if cfg.Dev.Enabled {
		mux.HandleFunc("POST /api/dev/tick/{id}", handler.ForceTick)
		mux.HandleFunc("POST /api/dev/pause/{id}", handler.PauseGame)
		mux.HandleFunc("POST /api/dev/resume/{id}", handler.ResumeGame)
		mux.HandleFunc("GET /api/dev/state/{id}", handler.DebugState)
	}

	// Adversary types
	mux.HandleFunc("GET /api/adversaries", handler.ListAdversaries)

	// Add CORS middleware
	return corsMiddleware(mux)
}

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
