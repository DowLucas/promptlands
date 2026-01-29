package actions_test

import (
	"strings"
	"testing"

	"github.com/lucas/promptlands/internal/game"
	"github.com/lucas/promptlands/internal/game/actions"
	"github.com/lucas/promptlands/internal/game/testutil"
)

func TestClaimHandler_ActionType(t *testing.T) {
	handler := actions.NewClaimHandler()

	if handler.ActionType() != game.ActionClaim {
		t.Errorf("expected ActionType to return %v, got %v", game.ActionClaim, handler.ActionType())
	}
}

func TestClaimHandler_ClaimUnownedArea(t *testing.T) {
	// Setup: create agent and world with enough space for radius claim
	pos := game.Position{X: 10, Y: 10}
	agent := testutil.NewTestAgent(pos)
	world := testutil.NewTestWorld(30)

	// Create claim action and context
	action := testutil.ClaimAction(agent.ID)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewClaimHandler()
	result := handler.Process(ctx)

	// Verify success
	testutil.AssertActionSuccess(t, result)

	// Should have claimed multiple tiles
	if !strings.Contains(result.Message, "claimed") {
		t.Errorf("expected message to contain 'claimed', got %q", result.Message)
	}

	// Verify the agent's tile is owned
	tile := world.GetTile(pos)
	if tile.OwnerID == nil {
		t.Fatal("expected agent's tile to be owned after claim")
	}
	if *tile.OwnerID != agent.ID {
		t.Errorf("expected tile owner to be %v, got %v", agent.ID, *tile.OwnerID)
	}

	// Verify ClaimedAt is set
	if result.ClaimedAt == nil {
		t.Fatal("expected ClaimedAt to be set")
	}
	if result.ClaimedAt.X != pos.X || result.ClaimedAt.Y != pos.Y {
		t.Errorf("expected ClaimedAt to be %v, got %v", pos, *result.ClaimedAt)
	}
}

func TestClaimHandler_CaptureEnemyTiles(t *testing.T) {
	pos := game.Position{X: 10, Y: 10}
	agent := testutil.NewTestAgent(pos)
	enemy := testutil.NewTestAgent(game.Position{X: 3, Y: 3})
	world := testutil.NewTestWorld(30)

	// Set the center tile as owned by enemy
	world.SetOwner(pos, &enemy.ID)

	action := testutil.ClaimAction(agent.ID)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewClaimHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	// Message should mention captured tiles
	if !strings.Contains(result.Message, "captured from enemies") {
		t.Errorf("expected message to mention captured tiles, got %q", result.Message)
	}

	// Verify the center tile is now owned by claiming agent
	tile := world.GetTile(pos)
	if tile.OwnerID == nil {
		t.Fatal("expected tile to be owned after capture")
	}
	if *tile.OwnerID != agent.ID {
		t.Errorf("expected tile owner to be %v, got %v", agent.ID, *tile.OwnerID)
	}
}

func TestClaimHandler_AllAlreadyOwned(t *testing.T) {
	pos := game.Position{X: 10, Y: 10}
	agent := testutil.NewTestAgent(pos)
	world := testutil.NewTestWorld(30)

	// Pre-own all tiles in a large radius around the agent
	for dy := -10; dy <= 10; dy++ {
		for dx := -10; dx <= 10; dx++ {
			tilePos := game.Position{X: pos.X + dx, Y: pos.Y + dy}
			if world.IsValidPosition(tilePos) {
				world.SetOwner(tilePos, &agent.ID)
			}
		}
	}

	action := testutil.ClaimAction(agent.ID)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewClaimHandler()
	result := handler.Process(ctx)

	// Should succeed but claim 0 new tiles
	testutil.AssertActionSuccess(t, result)
	if result.Message != "all tiles in radius already owned" {
		t.Errorf("expected 'all tiles in radius already owned', got %q", result.Message)
	}
}

func TestClaimHandler_ClaimsAllTerrainTypes(t *testing.T) {
	pos := game.Position{X: 10, Y: 10}
	agent := testutil.NewTestAgent(pos)
	world := testutil.NewTestWorld(30)

	// Place water around the agent (except agent's own tile)
	for dy := -5; dy <= 5; dy++ {
		for dx := -5; dx <= 5; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			tilePos := game.Position{X: pos.X + dx, Y: pos.Y + dy}
			if world.IsValidPosition(tilePos) {
				tile := world.GetTile(tilePos)
				tile.Terrain = game.TerrainWater
			}
		}
	}

	action := testutil.ClaimAction(agent.ID)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewClaimHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)
	// All terrain types are now claimable, so should claim more than just 1 tile
	if !strings.Contains(result.Message, "claimed") {
		t.Errorf("expected message to contain 'claimed', got %q", result.Message)
	}
	// With default claim radius of 4, agent should claim many tiles (including water)
	if result.Message == "claimed 1 tiles" {
		t.Errorf("expected more than 1 tile claimed since water is now claimable, got %q", result.Message)
	}
}

func TestClaimHandler_ClaimRadiusWithUpgrade(t *testing.T) {
	pos := game.Position{X: 15, Y: 15}
	balance := testutil.DefaultTestBalance()
	agent := testutil.NewTestAgentWithBalance(pos, balance)
	world := testutil.NewTestWorld(40)

	// Set claim level to 2 => radius should be 4 + 1 = 5
	agent.ClaimLevel = 2

	action := testutil.ClaimAction(agent.ID)
	ctx := testutil.NewTestActionContextWithBalance(agent, world, action, balance)

	handler := actions.NewClaimHandler()
	resultBase := handler.Process(ctx)

	// Reset: create agent at level 1 for comparison
	agent2 := testutil.NewTestAgentWithBalance(pos, balance)
	world2 := testutil.NewTestWorld(40)
	action2 := testutil.ClaimAction(agent2.ID)
	ctx2 := testutil.NewTestActionContextWithBalance(agent2, world2, action2, balance)

	resultSmaller := handler.Process(ctx2)

	testutil.AssertActionSuccess(t, resultBase)
	testutil.AssertActionSuccess(t, resultSmaller)

	// The upgraded agent should have claimed more tiles
	// We can't easily count from messages, but both should succeed
	// Upgraded claim radius is 5, base is 4 — so upgraded agent claims more
	if resultBase.Message == resultSmaller.Message {
		t.Error("expected different claim counts for different claim levels")
	}
}
