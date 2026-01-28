package game

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ActionType represents the type of action an agent can take
type ActionType string

const (
	ActionMove     ActionType = "MOVE"
	ActionClaim    ActionType = "CLAIM"
	ActionMessage  ActionType = "MESSAGE"
	ActionWait     ActionType = "WAIT"
	ActionHold     ActionType = "HOLD"
	ActionFight    ActionType = "FIGHT"
	ActionPickup   ActionType = "PICKUP"
	ActionDrop     ActionType = "DROP"
	ActionUse      ActionType = "USE"
	ActionPlace    ActionType = "PLACE"
	ActionCraft    ActionType = "CRAFT"
	ActionHarvest  ActionType = "HARVEST"
	ActionScan     ActionType = "SCAN"
	ActionInteract ActionType = "INTERACT"
	ActionUpgrade  ActionType = "UPGRADE"
)

// Direction represents a movement direction
type Direction string

const (
	DirNorth Direction = "north"
	DirSouth Direction = "south"
	DirEast  Direction = "east"
	DirWest  Direction = "west"
)

// Action represents an agent's action for a tick
type Action struct {
	Type       ActionType `json:"action"`
	AgentID    uuid.UUID  `json:"agent_id"`
	Params     ActionParams `json:"params,omitempty"`
	ReceivedAt time.Time  `json:"-"`
}

// ActionParams holds the parameters for different action types
type ActionParams struct {
	Direction   Direction  `json:"direction,omitempty"`    // For MOVE
	Target      *uuid.UUID `json:"target,omitempty"`       // For MESSAGE (nil = broadcast), FIGHT
	Message     string     `json:"message,omitempty"`      // For MESSAGE, INTERACT (obelisk)
	ItemID      string     `json:"item_id,omitempty"`      // For USE, PLACE, DROP
	SlotIndex   *int       `json:"slot_index,omitempty"`   // Alternative to ItemID for slot-based ops
	Quantity    int        `json:"quantity,omitempty"`     // For DROP
	RecipeID    string     `json:"recipe_id,omitempty"`    // For CRAFT
	UpgradeType string     `json:"upgrade_type,omitempty"` // For UPGRADE: vision, memory, strength, storage
}

// WaitAction returns a default wait action
func WaitAction(agentID uuid.UUID) Action {
	return Action{
		Type:       ActionWait,
		AgentID:    agentID,
		ReceivedAt: time.Now(),
	}
}

// MoveAction creates a move action
func MoveAction(agentID uuid.UUID, dir Direction) Action {
	return Action{
		Type:    ActionMove,
		AgentID: agentID,
		Params:  ActionParams{Direction: dir},
	}
}

// ClaimAction creates a claim action
func ClaimAction(agentID uuid.UUID) Action {
	return Action{
		Type:    ActionClaim,
		AgentID: agentID,
	}
}

// MessageAction creates a message action
func MessageAction(agentID uuid.UUID, target *uuid.UUID, message string) Action {
	return Action{
		Type:    ActionMessage,
		AgentID: agentID,
		Params:  ActionParams{Target: target, Message: message},
	}
}

// HoldAction creates a hold action (stay in place)
func HoldAction(agentID uuid.UUID) Action {
	return Action{
		Type:    ActionHold,
		AgentID: agentID,
	}
}

// FightAction creates a fight action
func FightAction(agentID uuid.UUID, target uuid.UUID) Action {
	return Action{
		Type:    ActionFight,
		AgentID: agentID,
		Params:  ActionParams{Target: &target},
	}
}

// PickupAction creates a pickup action
func PickupAction(agentID uuid.UUID) Action {
	return Action{
		Type:    ActionPickup,
		AgentID: agentID,
	}
}

// DropAction creates a drop action
func DropAction(agentID uuid.UUID, itemID string, quantity int) Action {
	return Action{
		Type:    ActionDrop,
		AgentID: agentID,
		Params:  ActionParams{ItemID: itemID, Quantity: quantity},
	}
}

// UseAction creates a use action
func UseAction(agentID uuid.UUID, itemID string) Action {
	return Action{
		Type:    ActionUse,
		AgentID: agentID,
		Params:  ActionParams{ItemID: itemID},
	}
}

// PlaceAction creates a place action
func PlaceAction(agentID uuid.UUID, itemID string) Action {
	return Action{
		Type:    ActionPlace,
		AgentID: agentID,
		Params:  ActionParams{ItemID: itemID},
	}
}

// CraftAction creates a craft action
func CraftAction(agentID uuid.UUID, recipeID string) Action {
	return Action{
		Type:    ActionCraft,
		AgentID: agentID,
		Params:  ActionParams{RecipeID: recipeID},
	}
}

// HarvestAction creates a harvest action
func HarvestAction(agentID uuid.UUID) Action {
	return Action{
		Type:    ActionHarvest,
		AgentID: agentID,
	}
}

// ScanAction creates a scan action
func ScanAction(agentID uuid.UUID) Action {
	return Action{
		Type:    ActionScan,
		AgentID: agentID,
	}
}

// InteractAction creates an interact action
func InteractAction(agentID uuid.UUID, message string) Action {
	return Action{
		Type:    ActionInteract,
		AgentID: agentID,
		Params:  ActionParams{Message: message},
	}
}

// UpgradeAction creates an upgrade action
func UpgradeAction(agentID uuid.UUID, upgradeType string) Action {
	return Action{
		Type:    ActionUpgrade,
		AgentID: agentID,
		Params:  ActionParams{UpgradeType: upgradeType},
	}
}

