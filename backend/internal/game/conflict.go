package game

import (
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// ConflictResolver handles simultaneous action conflicts
type ConflictResolver struct {
	mu      sync.Mutex
	actions []Action
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		actions: make([]Action, 0),
	}
}

// AddAction adds an action to be resolved
func (cr *ConflictResolver) AddAction(action Action) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.actions = append(cr.actions, action)
}

// AddActions adds multiple actions at once
func (cr *ConflictResolver) AddActions(actions []Action) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.actions = append(cr.actions, actions...)
}

// Resolve processes all actions with first-come priority
// Returns actions sorted by arrival time
func (cr *ConflictResolver) Resolve() []Action {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Sort by arrival time (first-come priority)
	sort.Slice(cr.actions, func(i, j int) bool {
		return cr.actions[i].ReceivedAt.Before(cr.actions[j].ReceivedAt)
	})

	result := cr.actions
	cr.actions = make([]Action, 0)
	return result
}

// Clear removes all pending actions
func (cr *ConflictResolver) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.actions = make([]Action, 0)
}

// ActionProcessor applies actions to the game state
type ActionProcessor struct {
	world          *World
	agents         map[uuid.UUID]*Agent
	worldObjects   *WorldObjectManager
	itemRegistry   *ItemRegistry
	recipeRegistry *RecipeRegistry
	currentTick    int
}

// NewActionProcessor creates a new action processor
func NewActionProcessor(world *World, agents map[uuid.UUID]*Agent) *ActionProcessor {
	return &ActionProcessor{
		world:  world,
		agents: agents,
	}
}

// NewActionProcessorFull creates an action processor with all dependencies
func NewActionProcessorFull(world *World, agents map[uuid.UUID]*Agent, worldObjects *WorldObjectManager, itemRegistry *ItemRegistry, recipeRegistry *RecipeRegistry, currentTick int) *ActionProcessor {
	return &ActionProcessor{
		world:          world,
		agents:         agents,
		worldObjects:   worldObjects,
		itemRegistry:   itemRegistry,
		recipeRegistry: recipeRegistry,
		currentTick:    currentTick,
	}
}

// Process applies an action and returns the result
func (ap *ActionProcessor) Process(action Action) ActionResult {
	agent, ok := ap.agents[action.AgentID]
	if !ok {
		return ActionResult{
			AgentID: action.AgentID,
			Action:  action.Type,
			Success: false,
			Message: "agent not found",
		}
	}

	// Dead agents can only wait
	if agent.IsDead {
		return ActionResult{
			AgentID: agent.ID,
			Action:  action.Type,
			Success: false,
			Message: "agent is dead",
		}
	}

	switch action.Type {
	case ActionMove:
		return ap.processMove(agent, action.Params.Direction)
	case ActionClaim:
		return ap.processClaim(agent)
	case ActionMessage:
		return ap.processMessage(agent, action.Params.Target, action.Params.Message)
	case ActionWait, ActionHold:
		return ActionResult{
			AgentID: action.AgentID,
			Action:  action.Type,
			Success: true,
			Message: "holding position",
		}
	case ActionFight:
		return ap.processFight(agent, action.Params.Target)
	case ActionPickup:
		return ap.processPickup(agent)
	case ActionDrop:
		return ap.processDrop(agent, action.Params.ItemID, action.Params.Quantity)
	case ActionUse:
		return ap.processUse(agent, action.Params.ItemID)
	case ActionPlace:
		return ap.processPlace(agent, action.Params.ItemID)
	case ActionCraft:
		return ap.processCraft(agent, action.Params.RecipeID)
	case ActionHarvest:
		return ap.processHarvest(agent)
	case ActionScan:
		return ap.processScan(agent)
	case ActionInteract:
		return ap.processInteract(agent, action.Params.Message)
	case ActionUpgrade:
		return ap.processUpgrade(agent, action.Params.UpgradeType)
	default:
		return ActionResult{
			AgentID: action.AgentID,
			Action:  action.Type,
			Success: false,
			Message: "unknown action type",
		}
	}
}

// processMove handles a move action
func (ap *ActionProcessor) processMove(agent *Agent, dir Direction) ActionResult {
	oldPos := agent.GetPosition()
	newPos := GetNewPosition(oldPos, dir)

	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionMove,
		OldPos:  &oldPos,
	}

	// Check if new position is valid
	if !ap.world.IsValidPosition(newPos) {
		result.Success = false
		result.Message = "cannot move out of bounds"
		return result
	}

	// Check if tile is passable (not water or mountain)
	tile := ap.world.GetTile(newPos)
	if tile.Terrain == TerrainWater || tile.Terrain == TerrainMountain {
		result.Success = false
		result.Message = "cannot move to impassable terrain"
		return result
	}

	// Check for blocking structures
	if ap.worldObjects != nil && ap.worldObjects.HasBlockingObject(newPos) {
		result.Success = false
		result.Message = "path blocked by structure"
		return result
	}

	// Check if another agent is already there (skip dead agents)
	for _, other := range ap.agents {
		if other.ID != agent.ID && !other.IsDead && other.GetPosition() == newPos {
			result.Success = false
			result.Message = "tile occupied by another agent"
			return result
		}
	}

	// Move successful
	agent.SetPosition(newPos)
	result.Success = true
	result.NewPos = &newPos
	result.Message = "moved successfully"

	// Add to memory
	agent.AddMemory("Moved " + string(dir) + " to (" + posString(newPos) + ")")

	return result
}

// posString formats a position for memory/display
func posString(pos Position) string {
	return fmt.Sprintf("%d,%d", pos.X, pos.Y)
}

// processClaim handles a claim action
func (ap *ActionProcessor) processClaim(agent *Agent) ActionResult {
	pos := agent.GetPosition()

	result := ActionResult{
		AgentID:   agent.ID,
		Action:    ActionClaim,
		ClaimedAt: &pos,
	}

	tile := ap.world.GetTile(pos)
	if tile == nil {
		result.Success = false
		result.Message = "invalid position"
		return result
	}

	// Check if already owned by this agent
	if tile.OwnerID != nil && *tile.OwnerID == agent.ID {
		result.Success = false
		result.Message = "already own this tile"
		return result
	}

	// Claim the tile
	ap.world.SetOwner(pos, &agent.ID)
	result.Success = true

	if tile.OwnerID != nil {
		result.Message = "captured tile from enemy"
		agent.AddMemory("Captured enemy tile at (" + posString(pos) + ")")
	} else {
		result.Message = "claimed new tile"
		agent.AddMemory("Claimed tile at (" + posString(pos) + ")")
	}

	return result
}

// processMessage handles a message action
func (ap *ActionProcessor) processMessage(agent *Agent, target *uuid.UUID, message string) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionMessage,
		Success: true,
	}

	if target == nil {
		result.Message = "broadcast message sent"
		agent.AddMemory("Broadcast: " + message)
	} else {
		if _, ok := ap.agents[*target]; ok {
			result.Message = "message sent"
			agent.AddMemory("Sent message to agent: " + message)
		} else {
			result.Success = false
			result.Message = "target agent not found"
		}
	}

	return result
}

// ProcessAll applies all actions and returns results
func (ap *ActionProcessor) ProcessAll(actions []Action) []ActionResult {
	results := make([]ActionResult, len(actions))
	for i, action := range actions {
		results[i] = ap.Process(action)
	}
	return results
}
