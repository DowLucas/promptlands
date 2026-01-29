package game

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
	"github.com/lucas/promptlands/internal/game/worldgen"
)

// GameStatus represents the current state of a game
type GameStatus string

const (
	StatusWaiting  GameStatus = "waiting"
	StatusRunning  GameStatus = "running"
	StatusFinished GameStatus = "finished"
)

// Engine manages a single game instance
type Engine struct {
	mu              sync.RWMutex
	ID              uuid.UUID
	config          config.GameConfig
	balance         config.BalanceConfig
	world           *World
	agents          map[uuid.UUID]*Agent
	messages        []GameMessage
	llmClient       LLMClient
	promptBuilder   PromptBuilder
	broadcaster     Broadcaster
	status          GameStatus
	tick            int
	cancel          context.CancelFunc
	resolver        *ConflictResolver
	worldObjects    *WorldObjectManager
	itemRegistry    *ItemRegistry
	recipeRegistry  *RecipeRegistry
	handlerRegistry *HandlerRegistry
	paused          bool // When true, tick loop doesn't run

	// Biome/loot registries for per-tick resource spawning
	biomeRegistry *worldgen.BiomeRegistry
	lootTables    *worldgen.LootTableRegistry
	biomeLoot     map[worldgen.BiomeType]worldgen.BiomeLootTable
	spawnRng      *rand.Rand
}

// GameMessage represents a message sent during the game
type GameMessage struct {
	Tick        int        `json:"tick"`
	FromAgentID uuid.UUID  `json:"from_agent_id"`
	ToAgentID   *uuid.UUID `json:"to_agent_id,omitempty"`
	Content     string     `json:"content"`
}

// Broadcaster interface for sending game updates
type Broadcaster interface {
	BroadcastToGame(gameID uuid.UUID, message interface{})
	BroadcastToGameWithVisibility(gameID uuid.UUID, baseUpdate interface{}, visibilityProvider func(playerAgentID uuid.UUID) []string, inventoryProvider func(playerAgentID uuid.UUID) interface{})
}

// LLMClient interface for getting actions from an LLM
type LLMClient interface {
	GetAction(ctx context.Context, agentID uuid.UUID, prompt string) (Action, error)
}

// PromptBuilder interface for building prompts
type PromptBuilder interface {
	BuildPrompt(ctx AgentContext) string
}

// NewEngine creates a new game engine with default (flat plains) terrain
func NewEngine(id uuid.UUID, cfg config.GameConfig, llmClient LLMClient, promptBuilder PromptBuilder, broadcaster Broadcaster) *Engine {
	return NewEngineWithSeed(id, cfg, config.DefaultBalanceConfig(), llmClient, promptBuilder, broadcaster, 0)
}

