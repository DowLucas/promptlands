package actions

import "github.com/lucas/promptlands/internal/game"

// WaitHandler handles WAIT actions
type WaitHandler struct{}

// NewWaitHandler creates a new wait handler
func NewWaitHandler() *WaitHandler {
	return &WaitHandler{}
}

// ActionType returns the type of action this handler processes
func (h *WaitHandler) ActionType() game.ActionType {
	return game.ActionWait
}

// Validate checks if the wait action can be performed
// Wait actions are always valid
func (h *WaitHandler) Validate(ctx *ActionContext) error {
	return nil
}

// Process executes the wait action and returns the result
func (h *WaitHandler) Process(ctx *ActionContext) game.ActionResult {
	return SuccessResult(ctx.Agent.ID, ctx.Action.Type, "holding position")
}
