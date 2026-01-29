package actions

import (
	"errors"
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// BuyHandler implements ActionHandler for the BUY action
type BuyHandler struct{}

// NewBuyHandler creates a new buy action handler
func NewBuyHandler() *BuyHandler {
	return &BuyHandler{}
}

// ActionType returns the type of action this handler processes
func (h *BuyHandler) ActionType() game.ActionType {
	return game.ActionBuy
}

// Validate checks if the buy action can be performed
func (h *BuyHandler) Validate(ctx *ActionContext) error {
	if ctx.ItemRegistry == nil {
		return errors.New("system not initialized")
	}

	itemID := ctx.Action.Params.ItemID
	if itemID == "" {
		return errors.New("item_id is required")
	}

	def := ctx.ItemRegistry.Get(itemID)
	if def == nil {
		return errors.New("unknown item")
	}

	cost := def.GetPropertyInt("coin_cost", 0)
	if cost <= 0 {
		return errors.New("item is not buyable")
	}

	if ctx.Agent.GetCoins() < cost {
		return fmt.Errorf("need %d coins (have %d)", cost, ctx.Agent.GetCoins())
	}

	if ctx.Agent.Inventory == nil {
		return errors.New("no inventory")
	}

	if ctx.Agent.Inventory.IsFull() {
		return errors.New("inventory full")
	}

	return nil
}

// Process executes the buy action and returns the result
func (h *BuyHandler) Process(ctx *ActionContext) game.ActionResult {
	if err := h.Validate(ctx); err != nil {
		return FailedResult(ctx.Agent.ID, game.ActionBuy, err.Error())
	}

	itemID := ctx.Action.Params.ItemID
	def := ctx.ItemRegistry.Get(itemID)
	cost := def.GetPropertyInt("coin_cost", 0)

	ctx.Agent.SpendCoins(cost)
	ctx.Agent.Inventory.AddItem(def.ID, 1)

	ctx.Agent.AddMemory(fmt.Sprintf("Bought %s for %d coins", def.Name, cost))

	return game.ActionResult{
		AgentID: ctx.Agent.ID,
		Action:  game.ActionBuy,
		Success: true,
		ItemID:  itemID,
		Message: fmt.Sprintf("bought %s for %d coins", def.Name, cost),
	}
}