// NewEngineWithSeed creates a new game engine with procedurally generated terrain
func NewEngineWithSeed(id uuid.UUID, cfg config.GameConfig, balance config.BalanceConfig, llmClient LLMClient, promptBuilder PromptBuilder, broadcaster Broadcaster, seed int64) *Engine {
	var world *World
	var enhancedTiles [][]worldgen.EnhancedTileData
	var mapConfig *worldgen.MapConfig
	mapSize := cfg.GetMapSize()

	if seed == 0 {
		// Use flat plains world when no seed
		world = NewWorld(mapSize)
	} else {
		// Generate procedural terrain with enhanced biome system
		mapConfig = worldgen.DefaultMapConfig()
		mapConfig.CustomSize = mapSize
		enhancedGen := worldgen.NewEnhancedWorldGenerator(seed, mapConfig)
		enhancedTiles = enhancedGen.Generate()

		// Convert worldgen.EnhancedTileData to game.Tile (preserving biome)
		tiles := make([][]*Tile, mapSize)
		for y := 0; y < mapSize; y++ {
			tiles[y] = make([]*Tile, mapSize)
			for x := 0; x < mapSize; x++ {
				tiles[y][x] = &Tile{
					Position: Position{X: x, Y: y},
					OwnerID:  nil,
					Terrain:  TerrainType(enhancedTiles[y][x].Terrain),
					Biome:    string(enhancedTiles[y][x].Biome),
				}
			}
		}

		world = NewWorldWithSeed(mapSize, seed, tiles)
	}

	// Initialize registries
	itemRegistry := DefaultItemRegistry()
	recipeRegistry := DefaultRecipeRegistry()
	worldObjects := NewWorldObjectManager()

	// Initialize biome/loot registries for per-tick resource spawning
	biomeRegistry := worldgen.DefaultBiomeRegistry()
	lootTables := worldgen.NewLootTableRegistry(seed + 3000)
	biomeLoot := worldgen.GetBiomeLootTables()

	engine := &Engine{
		ID:              id,
		config:          cfg,
		balance:         balance,
		world:           world,
		agents:          make(map[uuid.UUID]*Agent),
		messages:        make([]GameMessage, 0),
		llmClient:       llmClient,
		promptBuilder:   promptBuilder,
		broadcaster:     broadcaster,
		status:          StatusWaiting,
		tick:            0,
		resolver:        NewConflictResolver(),
		worldObjects:    worldObjects,
		itemRegistry:    itemRegistry,
		recipeRegistry:  recipeRegistry,
		handlerRegistry: nil, // Set via SetHandlerRegistry
		biomeRegistry:   biomeRegistry,
		lootTables:      lootTables,
		biomeLoot:       biomeLoot,
		spawnRng:        rand.New(rand.NewSource(seed + 4000)),
	}

	// Populate world with interactives (no initial resources — they spawn per-tick)
	if seed != 0 && enhancedTiles != nil && mapConfig != nil {
		populator := NewEnhancedWorldPopulator(seed, world, worldObjects, mapConfig, enhancedTiles)
		populator.PopulateInteractives()
	} else if seed != 0 {
		populator := NewWorldPopulator(seed, world, worldObjects)
		populator.PopulateInteractives()
	}

	return engine
}

// GetWorld returns the game world (for spawn position validation)
func (e *Engine) GetWorld() *World {
	return e.world
}

// AddAgent adds an agent to the game
func (e *Engine) AddAgent(agent *Agent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != StatusWaiting {
		return ErrGameAlreadyStarted
	}

	if len(e.agents) >= e.config.MaxPlayers {
		return ErrGameFull
	}

	// Initialize agent's inventory
	agent.InitInventory(e.itemRegistry)

	e.agents[agent.ID] = agent
	return nil
}

// RemoveAgent removes an agent from the game
func (e *Engine) RemoveAgent(agentID uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.agents, agentID)
}

// Start begins the game loop (unless paused)
func (e *Engine) Start() error {
	e.mu.Lock()
	if e.status != StatusWaiting {
		e.mu.Unlock()
		return ErrGameAlreadyStarted
	}

	if len(e.agents) == 0 {
		e.mu.Unlock()
		return ErrNoAgents
	}

	e.status = StatusRunning

	// If paused, don't start the tick loop
	if e.paused {
		log.Printf("Game %s started in paused mode (no tick loop)", e.ID)
		e.mu.Unlock()
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.mu.Unlock()

	go e.runLoop(ctx)
	return nil
}

// Resume starts the tick loop for a paused game
func (e *Engine) Resume() error {
	e.mu.Lock()
	if e.status != StatusRunning {
		e.mu.Unlock()
		return &GameError{"game not running"}
	}

	if !e.paused {
		e.mu.Unlock()
		return &GameError{"game not paused"}
	}

	e.paused = false
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.mu.Unlock()

	log.Printf("Game %s resumed", e.ID)
	go e.runLoop(ctx)
	return nil
}

// Pause stops the tick loop but keeps the game running
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancel != nil {
		e.cancel()
		e.cancel = nil
	}
	e.paused = true
	log.Printf("Game %s paused", e.ID)
}

