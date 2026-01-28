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
	mu             sync.RWMutex
	ID             uuid.UUID
	config         config.GameConfig
	world          *World
	agents         map[uuid.UUID]*Agent
	messages       []GameMessage
	llmClient      LLMClient
	promptBuilder  PromptBuilder
	broadcaster    Broadcaster
	status         GameStatus
	tick           int
	cancel         context.CancelFunc
	resolver       *ConflictResolver
	worldObjects   *WorldObjectManager
	itemRegistry   *ItemRegistry
	recipeRegistry *RecipeRegistry
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

	// Initialize registries
	itemRegistry := DefaultItemRegistry()
	recipeRegistry := DefaultRecipeRegistry()
	worldObjects := NewWorldObjectManager()

	engine := &Engine{
		ID:             id,
		config:         cfg,
		world:          world,
		agents:         make(map[uuid.UUID]*Agent),
		messages:       make([]GameMessage, 0),
		llmClient:      llmClient,
		promptBuilder:  promptBuilder,
		broadcaster:    broadcaster,
		status:         StatusWaiting,
		tick:           0,
		resolver:       NewConflictResolver(),
		worldObjects:   worldObjects,
		itemRegistry:   itemRegistry,
		recipeRegistry: recipeRegistry,
	}

	// Populate world with resources and interactives if using a seed
	if seed != 0 {
		populator := NewWorldPopulator(seed, world, worldObjects)
		populator.PopulateWorld()
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

	// Phase 1: Process respawns
	respawnedAgents := e.processRespawns(tick)

	// Phase 2: Passive energy income (before actions)
	e.processPassiveIncome()

	// Build context for each agent (skip dead agents)
	aliveAgents := make([]*Agent, 0, len(agents))
	for _, a := range agents {
		if !a.IsDead {
			aliveAgents = append(aliveAgents, a)
		}
	}
	contexts := e.buildAgentContexts(aliveAgents)

	// Fan-out LLM requests with timeout
	llmCtx, cancel := context.WithTimeout(ctx, e.config.TickDuration-2*time.Second)
	actions := e.requestActions(llmCtx, contexts)
	cancel()

	// Add actions to resolver
	e.resolver.AddActions(actions)

	// Resolve conflicts (first-come priority)
	orderedActions := e.resolver.Resolve()

	// Process actions with full processor
	processor := NewActionProcessorFull(e.world, e.agents, e.worldObjects, e.itemRegistry, e.recipeRegistry, tick)
	results := processor.ProcessAll(orderedActions)

	// Phase 3: Trigger traps after movement
	trapResults := e.processTrapTriggers(results)
	results = append(results, trapResults...)

	// Phase 4: Process despawns
	removedObjects := e.worldObjects.ProcessDespawns(tick)
	depletedResources := e.worldObjects.ProcessDepletedResources()
	removedObjects = append(removedObjects, depletedResources...)

	// Collect messages from this tick
	tickMessages := e.collectMessages(orderedActions)

	// Build tick update
	update := e.buildTickUpdate(tick, orderedActions, results, tickMessages, removedObjects, respawnedAgents)

	// Broadcast to all connected clients
	if e.broadcaster != nil {
		e.broadcaster.BroadcastToGame(e.ID, update)
	}

	// Check win condition
	if tick >= e.config.WinAfterTicks {
		e.endGame()
	}
}

// processRespawns handles agent respawns
func (e *Engine) processRespawns(tick int) []uuid.UUID {
	respawned := make([]uuid.UUID, 0)

	for _, agent := range e.agents {
		if agent.ShouldRespawn(tick) {
			// Find a valid spawn position at map edge
			pos := e.findEdgeSpawnPosition()
			agent.Respawn(pos)
			respawned = append(respawned, agent.ID)
			log.Printf("Agent %s respawned at (%d,%d)", agent.Name, pos.X, pos.Y)
		}
	}

	return respawned
}

