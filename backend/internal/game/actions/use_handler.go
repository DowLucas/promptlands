package actions

import (
	"errors"
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// UseHandler implements ActionHandler for the USE action.
// Handles both consumable items (potions) and placeable items (wall/beacon/trap).
type UseHandler struct{}

// NewUseHandler creates a new use action handler
func NewUseHandler() *UseHandler {
	return &UseHandler{}
}

// ActionType returns the type of action this handler processes
func (h *UseHandler) ActionType() game.ActionType {
	return game.ActionUse
}

// Validate checks if the use action can be performed
func (h *UseHandler) Validate(ctx *ActionContext) error {
	if ctx.Agent.Inventory == nil {
		return errors.New("system not initialized")
	}
	if ctx.ItemRegistry == nil {
		return errors.New("system not initialized")
	}

	itemID := ctx.Action.Params.ItemID

	if !ctx.Agent.Inventory.HasItems(itemID, 1) {
		return errors.New("item not in inventory")
	}

	def := ctx.ItemRegistry.Get(itemID)
	if def == nil {
		return errors.New("unknown item")
	}

	if def.Usable {
		return nil
	}

	if def.Placeable {
		return h.validatePlaceable(ctx, def)
	}

	return errors.New("item cannot be used")
}

// validatePlaceable checks placement-specific constraints
func (h *UseHandler) validatePlaceable(ctx *ActionContext, def *game.ItemDefinition) error {
	if ctx.WorldObjects == nil {
		return errors.New("system not initialized")
	}

	if def.EnergyCost > 0 && ctx.Agent.GetEnergy() < def.EnergyCost {
		return fmt.Errorf("need %d energy to place", def.EnergyCost)
	}

	pos := ctx.Agent.GetPosition()
	if existing := ctx.WorldObjects.GetStructureAt(pos); existing != nil {
		return errors.New("position already has a structure")
	}

	switch ctx.Action.Params.ItemID {
	case "wall", "beacon", "trap":
		// Valid structure types
	default:
		return errors.New("unknown structure type")
	}

	return nil
}

// Process executes the use action and returns the result
func (h *UseHandler) Process(ctx *ActionContext) game.ActionResult {
	if err := h.Validate(ctx); err != nil {
		return FailedResult(ctx.Agent.ID, game.ActionUse, err.Error())
	}

	itemID := ctx.Action.Params.ItemID
	def := ctx.ItemRegistry.Get(itemID)

	if def.Placeable {
		return h.processPlace(ctx, itemID, def)
	}

	return h.processConsumable(ctx, itemID, def)
}

// processConsumable handles usable/consumable items (potions etc.)
func (h *UseHandler) processConsumable(ctx *ActionContext, itemID string, def *game.ItemDefinition) game.ActionResult {
	effectApplied := false
	var message string

	if healAmount := def.GetPropertyInt("heal_amount", 0); healAmount > 0 {
		ctx.Agent.Heal(healAmount)
		effectApplied = true
		message = fmt.Sprintf("healed %d HP", healAmount)
	}

	if energyAmount := def.GetPropertyInt("energy_amount", 0); energyAmount > 0 {
		ctx.Agent.AddEnergy(energyAmount)
		effectApplied = true
		message = fmt.Sprintf("restored %d energy", energyAmount)
	}

	if !effectApplied {
		return FailedResult(ctx.Agent.ID, game.ActionUse, "item has no effect")
	}

	if def.Consumable {
		ctx.Agent.Inventory.RemoveItem(itemID, 1)
	}

	result := game.ActionResult{
		AgentID: ctx.Agent.ID,
		Action:  game.ActionUse,
		Success: true,
		ItemID:  itemID,
		Message: message,
	}

	ctx.Agent.AddMemory(fmt.Sprintf("Used %s", def.Name))
	return result
}

// processPlace handles placeable items (wall/beacon/trap)
func (h *UseHandler) processPlace(ctx *ActionContext, itemID string, def *game.ItemDefinition) game.ActionResult {
	pos := ctx.Agent.GetPosition()

	var structureType game.StructureType
	switch itemID {
	case "wall":
		structureType = game.StructureWall
	case "beacon":
		structureType = game.StructureBeacon
	case "trap":
		structureType = game.StructureTrap
	}

	ctx.Agent.Inventory.RemoveItem(itemID, 1)
	if def.EnergyCost > 0 {
		ctx.Agent.SpendEnergy(def.EnergyCost)
	}

	structure := game.NewStructure(structureType, pos, ctx.Agent.ID, ctx.CurrentTick)
	ctx.WorldObjects.Add(structure)

	result := game.ActionResult{
		AgentID: ctx.Agent.ID,
		Action:  game.ActionUse,
		Success: true,
		Placed:  string(structureType),
		ItemID:  itemID,
		Message: fmt.Sprintf("placed %s", def.Name),
	}

	ctx.Agent.AddMemory(fmt.Sprintf("Placed %s at (%d,%d)", def.Name, pos.X, pos.Y))
	return result
}