// Stop ends the game
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cancel != nil {
		e.cancel()
	}
	e.status = StatusFinished
}

// runLoop is the main game loop
func (e *Engine) runLoop(ctx context.Context) {
	ticker := time.NewTicker(e.config.TickDuration)
	defer ticker.Stop()

	log.Printf("Game %s started with %d agents", e.ID, len(e.agents))

	for {
		select {
		case <-ctx.Done():
			log.Printf("Game %s stopped", e.ID)
			return
		case <-ticker.C:
			e.processTick(ctx)
		}
	}
}

// TickUpdate represents the changes in a single tick
type TickUpdate struct {
	Type     string         `json:"type"`
	Tick     int            `json:"tick"`
	Changes  TickChanges    `json:"changes"`
	GameID   uuid.UUID      `json:"game_id"`
}

// TickChanges contains all changes from a tick
type TickChanges struct {
	Tiles           []TileChange          `json:"tiles"`
	Agents          []AgentSnapshot       `json:"agents"`
	Messages        []GameMessage         `json:"messages"`
	Results         []ActionResult        `json:"results"`
	ObjectsAdded    []WorldObjectSnapshot `json:"objects_added,omitempty"`
	ObjectsRemoved  []uuid.UUID           `json:"objects_removed,omitempty"`
	Respawned       []uuid.UUID           `json:"respawned,omitempty"`
	VisibleTiles    []string              `json:"visible_tiles,omitempty"`    // Per-player fog of war
	PlayerInventory *InventorySnapshot    `json:"player_inventory,omitempty"` // Per-player inventory
}

// TileChange represents a tile ownership change
type TileChange struct {
	X       int        `json:"x"`
	Y       int        `json:"y"`
	OwnerID *uuid.UUID `json:"owner_id"`
}

// endGame finishes the game and determines winner
func (e *Engine) endGame() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.status = StatusFinished
	if e.cancel != nil {
		e.cancel()
	}

	// Determine winner
	ownership := e.world.GetOwnershipMap()
	var winnerID uuid.UUID
	maxTiles := 0

	for agentID, count := range ownership {
		if count > maxTiles {
			maxTiles = count
			winnerID = agentID
		}
	}

	log.Printf("Game %s finished. Winner: %s with %d tiles", e.ID, winnerID, maxTiles)

	// Broadcast game end
	if e.broadcaster != nil {
		e.broadcaster.BroadcastToGame(e.ID, map[string]interface{}{
			"type":     "game_over",
			"game_id":  e.ID,
			"winner":   winnerID,
			"scores":   ownership,
		})
	}
}

// GetStatus returns the current game status
func (e *Engine) GetStatus() GameStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// GetTick returns the current tick number
func (e *Engine) GetTick() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tick
}

// GetFullState returns the complete game state
func (e *Engine) GetFullState() FullGameState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	agents := make([]AgentSnapshot, 0, len(e.agents))
	for _, agent := range e.agents {
		agents = append(agents, agent.Snapshot())
	}

	// Get all world objects (excluding hidden traps)
	worldObjects := e.worldObjects.Snapshot()

	return FullGameState{
		Type:         "full_state",
		GameID:       e.ID,
		Tick:         e.tick,
		Status:       e.status,
		World:        e.world.Snapshot(),
		Agents:       agents,
		Messages:     e.messages,
		WorldObjects: worldObjects,
	}
}

// GetFullStateForPlayer returns the complete game state with visible tiles calculated for a specific player
func (e *Engine) GetFullStateForPlayer(playerAgentID uuid.UUID) FullGameState {
	state := e.GetFullState()
	state.VisibleTiles = e.getVisibleTilesForPlayer(playerAgentID)
	if inv := e.getPlayerInventory(playerAgentID); inv != nil {
		state.PlayerInventory = inv.(*InventorySnapshot)
	}
	log.Printf("GetFullStateForPlayer: playerAgentID=%s, visibleTiles=%d", playerAgentID, len(state.VisibleTiles))
	return state
}

