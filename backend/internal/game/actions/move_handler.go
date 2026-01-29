package actions

import (
	"fmt"

	"github.com/lucas/promptlands/internal/game"
)

// MoveHandler handles MOVE actions
type MoveHandler struct{}

// NewMoveHandler creates a new move handler
func NewMoveHandler() *MoveHandler {
	return &MoveHandler{}
}

// ActionType returns the type of action this handler processes
func (h *MoveHandler) ActionType() game.ActionType {
	return game.ActionMove
}

// Validate checks if the move action has a valid direction
func (h *MoveHandler) Validate(ctx *ActionContext) error {
	dir := ctx.Action.Params.Direction
	if dir != game.DirNorth && dir != game.DirSouth && dir != game.DirEast && dir != game.DirWest {
		return fmt.Errorf("invalid direction")
	}
	return nil
}

// Process executes the move action, walking up to N steps in one direction
func (h *MoveHandler) Process(ctx *ActionContext) game.ActionResult {
	dir := ctx.Action.Params.Direction
	oldPos := ctx.Agent.GetPosition()

	result := game.ActionResult{
		AgentID: ctx.Agent.ID,
		Action:  game.ActionMove,
		OldPos:  &oldPos,
	}

	if err := h.Validate(ctx); err != nil {
		result.Success = false
		result.Message = err.Error()
		return result
	}

	// Determine max steps from agent's effective move speed
	baseMoveSpeed := 3
	if ctx.Balance != nil {
		baseMoveSpeed = ctx.Balance.Agent.DefaultMoveSpeed
	}
	maxSteps := ctx.Agent.GetEffectiveMoveSpeed(baseMoveSpeed)

	// Allow LLM to request fewer steps
	requestedSteps := ctx.Action.Params.Steps
	if requestedSteps > 0 && requestedSteps < maxSteps {
		maxSteps = requestedSteps
	}

	// Walk step by step, stop on first blocked tile
	currentPos := oldPos
	stepsWalked := 0
	for step := 0; step < maxSteps; step++ {
		nextPos := game.GetNewPosition(currentPos, dir)

		if !ctx.World.IsValidPosition(nextPos) {
			break
		}
		if ctx.WorldObjects != nil && ctx.WorldObjects.HasBlockingObject(nextPos) {
			break
		}

		// Check for agent collision — stop if tile is occupied
		occupied := false
		for _, other := range ctx.Agents {
			if other.ID != ctx.Agent.ID && !other.IsDead && other.GetPosition() == nextPos {
				occupied = true
				break
			}
		}
		if occupied {
			break
		}

		currentPos = nextPos
		stepsWalked++
	}

	if stepsWalked == 0 {
		result.Success = false
		result.Message = "path blocked"
		return result
	}

	ctx.Agent.SetPosition(currentPos)
	result.Success = true
	result.NewPos = &currentPos
	result.Message = fmt.Sprintf("moved %d steps %s", stepsWalked, dir)
	ctx.Agent.AddMemory(fmt.Sprintf("Moved %s %d steps to (%d,%d)", dir, stepsWalked, currentPos.X, currentPos.Y))
	return result
}
