package game

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
	"github.com/lucas/promptlands/internal/db"
	"github.com/lucas/promptlands/internal/game/worldgen"
)

// Manager handles multiple game instances
type Manager struct {
	mu              sync.RWMutex
	games           map[uuid.UUID]*Engine
	config          config.GameConfig
	balance         config.BalanceConfig
	llmClient       LLMClient
	promptBuilder   PromptBuilder
	hub             Broadcaster
	postgres        *db.Postgres
	redis           *db.Redis
	handlerRegistry *HandlerRegistry
	pauseByDefault  bool // When true, new games start paused
}

// NewManager creates a new game manager
func NewManager(cfg config.GameConfig, llmClient LLMClient, promptBuilder PromptBuilder, hub Broadcaster, postgres *db.Postgres, redis *db.Redis) *Manager {
	return NewManagerWithBalance(cfg, config.DefaultBalanceConfig(), llmClient, promptBuilder, hub, postgres, redis)
}

// NewManagerWithBalance creates a new game manager with explicit balance config
func NewManagerWithBalance(cfg config.GameConfig, balance config.BalanceConfig, llmClient LLMClient, promptBuilder PromptBuilder, hub Broadcaster, postgres *db.Postgres, redis *db.Redis) *Manager {
	return &Manager{
		games:           make(map[uuid.UUID]*Engine),
		config:          cfg,
		balance:         balance,
		llmClient:       llmClient,
		promptBuilder:   promptBuilder,
		hub:             hub,
		postgres:        postgres,
		redis:           redis,
		handlerRegistry: nil, // Set via SetHandlerRegistry
	}
}

// SetHandlerRegistry sets the handler registry for all new games
func (m *Manager) SetHandlerRegistry(registry *HandlerRegistry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlerRegistry = registry
}

// SetPauseByDefault sets whether new games start paused
func (m *Manager) SetPauseByDefault(paused bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pauseByDefault = paused
}

// PauseGame pauses a running game
func (m *Manager) PauseGame(gameID uuid.UUID) error {
	m.mu.RLock()
	game, ok := m.games[gameID]
	m.mu.RUnlock()

	if !ok {
		return ErrGameNotFound
	}

	game.Pause()
	return nil
}

// ResumeGame resumes a paused game
func (m *Manager) ResumeGame(gameID uuid.UUID) error {
	m.mu.RLock()
	game, ok := m.games[gameID]
	m.mu.RUnlock()

	if !ok {
		return ErrGameNotFound
	}

	return game.Resume()
}

// CreateGame creates a new game instance with optional seed
func (m *Manager) CreateGame() (*Engine, error) {
	return m.CreateGameWithSeed(0)
}

// CreateGameWithSeed creates a new game instance with a specific seed
// If seed is 0, a random seed will be generated
func (m *Manager) CreateGameWithSeed(seed int64) (*Engine, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate random seed if not provided
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	gameID := uuid.New()
	engine := NewEngineWithSeed(gameID, m.config, m.balance, m.llmClient, m.promptBuilder, m.hub, seed)
	engine.SetHandlerRegistry(m.handlerRegistry)
	if m.pauseByDefault {
		engine.SetPaused(true)
	}
	m.games[gameID] = engine

	return engine, nil
}

// CreateSingleplayerGame creates a game with AI adversaries
func (m *Manager) CreateSingleplayerGame(playerPrompt string, adversaryTypes []string) (*Engine, uuid.UUID, error) {
	return m.CreateSingleplayerGameWithSeed(playerPrompt, adversaryTypes, 0, "")
}

// CreateSingleplayerGameWithSeed creates a game with AI adversaries and a specific seed.
// If seed is 0, a random seed will be generated.
// If mapSizeOverride is non-empty, it overrides the config's map size (e.g. "tiny", "small", "medium", "large", "huge", "massive").
func (m *Manager) CreateSingleplayerGameWithSeed(playerPrompt string, adversaryTypes []string, seed int64, mapSizeOverride string) (*Engine, uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate random seed if not provided
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	// Apply map size override from the request if provided
	cfg := m.config
	if mapSizeOverride != "" {
		cfg.Map.Size = mapSizeOverride
		cfg.MapSize = 0 // Clear direct map_size so Map.Size takes effect
	}

	gameID := uuid.New()
	engine := NewEngineWithSeed(gameID, cfg, m.balance, m.llmClient, m.promptBuilder, m.hub, seed)
	engine.SetHandlerRegistry(m.handlerRegistry)
	if m.pauseByDefault {
		engine.SetPaused(true)
	}

	// Generate spawn positions with passability check
	positions := generateSpawnPositionsForWorld(engine.GetWorld(), len(adversaryTypes)+1)

	// Add player agent
	playerAgent := NewAgentWithBalance(gameID, "Player", playerPrompt, positions[0], m.config.MaxMemoryItems, &m.balance)
	playerAgent.InitInventory(engine.itemRegistry)
	engine.agents[playerAgent.ID] = playerAgent

	// Add adversary agents
	for i, advType := range adversaryTypes {
		adversary := NewAdversaryAgentWithBalance(gameID, advType, positions[i+1], m.config.MaxMemoryItems, &m.balance)
		adversary.InitInventory(engine.itemRegistry)
		engine.agents[adversary.ID] = adversary
	}

	m.games[gameID] = engine

	return engine, playerAgent.ID, nil
}

