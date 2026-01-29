package actions_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/game"
	"github.com/lucas/promptlands/internal/game/actions"
	"github.com/lucas/promptlands/internal/game/testutil"
)

func TestFightHandler_ActionType(t *testing.T) {
	handler := actions.NewFightHandler()

	if handler.ActionType() != game.ActionFight {
		t.Errorf("expected ActionType() to return ActionFight, got %v", handler.ActionType())
	}
}

func TestFightHandler_NoTarget(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})

	// Create action with no target (nil)
	action := testutil.FightAction(agent.ID, nil)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "no target specified")
}

func TestFightHandler_TargetNotFound(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})

	// Create action with a non-existent target ID
	nonExistentID := uuid.New()
	action := testutil.FightAction(agent.ID, &nonExistentID)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "target not found")
}

func TestFightHandler_TargetNotAdjacent(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})
	target := testutil.NewTestAgent(game.Position{X: 8, Y: 8}) // 3 tiles away

	// Create action with a target that is too far away
	action := testutil.FightAction(agent.ID, &target.ID)

	// Use NewTestActionContextFull to include both agents in the agents map
	agents := map[uuid.UUID]*game.Agent{
		agent.ID:  agent,
		target.ID: target,
	}
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	worldObjects := game.NewWorldObjectManager()

	ctx := testutil.NewTestActionContextFull(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		1, // currentTick
		action,
		nil,
	)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "target not adjacent")
}

func TestFightHandler_TargetNotAdjacent_TwoTilesAway(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})
	target := testutil.NewTestAgent(game.Position{X: 5, Y: 7}) // 2 tiles away vertically

	action := testutil.FightAction(agent.ID, &target.ID)

	agents := map[uuid.UUID]*game.Agent{
		agent.ID:  agent,
		target.ID: target,
	}
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	worldObjects := game.NewWorldObjectManager()

	ctx := testutil.NewTestActionContextFull(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		1,
		action,
		nil,
	)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "target not adjacent")
}

func TestFightHandler_DealsDamage(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})
	target := testutil.NewTestAgent(game.Position{X: 5, Y: 6}) // Adjacent (1 tile away)

	initialHP := target.GetHP()
	expectedDamage := agent.GetEffectiveStrength()

	action := testutil.FightAction(agent.ID, &target.ID)

	agents := map[uuid.UUID]*game.Agent{
		agent.ID:  agent,
		target.ID: target,
	}
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	worldObjects := game.NewWorldObjectManager()

	ctx := testutil.NewTestActionContextFull(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		1,
		action,
		nil,
	)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)

	// Verify damage was dealt
	if result.DamageDealt != expectedDamage {
		t.Errorf("expected damage dealt to be %d, got %d", expectedDamage, result.DamageDealt)
	}

	// Verify target HP was reduced
	expectedHP := initialHP - expectedDamage
	if target.GetHP() != expectedHP {
		t.Errorf("expected target HP to be %d, got %d", expectedHP, target.GetHP())
	}

	// Verify target ID is set in result
	if result.TargetID == nil || *result.TargetID != target.ID {
		t.Errorf("expected TargetID to be %v, got %v", target.ID, result.TargetID)
	}
}

func TestFightHandler_DealsDamage_DiagonallyAdjacent(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})
	target := testutil.NewTestAgent(game.Position{X: 6, Y: 6}) // Diagonally adjacent

	initialHP := target.GetHP()
	expectedDamage := agent.GetEffectiveStrength()

	action := testutil.FightAction(agent.ID, &target.ID)

	agents := map[uuid.UUID]*game.Agent{
		agent.ID:  agent,
		target.ID: target,
	}
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	worldObjects := game.NewWorldObjectManager()

	ctx := testutil.NewTestActionContextFull(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		1,
		action,
		nil,
	)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)

	// Verify damage was dealt
	expectedHP := initialHP - expectedDamage
	if target.GetHP() != expectedHP {
		t.Errorf("expected target HP to be %d, got %d", expectedHP, target.GetHP())
	}
}

