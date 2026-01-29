package actions

import (
	"errors"
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// FightHandler implements ActionHandler for the FIGHT action
type FightHandler struct{}

// NewFightHandler creates a new fight action handler
func NewFightHandler() *FightHandler {
	return &FightHandler{}
}

// ActionType returns the type of action this handler processes
func (h *FightHandler) ActionType() game.ActionType {
	return game.ActionFight
}

// Validate checks if the fight action can be performed
func (h *FightHandler) Validate(ctx *ActionContext) error {
	// Check if target is specified
	if ctx.Action.Params.Target == nil {
		return errors.New("no target specified")
	}

	targetID := *ctx.Action.Params.Target

	// Check if target exists and is alive
	target, ok := ctx.Agents[targetID]
	if !ok {
		return errors.New("target not found")
	}
	if target.IsDead {
		return errors.New("target is already dead")
	}

	// Check adjacency (within 1 tile)
	agentPos := ctx.Agent.GetPosition()
	targetPos := target.GetPosition()
	dx := agentPos.X - targetPos.X
	dy := agentPos.Y - targetPos.Y

	if dx < -1 || dx > 1 || dy < -1 || dy > 1 || (dx == 0 && dy == 0) {
		return errors.New("target not adjacent")
	}

	return nil
}

// Process executes the fight action and returns the result
func (h *FightHandler) Process(ctx *ActionContext) game.ActionResult {
	// Validate first
	if err := h.Validate(ctx); err != nil {
		return FailedResult(ctx.Agent.ID, game.ActionFight, err.Error())
	}

	targetID := *ctx.Action.Params.Target
	target := ctx.Agents[targetID]

	// Calculate damage from effective strength
	damage := ctx.Agent.GetEffectiveStrength()

	// Apply weapon damage bonus
	if ctx.Agent.Inventory != nil && ctx.ItemRegistry != nil {
		weapon := ctx.Agent.Inventory.GetEquipped(game.SlotWeapon)
		if weapon != nil {
			if def := ctx.ItemRegistry.Get(weapon.DefinitionID); def != nil {
				damage += def.GetPropertyInt("damage_bonus", 0)
			}
		}
	}

	// Apply target armor defense reduction
	if target.Inventory != nil && ctx.ItemRegistry != nil {
		armor := target.Inventory.GetEquipped(game.SlotArmor)
		if armor != nil {
			if def := ctx.ItemRegistry.Get(armor.DefinitionID); def != nil {
				damage -= def.GetPropertyInt("defense_bonus", 0)
				if damage < 1 {
					damage = 1
				}
			}
		}
	}

	// Apply damage
	killed := target.TakeDamage(damage)

	result := game.ActionResult{
		AgentID:     ctx.Agent.ID,
		Action:      game.ActionFight,
		Success:     true,
		TargetID:    &targetID,
		DamageDealt: damage,
	}

	if killed {
		target.Kill(ctx.CurrentTick)

		// Clear target's tiles on death
		ownedTiles := ctx.World.GetOwnedTiles(target.ID)
		for _, pos := range ownedTiles {
			ctx.World.SetOwner(pos, nil)
		}

		// Clear inventory
		if target.Inventory != nil {
			target.Inventory.Clear()
		}

		result.Message = fmt.Sprintf("killed %s", target.Name)
		ctx.Agent.AddMemory(fmt.Sprintf("Killed %s in combat", target.Name))
	} else {
		result.Message = fmt.Sprintf("dealt %d damage to %s (HP: %d/%d)", damage, target.Name, target.GetHP(), target.MaxHP)
		ctx.Agent.AddMemory(fmt.Sprintf("Attacked %s for %d damage", target.Name, damage))
	}

	return result
}
