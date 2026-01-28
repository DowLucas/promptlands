package llm

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/game"
)

// DefaultPromptBuilder implements game.PromptBuilder
type DefaultPromptBuilder struct{}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *DefaultPromptBuilder {
	return &DefaultPromptBuilder{}
}

// BuildPrompt creates the full prompt for an agent (implements game.PromptBuilder)
func (b *DefaultPromptBuilder) BuildPrompt(ctx game.AgentContext) string {
	return buildPrompt(ctx)
}

// buildPrompt creates the full prompt for an agent
func buildPrompt(ctx game.AgentContext) string {
	var sb strings.Builder

	// System instructions (player's custom prompt)
	sb.WriteString("[System Instructions]\n")
	sb.WriteString(ctx.Agent.SystemPrompt)
	sb.WriteString("\n\n")

	// Current state
	sb.WriteString("[Current State]\n")
	pos := ctx.Agent.GetPosition()
	sb.WriteString(fmt.Sprintf("Position: (%d, %d)\n", pos.X, pos.Y))
	sb.WriteString(fmt.Sprintf("HP: %d/%d\n", ctx.Agent.GetHP(), ctx.Agent.MaxHP))
	sb.WriteString(fmt.Sprintf("Energy: %d/%d (+%d/tick from %d owned tiles)\n", ctx.Agent.GetEnergy(), ctx.Agent.MaxEnergy, ctx.EnergyPerTick, ctx.OwnedCount))
	sb.WriteString(fmt.Sprintf("World size: %dx%d\n", ctx.WorldSize, ctx.WorldSize))
	sb.WriteString(fmt.Sprintf("Tick: %d\n", ctx.CurrentTick))

	// Upgrades
	sb.WriteString(fmt.Sprintf("Upgrades: Vision %d, Memory %d, Strength %d, Storage %d\n",
		ctx.Agent.VisionLevel, ctx.Agent.MemoryLevel, ctx.Agent.StrengthLevel, ctx.Agent.StorageLevel))

	// Tell agent about their current tile
	if ctx.CurrentTileOwned {
		sb.WriteString("Current tile: YOU OWN THIS TILE\n")
	} else if ctx.CurrentTileEnemy {
		sb.WriteString("Current tile: ENEMY TERRITORY (claim to capture!)\n")
	} else {
		sb.WriteString("Current tile: UNCLAIMED (you can claim it)\n")
	}

	// Tell agent valid moves
	validMoves := []string{}
	if pos.X > 0 {
		validMoves = append(validMoves, "west")
	}
	if pos.X < ctx.WorldSize-1 {
		validMoves = append(validMoves, "east")
	}
	if pos.Y > 0 {
		validMoves = append(validMoves, "north")
	}
	if pos.Y < ctx.WorldSize-1 {
		validMoves = append(validMoves, "south")
	}
	sb.WriteString(fmt.Sprintf("Valid moves: %s\n", strings.Join(validMoves, ", ")))
	sb.WriteString("\n")

	// Inventory
	if ctx.Agent.Inventory != nil {
		sb.WriteString("[Inventory]\n")
		items := ctx.Agent.Inventory.GetItemSummary()
		if len(items) == 0 {
			sb.WriteString("(empty)\n")
		} else {
			for _, item := range items {
				sb.WriteString(fmt.Sprintf("- %s x%d\n", item.Name, item.Quantity))
			}
		}
		sb.WriteString("\n")
	}

	// Memory
	memory := ctx.Agent.GetMemory()
	if len(memory) > 0 {
		sb.WriteString("[Your Memory]\n")
		for _, m := range memory {
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
		sb.WriteString("\n")
	}

	// Visible world
	sb.WriteString("[Visible World]\n")
	sb.WriteString(formatVisibleWorld(ctx.VisibleTiles, pos, ctx.Agent.ID))
	sb.WriteString("\n")

	// Visible objects
	if len(ctx.VisibleObjects) > 0 {
		sb.WriteString("[Visible Objects]\n")
		for _, obj := range ctx.VisibleObjects {
			sb.WriteString(formatWorldObject(obj, pos))
		}
		sb.WriteString("\n")
	}

	// Visible agents
	if len(ctx.VisibleAgents) > 0 {
		sb.WriteString("[Visible Agents]\n")
		for _, agent := range ctx.VisibleAgents {
			sb.WriteString(fmt.Sprintf("- %s at (%d,%d) HP:%d/%d [%s]\n",
				agent.Name, agent.Position.X, agent.Position.Y, agent.HP, agent.MaxHP, agent.ID))
		}
		sb.WriteString("\n")
	}

	// Incoming messages
	if len(ctx.Messages) > 0 {
		sb.WriteString("[Incoming Messages]\n")
		for _, msg := range ctx.Messages {
			msgType := "private"
			if msg.IsBroadcast {
				msgType = "broadcast"
			}
			sb.WriteString(fmt.Sprintf("From %s (%s): %s\n", msg.FromAgentName, msgType, msg.Content))
		}
		sb.WriteString("\n")
	}

	// Available actions
	sb.WriteString("[Available Actions]\n")
	sb.WriteString("- MOVE <direction>: Move one tile (north/south/east/west)\n")
	sb.WriteString("- HOLD: Stay in place\n")
	sb.WriteString("- CLAIM: Claim the tile you're on\n")
	sb.WriteString("- FIGHT <target>: Attack adjacent agent (1+strength damage)\n")
	sb.WriteString("- PICKUP: Pick up dropped item at your position\n")
	sb.WriteString("- DROP <item_id> <quantity>: Drop item from inventory\n")
	sb.WriteString("- USE <item_id>: Use consumable item\n")
	sb.WriteString("- PLACE <item_id>: Build structure (costs energy)\n")
	sb.WriteString("- CRAFT <recipe_id>: Craft item from materials\n")
	sb.WriteString("- HARVEST: Gather from resource node at your position\n")
	sb.WriteString("- SCAN: Extended vision this turn (costs 2 energy)\n")
	sb.WriteString("- INTERACT <message?>: Use shrine/portal/cache/obelisk\n")
	sb.WriteString("- UPGRADE <type>: Upgrade vision/memory/strength/storage\n")
	sb.WriteString("- MESSAGE <target?> <text>: Send message (null target = broadcast)\n")
	sb.WriteString("\n")

	// Upgrade costs
	sb.WriteString("[Upgrade Costs]\n")
	sb.WriteString("Level 1→2: 10 energy | 2→3: 20 | 3→4: 35 | 4→5: 55\n")
	sb.WriteString("Vision/Memory: max level 5 | Strength/Storage: max level 3\n")
	sb.WriteString("\n")

	// Response format
	sb.WriteString("[Response Format]\n")
	sb.WriteString("Respond with valid JSON:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"action\": \"ACTION_TYPE\",\n")
	sb.WriteString("  \"direction\": \"...\",     // MOVE only\n")
	sb.WriteString("  \"target\": \"agent-id\",   // FIGHT, MESSAGE\n")
	sb.WriteString("  \"item_id\": \"...\",       // USE, PLACE, DROP\n")
	sb.WriteString("  \"quantity\": 1,          // DROP\n")
	sb.WriteString("  \"recipe_id\": \"...\",     // CRAFT\n")
	sb.WriteString("  \"upgrade_type\": \"...\",  // UPGRADE\n")
	sb.WriteString("  \"message\": \"...\"        // MESSAGE, INTERACT\n")
	sb.WriteString("}\n\n")

	sb.WriteString("Examples:\n")
	sb.WriteString(`{"action": "MOVE", "direction": "north"}` + "\n")
	sb.WriteString(`{"action": "CLAIM"}` + "\n")
	sb.WriteString(`{"action": "FIGHT", "target": "agent-uuid"}` + "\n")
	sb.WriteString(`{"action": "HARVEST"}` + "\n")
	sb.WriteString(`{"action": "CRAFT", "recipe_id": "craft_wall"}` + "\n")
	sb.WriteString(`{"action": "USE", "item_id": "health_potion"}` + "\n")
	sb.WriteString(`{"action": "UPGRADE", "upgrade_type": "vision"}` + "\n")
	sb.WriteString("\n")

	sb.WriteString("Choose your action:")

	return sb.String()
}

// formatWorldObject formats a world object for the prompt
func formatWorldObject(obj *game.WorldObject, agentPos game.Position) string {
	pos := obj.Position
	dist := fmt.Sprintf("(%d,%d)", pos.X, pos.Y)

	switch obj.Type {
	case game.ObjectResource:
		return fmt.Sprintf("- Resource Node at %s: %s (%d remaining)\n", dist, obj.ResourceType, obj.Remaining)
	case game.ObjectStructure:
		owner := "enemy"
		if obj.OwnerID != nil {
			owner = "yours"
		}
		return fmt.Sprintf("- %s at %s (%s, HP:%d/%d)\n", obj.StructureType, dist, owner, obj.HP, obj.MaxHP)
	case game.ObjectInteractive:
		switch obj.InteractiveType {
		case game.InteractiveShrine:
			return fmt.Sprintf("- Shrine at %s (grants +1 max HP)\n", dist)
		case game.InteractiveCache:
			if obj.Activated {
				return ""
			}
			return fmt.Sprintf("- Cache at %s (contains energy)\n", dist)
		case game.InteractivePortal:
			return fmt.Sprintf("- Portal at %s\n", dist)
		case game.InteractiveObelisk:
			if obj.Message != "" {
				return fmt.Sprintf("- Obelisk at %s (has message)\n", dist)
			}
			return fmt.Sprintf("- Obelisk at %s (blank)\n", dist)
		}
	case game.ObjectDroppedItem:
		if obj.Item != nil {
			return fmt.Sprintf("- Dropped item at %s: %s x%d\n", dist, obj.Item.DefinitionID, obj.Item.Quantity)
		}
	}
	return ""
}

// formatVisibleWorld creates a text representation of visible tiles
func formatVisibleWorld(tiles []*game.Tile, agentPos game.Position, agentID uuid.UUID) string {
	if len(tiles) == 0 {
		return "No tiles visible.\n"
	}

	// Find bounds
	minX, maxX := tiles[0].Position.X, tiles[0].Position.X
	minY, maxY := tiles[0].Position.Y, tiles[0].Position.Y

	for _, t := range tiles {
		if t.Position.X < minX {
			minX = t.Position.X
		}
		if t.Position.X > maxX {
			maxX = t.Position.X
		}
		if t.Position.Y < minY {
			minY = t.Position.Y
		}
		if t.Position.Y > maxY {
			maxY = t.Position.Y
		}
	}

	// Create tile map for lookup
	tileMap := make(map[game.Position]*game.Tile)
	for _, t := range tiles {
		tileMap[t.Position] = t
	}

	var sb strings.Builder
	sb.WriteString("Legend: @ = you, O = your tile, X = enemy tile, . = unclaimed, # = obstacle\n")
	sb.WriteString(fmt.Sprintf("Coordinates shown for (%d,%d) to (%d,%d)\n\n", minX, minY, maxX, maxY))

	// Build grid
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			pos := game.Position{X: x, Y: y}

			// Check if this is the agent's position
			if pos == agentPos {
				sb.WriteString("@ ")
				continue
			}

			tile, ok := tileMap[pos]
			if !ok {
				sb.WriteString("? ") // Unknown/fog of war
				continue
			}

			// Check terrain
			switch tile.Terrain {
			case game.TerrainWater, game.TerrainMountain:
				sb.WriteString("# ")
				continue
			}

			// Check ownership
			if tile.OwnerID == nil {
				sb.WriteString(". ")
			} else if *tile.OwnerID == agentID {
				sb.WriteString("O ")
			} else {
				sb.WriteString("X ")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
