package actions

import (
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// ClaimHandler handles the CLAIM action for claiming tiles in a radius
type ClaimHandler struct{}

// NewClaimHandler creates a new claim action handler
func NewClaimHandler() *ClaimHandler {
	return &ClaimHandler{}
}

// ActionType returns the type of action this handler processes
func (h *ClaimHandler) ActionType() game.ActionType {
	return game.ActionClaim
}

// Validate checks if the claim action can be performed
func (h *ClaimHandler) Validate(ctx *ActionContext) error {
	pos := ctx.Agent.GetPosition()
	if !ctx.World.IsValidPosition(pos) {
		return fmt.Errorf("invalid position")
	}
	return nil
}

// Process executes the claim action, claiming all passable tiles in radius
func (h *ClaimHandler) Process(ctx *ActionContext) game.ActionResult {
	pos := ctx.Agent.GetPosition()

	result := game.ActionResult{
		AgentID:   ctx.Agent.ID,
		Action:    game.ActionClaim,
		ClaimedAt: &pos,
	}

	if err := h.Validate(ctx); err != nil {
		result.Success = false
		result.Message = err.Error()
		return result
	}

	baseClaimRadius := 4
	if ctx.Balance != nil {
		baseClaimRadius = ctx.Balance.Agent.DefaultClaimRadius
	}
	radius := ctx.Agent.GetEffectiveClaimRadius(baseClaimRadius)

	claimedCount := 0
	capturedCount := 0
	var claimedTiles []game.Position

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			// Circular radius check (Euclidean distance)
			if dx*dx+dy*dy > radius*radius {
				continue
			}

			tilePos := game.Position{X: pos.X + dx, Y: pos.Y + dy}
			if !ctx.World.IsValidPosition(tilePos) {
				continue
			}

			tile := ctx.World.GetTile(tilePos)
			if tile == nil {
				continue
			}
			// Skip already owned by this agent
			if tile.OwnerID != nil && *tile.OwnerID == ctx.Agent.ID {
				continue
			}

			wasEnemy := tile.OwnerID != nil
			ctx.World.SetOwner(tilePos, &ctx.Agent.ID)
			claimedTiles = append(claimedTiles, tilePos)

			if wasEnemy {
				capturedCount++
			} else {
				claimedCount++
			}
		}
	}

	total := claimedCount + capturedCount
	if total == 0 {
		result.Success = true
		result.Message = "all tiles in radius already owned"
		return result
	}

	result.Success = true
	result.ClaimedTiles = claimedTiles
	if capturedCount > 0 {
		result.Message = fmt.Sprintf("claimed %d tiles (%d captured from enemies)", total, capturedCount)
	} else {
		result.Message = fmt.Sprintf("claimed %d tiles", total)
	}
	ctx.Agent.AddMemory(fmt.Sprintf("Claimed %d tiles (radius %d) at (%d,%d)", total, radius, pos.X, pos.Y))
	return result
}