func TestFightHandler_KillsTarget(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})
	target := testutil.NewTestAgent(game.Position{X: 5, Y: 6}) // Adjacent

	// Give target some owned tiles
	targetPos := target.GetPosition()
	world.SetOwner(targetPos, &target.ID)
	adjacentPos := game.Position{X: targetPos.X + 1, Y: targetPos.Y}
	world.SetOwner(adjacentPos, &target.ID)

	// Verify tiles are owned before attack
	ownedTilesBefore := world.GetOwnedTiles(target.ID)
	if len(ownedTilesBefore) != 2 {
		t.Fatalf("expected target to own 2 tiles before attack, got %d", len(ownedTilesBefore))
	}

	// Give target some inventory items
	itemRegistry := game.DefaultItemRegistry()
	target.InitInventory(itemRegistry)
	target.Inventory.AddItem("wood", 5)
	target.Inventory.AddItem("stone", 3)

	// Verify inventory has items before attack
	woodBefore := target.Inventory.GetItemCount("wood")
	stoneBefore := target.Inventory.GetItemCount("stone")
	if woodBefore != 5 || stoneBefore != 3 {
		t.Fatalf("expected target inventory to have wood:5, stone:3, got wood:%d, stone:%d", woodBefore, stoneBefore)
	}

	// Set target HP low enough to be killed in one hit
	damage := agent.GetEffectiveStrength()
	target.SetHP(damage) // Set HP equal to damage so one hit kills

	action := testutil.FightAction(agent.ID, &target.ID)

	agents := map[uuid.UUID]*game.Agent{
		agent.ID:  agent,
		target.ID: target,
	}
	recipeRegistry := game.DefaultRecipeRegistry()
	worldObjects := game.NewWorldObjectManager()

	ctx := testutil.NewTestActionContextFull(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		1,
		action,
		nil,
	)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionSuccess(t, result)

	// Verify target is dead
	if !target.IsDead {
		t.Error("expected target to be dead after lethal damage")
	}

	// Verify target HP is 0
	if target.GetHP() != 0 {
		t.Errorf("expected target HP to be 0, got %d", target.GetHP())
	}

	// Verify tiles are cleared
	ownedTilesAfter := world.GetOwnedTiles(target.ID)
	if len(ownedTilesAfter) != 0 {
		t.Errorf("expected target to own 0 tiles after death, got %d", len(ownedTilesAfter))
	}

	// Verify inventory is cleared
	woodAfter := target.Inventory.GetItemCount("wood")
	stoneAfter := target.Inventory.GetItemCount("stone")
	if woodAfter != 0 || stoneAfter != 0 {
		t.Errorf("expected target inventory to be cleared, got wood:%d, stone:%d", woodAfter, stoneAfter)
	}

	// Verify message contains "killed"
	if result.Message == "" {
		t.Error("expected result message to be set")
	}
}

func TestFightHandler_TargetAlreadyDead(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})
	target := testutil.NewTestAgent(game.Position{X: 5, Y: 6}) // Adjacent

	// Mark target as already dead
	target.Kill(0)

	action := testutil.FightAction(agent.ID, &target.ID)

	agents := map[uuid.UUID]*game.Agent{
		agent.ID:  agent,
		target.ID: target,
	}
	itemRegistry := game.DefaultItemRegistry()
	recipeRegistry := game.DefaultRecipeRegistry()
	worldObjects := game.NewWorldObjectManager()

	ctx := testutil.NewTestActionContextFull(
		agent,
		world,
		worldObjects,
		itemRegistry,
		recipeRegistry,
		agents,
		1,
		action,
		nil,
	)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "target is already dead")
}

func TestFightHandler_CannotFightSelf(t *testing.T) {
	world := testutil.NewTestWorld(10)
	agent := testutil.NewTestAgent(game.Position{X: 5, Y: 5})

	// Create action targeting self
	action := testutil.FightAction(agent.ID, &agent.ID)
	ctx := testutil.NewTestActionContext(agent, world, action)

	handler := actions.NewFightHandler()
	result := handler.Process(ctx)

	// Should fail because dx == 0 && dy == 0 (same position)
	testutil.AssertActionFailed(t, result)
	testutil.AssertActionFailedWithMessage(t, result, "target not adjacent")
}
