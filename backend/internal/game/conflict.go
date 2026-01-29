package game

import (
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
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
	balance        *config.BalanceConfig
	registry       *HandlerRegistry
}

// NewActionProcessor creates a new action processor with all dependencies
func NewActionProcessor(world *World, agents map[uuid.UUID]*Agent, worldObjects *WorldObjectManager, itemRegistry *ItemRegistry, recipeRegistry *RecipeRegistry, currentTick int, balance *config.BalanceConfig, registry *HandlerRegistry) *ActionProcessor {
	return &ActionProcessor{
		world:          world,
		agents:         agents,
		worldObjects:   worldObjects,
		itemRegistry:   itemRegistry,
		recipeRegistry: recipeRegistry,
		currentTick:    currentTick,
		balance:        balance,
		registry:       registry,
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

	handler, found := ap.registry.Get(action.Type)
	if !found {
		return ActionResult{
			AgentID: action.AgentID,
			Action:  action.Type,
			Success: false,
			Message: "unknown action type",
		}
	}

	ctx := NewActionContext(
		agent,
		ap.world,
		ap.worldObjects,
		ap.itemRegistry,
		ap.recipeRegistry,
		ap.agents,
		ap.currentTick,
		action,
		ap.balance,
	)
	return handler.Process(ctx)
}

// ProcessAll applies all actions and returns results
func (ap *ActionProcessor) ProcessAll(actions []Action) []ActionResult {
	results := make([]ActionResult, len(actions))
	for i, action := range actions {
		results[i] = ap.Process(action)
		results[i].Reasoning = action.Reasoning
	}
	return results
}
