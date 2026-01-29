package testutil

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
	"github.com/lucas/promptlands/internal/game"
	"github.com/lucas/promptlands/internal/game/actions"
)

// NewTestWorld creates a simple test world with flat plains terrain
func NewTestWorld(size int) *game.World {
	return game.NewWorld(size)
}

// NewTestAgent creates a test agent at a given position
func NewTestAgent(pos game.Position) *game.Agent {
	return NewTestAgentWithBalance(pos, nil)
}

// NewTestAgentWithBalance creates a test agent with balance config
func NewTestAgentWithBalance(pos game.Position, balance *config.BalanceConfig) *game.Agent {
	gameID := uuid.New()
	name := "TestAgent"
	systemPrompt := "You are a test agent."
	maxMemory := 10

	return game.NewAgentWithBalance(gameID, name, systemPrompt, pos, maxMemory, balance)
}

// NewTestActionContext creates a context for testing handlers
func NewTestActionContext(agent *game.Agent, world *game.World, action game.Action) *actions.ActionContext {
	return NewTestActionContextWithBalance(agent, world, action, nil)
}

// NewTestActionContextWithBalance creates a context for testing handlers with balance config
func NewTestActionContextWithBalance(agent *game.Agent, world *game.World, action game.Action, balance *config.BalanceConfig) *actions.ActionContext {
	worldObjects := game.NewWorldObjectManager()
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	agents := map[uuid.UUID]*game.Agent{
		agent.ID: agent,
	}

	// Initialize agent inventory if not already initialized
	if agent.Inventory == nil {
		agent.InitInventory(itemRegistry)
	}

	return actions.NewActionContext(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		1, // currentTick
		action,
		balance,
	)
}

// NewTestActionContextFull creates a fully customized context for testing handlers
func NewTestActionContextFull(
	agent *game.Agent,
	world *game.World,
	worldObjects *game.WorldObjectManager,
	itemRegistry *game.ItemRegistry,
	recipeRegistry *game.RecipeRegistry,
	agents map[uuid.UUID]*game.Agent,
	currentTick int,
	action game.Action,
	balance *config.BalanceConfig,
) *actions.ActionContext {
	// Initialize agent inventory if not already initialized
	if agent.Inventory == nil && itemRegistry != nil {
		agent.InitInventory(itemRegistry)
	}

	return actions.NewActionContext(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		currentTick,
		action,
		balance,
	)
}

// AssertAgentAt asserts an agent is at a specific position
func AssertAgentAt(t *testing.T, agent *game.Agent, pos game.Position) {
	t.Helper()
	actual := agent.GetPosition()
	if actual.X != pos.X || actual.Y != pos.Y {
		t.Errorf("expected agent at (%d, %d), got (%d, %d)", pos.X, pos.Y, actual.X, actual.Y)
	}
}

// AssertAgentHP asserts an agent has specific HP
func AssertAgentHP(t *testing.T, agent *game.Agent, hp int) {
	t.Helper()
	actual := agent.GetHP()
	if actual != hp {
		t.Errorf("expected agent HP %d, got %d", hp, actual)
	}
}

// AssertAgentEnergy asserts an agent has specific energy
func AssertAgentEnergy(t *testing.T, agent *game.Agent, energy int) {
	t.Helper()
	actual := agent.GetEnergy()
	if actual != energy {
		t.Errorf("expected agent energy %d, got %d", energy, actual)
	}
}

// AssertActionSuccess asserts an action result was successful
func AssertActionSuccess(t *testing.T, result game.ActionResult) {
	t.Helper()
	if !result.Success {
		t.Errorf("expected action to succeed, but it failed: %s", result.Message)
	}
}

// AssertActionFailed asserts an action result failed
func AssertActionFailed(t *testing.T, result game.ActionResult) {
	t.Helper()
	if result.Success {
		t.Errorf("expected action to fail, but it succeeded: %s", result.Message)
	}
}

// AssertActionFailedWithMessage asserts an action result failed with a specific message
func AssertActionFailedWithMessage(t *testing.T, result game.ActionResult, expectedMsg string) {
	t.Helper()
	if result.Success {
		t.Errorf("expected action to fail, but it succeeded: %s", result.Message)
		return
	}
	if result.Message != expectedMsg {
		t.Errorf("expected failure message %q, got %q", expectedMsg, result.Message)
	}
}

// DefaultTestBalance returns the default balance config for testing
func DefaultTestBalance() *config.BalanceConfig {
	balance := config.DefaultBalanceConfig()
	return &balance
}

// MoveAction creates a move action for testing
func MoveAction(agentID uuid.UUID, direction game.Direction) game.Action {
	return game.Action{
		Type:    game.ActionMove,
		AgentID: agentID,
		Params: game.ActionParams{
			Direction: direction,
		},
	}
}

// ClaimAction creates a claim action for testing
func ClaimAction(agentID uuid.UUID) game.Action {
	return game.Action{
		Type:    game.ActionClaim,
		AgentID: agentID,
	}
}

// WaitAction creates a wait action for testing
func WaitAction(agentID uuid.UUID) game.Action {
	return game.WaitAction(agentID)
}

// FightAction creates a fight action for testing
func FightAction(agentID uuid.UUID, targetID *uuid.UUID) game.Action {
	return game.Action{
		Type:    game.ActionFight,
		AgentID: agentID,
		Params: game.ActionParams{
			Target: targetID,
		},
	}
}

// UpgradeAction creates an upgrade action for testing
func UpgradeAction(agentID uuid.UUID, upgradeType string) game.Action {
	return game.Action{
		Type:    game.ActionUpgrade,
		AgentID: agentID,
		Params: game.ActionParams{
			UpgradeType: upgradeType,
		},
	}
}

// HarvestAction creates a harvest action for testing
func HarvestAction(agentID uuid.UUID) game.Action {
	return game.Action{
		Type:    game.ActionHarvest,
		AgentID: agentID,
	}
}

// PickupAction creates a pickup action for testing
func PickupAction(agentID uuid.UUID, itemID string) game.Action {
	return game.Action{
		Type:    game.ActionPickup,
		AgentID: agentID,
		Params: game.ActionParams{
			ItemID: itemID,
		},
	}
}

// UseAction creates a use action for testing
func UseAction(agentID uuid.UUID, itemID string) game.Action {
	return game.Action{
		Type:    game.ActionUse,
		AgentID: agentID,
		Params: game.ActionParams{
			ItemID: itemID,
		},
	}
}

// MessageAction creates a message action for testing
func MessageAction(agentID uuid.UUID, targetID *uuid.UUID, message string) game.Action {
	return game.Action{
		Type:    game.ActionMessage,
		AgentID: agentID,
		Params: game.ActionParams{
			Target:  targetID,
			Message: message,
		},
	}
}
