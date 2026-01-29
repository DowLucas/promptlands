package actions_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
	"github.com/lucas/promptlands/internal/game"
	"github.com/lucas/promptlands/internal/game/actions"
	"github.com/lucas/promptlands/internal/game/testutil"
)

func TestMoveHandler_ActionType(t *testing.T) {
	handler := actions.NewMoveHandler()

	if handler.ActionType() != game.ActionMove {
		t.Errorf("expected ActionType to return %v, got %v", game.ActionMove, handler.ActionType())
	}
}

func TestMoveHandler_SingleStep(t *testing.T) {
	// Use steps=1 to test single-step move behavior
	tests := []struct {
		name      string
		direction game.Direction
		startPos  game.Position
		endPos    game.Position
	}{
		{
			name:      "move north 1 step",
			direction: game.DirNorth,
			startPos:  game.Position{X: 5, Y: 5},
			endPos:    game.Position{X: 5, Y: 4},
		},
		{
			name:      "move south 1 step",
			direction: game.DirSouth,
			startPos:  game.Position{X: 5, Y: 5},
			endPos:    game.Position{X: 5, Y: 6},
		},
		{
			name:      "move east 1 step",
			direction: game.DirEast,
			startPos:  game.Position{X: 5, Y: 5},
			endPos:    game.Position{X: 6, Y: 5},
		},
		{
			name:      "move west 1 step",
			direction: game.DirWest,
			startPos:  game.Position{X: 5, Y: 5},
			endPos:    game.Position{X: 4, Y: 5},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agent := testutil.NewTestAgent(tc.startPos)
			world := testutil.NewTestWorld(10)

			action := game.Action{
				Type:    game.ActionMove,
				AgentID: agent.ID,
				Params:  game.ActionParams{Direction: tc.direction, Steps: 1},
			}
			ctx := testutil.NewTestActionContext(agent, world, action)

			handler := actions.NewMoveHandler()
			result := handler.Process(ctx)

			testutil.AssertActionSuccess(t, result)
			testutil.AssertAgentAt(t, agent, tc.endPos)

			if result.OldPos == nil {
				t.Fatal("expected OldPos to be set")
			}
			if *result.OldPos != tc.startPos {
				t.Errorf("expected OldPos to be %v, got %v", tc.startPos, *result.OldPos)
			}
			if result.NewPos == nil {
				t.Fatal("expected NewPos to be set")
			}
			if *result.NewPos != tc.endPos {
				t.Errorf("expected NewPos to be %v, got %v", tc.endPos, *result.NewPos)
			}
		})
	}
}

func TestMoveHandler_MultiStep(t *testing.T) {
	// Default move speed is 3, so agent should move 3 tiles
	startPos := game.Position{X: 5, Y: 5}
	agent := testutil.NewTestAgent(startPos)
	world := testutil.NewTestWorld(20)

	action := testutil.MoveAction(agent.ID, game.DirEast)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	// With default move speed of 3, agent should move 3 tiles east
	expectedPos := game.Position{X: 8, Y: 5}
	testutil.AssertAgentAt(t, agent, expectedPos)

	if result.Message != "moved 3 steps east" {
		t.Errorf("expected message 'moved 3 steps east', got %q", result.Message)
	}
}

func TestMoveHandler_MultiStepWithBalanceConfig(t *testing.T) {
	startPos := game.Position{X: 5, Y: 5}
	balance := testutil.DefaultTestBalance()
	agent := testutil.NewTestAgentWithBalance(startPos, balance)
	world := testutil.NewTestWorld(20)

	action := testutil.MoveAction(agent.ID, game.DirSouth)
	ctx := testutil.NewTestActionContextWithBalance(agent, world, action, balance)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	// DefaultMoveSpeed is 3, agent at speed level 1 => 3 steps
	expectedPos := game.Position{X: 5, Y: 8}
	testutil.AssertAgentAt(t, agent, expectedPos)
}

func TestMoveHandler_RequestedFewerSteps(t *testing.T) {
	startPos := game.Position{X: 5, Y: 5}
	agent := testutil.NewTestAgent(startPos)
	world := testutil.NewTestWorld(20)

	// Request only 2 steps (less than max of 3)
	action := game.Action{
		Type:    game.ActionMove,
		AgentID: agent.ID,
		Params:  game.ActionParams{Direction: game.DirEast, Steps: 2},
	}
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	expectedPos := game.Position{X: 7, Y: 5}
	testutil.AssertAgentAt(t, agent, expectedPos)

	if result.Message != "moved 2 steps east" {
		t.Errorf("expected message 'moved 2 steps east', got %q", result.Message)
	}
}

