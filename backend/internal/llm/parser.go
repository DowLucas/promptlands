package llm

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/game"
)

// ExtractJSON attempts to extract JSON from LLM response text
// LLMs sometimes wrap JSON in markdown code blocks or add extra text
func ExtractJSON(text string) string {
	text = strings.TrimSpace(text)

	// Try to find JSON in markdown code blocks
	codeBlockPattern := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	if matches := codeBlockPattern.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find raw JSON object
	jsonPattern := regexp.MustCompile(`\{[^{}]*\}`)
	if match := jsonPattern.FindString(text); match != "" {
		return match
	}

	// Return original if no patterns match
	return text
}

// ParseActionFromText parses an action from potentially messy LLM output
func ParseActionFromText(agentID uuid.UUID, text string) (game.Action, error) {
	jsonStr := ExtractJSON(text)
	return game.ParseAction(agentID, []byte(jsonStr))
}

// ValidateAction checks if an action is valid for the current game state
func ValidateAction(action game.Action, agent *game.Agent, world *game.World) (bool, string) {
	switch action.Type {
	case game.ActionMove:
		if !isValidDirection(action.Params.Direction) {
			return false, "invalid direction"
		}
		newPos := game.GetNewPosition(agent.GetPosition(), action.Params.Direction)
		if !world.IsValidPosition(newPos) {
			return false, "would move out of bounds"
		}
		tile := world.GetTile(newPos)
		if tile != nil && (tile.Terrain == game.TerrainWater || tile.Terrain == game.TerrainMountain) {
			return false, "cannot move to impassable terrain"
		}
		return true, ""

	case game.ActionClaim:
		pos := agent.GetPosition()
		tile := world.GetTile(pos)
		if tile == nil {
			return false, "invalid position"
		}
		// Can claim even if already owned (capture mechanic)
		return true, ""

	case game.ActionMessage:
		if action.Params.Message == "" {
			return false, "message cannot be empty"
		}
		if len(action.Params.Message) > 500 {
			return false, "message too long (max 500 chars)"
		}
		return true, ""

	case game.ActionWait:
		return true, ""

	default:
		return false, "unknown action type"
	}
}

func isValidDirection(dir game.Direction) bool {
	switch dir {
	case game.DirNorth, game.DirSouth, game.DirEast, game.DirWest:
		return true
	default:
		return false
	}
}

// ActionStats tracks LLM response statistics
type ActionStats struct {
	TotalRequests   int
	SuccessfulParse int
	ParseErrors     int
	Timeouts        int
	ValidationFails int
}

// NewActionStats creates a new stats tracker
func NewActionStats() *ActionStats {
	return &ActionStats{}
}

// RecordSuccess records a successful parse
func (s *ActionStats) RecordSuccess() {
	s.TotalRequests++
	s.SuccessfulParse++
}

// RecordParseError records a parse error
func (s *ActionStats) RecordParseError() {
	s.TotalRequests++
	s.ParseErrors++
}

// RecordTimeout records a timeout
func (s *ActionStats) RecordTimeout() {
	s.TotalRequests++
	s.Timeouts++
}

// RecordValidationFail records a validation failure
func (s *ActionStats) RecordValidationFail() {
	s.TotalRequests++
	s.ValidationFails++
}

// SuccessRate returns the success rate as a percentage
func (s *ActionStats) SuccessRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.SuccessfulParse) / float64(s.TotalRequests) * 100
}

// MarshalJSON implements json.Marshaler
func (s *ActionStats) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"total_requests":    s.TotalRequests,
		"successful_parse":  s.SuccessfulParse,
		"parse_errors":      s.ParseErrors,
		"timeouts":          s.Timeouts,
		"validation_fails":  s.ValidationFails,
		"success_rate":      s.SuccessRate(),
	})
}
