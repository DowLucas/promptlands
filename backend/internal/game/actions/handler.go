// Package actions provides action handler implementations for the game.
// The core interfaces (ActionHandler, ActionContext, HandlerRegistry) are
// defined in the game package to avoid import cycles.
package actions

import (
	"github.com/lucas/promptlands/internal/game"
)

// Re-export types from game package for convenience
type (
	ActionHandler  = game.ActionHandler
	ActionContext  = game.ActionContext
	HandlerRegistry = game.HandlerRegistry
)

// NewActionContext is a convenience function that delegates to game.NewActionContext
var NewActionContext = game.NewActionContext

// NewHandlerRegistry is a convenience function that delegates to game.NewHandlerRegistry
var NewHandlerRegistry = game.NewHandlerRegistry

// FailedResult is a convenience function that delegates to game.FailedResult
var FailedResult = game.FailedResult

// SuccessResult is a convenience function that delegates to game.SuccessResult
var SuccessResult = game.SuccessResult