// getPlayerInventory returns the inventory snapshot for a specific player agent
func (e *Engine) getPlayerInventory(playerAgentID uuid.UUID) interface{} {
	e.mu.RLock()
	agent := e.agents[playerAgentID]
	e.mu.RUnlock()

	if agent == nil || agent.Inventory == nil {
		return nil
	}

	snapshot := agent.Inventory.Snapshot()
	return &snapshot
}

// getVisibleTilesForPlayer calculates visible tiles for a specific player
func (e *Engine) getVisibleTilesForPlayer(playerAgentID uuid.UUID) []string {
	e.mu.RLock()
	agent := e.agents[playerAgentID]
	agentCount := len(e.agents)
	e.mu.RUnlock()

	if agent == nil {
		log.Printf("getVisibleTilesForPlayer: agent not found for ID %s (total agents: %d)", playerAgentID, agentCount)
		return nil
	}
	if agent.IsDead {
		log.Printf("getVisibleTilesForPlayer: agent %s is dead", playerAgentID)
		return nil
	}

	pos := agent.GetPosition()
	visionRadius := CalculateEffectiveVisionRadius(agent, e.config.VisionRadius, e.worldObjects)

	visibleTiles := e.world.GetVisibleTiles(pos, visionRadius)
	result := make([]string, 0, len(visibleTiles))
	for _, tile := range visibleTiles {
		result = append(result, fmt.Sprintf("%d,%d", tile.Position.X, tile.Position.Y))
	}
	return result
}

// GetWorldObjects returns the world object manager
func (e *Engine) GetWorldObjects() *WorldObjectManager {
	return e.worldObjects
}

// GetItemRegistry returns the item registry
func (e *Engine) GetItemRegistry() *ItemRegistry {
	return e.itemRegistry
}

// GetRecipeRegistry returns the recipe registry
func (e *Engine) GetRecipeRegistry() *RecipeRegistry {
	return e.recipeRegistry
}

// GetBalance returns the balance configuration
func (e *Engine) GetBalance() *config.BalanceConfig {
	return &e.balance
}

// SetHandlerRegistry sets the handler registry for action processing
func (e *Engine) SetHandlerRegistry(registry *HandlerRegistry) {
	e.handlerRegistry = registry
}

// SetPaused sets whether the game is paused (no tick loop)
func (e *Engine) SetPaused(paused bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = paused
}

// IsPaused returns whether the game is paused
func (e *Engine) IsPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
}

// FullGameState represents the complete game state
type FullGameState struct {
	Type            string                `json:"type"`
	GameID          uuid.UUID             `json:"game_id"`
	Tick            int                   `json:"tick"`
	Status          GameStatus            `json:"status"`
	World           WorldSnapshot         `json:"world"`
	Agents          []AgentSnapshot       `json:"agents"`
	Messages        []GameMessage         `json:"messages"`
	WorldObjects    []WorldObjectSnapshot `json:"world_objects,omitempty"`
	VisibleTiles    []string              `json:"visible_tiles,omitempty"`    // For player's fog of war
	PlayerInventory *InventorySnapshot    `json:"player_inventory,omitempty"` // For player's inventory
}

// ForceTick manually triggers the next tick (for dev/testing)
// Works even when game is paused
func (e *Engine) ForceTick() {
	e.mu.RLock()
	status := e.status
	e.mu.RUnlock()

	if status == StatusRunning {
		e.processTick(context.Background())
	}
}

// Game errors
var (
	ErrGameAlreadyStarted = &GameError{"game already started"}
	ErrGameFull           = &GameError{"game is full"}
	ErrNoAgents           = &GameError{"no agents in game"}
	ErrGameNotFound       = &GameError{"game not found"}
)

// GameError represents a game-related error
type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}
