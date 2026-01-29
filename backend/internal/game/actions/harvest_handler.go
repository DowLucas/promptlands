package actions

import (
	"errors"
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// HarvestHandler implements ActionHandler for the HARVEST action
type HarvestHandler struct{}

// NewHarvestHandler creates a new harvest action handler
func NewHarvestHandler() *HarvestHandler {
	return &HarvestHandler{}
}

// ActionType returns the type of action this handler processes
func (h *HarvestHandler) ActionType() game.ActionType {
	return game.ActionHarvest
}

// Validate checks if the harvest action can be performed
func (h *HarvestHandler) Validate(ctx *ActionContext) error {
	// Check if world objects manager is available
	if ctx.WorldObjects == nil {
		return errors.New("system not initialized")
	}

	// Check if agent has inventory
	if ctx.Agent.Inventory == nil {
		return errors.New("system not initialized")
	}

	// Check if there's a resource at the agent's position
	pos := ctx.Agent.GetPosition()
	resource := ctx.WorldObjects.GetResourceAt(pos)

	if resource == nil {
		return errors.New("no resource to harvest")
	}

	if resource.Remaining <= 0 {
		return errors.New("resource depleted")
	}

	return nil
}

// Process executes the harvest action and returns the result
func (h *HarvestHandler) Process(ctx *ActionContext) game.ActionResult {
	// Validate first
	if err := h.Validate(ctx); err != nil {
		return FailedResult(ctx.Agent.ID, game.ActionHarvest, err.Error())
	}

	pos := ctx.Agent.GetPosition()
	resource := ctx.WorldObjects.GetResourceAt(pos)

	// Harvest 1 unit
	harvested := resource.Harvest(1)
	if harvested == 0 {
		return FailedResult(ctx.Agent.ID, game.ActionHarvest, "resource depleted")
	}

	// Map resource type to item ID
	itemID := string(resource.ResourceType)

	// Add to inventory
	remaining := ctx.Agent.Inventory.AddItem(itemID, harvested)
	if remaining == harvested {
		// Inventory full, put resource back
		resource.Remaining += harvested
		return FailedResult(ctx.Agent.ID, game.ActionHarvest, "inventory full")
	}

	actualHarvested := harvested - remaining

	result := game.ActionResult{
		AgentID:      ctx.Agent.ID,
		Action:       game.ActionHarvest,
		Success:      true,
		Harvested:    itemID,
		ItemQuantity: actualHarvested,
		Message:      fmt.Sprintf("harvested %d %s", actualHarvested, itemID),
	}

	ctx.Agent.AddMemory(fmt.Sprintf("Harvested %d %s", actualHarvested, itemID))

	return result
}
