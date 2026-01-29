package game

import (
	"context"
	"log"
	"sync"
	"time"
)

// buildAgentContexts creates context for each agent
func (e *Engine) buildAgentContexts(agents []*Agent) []AgentContext {
	contexts := make([]AgentContext, len(agents))

	e.mu.RLock()
	tickMessages := e.getMessagesForTick(e.tick - 1)
	e.mu.RUnlock()

	for i, agent := range agents {
		pos := agent.GetPosition()
		currentTile := e.world.GetTile(pos)

		// Determine current tile ownership status
		currentTileOwned := false
		currentTileEnemy := false
		if currentTile != nil && currentTile.OwnerID != nil {
			if *currentTile.OwnerID == agent.ID {
				currentTileOwned = true
			} else {
				currentTileEnemy = true
			}
		}

		// Calculate effective vision (base + upgrades + beacons)
		visionRadius := CalculateEffectiveVisionRadius(agent, e.config.VisionRadius, e.worldObjects)

		// Calculate energy income per tick
		ownedTiles := e.world.GetOwnedTiles(agent.ID)
		energyPerTick := 0
		for _, tilePos := range ownedTiles {
			tile := e.world.GetTile(tilePos)
			if tile != nil {
				switch tile.Terrain {
				case TerrainPlains:
					energyPerTick += 1
				case TerrainForest:
					energyPerTick += 2
				}
			}
		}

		// Get visible objects (excluding hidden traps from other players)
		visibleObjects := e.worldObjects.GetVisibleObjects(pos, visionRadius, &agent.ID)

		// Get visible agents
		visibleAgents := make([]*AgentSnapshot, 0)
		for _, other := range e.agents {
			if other.ID == agent.ID || other.IsDead {
				continue
			}
			otherPos := other.GetPosition()
			dx := otherPos.X - pos.X
			dy := otherPos.Y - pos.Y
			if dx >= -visionRadius && dx <= visionRadius && dy >= -visionRadius && dy <= visionRadius {
				snap := other.Snapshot()
				visibleAgents = append(visibleAgents, &snap)
			}
		}

		// Get visible tiles for this agent
		visibleTiles := e.world.GetVisibleTiles(pos, visionRadius)

		// Update agent's explored tiles (fog of war persistence)
		agent.UpdateExploredTiles(visibleTiles)

		// Get current biome
		currentBiome := ""
		if currentTile != nil {
			currentBiome = currentTile.Biome
		}

		contexts[i] = AgentContext{
			Agent:            agent,
			VisibleTiles:     visibleTiles,
			VisibleObjects:   visibleObjects,
			VisibleAgents:    visibleAgents,
			OwnedCount:       len(ownedTiles),
			Messages:         e.filterMessagesForAgent(tickMessages, agent.ID),
			CurrentTick:      e.tick,
			WorldSize:        e.config.GetMapSize(),
			CurrentTileOwned: currentTileOwned,
			CurrentTileEnemy: currentTileEnemy,
			EnergyPerTick:    energyPerTick,
			MoveSpeed:        agent.GetEffectiveMoveSpeed(e.balance.Agent.DefaultMoveSpeed),
			ClaimRadius:      agent.GetEffectiveClaimRadius(e.balance.Agent.DefaultClaimRadius),
			CurrentBiome:     currentBiome,
		}
	}

	return contexts
}

// requestActions gets actions from all agents in parallel
func (e *Engine) requestActions(ctx context.Context, contexts []AgentContext) []Action {
	var wg sync.WaitGroup
	actions := make([]Action, len(contexts))

	for i, agentCtx := range contexts {
		wg.Add(1)
		go func(idx int, actx AgentContext) {
			defer wg.Done()

			prompt := e.promptBuilder.BuildPrompt(actx)
			action, err := e.llmClient.GetAction(ctx, actx.Agent.ID, prompt)
			if err != nil {
				log.Printf("LLM error for agent %s: %v", actx.Agent.Name, err)
				action = WaitAction(actx.Agent.ID)
			}
			action.ReceivedAt = time.Now()
			actions[idx] = action
		}(i, agentCtx)
	}

	wg.Wait()
	return actions
}
