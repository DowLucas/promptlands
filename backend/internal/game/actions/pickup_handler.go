package actions

import (
	"errors"
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// PickupHandler implements ActionHandler for the PICKUP action
type PickupHandler struct{}

// NewPickupHandler creates a new pickup action handler
func NewPickupHandler() *PickupHandler {
	return &PickupHandler{}
}

// ActionType returns the type of action this handler processes
func (h *PickupHandler) ActionType() game.ActionType {
	return game.ActionPickup
}

// Validate checks if the pickup action can be performed
func (h *PickupHandler) Validate(ctx *ActionContext) error {
	// Check if world objects manager is available
	if ctx.WorldObjects == nil {
		return errors.New("system not initialized")
	}

	// Check if agent has inventory
	if ctx.Agent.Inventory == nil {
		return errors.New("system not initialized")
	}

	// Check if there are dropped items at the agent's position
	pos := ctx.Agent.GetPosition()
	droppedItems := ctx.WorldObjects.GetDroppedItemsAt(pos)

	if len(droppedItems) == 0 {
		return errors.New("no items to pick up")
	}

	return nil
}

// Process executes the pickup action and returns the result
func (h *PickupHandler) Process(ctx *ActionContext) game.ActionResult {
	// Validate first
	if err := h.Validate(ctx); err != nil {
		return FailedResult(ctx.Agent.ID, game.ActionPickup, err.Error())
	}

	pos := ctx.Agent.GetPosition()
	droppedItems := ctx.WorldObjects.GetDroppedItemsAt(pos)

	// Pick up the first item
	obj := droppedItems[0]
	if obj.Item == nil {
		return FailedResult(ctx.Agent.ID, game.ActionPickup, "invalid item")
	}

	// Try to add to inventory
	remaining := ctx.Agent.Inventory.AddItem(obj.Item.DefinitionID, obj.Item.Quantity)
	if remaining == obj.Item.Quantity {
		return FailedResult(ctx.Agent.ID, game.ActionPickup, "inventory full")
	}

	pickedUp := obj.Item.Quantity - remaining

	result := game.ActionResult{
		AgentID:      ctx.Agent.ID,
		Action:       game.ActionPickup,
		Success:      true,
		ItemID:       obj.Item.DefinitionID,
		ItemQuantity: pickedUp,
	}

	if remaining > 0 {
		// Partial pickup - update the dropped item quantity
		obj.Item.Quantity = remaining
		result.Message = fmt.Sprintf("picked up %d %s (partial)", pickedUp, obj.Item.DefinitionID)
	} else {
		// Full pickup - remove the dropped item from the world
		ctx.WorldObjects.Remove(obj.ID)
		result.Message = fmt.Sprintf("picked up %d %s", pickedUp, obj.Item.DefinitionID)
	}

	ctx.Agent.AddMemory(fmt.Sprintf("Picked up %d %s", pickedUp, obj.Item.DefinitionID))

	return result
}
