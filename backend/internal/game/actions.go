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
	ActionMove    ActionType = "MOVE"
	ActionClaim   ActionType = "CLAIM"
	ActionMessage ActionType = "MESSAGE"
	ActionWait    ActionType = "WAIT"
	ActionFight   ActionType = "FIGHT"
	ActionPickup  ActionType = "PICKUP"
	ActionUse     ActionType = "USE"
	ActionHarvest ActionType = "HARVEST"
	ActionUpgrade ActionType = "UPGRADE"
	ActionBuy     ActionType = "BUY"

	// ActionInteract is used internally for auto-activation results (not a player action)
	ActionInteract ActionType = "INTERACT"
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
	Type       ActionType   `json:"action"`
	AgentID    uuid.UUID    `json:"agent_id"`
	Params     ActionParams `json:"params,omitempty"`
	Reasoning  string       `json:"reasoning,omitempty"`
	ReceivedAt time.Time    `json:"-"`
}

// ActionParams holds the parameters for different action types
type ActionParams struct {
	Direction   Direction  `json:"direction,omitempty"`    // For MOVE
	Steps       int        `json:"steps,omitempty"`        // For MOVE: number of steps (0 = use max)
	Target      *uuid.UUID `json:"target,omitempty"`       // For MESSAGE (nil = broadcast), FIGHT
	Message     string     `json:"message,omitempty"`      // For MESSAGE
	ItemID      string     `json:"item_id,omitempty"`      // For USE
	UpgradeType string     `json:"upgrade_type,omitempty"` // For UPGRADE: vision, memory, strength, storage, speed, claim
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

// UseAction creates a use action
func UseAction(agentID uuid.UUID, itemID string) Action {
	return Action{
		Type:    ActionUse,
		AgentID: agentID,
		Params:  ActionParams{ItemID: itemID},
	}
}

// HarvestAction creates a harvest action
func HarvestAction(agentID uuid.UUID) Action {
	return Action{
		Type:    ActionHarvest,
		AgentID: agentID,
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

// BuyAction creates a buy action
func BuyAction(agentID uuid.UUID, itemID string) Action {
	return Action{
		Type:    ActionBuy,
		AgentID: agentID,
		Params:  ActionParams{ItemID: itemID},
	}
}

// ParseAction parses a JSON action from LLM response
func ParseAction(agentID uuid.UUID, data []byte) (Action, error) {
	var raw struct {
		Action      string `json:"action"`
		Direction   string `json:"direction,omitempty"`
		Steps       int    `json:"steps,omitempty"`
		Target      string `json:"target,omitempty"`
		Message     string `json:"message,omitempty"`
		ItemID      string `json:"item_id,omitempty"`
		UpgradeType string `json:"upgrade_type,omitempty"`
		Reasoning   string `json:"reasoning,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return WaitAction(agentID), fmt.Errorf("invalid JSON: %w", err)
	}

	action := Action{
		AgentID:    agentID,
		Reasoning:  raw.Reasoning,
		ReceivedAt: time.Now(),
	}

	// Backwards compatibility: map removed actions to their replacements
	actionType := ActionType(raw.Action)
	switch actionType {
	case "HOLD":
		actionType = ActionWait
	case "PLACE":
		actionType = ActionUse
	}

	switch actionType {
	case ActionMove:
		action.Type = ActionMove
		dir := Direction(raw.Direction)
		if !isValidDirection(dir) {
			return WaitAction(agentID), fmt.Errorf("invalid direction: %s", raw.Direction)
		}
		action.Params.Direction = dir
		action.Params.Steps = raw.Steps

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

	case ActionUse:
		action.Type = ActionUse
		if raw.ItemID == "" {
			return WaitAction(agentID), fmt.Errorf("USE requires item_id")
		}
		action.Params.ItemID = raw.ItemID

	case ActionHarvest:
		action.Type = ActionHarvest

	case ActionUpgrade:
		action.Type = ActionUpgrade
		if raw.UpgradeType == "" {
			return WaitAction(agentID), fmt.Errorf("UPGRADE requires upgrade_type")
		}
		action.Params.UpgradeType = raw.UpgradeType

	case ActionBuy:
		action.Type = ActionBuy
		if raw.ItemID == "" {
			return WaitAction(agentID), fmt.Errorf("BUY requires item_id")
		}
		action.Params.ItemID = raw.ItemID

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
	Reasoning    string     `json:"reasoning,omitempty"`
	OldPos       *Position  `json:"old_pos,omitempty"`
	NewPos       *Position  `json:"new_pos,omitempty"`
	ClaimedAt    *Position  `json:"claimed_at,omitempty"`
	ClaimedTiles []Position `json:"-"`                       // All tiles claimed (not serialized to client)
	TargetID     *uuid.UUID `json:"target_id,omitempty"`     // For FIGHT
	DamageDealt  int        `json:"damage_dealt,omitempty"`  // For FIGHT
	ItemID       string     `json:"item_id,omitempty"`       // For item-related actions
	ItemQuantity int        `json:"item_quantity,omitempty"` // For HARVEST, PICKUP
	Harvested    string     `json:"harvested,omitempty"`     // Resource type harvested
	Placed       string     `json:"placed,omitempty"`        // Structure placed
	Upgraded     string     `json:"upgraded,omitempty"`      // Upgrade type
	NewLevel     int        `json:"new_level,omitempty"`     // New upgrade level
}
