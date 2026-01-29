package game

import (
	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
)

// ActionHandler interface that all action handlers must implement
type ActionHandler interface {
	// ActionType returns the type of action this handler processes
	ActionType() ActionType

	// Validate checks if the action can be performed in the current context
	// Returns nil if valid, or an error describing why the action is invalid
	Validate(ctx *ActionContext) error

	// Process executes the action and returns the result
	Process(ctx *ActionContext) ActionResult
}

// ActionContext contains all dependencies needed to process an action
type ActionContext struct {
	Agent          *Agent
	World          *World
	WorldObjects   *WorldObjectManager
	ItemRegistry   *ItemRegistry
	RecipeRegistry *RecipeRegistry
	Agents         map[uuid.UUID]*Agent
	CurrentTick    int
	Action         Action
	Balance        *config.BalanceConfig
}

// NewActionContext creates a new action context with all required dependencies
func NewActionContext(
	agent *Agent,
	world *World,
	worldObjects *WorldObjectManager,
	itemRegistry *ItemRegistry,
	recipeRegistry *RecipeRegistry,
	agents map[uuid.UUID]*Agent,
	currentTick int,
	action Action,
	balance *config.BalanceConfig,
) *ActionContext {
	return &ActionContext{
		Agent:          agent,
		World:          world,
		WorldObjects:   worldObjects,
		ItemRegistry:   itemRegistry,
		RecipeRegistry: recipeRegistry,
		Agents:         agents,
		CurrentTick:    currentTick,
		Action:         action,
		Balance:        balance,
	}
}

// HandlerRegistry manages all action handlers
type HandlerRegistry struct {
	handlers map[ActionType]ActionHandler
}

// NewHandlerRegistry creates a new empty handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[ActionType]ActionHandler),
	}
}

// Register adds a handler to the registry
func (r *HandlerRegistry) Register(h ActionHandler) {
	r.handlers[h.ActionType()] = h
}

// Get retrieves a handler for the given action type
func (r *HandlerRegistry) Get(t ActionType) (ActionHandler, bool) {
	h, ok := r.handlers[t]
	return h, ok
}

// Has checks if a handler exists for the given action type
func (r *HandlerRegistry) Has(t ActionType) bool {
	_, ok := r.handlers[t]
	return ok
}

// FailedResult creates a failed action result with the given message
func FailedResult(agentID uuid.UUID, actionType ActionType, message string) ActionResult {
	return ActionResult{
		AgentID: agentID,
		Action:  actionType,
		Success: false,
		Message: message,
	}
}

// SuccessResult creates a successful action result with the given message
func SuccessResult(agentID uuid.UUID, actionType ActionType, message string) ActionResult {
	return ActionResult{
		AgentID: agentID,
		Action:  actionType,
		Success: true,
		Message: message,
	}
}
