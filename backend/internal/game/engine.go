package game

import (
	"context"
	"log"
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
	mu            sync.RWMutex
	ID            uuid.UUID
	config        config.GameConfig
	world         *World
	agents        map[uuid.UUID]*Agent
	messages      []GameMessage
	llmClient     LLMClient
	promptBuilder PromptBuilder
	broadcaster   Broadcaster
	status        GameStatus
	tick          int
	cancel        context.CancelFunc
	resolver      *ConflictResolver
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
	return NewEngineWithSeed(id, cfg, llmClient, promptBuilder, broadcaster, 0)
}

// NewEngineWithSeed creates a new game engine with procedurally generated terrain
func NewEngineWithSeed(id uuid.UUID, cfg config.GameConfig, llmClient LLMClient, promptBuilder PromptBuilder, broadcaster Broadcaster, seed int64) *Engine {
	var world *World
	if seed == 0 {
		// Use flat plains world when no seed
		world = NewWorld(cfg.MapSize)
	} else {
		// Generate procedural terrain with seed
		gen := worldgen.NewWorldGenerator(seed)
		tileData := gen.Generate(cfg.MapSize)

		// Convert worldgen.TileData to game.Tile
		tiles := make([][]*Tile, cfg.MapSize)
		for y := 0; y < cfg.MapSize; y++ {
			tiles[y] = make([]*Tile, cfg.MapSize)
			for x := 0; x < cfg.MapSize; x++ {
				tiles[y][x] = &Tile{
					Position: Position{X: x, Y: y},
					OwnerID:  nil,
					Terrain:  TerrainType(tileData[y][x].Terrain),
				}
			}
		}

		world = NewWorldWithSeed(cfg.MapSize, seed, tiles)
	}

	return &Engine{
		ID:            id,
		config:        cfg,
		world:         world,
		agents:        make(map[uuid.UUID]*Agent),
		messages:      make([]GameMessage, 0),
		llmClient:     llmClient,
		promptBuilder: promptBuilder,
		broadcaster:   broadcaster,
		status:        StatusWaiting,
		tick:          0,
		resolver:      NewConflictResolver(),
	}
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

	e.agents[agent.ID] = agent
	return nil
}

// RemoveAgent removes an agent from the game
func (e *Engine) RemoveAgent(agentID uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.agents, agentID)
}

// Start begins the game loop
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
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.mu.Unlock()

	go e.runLoop(ctx)
	return nil
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

// processTick handles a single game tick
func (e *Engine) processTick(ctx context.Context) {
	e.mu.Lock()
	e.tick++
	tick := e.tick
	agents := make([]*Agent, 0, len(e.agents))
	for _, a := range e.agents {
		agents = append(agents, a)
	}
	e.mu.Unlock()

	log.Printf("Game %s: Processing tick %d", e.ID, tick)

	// Build context for each agent
	contexts := e.buildAgentContexts(agents)

	// Fan-out LLM requests with timeout
	llmCtx, cancel := context.WithTimeout(ctx, e.config.TickDuration-2*time.Second)
	actions := e.requestActions(llmCtx, contexts)
	cancel()

	// Add actions to resolver
	e.resolver.AddActions(actions)

	// Resolve conflicts (first-come priority)
	orderedActions := e.resolver.Resolve()

	// Process actions
	processor := NewActionProcessor(e.world, e.agents)
	results := processor.ProcessAll(orderedActions)

	// Collect messages from this tick
	tickMessages := e.collectMessages(orderedActions)

	// Build tick update
	update := e.buildTickUpdate(tick, orderedActions, results, tickMessages)

	// Broadcast to all connected clients
	if e.broadcaster != nil {
		e.broadcaster.BroadcastToGame(e.ID, update)
	}

	// Check win condition
	if tick >= e.config.WinAfterTicks {
		e.endGame()
	}
}

// buildAgentContexts creates context for each agent
func (e *Engine) buildAgentContexts(agents []*Agent) []AgentContext {
	contexts := make([]AgentContext, len(agents))

	e.mu.RLock()
	tickMessages := e.getMessagesForTick(e.tick - 1)
	e.mu.RUnlock()

	for i, agent := range agents {
		pos := agent.GetPosition()
		currentTile := e.world.GetTile(pos)

		// Determine current tile ownership status
		currentTileOwned := false
		currentTileEnemy := false
		if currentTile != nil && currentTile.OwnerID != nil {
			if *currentTile.OwnerID == agent.ID {
				currentTileOwned = true
			} else {
				currentTileEnemy = true
			}
		}

		contexts[i] = AgentContext{
			Agent:            agent,
			VisibleTiles:     e.world.GetVisibleTiles(pos, e.config.VisionRadius),
			OwnedCount:       e.world.CountOwnedTiles(agent.ID),
			Messages:         e.filterMessagesForAgent(tickMessages, agent.ID),
			CurrentTick:      e.tick,
			WorldSize:        e.config.MapSize,
			CurrentTileOwned: currentTileOwned,
			CurrentTileEnemy: currentTileEnemy,
		}
	}

	return contexts
}

