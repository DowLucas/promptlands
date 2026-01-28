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
	sb.WriteString(fmt.Sprintf("Your position: (%d, %d)\n", pos.X, pos.Y))
	sb.WriteString(fmt.Sprintf("World size: %dx%d (coordinates 0 to %d)\n", ctx.WorldSize, ctx.WorldSize, ctx.WorldSize-1))
	sb.WriteString(fmt.Sprintf("Tiles you own: %d\n", ctx.OwnedCount))
	sb.WriteString(fmt.Sprintf("Current tick: %d\n", ctx.CurrentTick))

	// Tell agent about their current tile
	if ctx.CurrentTileOwned {
		sb.WriteString("Current tile: YOU OWN THIS TILE (no need to claim)\n")
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
	sb.WriteString(fmt.Sprintf("Valid move directions: %s\n", strings.Join(validMoves, ", ")))
	sb.WriteString("\n")

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
	sb.WriteString("- MOVE: Move one tile in a direction\n")
	sb.WriteString("  Parameters: direction (north, south, east, west)\n")
	sb.WriteString("- CLAIM: Claim or capture the tile you're standing on\n")
	sb.WriteString("  Parameters: none\n")
	sb.WriteString("- MESSAGE: Send a message to another agent or broadcast\n")
	sb.WriteString("  Parameters: target (agent ID or null for broadcast), message (text)\n")
	sb.WriteString("- WAIT: Do nothing this turn\n")
	sb.WriteString("  Parameters: none\n")
	sb.WriteString("\n")

	// Response format
	sb.WriteString("[Response Format]\n")
	sb.WriteString("You MUST respond with valid JSON in this exact format:\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"action\": \"MOVE\" | \"CLAIM\" | \"MESSAGE\" | \"WAIT\",\n")
	sb.WriteString("  \"direction\": \"north\" | \"south\" | \"east\" | \"west\",  // only for MOVE\n")
	sb.WriteString("  \"target\": \"agent-id\" | null,  // only for MESSAGE, null=broadcast\n")
	sb.WriteString("  \"message\": \"your message\"    // only for MESSAGE\n")
	sb.WriteString("}\n\n")

	sb.WriteString("Example responses:\n")
	sb.WriteString(`{"action": "MOVE", "direction": "north"}`)
	sb.WriteString("\n")
	sb.WriteString(`{"action": "CLAIM"}`)
	sb.WriteString("\n")
	sb.WriteString(`{"action": "MESSAGE", "target": null, "message": "Hello everyone!"}`)
	sb.WriteString("\n")
	sb.WriteString(`{"action": "WAIT"}`)
	sb.WriteString("\n\n")

	sb.WriteString("Now choose your action:")

	return sb.String()
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