// ParseAction parses a JSON action from LLM response
func ParseAction(agentID uuid.UUID, data []byte) (Action, error) {
	var raw struct {
		Action      string `json:"action"`
		Direction   string `json:"direction,omitempty"`
		Target      string `json:"target,omitempty"`
		Message     string `json:"message,omitempty"`
		ItemID      string `json:"item_id,omitempty"`
		Quantity    int    `json:"quantity,omitempty"`
		RecipeID    string `json:"recipe_id,omitempty"`
		UpgradeType string `json:"upgrade_type,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return WaitAction(agentID), fmt.Errorf("invalid JSON: %w", err)
	}

	action := Action{
		AgentID:    agentID,
		ReceivedAt: time.Now(),
	}

	switch ActionType(raw.Action) {
	case ActionMove:
		action.Type = ActionMove
		dir := Direction(raw.Direction)
		if !isValidDirection(dir) {
			return WaitAction(agentID), fmt.Errorf("invalid direction: %s", raw.Direction)
		}
		action.Params.Direction = dir

	case ActionClaim:
		action.Type = ActionClaim

	case ActionMessage:
		action.Type = ActionMessage
		action.Params.Message = raw.Message
		if raw.Target != "" {
			targetID, err := uuid.Parse(raw.Target)
			if err != nil {
				// Invalid target, send as broadcast
				action.Params.Target = nil
			} else {
				action.Params.Target = &targetID
			}
		}

	case ActionWait:
		action.Type = ActionWait

	case ActionHold:
		action.Type = ActionHold

	case ActionFight:
		action.Type = ActionFight
		if raw.Target == "" {
			return WaitAction(agentID), fmt.Errorf("FIGHT requires target")
		}
		targetID, err := uuid.Parse(raw.Target)
		if err != nil {
			return WaitAction(agentID), fmt.Errorf("invalid target ID: %s", raw.Target)
		}
		action.Params.Target = &targetID

	case ActionPickup:
		action.Type = ActionPickup

	case ActionDrop:
		action.Type = ActionDrop
		if raw.ItemID == "" {
			return WaitAction(agentID), fmt.Errorf("DROP requires item_id")
		}
		action.Params.ItemID = raw.ItemID
		action.Params.Quantity = raw.Quantity
		if action.Params.Quantity <= 0 {
			action.Params.Quantity = 1
		}

	case ActionUse:
		action.Type = ActionUse
		if raw.ItemID == "" {
			return WaitAction(agentID), fmt.Errorf("USE requires item_id")
		}
		action.Params.ItemID = raw.ItemID

	case ActionPlace:
		action.Type = ActionPlace
		if raw.ItemID == "" {
			return WaitAction(agentID), fmt.Errorf("PLACE requires item_id")
		}
		action.Params.ItemID = raw.ItemID

	case ActionCraft:
		action.Type = ActionCraft
		if raw.RecipeID == "" {
			return WaitAction(agentID), fmt.Errorf("CRAFT requires recipe_id")
		}
		action.Params.RecipeID = raw.RecipeID

	case ActionHarvest:
		action.Type = ActionHarvest

	case ActionScan:
		action.Type = ActionScan

	case ActionInteract:
		action.Type = ActionInteract
		action.Params.Message = raw.Message

	case ActionUpgrade:
		action.Type = ActionUpgrade
		if raw.UpgradeType == "" {
			return WaitAction(agentID), fmt.Errorf("UPGRADE requires upgrade_type")
		}
		action.Params.UpgradeType = raw.UpgradeType

	default:
		return WaitAction(agentID), fmt.Errorf("unknown action: %s", raw.Action)
	}

	return action, nil
}

// isValidDirection checks if a direction is valid
func isValidDirection(dir Direction) bool {
	switch dir {
	case DirNorth, DirSouth, DirEast, DirWest:
		return true
	default:
		return false
	}
}

// GetNewPosition calculates the new position after a move
func GetNewPosition(current Position, dir Direction) Position {
	switch dir {
	case DirNorth:
		return Position{X: current.X, Y: current.Y - 1}
	case DirSouth:
		return Position{X: current.X, Y: current.Y + 1}
	case DirEast:
		return Position{X: current.X + 1, Y: current.Y}
	case DirWest:
		return Position{X: current.X - 1, Y: current.Y}
	default:
		return current
	}
}

// ActionResult represents the outcome of applying an action
type ActionResult struct {
	AgentID      uuid.UUID  `json:"agent_id"`
	Action       ActionType `json:"action"`
	Success      bool       `json:"success"`
	Message      string     `json:"message,omitempty"`
	OldPos       *Position  `json:"old_pos,omitempty"`
	NewPos       *Position  `json:"new_pos,omitempty"`
	ClaimedAt    *Position  `json:"claimed_at,omitempty"`
	TargetID     *uuid.UUID `json:"target_id,omitempty"`     // For FIGHT
	DamageDealt  int        `json:"damage_dealt,omitempty"`  // For FIGHT
	ItemID       string     `json:"item_id,omitempty"`       // For item-related actions
	ItemQuantity int        `json:"item_quantity,omitempty"` // For DROP, HARVEST, CRAFT
	Harvested    string     `json:"harvested,omitempty"`     // Resource type harvested
	Crafted      string     `json:"crafted,omitempty"`       // Item crafted
	Placed       string     `json:"placed,omitempty"`        // Structure placed
	Upgraded     string     `json:"upgraded,omitempty"`      // Upgrade type
	NewLevel     int        `json:"new_level,omitempty"`     // New upgrade level
}
