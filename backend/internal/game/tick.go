package game

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/game/worldgen"
)

// processTick handles a single game tick
func (e *Engine) processTick(ctx context.Context) {
	e.mu.Lock()
	e.tick++
	tick := e.tick
	agents := make([]*Agent, 0, len(e.agents))
	for _, a := range e.agents {
		agents = append(agents, a)
	}
	e.mu.Unlock()

	log.Printf("Game %s: Processing tick %d", e.ID, tick)

	// Phase 1: Process respawns
	respawnedAgents := e.processRespawns(tick)

	// Phase 2: Passive energy income (before actions)
	e.processPassiveIncome()

	// Phase 2b: Spawn new resources based on biomes
	spawnedObjects := e.processResourceSpawning()

	// Phase 2c: Auto-absorb resources from owned territory
	absorptionResults := e.processResourceAbsorption()

	// Build context for each agent (skip dead agents)
	aliveAgents := make([]*Agent, 0, len(agents))
	for _, a := range agents {
		if !a.IsDead {
			aliveAgents = append(aliveAgents, a)
		}
	}
	contexts := e.buildAgentContexts(aliveAgents)

	// Fan-out LLM requests with timeout
	llmCtx, cancel := context.WithTimeout(ctx, e.config.TickDuration-2*time.Second)
	actions := e.requestActions(llmCtx, contexts)
	cancel()

	// Add actions to resolver
	e.resolver.AddActions(actions)

	// Resolve conflicts (first-come priority)
	orderedActions := e.resolver.Resolve()

	// Process actions with full processor
	processor := NewActionProcessor(e.world, e.agents, e.worldObjects, e.itemRegistry, e.recipeRegistry, tick, &e.balance, e.handlerRegistry)
	results := processor.ProcessAll(orderedActions)

	// Append territory resource absorption results
	results = append(results, absorptionResults...)

	// Phase 3: Trigger traps after movement
	trapResults := e.processTrapTriggers(results)
	results = append(results, trapResults...)

	// Phase 3b: Auto-activate interactive objects (shrines, caches, portals)
	interactiveResults := e.processInteractiveActivation(results)
	results = append(results, interactiveResults...)

	// Phase 4: Process despawns
	removedObjects := e.worldObjects.ProcessDespawns(tick)
	depletedResources := e.worldObjects.ProcessDepletedResources()
	removedObjects = append(removedObjects, depletedResources...)

	// Collect messages from this tick
	tickMessages := e.collectMessages(orderedActions)

	// Build tick update
	update := e.buildTickUpdate(tick, orderedActions, results, tickMessages, removedObjects, respawnedAgents, spawnedObjects)

	// Broadcast to all connected clients with per-player visibility and inventory
	if e.broadcaster != nil {
		e.broadcaster.BroadcastToGameWithVisibility(e.ID, update, e.getVisibleTilesForPlayer, e.getPlayerInventory)
	}

	// Check win condition
	if tick >= e.config.WinAfterTicks {
		e.endGame()
	}
}

// processRespawns handles agent respawns
func (e *Engine) processRespawns(tick int) []uuid.UUID {
	respawned := make([]uuid.UUID, 0)

	for _, agent := range e.agents {
		if agent.ShouldRespawn(tick) {
			// Find a valid spawn position at map edge
			pos := e.findEdgeSpawnPosition()
			agent.Respawn(pos)
			respawned = append(respawned, agent.ID)
			log.Printf("Agent %s respawned at (%d,%d)", agent.Name, pos.X, pos.Y)
		}
	}

	return respawned
}

// findEdgeSpawnPosition finds a valid spawn position at the map edge
func (e *Engine) findEdgeSpawnPosition() Position {
	size := e.world.Size()
	edges := []Position{}

	// Collect all passable edge positions
	for x := 0; x < size; x++ {
		// Top edge
		pos := Position{X: x, Y: 0}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
		// Bottom edge
		pos = Position{X: x, Y: size - 1}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
	}
	for y := 1; y < size-1; y++ {
		// Left edge
		pos := Position{X: 0, Y: y}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
		// Right edge
		pos = Position{X: size - 1, Y: y}
		if e.isValidSpawnPosition(pos) {
			edges = append(edges, pos)
		}
	}

	if len(edges) == 0 {
		// Fallback to center if no edges available
		return Position{X: size / 2, Y: size / 2}
	}

	// Pick a random edge position
	return edges[time.Now().UnixNano()%int64(len(edges))]
}

