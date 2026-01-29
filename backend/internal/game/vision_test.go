package game

import (
	"testing"

	"github.com/google/uuid"
)

func TestCalculateEffectiveVisionRadius_BaseVision(t *testing.T) {
	agent := &Agent{
		ID:          uuid.New(),
		VisionLevel: 1, // No upgrade bonus
	}
	agent.SetPosition(Position{X: 10, Y: 10})

	wom := NewWorldObjectManager()

	got := CalculateEffectiveVisionRadius(agent, 3, wom)
	if got != 3 {
		t.Errorf("expected base vision 3, got %d", got)
	}
}

func TestCalculateEffectiveVisionRadius_WithUpgrade(t *testing.T) {
	agent := &Agent{
		ID:          uuid.New(),
		VisionLevel: 3, // +2 bonus
	}
	agent.SetPosition(Position{X: 10, Y: 10})

	wom := NewWorldObjectManager()

	got := CalculateEffectiveVisionRadius(agent, 3, wom)
	// base 3 + (3-1) = 5
	if got != 5 {
		t.Errorf("expected vision 5 with level 3 upgrade, got %d", got)
	}
}

func TestCalculateEffectiveVisionRadius_WithBeaconInRange(t *testing.T) {
	agentID := uuid.New()
	agent := &Agent{
		ID:          agentID,
		VisionLevel: 1,
	}
	agent.SetPosition(Position{X: 10, Y: 10})

	wom := NewWorldObjectManager()
	// Place beacon within vision range (base 3)
	beacon := NewStructure(StructureBeacon, Position{X: 11, Y: 10}, agentID, 1)
	wom.Add(beacon)

	got := CalculateEffectiveVisionRadius(agent, 3, wom)
	// base 3 + beacon bonus 2 = 5
	if got != 5 {
		t.Errorf("expected vision 5 with beacon bonus, got %d", got)
	}
}

func TestCalculateEffectiveVisionRadius_BeaconOutOfRange(t *testing.T) {
	agentID := uuid.New()
	agent := &Agent{
		ID:          agentID,
		VisionLevel: 1,
	}
	agent.SetPosition(Position{X: 10, Y: 10})

	wom := NewWorldObjectManager()
	// Place beacon far outside vision range (base 3)
	beacon := NewStructure(StructureBeacon, Position{X: 50, Y: 50}, agentID, 1)
	wom.Add(beacon)

	got := CalculateEffectiveVisionRadius(agent, 3, wom)
	// Beacon out of range, no bonus
	if got != 3 {
		t.Errorf("expected base vision 3 with out-of-range beacon, got %d", got)
	}
}

func TestCalculateEffectiveVisionRadius_OtherAgentBeaconIgnored(t *testing.T) {
	agentID := uuid.New()
	otherID := uuid.New()
	agent := &Agent{
		ID:          agentID,
		VisionLevel: 1,
	}
	agent.SetPosition(Position{X: 10, Y: 10})

	wom := NewWorldObjectManager()
	// Place beacon owned by other agent at adjacent position
	beacon := NewStructure(StructureBeacon, Position{X: 11, Y: 10}, otherID, 1)
	wom.Add(beacon)

	got := CalculateEffectiveVisionRadius(agent, 3, wom)
	// Other agent's beacon should not add bonus
	if got != 3 {
		t.Errorf("expected base vision 3 (other agent's beacon), got %d", got)
	}
}