func TestMoveHandler_StopsAtObstacle(t *testing.T) {
	// Agent starts at (5,5) moving east, wall at (7,5) should stop after 1 step
	startPos := game.Position{X: 5, Y: 5}
	agent := testutil.NewTestAgent(startPos)
	world := testutil.NewTestWorld(20)

	// Place a blocking structure at (7,5) — agent should move to (6,5) then stop
	worldObjects := game.NewWorldObjectManager()
	wall := game.NewStructure(game.StructureWall, game.Position{X: 7, Y: 5}, agent.ID, 1)
	worldObjects.Add(wall)

	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	agents := map[uuid.UUID]*game.Agent{
		agent.ID: agent,
	}
	agent.InitInventory(itemRegistry)

	action := testutil.MoveAction(agent.ID, game.DirEast)
	ctx := testutil.NewTestActionContextFull(
		agent, world, worldObjects, itemRegistry, recipeRegistry,
		agents, 1, action, nil,
	)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	// Should stop before wall: moved 1 step to (6,5)
	expectedPos := game.Position{X: 6, Y: 5}
	testutil.AssertAgentAt(t, agent, expectedPos)
}

func TestMoveHandler_OutOfBounds(t *testing.T) {
	tests := []struct {
		name      string
		direction game.Direction
		startPos  game.Position
	}{
		{
			name:      "move north at top edge",
			direction: game.DirNorth,
			startPos:  game.Position{X: 5, Y: 0},
		},
		{
			name:      "move south at bottom edge",
			direction: game.DirSouth,
			startPos:  game.Position{X: 5, Y: 9},
		},
		{
			name:      "move east at right edge",
			direction: game.DirEast,
			startPos:  game.Position{X: 9, Y: 5},
		},
		{
			name:      "move west at left edge",
			direction: game.DirWest,
			startPos:  game.Position{X: 0, Y: 5},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			agent := testutil.NewTestAgent(tc.startPos)
			world := testutil.NewTestWorld(10)

			action := testutil.MoveAction(agent.ID, tc.direction)
			ctx := testutil.NewTestActionContext(agent, world, action)

			handler := actions.NewMoveHandler()
			result := handler.Process(ctx)

			testutil.AssertActionFailed(t, result)
			testutil.AssertActionFailedWithMessage(t, result, "path blocked")

			// Agent should not have moved
			testutil.AssertAgentAt(t, agent, tc.startPos)
		})
	}
}

func TestMoveHandler_CanMoveOnAllTerrain(t *testing.T) {
	// Agents can now move on all terrain types including water and mountain
	tests := []struct {
		name    string
		terrain game.TerrainType
	}{
		{
			name:    "move on water",
			terrain: game.TerrainWater,
		},
		{
			name:    "move on mountain",
			terrain: game.TerrainMountain,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startPos := game.Position{X: 5, Y: 5}
			targetPos := game.Position{X: 6, Y: 5} // Moving east
			agent := testutil.NewTestAgent(startPos)
			world := testutil.NewTestWorld(10)

			tile := world.GetTile(targetPos)
			tile.Terrain = tc.terrain

			action := game.Action{
				Type:    game.ActionMove,
				AgentID: agent.ID,
				Params:  game.ActionParams{Direction: game.DirEast, Steps: 1},
			}
			ctx := testutil.NewTestActionContext(agent, world, action)

			handler := actions.NewMoveHandler()
			result := handler.Process(ctx)

			testutil.AssertActionSuccess(t, result)
			testutil.AssertAgentAt(t, agent, targetPos)
		})
	}
}

func TestMoveHandler_BlockedByStructure(t *testing.T) {
	startPos := game.Position{X: 5, Y: 5}
	targetPos := game.Position{X: 6, Y: 5}
	agent := testutil.NewTestAgent(startPos)
	world := testutil.NewTestWorld(10)

	worldObjects := game.NewWorldObjectManager()
	wall := game.NewStructure(game.StructureWall, targetPos, agent.ID, 1)
	worldObjects.Add(wall)

	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	agents := map[uuid.UUID]*game.Agent{
		agent.ID: agent,
	}
	agent.InitInventory(itemRegistry)

	action := game.Action{
		Type:    game.ActionMove,
		AgentID: agent.ID,
		Params:  game.ActionParams{Direction: game.DirEast, Steps: 1},
	}
	ctx := testutil.NewTestActionContextFull(
		agent, world, worldObjects, itemRegistry, recipeRegistry,
		agents, 1, action, nil,
	)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "path blocked")

	testutil.AssertAgentAt(t, agent, startPos)
}