// isValidSpawnPosition checks if a position is valid for spawning
func (e *Engine) isValidSpawnPosition(pos Position) bool {
	tile := e.world.GetTile(pos)
	if tile == nil {
		return false
	}
	if tile.Terrain == TerrainWater || tile.Terrain == TerrainMountain {
		return false
	}
	// Check for blocking structures
	if e.worldObjects.HasBlockingObject(pos) {
		return false
	}
	// Check for other agents
	for _, agent := range e.agents {
		if !agent.IsDead && agent.GetPosition() == pos {
			return false
		}
	}
	return true
}

// processResourceAbsorption auto-harvests resources on owned tiles into agent inventories
func (e *Engine) processResourceAbsorption() []ActionResult {
	var results []ActionResult

	for _, agent := range e.agents {
		if agent.IsDead {
			continue
		}
		if agent.Inventory == nil {
			continue
		}

		ownedTiles := e.world.GetOwnedTiles(agent.ID)
		for _, pos := range ownedTiles {
			resource := e.worldObjects.GetResourceAt(pos)
			if resource == nil || resource.Remaining <= 0 {
				continue
			}

			harvested := resource.Harvest(1)
			if harvested == 0 {
				continue
			}

			itemID := string(resource.ResourceType)
			remaining := agent.Inventory.AddItem(itemID, harvested)
			if remaining == harvested {
				// Inventory full, put resource back
				resource.Remaining += harvested
				continue
			}

			actual := harvested - remaining
			results = append(results, ActionResult{
				AgentID:      agent.ID,
				Action:       ActionHarvest,
				Success:      true,
				Harvested:    itemID,
				ItemQuantity: actual,
				Message:      fmt.Sprintf("territory absorbed %d %s", actual, itemID),
			})
		}
	}

	return results
}

// processPassiveIncome gives agents energy based on owned tiles
func (e *Engine) processPassiveIncome() {
	for _, agent := range e.agents {
		if agent.IsDead {
			continue
		}

		ownedTiles := e.world.GetOwnedTiles(agent.ID)
		energyGain := 0

		for _, pos := range ownedTiles {
			tile := e.world.GetTile(pos)
			if tile == nil {
				continue
			}
			switch tile.Terrain {
			case TerrainPlains:
				energyGain += 1
			case TerrainForest:
				energyGain += 2
			}
		}

		if energyGain > 0 {
			agent.AddEnergy(energyGain)
		}
	}
}

// processResourceSpawning spawns resource nodes each tick based on biome types
func (e *Engine) processResourceSpawning() []WorldObjectSnapshot {
	if e.biomeRegistry == nil || e.lootTables == nil || e.spawnRng == nil {
		return nil
	}

	mapSize := e.world.Size()

	// Check resource cap: max resources = mapSize * mapSize / 100
	maxResources := (mapSize * mapSize) / 100
	currentResources := e.worldObjects.CountResourceNodes()
	if currentResources >= maxResources {
		return nil
	}

	// Calculate spawn attempts: mapSize * mapSize / 10000, clamped [5, 500]
	attempts := (mapSize * mapSize) / 10000
	if attempts < 5 {
		attempts = 5
	}
	if attempts > 500 {
		attempts = 500
	}

	// Apply resource spawn rate multiplier from config
	spawnRate := e.config.ResourceSpawnRate
	if spawnRate <= 0 {
		spawnRate = 1.0
	}
	attempts = int(float64(attempts) * spawnRate)
	if attempts < 1 {
		attempts = 1
	}

	var spawned []WorldObjectSnapshot

	for i := 0; i < attempts; i++ {
		// Stop if we've hit the cap
		if currentResources+len(spawned) >= maxResources {
			break
		}

		// Pick a random tile
		x := e.spawnRng.Intn(mapSize)
		y := e.spawnRng.Intn(mapSize)
		pos := Position{X: x, Y: y}

		tile := e.world.GetTile(pos)
		if tile == nil {
			continue
		}

		// Get biome properties
		biomeType := worldgen.BiomeType(tile.Biome)
		biomeProps, ok := e.biomeRegistry.GetBiome(biomeType)
		if !ok || !biomeProps.Passable || biomeProps.ResourceChance <= 0 {
			continue
		}

		// Roll against biome's resource chance
		if e.spawnRng.Float64() > biomeProps.ResourceChance {
			continue
		}

		// Skip if tile already has a resource node or structure
		if e.worldObjects.GetResourceAt(pos) != nil {
			continue
		}
		if e.worldObjects.GetStructureAt(pos) != nil {
			continue
		}

		// Roll biome's loot table to determine resource type
		biomeLootConfig, hasLoot := e.biomeLoot[biomeType]
		if !hasLoot {
			continue
		}

		lootResults := e.lootTables.Roll(biomeLootConfig.ResourceTable, 1.0)
		if len(lootResults) == 0 {
			continue
		}

		result := lootResults[0]
		resourceType := itemToResourceType(result.ItemID)
		if resourceType == "" {
			continue
		}

		node := NewResourceNode(resourceType, pos, result.Quantity)
		e.worldObjects.Add(node)
		spawned = append(spawned, node.Snapshot())
	}

	if len(spawned) > 0 {
		log.Printf("Game %s: Spawned %d resource nodes (total: %d/%d)", e.ID, len(spawned), currentResources+len(spawned), maxResources)
	}

	return spawned
}