// GetGame returns a game by ID
func (m *Manager) GetGame(gameID uuid.UUID) (*Engine, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	game, ok := m.games[gameID]
	if !ok {
		return nil, ErrGameNotFound
	}
	return game, nil
}

// ListGames returns all active games
func (m *Manager) ListGames() []*GameInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	games := make([]*GameInfo, 0, len(m.games))
	for id, engine := range m.games {
		games = append(games, &GameInfo{
			ID:          id,
			Status:      engine.GetStatus(),
			PlayerCount: len(engine.agents),
			MaxPlayers:  m.config.MaxPlayers,
			Tick:        engine.GetTick(),
		})
	}
	return games
}

// JoinGame adds a player agent to an existing game
func (m *Manager) JoinGame(gameID uuid.UUID, playerName, systemPrompt string) (*Agent, error) {
	m.mu.RLock()
	game, ok := m.games[gameID]
	m.mu.RUnlock()

	if !ok {
		return nil, ErrGameNotFound
	}

	// Find available spawn position
	pos := findAvailableSpawnPosition(game, m.config.GetMapSize())

	agent := NewAgentWithBalance(gameID, playerName, systemPrompt, pos, m.config.MaxMemoryItems, &m.balance)
	if err := game.AddAgent(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// StartGame starts a game
func (m *Manager) StartGame(gameID uuid.UUID) error {
	m.mu.RLock()
	game, ok := m.games[gameID]
	m.mu.RUnlock()

	if !ok {
		return ErrGameNotFound
	}

	return game.Start()
}

// StopGame stops a game
func (m *Manager) StopGame(gameID uuid.UUID) error {
	m.mu.RLock()
	game, ok := m.games[gameID]
	m.mu.RUnlock()

	if !ok {
		return ErrGameNotFound
	}

	game.Stop()
	return nil
}

// StopAll stops all running games
func (m *Manager) StopAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, game := range m.games {
		game.Stop()
	}
}

// RemoveGame removes a finished game
func (m *Manager) RemoveGame(gameID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.games, gameID)
}

// ForceTick triggers a tick for a specific game (dev only)
func (m *Manager) ForceTick(gameID uuid.UUID) error {
	m.mu.RLock()
	game, ok := m.games[gameID]
	m.mu.RUnlock()

	if !ok {
		return ErrGameNotFound
	}

	game.ForceTick()
	return nil
}

// GameInfo contains summary information about a game
type GameInfo struct {
	ID          uuid.UUID  `json:"id"`
	Status      GameStatus `json:"status"`
	PlayerCount int        `json:"player_count"`
	MaxPlayers  int        `json:"max_players"`
	Tick        int        `json:"tick"`
}

// generateSpawnPositions creates evenly distributed spawn positions (legacy, no terrain check)
func generateSpawnPositions(mapSize, count int) []Position {
	positions := make([]Position, count)

	if count == 1 {
		positions[0] = Position{X: mapSize / 2, Y: mapSize / 2}
		return positions
	}

	// Place agents in corners and edges for better distribution
	cornerOffsets := []Position{
		{X: 2, Y: 2},                         // Top-left
		{X: mapSize - 3, Y: 2},               // Top-right
		{X: 2, Y: mapSize - 3},               // Bottom-left
		{X: mapSize - 3, Y: mapSize - 3},     // Bottom-right
		{X: mapSize / 2, Y: 2},               // Top-center
		{X: mapSize / 2, Y: mapSize - 3},     // Bottom-center
		{X: 2, Y: mapSize / 2},               // Left-center
		{X: mapSize - 3, Y: mapSize / 2},     // Right-center
	}

	for i := 0; i < count && i < len(cornerOffsets); i++ {
		positions[i] = cornerOffsets[i]
	}

	// If more agents than preset positions, randomize remaining
	for i := len(cornerOffsets); i < count; i++ {
		positions[i] = Position{
			X: rand.Intn(mapSize-4) + 2,
			Y: rand.Intn(mapSize-4) + 2,
		}
	}

	return positions
}

