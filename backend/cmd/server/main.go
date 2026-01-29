package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucas/promptlands/internal/api"
	"github.com/lucas/promptlands/internal/config"
	"github.com/lucas/promptlands/internal/db"
	"github.com/lucas/promptlands/internal/game"
	"github.com/lucas/promptlands/internal/game/actions"
	"github.com/lucas/promptlands/internal/llm"
	"github.com/lucas/promptlands/internal/ws"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	devMode := flag.Bool("dev", false, "enable development mode with mock LLM")
	noDB := flag.Bool("no-db", false, "run without database (in-memory only)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("Failed to load config from %s, using defaults: %v", *configPath, err)
		cfg = config.Default()
	}

	if *devMode {
		cfg.Dev.Enabled = true
		cfg.Dev.MockLLM = true
		log.Println("Development mode enabled with mock LLM")
	}

	// Initialize database connections
	var postgres *db.Postgres
	var redis *db.Redis

	if *noDB || cfg.Dev.Enabled {
		log.Println("Running without database (in-memory mode)")
	} else {
		var err error
		postgres, err = db.NewPostgres(cfg.Database.PostgresURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to PostgreSQL: %v", err)
		}

		redis, err = db.NewRedis(cfg.Database.RedisURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to Redis: %v", err)
		}
	}
	defer postgres.Close()
	defer redis.Close()

	// Initialize LLM client
	var llmClient game.LLMClient
	if cfg.Dev.MockLLM {
		llmClient = llm.NewMockClient()
	} else {
		llmClient = llm.NewGeminiClient(cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.Timeout)
	}

	// Initialize prompt builder
	promptBuilder := llm.NewPromptBuilder()

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize handler registry
	handlerRegistry := game.NewHandlerRegistry()
	actions.RegisterAllHandlers(handlerRegistry)

	// Initialize game manager with balance config
	gameManager := game.NewManagerWithBalance(cfg.Game, cfg.Balance, llmClient, promptBuilder, hub, postgres, redis)
	gameManager.SetHandlerRegistry(handlerRegistry)

	// Set pause mode if configured
	if cfg.Dev.PauseTick {
		gameManager.SetPauseByDefault(true)
		log.Println("Pause tick enabled: games will start paused (use ForceTick or Resume)")
	}

	// Set up HTTP routes
	router := api.NewRouter(gameManager, hub, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop all running games
	gameManager.StopAll()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