// findEdgeSpawnPosition finds a valid spawn position at the map edge
func (e *Engine) findEdgeSpawnPosition() Position {
	size := e.world.Size()
	edges := []Position{}

	// Collect all passable edge positions
	for x := 0; x < size; x++ {
		// Top edge
		pos := Position{X: x, Y: 0}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
		// Bottom edge
		pos = Position{X: x, Y: size - 1}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
	}
	for y := 1; y < size-1; y++ {
		// Left edge
		pos := Position{X: 0, Y: y}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
		// Right edge
		pos = Position{X: size - 1, Y: y}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
	}

	if len(edges) == 0 {
		// Fallback to center if no edges available
		return Position{X: size / 2, Y: size / 2}
	}

	// Pick a random edge position
	return edges[time.Now().UnixNano()%int64(len(edges))]
}

// isValidSpawnPosition checks if a position is valid for spawning
func (e *Engine) isValidSpawnPosition(pos Position) bool {
	tile := e.world.GetTile(pos)
	if tile == nil {
		return false
	}
	if tile.Terrain == TerrainWater || tile.Terrain == TerrainMountain {
		return false
	}
	// Check for blocking structures
	if e.worldObjects.HasBlockingObject(pos) {
		return false
	}
	// Check for other agents
	for _, agent := range e.agents {
		if !agent.IsDead && agent.GetPosition() == pos {
			return false
		}
	}
	return true
}

// processPassiveIncome gives agents energy based on owned tiles
func (e *Engine) processPassiveIncome() {
	for _, agent := range e.agents {
		if agent.IsDead {
			continue
		}

		ownedTiles := e.world.GetOwnedTiles(agent.ID)
		energyGain := 0

		for _, pos := range ownedTiles {
			tile := e.world.GetTile(pos)
			if tile == nil {
				continue
			}
			switch tile.Terrain {
			case TerrainPlains:
				energyGain += 1
			case TerrainForest:
				energyGain += 2
			}
		}

		if energyGain > 0 {
			agent.AddEnergy(energyGain)
		}
	}
}