func TestMoveHandler_TileOccupied(t *testing.T) {
	startPos := game.Position{X: 5, Y: 5}
	targetPos := game.Position{X: 6, Y: 5}
	agent := testutil.NewTestAgent(startPos)
	otherAgent := testutil.NewTestAgent(targetPos)
	world := testutil.NewTestWorld(10)

	worldObjects := game.NewWorldObjectManager()
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	agents := map[uuid.UUID]*game.Agent{
		agent.ID:      agent,
		otherAgent.ID: otherAgent,
	}
	agent.InitInventory(itemRegistry)

	action := game.Action{
		Type:    game.ActionMove,
		AgentID: agent.ID,
		Params:  game.ActionParams{Direction: game.DirEast, Steps: 1},
	}
	ctx := testutil.NewTestActionContextFull(
		agent, world, worldObjects, itemRegistry, recipeRegistry,
		agents, 1, action, nil,
	)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "path blocked")

	testutil.AssertAgentAt(t, agent, startPos)
}

func TestMoveHandler_CanMoveToTileWithDeadAgent(t *testing.T) {
	startPos := game.Position{X: 5, Y: 5}
	targetPos := game.Position{X: 6, Y: 5}
	agent := testutil.NewTestAgent(startPos)
	deadAgent := testutil.NewTestAgent(targetPos)
	deadAgent.IsDead = true
	world := testutil.NewTestWorld(20)

	worldObjects := game.NewWorldObjectManager()
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	agents := map[uuid.UUID]*game.Agent{
		agent.ID:     agent,
		deadAgent.ID: deadAgent,
	}
	agent.InitInventory(itemRegistry)

	action := game.Action{
		Type:    game.ActionMove,
		AgentID: agent.ID,
		Params:  game.ActionParams{Direction: game.DirEast, Steps: 1},
	}
	ctx := testutil.NewTestActionContextFull(
		agent, world, worldObjects, itemRegistry, recipeRegistry,
		agents, 1, action, nil,
	)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	testutil.AssertAgentAt(t, agent, targetPos)
}

func TestMoveHandler_CanMoveOnPassableTerrain(t *testing.T) {
	tests := []struct {
		name    string
		terrain game.TerrainType
	}{
		{
			name:    "move on plains",
			terrain: game.TerrainPlains,
		},
		{
			name:    "move on forest",
			terrain: game.TerrainForest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startPos := game.Position{X: 5, Y: 5}
			targetPos := game.Position{X: 6, Y: 5}
			agent := testutil.NewTestAgent(startPos)
			world := testutil.NewTestWorld(20)

			tile := world.GetTile(targetPos)
			tile.Terrain = tc.terrain

			action := game.Action{
				Type:    game.ActionMove,
				AgentID: agent.ID,
				Params:  game.ActionParams{Direction: game.DirEast, Steps: 1},
			}
			ctx := testutil.NewTestActionContext(agent, world, action)

			handler := actions.NewMoveHandler()
			result := handler.Process(ctx)

			testutil.AssertActionSuccess(t, result)
			testutil.AssertAgentAt(t, agent, targetPos)
		})
	}
}

func TestMoveHandler_SpeedUpgrade(t *testing.T) {
	startPos := game.Position{X: 5, Y: 5}
	balance := testutil.DefaultTestBalance()
	agent := testutil.NewTestAgentWithBalance(startPos, balance)
	world := testutil.NewTestWorld(20)

	// Upgrade speed to level 2 => base 3 + 1 = 4 steps
	agent.SpeedLevel = 2

	action := testutil.MoveAction(agent.ID, game.DirEast)
	ctx := testutil.NewTestActionContextWithBalance(agent, world, action, balance)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	expectedPos := game.Position{X: 9, Y: 5} // 5 + 4 = 9
	testutil.AssertAgentAt(t, agent, expectedPos)

	if result.Message != "moved 4 steps east" {
		t.Errorf("expected 'moved 4 steps east', got %q", result.Message)
	}
}

func TestMoveHandler_CustomMoveSpeed(t *testing.T) {
	startPos := game.Position{X: 5, Y: 5}
	balance := testutil.DefaultTestBalance()
	balance.Agent.DefaultMoveSpeed = 5
	agent := testutil.NewTestAgentWithBalance(startPos, balance)
	world := testutil.NewTestWorld(20)

	action := testutil.MoveAction(agent.ID, game.DirEast)
	ctx := testutil.NewTestActionContextWithBalance(agent, world, action, balance)

	handler := actions.NewMoveHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	expectedPos := game.Position{X: 10, Y: 5} // 5 + 5 = 10
	testutil.AssertAgentAt(t, agent, expectedPos)
}

func singleStepBalance() *config.BalanceConfig {
	b := config.DefaultBalanceConfig()
	b.Agent.DefaultMoveSpeed = 1
	return &b
}
