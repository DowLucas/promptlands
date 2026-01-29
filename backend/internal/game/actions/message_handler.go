package actions

import (
	"errors"

	"github.com/lucas/promptlands/internal/game"
)

// MessageHandler implements ActionHandler for the MESSAGE action
type MessageHandler struct{}

// NewMessageHandler creates a new message action handler
func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

// ActionType returns the type of action this handler processes
func (h *MessageHandler) ActionType() game.ActionType {
	return game.ActionMessage
}

// Validate checks if the message action can be performed
func (h *MessageHandler) Validate(ctx *ActionContext) error {
	// If a target is specified, validate it exists
	if ctx.Action.Params.Target != nil {
		targetID := *ctx.Action.Params.Target
		if _, ok := ctx.Agents[targetID]; !ok {
			return errors.New("target agent not found")
		}
	}

	return nil
}

// Process executes the message action and returns the result
func (h *MessageHandler) Process(ctx *ActionContext) game.ActionResult {
	// Validate first
	if err := h.Validate(ctx); err != nil {
		return FailedResult(ctx.Agent.ID, game.ActionMessage, err.Error())
	}

	message := ctx.Action.Params.Message

	result := game.ActionResult{
		AgentID: ctx.Agent.ID,
		Action:  game.ActionMessage,
		Success: true,
	}

	if ctx.Action.Params.Target == nil {
		// Broadcast message (no target)
		result.Message = "broadcast message sent"
		ctx.Agent.AddMemory("Broadcast: " + message)
	} else {
		// Direct message (with target)
		result.Message = "message sent"
		ctx.Agent.AddMemory("Sent message to agent: " + message)
	}

	return result
}
