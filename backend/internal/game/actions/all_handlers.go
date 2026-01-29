package actions

import "github.com/lucas/promptlands/internal/game"

// RegisterAllHandlers registers all action handlers with the given registry
func RegisterAllHandlers(registry *game.HandlerRegistry) {
	registry.Register(NewMoveHandler())
	registry.Register(NewClaimHandler())
	registry.Register(NewFightHandler())
	registry.Register(NewHarvestHandler())
	registry.Register(NewUpgradeHandler())
	registry.Register(NewPickupHandler())
	registry.Register(NewUseHandler())
	registry.Register(NewMessageHandler())
	registry.Register(NewWaitHandler())
	registry.Register(NewBuyHandler())
}