// itemToResourceType converts a loot table item ID to a ResourceType
func itemToResourceType(itemID string) ResourceType {
	switch itemID {
	case "wood", "ancient_wood":
		return ResourceWood
	case "stone", "frost_stone", "null_stone", "rune_stone":
		return ResourceStone
	case "crystal", "ice_crystal", "fire_crystal", "resonant_crystal", "void_crystal", "sun_crystal", "energy_crystal":
		return ResourceCrystal
	case "herb", "dark_herb", "frozen_herb", "cactus_fruit":
		return ResourceHerb
	default:
		// For special biome resources, map to crystal
		return ResourceCrystal
	}
}

// processTrapTriggers checks for and triggers traps after movement
func (e *Engine) processTrapTriggers(results []ActionResult) []ActionResult {
	trapResults := make([]ActionResult, 0)

	for _, result := range results {
		if result.Action == ActionMove && result.Success && result.NewPos != nil {
			agent := e.agents[result.AgentID]
			if agent == nil || agent.IsDead {
				continue
			}

			// Check for traps at new position
			traps := e.worldObjects.GetTrapsAt(*result.NewPos)
			for _, trap := range traps {
				// Don't trigger own traps
				if trap.OwnerID != nil && *trap.OwnerID == agent.ID {
					continue
				}

				// Trigger trap
				damage := trap.Damage
				if damage == 0 {
					damage = 1
				}

				killed := agent.TakeDamage(damage)
				trapResult := ActionResult{
					AgentID:     agent.ID,
					Action:      ActionFight, // Use FIGHT to indicate damage
					Success:     true,
					DamageDealt: damage,
					Message:     "triggered a trap",
				}

				if killed {
					agent.Kill(e.tick)
					// Clear tiles on death
					ownedTiles := e.world.GetOwnedTiles(agent.ID)
					for _, pos := range ownedTiles {
						e.world.SetOwner(pos, nil)
					}
					if agent.Inventory != nil {
						agent.Inventory.Clear()
					}
					trapResult.Message = "killed by a trap"
				}

				trapResults = append(trapResults, trapResult)

				// Remove the trap after triggering
				e.worldObjects.Remove(trap.ID)
			}
		}
	}

	return trapResults
}

// processInteractiveActivation auto-activates interactive objects when agents step on them
func (e *Engine) processInteractiveActivation(results []ActionResult) []ActionResult {
	var interactiveResults []ActionResult

	for _, result := range results {
		if result.Action != ActionMove || !result.Success || result.NewPos == nil {
			continue
		}

		agent := e.agents[result.AgentID]
		if agent == nil || agent.IsDead {
			continue
		}

		interactive := e.worldObjects.GetInteractiveAt(*result.NewPos)
		if interactive == nil {
			continue
		}

		switch interactive.InteractiveType {
		case InteractiveShrine:
			if !interactive.CanBeActivatedBy(agent.ID) {
				continue
			}
			interactive.Activate(agent.ID)
			agent.MaxHP += interactive.HPReward
			agent.Heal(interactive.HPReward)
			agent.AddMemory("Activated a shrine, gained +1 max HP")
			interactiveResults = append(interactiveResults, ActionResult{
				AgentID: agent.ID,
				Action:  ActionInteract,
				Success: true,
				Message: fmt.Sprintf("activated shrine - max HP increased to %d", agent.MaxHP),
			})

		case InteractiveCache:
			if !interactive.CanBeActivatedBy(agent.ID) {
				continue
			}
			interactive.Activate(agent.ID)
			agent.AddEnergy(interactive.EnergyReward)
			e.worldObjects.Remove(interactive.ID)
			agent.AddMemory(fmt.Sprintf("Found a cache with %d energy", interactive.EnergyReward))
			interactiveResults = append(interactiveResults, ActionResult{
				AgentID: agent.ID,
				Action:  ActionInteract,
				Success: true,
				Message: fmt.Sprintf("found cache - gained %d energy", interactive.EnergyReward),
			})

		case InteractivePortal:
			if interactive.Destination == nil {
				continue
			}
			agent.SetPosition(*interactive.Destination)
			agent.AddMemory(fmt.Sprintf("Used portal to (%d,%d)", interactive.Destination.X, interactive.Destination.Y))
			interactiveResults = append(interactiveResults, ActionResult{
				AgentID: agent.ID,
				Action:  ActionInteract,
				Success: true,
				NewPos:  interactive.Destination,
				Message: fmt.Sprintf("teleported to (%d,%d)", interactive.Destination.X, interactive.Destination.Y),
			})
		}
		// Obelisk: passive flavor only, skip auto-activation
	}

	return interactiveResults
}