// processTrapTriggers checks for and triggers traps after movement
func (e *Engine) processTrapTriggers(results []ActionResult) []ActionResult {
	trapResults := make([]ActionResult, 0)

	for _, result := range results {
		if result.Action == ActionMove && result.Success && result.NewPos != nil {
			agent := e.agents[result.AgentID]
			if agent == nil || agent.IsDead {
				continue
			}

			// Check for traps at new position
			traps := e.worldObjects.GetTrapsAt(*result.NewPos)
			for _, trap := range traps {
				// Don't trigger own traps
				if trap.OwnerID != nil && *trap.OwnerID == agent.ID {
					continue
				}

				// Trigger trap
				damage := trap.Damage
				if damage == 0 {
					damage = 1
				}

				killed := agent.TakeDamage(damage)
				trapResult := ActionResult{
					AgentID:     agent.ID,
					Action:      ActionFight, // Use FIGHT to indicate damage
					Success:     true,
					DamageDealt: damage,
					Message:     "triggered a trap",
				}

				if killed {
					agent.Kill(e.tick)
					// Clear tiles on death
					ownedTiles := e.world.GetOwnedTiles(agent.ID)
					for _, pos := range ownedTiles {
						e.world.SetOwner(pos, nil)
					}
					if agent.Inventory != nil {
						agent.Inventory.Clear()
					}
					trapResult.Message = "killed by a trap"
				}

				trapResults = append(trapResults, trapResult)

				// Remove the trap after triggering
				e.worldObjects.Remove(trap.ID)
			}
		}
	}

	return trapResults
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

		// Calculate effective vision (base + upgrades + beacons)
		visionRadius := agent.GetEffectiveVision(e.config.VisionRadius)

		// Add vision bonus from nearby beacons
		beacons := e.worldObjects.GetByOwner(agent.ID)
		for _, beacon := range beacons {
			if beacon.Type == ObjectStructure && beacon.StructureType == StructureBeacon {
				// Check if beacon is within vision range
				dx := beacon.Position.X - pos.X
				dy := beacon.Position.Y - pos.Y
				if dx >= -visionRadius && dx <= visionRadius && dy >= -visionRadius && dy <= visionRadius {
					visionRadius += beacon.VisionBonus
				}
			}
		}

		// Calculate energy income per tick
		ownedTiles := e.world.GetOwnedTiles(agent.ID)
		energyPerTick := 0
		for _, tilePos := range ownedTiles {
			tile := e.world.GetTile(tilePos)
			if tile != nil {
				switch tile.Terrain {
				case TerrainPlains:
					energyPerTick += 1
				case TerrainForest:
					energyPerTick += 2
				}
			}
		}

		// Get visible objects (excluding hidden traps from other players)
		visibleObjects := e.worldObjects.GetVisibleObjects(pos, visionRadius, &agent.ID)

		// Get visible agents
		visibleAgents := make([]*AgentSnapshot, 0)
		for _, other := range e.agents {
			if other.ID == agent.ID || other.IsDead {
				continue
			}
			otherPos := other.GetPosition()
			dx := otherPos.X - pos.X
			dy := otherPos.Y - pos.Y
			if dx >= -visionRadius && dx <= visionRadius && dy >= -visionRadius && dy <= visionRadius {
				snap := other.Snapshot()
				visibleAgents = append(visibleAgents, &snap)
			}
		}

		contexts[i] = AgentContext{
			Agent:            agent,
			VisibleTiles:     e.world.GetVisibleTiles(pos, visionRadius),
			VisibleObjects:   visibleObjects,
			VisibleAgents:    visibleAgents,
			OwnedCount:       len(ownedTiles),
			Messages:         e.filterMessagesForAgent(tickMessages, agent.ID),
			CurrentTick:      e.tick,
			WorldSize:        e.config.MapSize,
			CurrentTileOwned: currentTileOwned,
			CurrentTileEnemy: currentTileEnemy,
			EnergyPerTick:    energyPerTick,
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
	Tiles          []TileChange          `json:"tiles"`
	Agents         []AgentSnapshot       `json:"agents"`
	Messages       []GameMessage         `json:"messages"`
	Results        []ActionResult        `json:"results"`
	ObjectsAdded   []WorldObjectSnapshot `json:"objects_added,omitempty"`
	ObjectsRemoved []uuid.UUID           `json:"objects_removed,omitempty"`
	Respawned      []uuid.UUID           `json:"respawned,omitempty"`
}

// TileChange represents a tile ownership change
type TileChange struct {
	X       int        `json:"x"`
	Y       int        `json:"y"`
	OwnerID *uuid.UUID `json:"owner_id"`
}

// buildTickUpdate creates the tick update message
func (e *Engine) buildTickUpdate(tick int, actions []Action, results []ActionResult, messages []GameMessage, removedObjects []uuid.UUID, respawnedAgents []uuid.UUID) TickUpdate {
	// Collect tile changes from claim actions and death-related changes
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

	// Collect newly added objects (structures placed this tick)
	addedObjects := make([]WorldObjectSnapshot, 0)
	for _, result := range results {
		if result.Action == ActionPlace && result.Success {
			// Find the structure that was just placed
			agent := e.agents[result.AgentID]
			if agent != nil {
				pos := agent.GetPosition()
				structure := e.worldObjects.GetStructureAt(pos)
				if structure != nil && structure.CreatedTick == tick {
					addedObjects = append(addedObjects, structure.Snapshot())
				}
			}
		}
	}

	return TickUpdate{
		Type:   "tick",
		Tick:   tick,
		GameID: e.ID,
		Changes: TickChanges{
			Tiles:          tileChanges,
			Agents:         agentSnapshots,
			Messages:       messages,
			Results:        results,
			ObjectsAdded:   addedObjects,
			ObjectsRemoved: removedObjects,
			Respawned:      respawnedAgents,
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

// FullGameState represents the complete game state
type FullGameState struct {
	Type         string                `json:"type"`
	GameID       uuid.UUID             `json:"game_id"`
	Tick         int                   `json:"tick"`
	Status       GameStatus            `json:"status"`
	World        WorldSnapshot         `json:"world"`
	Agents       []AgentSnapshot       `json:"agents"`
	Messages     []GameMessage         `json:"messages"`
	WorldObjects []WorldObjectSnapshot `json:"world_objects,omitempty"`
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