// requestActions gets actions from all agents in parallel
func (e *Engine) requestActions(ctx context.Context, contexts []AgentContext) []Action {
	var wg sync.WaitGroup
	actions := make([]Action, len(contexts))

	for i, agentCtx := range contexts {
		wg.Add(1)
		go func(idx int, actx AgentContext) {
			defer wg.Done()

			prompt := e.promptBuilder.BuildPrompt(actx)
			action, err := e.llmClient.GetAction(ctx, actx.Agent.ID, prompt)
			if err != nil {
				log.Printf("LLM error for agent %s: %v", actx.Agent.Name, err)
				action = WaitAction(actx.Agent.ID)
			}
			action.ReceivedAt = time.Now()
			actions[idx] = action
		}(i, agentCtx)
	}

	wg.Wait()
	return actions
}

// collectMessages extracts messages from actions
func (e *Engine) collectMessages(actions []Action) []GameMessage {
	messages := make([]GameMessage, 0)

	for _, action := range actions {
		if action.Type == ActionMessage && action.Params.Message != "" {
			msg := GameMessage{
				Tick:        e.tick,
				FromAgentID: action.AgentID,
				ToAgentID:   action.Params.Target,
				Content:     action.Params.Message,
			}
			messages = append(messages, msg)
		}
	}

	e.mu.Lock()
	e.messages = append(e.messages, messages...)
	e.mu.Unlock()

	return messages
}

// getMessagesForTick returns messages from a specific tick
func (e *Engine) getMessagesForTick(tick int) []GameMessage {
	result := make([]GameMessage, 0)
	for _, msg := range e.messages {
		if msg.Tick == tick {
			result = append(result, msg)
		}
	}
	return result
}

// filterMessagesForAgent filters messages that an agent should receive
func (e *Engine) filterMessagesForAgent(messages []GameMessage, agentID uuid.UUID) []IncomingMessage {
	result := make([]IncomingMessage, 0)

	for _, msg := range messages {
		// Skip own messages
		if msg.FromAgentID == agentID {
			continue
		}

		// Include if broadcast or targeted to this agent
		if msg.ToAgentID == nil || *msg.ToAgentID == agentID {
			fromAgent, ok := e.agents[msg.FromAgentID]
			fromName := "Unknown"
			if ok {
				fromName = fromAgent.Name
			}

			result = append(result, IncomingMessage{
				FromAgentID:   msg.FromAgentID,
				FromAgentName: fromName,
				Content:       msg.Content,
				IsBroadcast:   msg.ToAgentID == nil,
			})
		}
	}

	return result
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
	Tiles    []TileChange    `json:"tiles"`
	Agents   []AgentSnapshot `json:"agents"`
	Messages []GameMessage   `json:"messages"`
	Results  []ActionResult  `json:"results"`
}

// TileChange represents a tile ownership change
type TileChange struct {
	X       int        `json:"x"`
	Y       int        `json:"y"`
	OwnerID *uuid.UUID `json:"owner_id"`
}

// buildTickUpdate creates the tick update message
func (e *Engine) buildTickUpdate(tick int, actions []Action, results []ActionResult, messages []GameMessage) TickUpdate {
	// Collect tile changes from claim actions
	tileChanges := make([]TileChange, 0)
	for _, result := range results {
		if result.Action == ActionClaim && result.Success && result.ClaimedAt != nil {
			agent := e.agents[result.AgentID]
			if agent != nil {
				tileChanges = append(tileChanges, TileChange{
					X:       result.ClaimedAt.X,
					Y:       result.ClaimedAt.Y,
					OwnerID: &agent.ID,
				})
			}
		}
	}

	// Collect agent snapshots
	agentSnapshots := make([]AgentSnapshot, 0, len(e.agents))
	for _, agent := range e.agents {
		agentSnapshots = append(agentSnapshots, agent.Snapshot())
	}

	return TickUpdate{
		Type:   "tick",
		Tick:   tick,
		GameID: e.ID,
		Changes: TickChanges{
			Tiles:    tileChanges,
			Agents:   agentSnapshots,
			Messages: messages,
			Results:  results,
		},
	}
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

	return FullGameState{
		Type:     "full_state",
		GameID:   e.ID,
		Tick:     e.tick,
		Status:   e.status,
		World:    e.world.Snapshot(),
		Agents:   agents,
		Messages: e.messages,
	}
}

// FullGameState represents the complete game state
type FullGameState struct {
	Type     string          `json:"type"`
	GameID   uuid.UUID       `json:"game_id"`
	Tick     int             `json:"tick"`
	Status   GameStatus      `json:"status"`
	World    WorldSnapshot   `json:"world"`
	Agents   []AgentSnapshot `json:"agents"`
	Messages []GameMessage   `json:"messages"`
}

// ForceTick manually triggers the next tick (for dev/testing)
func (e *Engine) ForceTick() {
	if e.status == StatusRunning {
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