// collectMessages extracts messages from actions
func (e *Engine) collectMessages(actions []Action) []GameMessage {
	messages := make([]GameMessage, 0)

	for _, action := range actions {
		if action.Type == ActionMessage && action.Params.Message != "" {
			msg := GameMessage{
				Tick:        e.tick,
				FromAgentID: action.AgentID,
				ToAgentID:   action.Params.Target,
				Content:     action.Params.Message,
			}
			messages = append(messages, msg)
		}
	}

	e.mu.Lock()
	e.messages = append(e.messages, messages...)
	e.mu.Unlock()

	return messages
}

// getMessagesForTick returns messages from a specific tick
func (e *Engine) getMessagesForTick(tick int) []GameMessage {
	result := make([]GameMessage, 0)
	for _, msg := range e.messages {
		if msg.Tick == tick {
			result = append(result, msg)
		}
	}
	return result
}

// filterMessagesForAgent filters messages that an agent should receive
func (e *Engine) filterMessagesForAgent(messages []GameMessage, agentID uuid.UUID) []IncomingMessage {
	result := make([]IncomingMessage, 0)

	for _, msg := range messages {
		// Skip own messages
		if msg.FromAgentID == agentID {
			continue
		}

		// Include if broadcast or targeted to this agent
		if msg.ToAgentID == nil || *msg.ToAgentID == agentID {
			fromAgent, ok := e.agents[msg.FromAgentID]
			fromName := "Unknown"
			if ok {
				fromName = fromAgent.Name
			}

			result = append(result, IncomingMessage{
				FromAgentID:   msg.FromAgentID,
				FromAgentName: fromName,
				Content:       msg.Content,
				IsBroadcast:   msg.ToAgentID == nil,
			})
		}
	}

	return result
}

// buildTickUpdate creates the tick update message
func (e *Engine) buildTickUpdate(tick int, actions []Action, results []ActionResult, messages []GameMessage, removedObjects []uuid.UUID, respawnedAgents []uuid.UUID, spawnedResources []WorldObjectSnapshot) TickUpdate {
	// Collect tile changes from claim actions and death-related changes
	tileChanges := make([]TileChange, 0)
	for _, result := range results {
		if result.Action == ActionClaim && result.Success {
			agent := e.agents[result.AgentID]
			if agent == nil {
				continue
			}
			for _, tilePos := range result.ClaimedTiles {
				tileChanges = append(tileChanges, TileChange{
					X:       tilePos.X,
					Y:       tilePos.Y,
					OwnerID: &agent.ID,
				})
			}
		}
	}

	// Collect agent snapshots
	agentSnapshots := make([]AgentSnapshot, 0, len(e.agents))
	for _, agent := range e.agents {
		agentSnapshots = append(agentSnapshots, agent.Snapshot())
	}

	// Collect newly added objects (structures placed this tick via USE + spawned resources)
	addedObjects := make([]WorldObjectSnapshot, 0, len(spawnedResources))
	addedObjects = append(addedObjects, spawnedResources...)
	for _, result := range results {
		if result.Placed != "" && result.Success {
			// Find the structure that was just placed
			agent := e.agents[result.AgentID]
			if agent != nil {
				pos := agent.GetPosition()
				structure := e.worldObjects.GetStructureAt(pos)
				if structure != nil && structure.CreatedTick == tick {
					addedObjects = append(addedObjects, structure.Snapshot())
				}
			}
		}
	}

	return TickUpdate{
		Type:   "tick",
		Tick:   tick,
		GameID: e.ID,
		Changes: TickChanges{
			Tiles:          tileChanges,
			Agents:         agentSnapshots,
			Messages:       messages,
			Results:        results,
			ObjectsAdded:   addedObjects,
			ObjectsRemoved: removedObjects,
			Respawned:      respawnedAgents,
		},
	}
}
