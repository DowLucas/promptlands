package actions

import (
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// UpgradeHandler handles the UPGRADE action for upgrading agent abilities
type UpgradeHandler struct{}

// NewUpgradeHandler creates a new upgrade action handler
func NewUpgradeHandler() *UpgradeHandler {
	return &UpgradeHandler{}
}

// ActionType returns the type of action this handler processes
func (h *UpgradeHandler) ActionType() game.ActionType {
	return game.ActionUpgrade
}

// Validate checks if the upgrade action can be performed
func (h *UpgradeHandler) Validate(ctx *ActionContext) error {
	upgradeType := ctx.Action.Params.UpgradeType

	canUpgrade, _, reason := ctx.Agent.CanUpgrade(upgradeType)
	if !canUpgrade {
		return fmt.Errorf("%s", reason)
	}

	return nil
}

// Process executes the upgrade action
func (h *UpgradeHandler) Process(ctx *ActionContext) game.ActionResult {
	upgradeType := ctx.Action.Params.UpgradeType

	result := game.ActionResult{
		AgentID: ctx.Agent.ID,
		Action:  game.ActionUpgrade,
	}

	// Run validation to get cost
	canUpgrade, cost, reason := ctx.Agent.CanUpgrade(upgradeType)
	if !canUpgrade {
		result.Success = false
		result.Message = reason
		return result
	}

	// Apply the upgrade
	if !ctx.Agent.ApplyUpgrade(upgradeType) {
		result.Success = false
		result.Message = "upgrade failed"
		return result
	}

	// Get the new level after upgrade
	var newLevel int
	switch upgradeType {
	case "vision":
		newLevel = ctx.Agent.VisionLevel
	case "memory":
		newLevel = ctx.Agent.MemoryLevel
	case "strength":
		newLevel = ctx.Agent.StrengthLevel
	case "storage":
		newLevel = ctx.Agent.StorageLevel
	case "speed":
		newLevel = ctx.Agent.SpeedLevel
	case "claim":
		newLevel = ctx.Agent.ClaimLevel
	}

	result.Success = true
	result.Upgraded = upgradeType
	result.NewLevel = newLevel
	result.Message = fmt.Sprintf("upgraded %s to level %d (cost: %d energy)", upgradeType, newLevel, cost)
	ctx.Agent.AddMemory(fmt.Sprintf("Upgraded %s to level %d", upgradeType, newLevel))

	return result
}