// generateSpawnPositionsForWorld creates random spawn positions on passable terrain,
// spread across the map with minimum distance between spawns.
func generateSpawnPositionsForWorld(world *World, count int) []Position {
	mapSize := world.Size()
	positions := make([]Position, count)
	usedPositions := make(map[Position]bool)

	// Minimum distance between spawns scales with map size
	minDist := mapSize / (count + 1)
	if minDist < 5 {
		minDist = 5
	}

	for i := 0; i < count; i++ {
		pos := findRandomSpreadPassable(world, usedPositions, positions[:i], minDist)
		positions[i] = pos
		usedPositions[pos] = true
	}

	return positions
}

// findRandomSpreadPassable finds a random passable tile that is at least minDist
// from all existing spawn positions. Falls back to any passable tile if spacing
// cannot be satisfied.
func findRandomSpreadPassable(world *World, used map[Position]bool, existing []Position, minDist int) Position {
	mapSize := world.Size()
	margin := 4 // stay away from edges

	// Try with spacing constraint first (500 attempts)
	for attempts := 0; attempts < 500; attempts++ {
		pos := Position{
			X: rand.Intn(mapSize-margin*2) + margin,
			Y: rand.Intn(mapSize-margin*2) + margin,
		}

		tile := world.GetTile(pos)
		if tile == nil || !worldgen.IsPassableString(string(tile.Terrain)) || used[pos] {
			continue
		}

		// Check minimum distance to all existing spawns
		tooClose := false
		for _, ep := range existing {
			dx := pos.X - ep.X
			dy := pos.Y - ep.Y
			if dx*dx+dy*dy < minDist*minDist {
				tooClose = true
				break
			}
		}
		if !tooClose {
			return pos
		}
	}

	// Fall back to any passable tile (ignore spacing)
	return findRandomPassable(world, used)
}

// findNearestPassable finds the nearest passable tile to the given position using BFS
func findNearestPassable(world *World, start Position, used map[Position]bool) Position {
	// Check if start is already passable
	tile := world.GetTile(start)
	if tile != nil && worldgen.IsPassableString(string(tile.Terrain)) && !used[start] {
		return start
	}

	// BFS to find nearest passable tile
	visited := make(map[Position]bool)
	queue := []Position{start}
	visited[start] = true

	// Direction offsets for 8-directional search
	directions := []Position{
		{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0},
		{X: -1, Y: -1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: 1, Y: 1},
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dir := range directions {
			next := Position{X: current.X + dir.X, Y: current.Y + dir.Y}

			if visited[next] || !world.IsValidPosition(next) {
				continue
			}

			visited[next] = true
			tile := world.GetTile(next)

			if tile != nil && worldgen.IsPassableString(string(tile.Terrain)) && !used[next] {
				return next
			}

			queue = append(queue, next)
		}
	}

	// Fallback: return center (shouldn't happen with reasonable maps)
	return Position{X: world.Size() / 2, Y: world.Size() / 2}
}

// findRandomPassable finds a random passable tile not in the used set
func findRandomPassable(world *World, used map[Position]bool) Position {
	mapSize := world.Size()

	// Try random positions
	for attempts := 0; attempts < 1000; attempts++ {
		pos := Position{
			X: rand.Intn(mapSize-4) + 2,
			Y: rand.Intn(mapSize-4) + 2,
		}

		tile := world.GetTile(pos)
		if tile != nil && worldgen.IsPassableString(string(tile.Terrain)) && !used[pos] {
			return pos
		}
	}

	// Fallback: find any passable tile via BFS from center
	return findNearestPassable(world, Position{X: mapSize / 2, Y: mapSize / 2}, used)
}

// findAvailableSpawnPosition finds a position not occupied by other agents and on passable terrain
func findAvailableSpawnPosition(game *Engine, mapSize int) Position {
	occupied := make(map[Position]bool)
	for _, agent := range game.agents {
		occupied[agent.GetPosition()] = true
	}

	world := game.GetWorld()

	// Try preset positions first
	presets := generateSpawnPositions(mapSize, 8)
	for _, pos := range presets {
		tile := world.GetTile(pos)
		if !occupied[pos] && tile != nil && worldgen.IsPassableString(string(tile.Terrain)) {
			return pos
		}
	}

	// Fall back to random passable positions
	for attempts := 0; attempts < 100; attempts++ {
		pos := Position{
			X: rand.Intn(mapSize-4) + 2,
			Y: rand.Intn(mapSize-4) + 2,
		}
		tile := world.GetTile(pos)
		if !occupied[pos] && tile != nil && worldgen.IsPassableString(string(tile.Terrain)) {
			return pos
		}
	}

	// Last resort: find nearest passable to center
	return findNearestPassable(world, Position{X: mapSize / 2, Y: mapSize / 2}, occupied)
}
