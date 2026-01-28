package game

import (
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// ConflictResolver handles simultaneous action conflicts
type ConflictResolver struct {
	mu      sync.Mutex
	actions []Action
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		actions: make([]Action, 0),
	}
}

// AddAction adds an action to be resolved
func (cr *ConflictResolver) AddAction(action Action) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.actions = append(cr.actions, action)
}

// AddActions adds multiple actions at once
func (cr *ConflictResolver) AddActions(actions []Action) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.actions = append(cr.actions, actions...)
}

// Resolve processes all actions with first-come priority
// Returns actions sorted by arrival time
func (cr *ConflictResolver) Resolve() []Action {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Sort by arrival time (first-come priority)
	sort.Slice(cr.actions, func(i, j int) bool {
		return cr.actions[i].ReceivedAt.Before(cr.actions[j].ReceivedAt)
	})

	result := cr.actions
	cr.actions = make([]Action, 0)
	return result
}

// Clear removes all pending actions
func (cr *ConflictResolver) Clear() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.actions = make([]Action, 0)
}

// ActionProcessor applies actions to the game state
type ActionProcessor struct {
	world          *World
	agents         map[uuid.UUID]*Agent
	worldObjects   *WorldObjectManager
	itemRegistry   *ItemRegistry
	recipeRegistry *RecipeRegistry
	currentTick    int
}

// NewActionProcessor creates a new action processor
func NewActionProcessor(world *World, agents map[uuid.UUID]*Agent) *ActionProcessor {
	return &ActionProcessor{
		world:  world,
		agents: agents,
	}
}

// NewActionProcessorFull creates an action processor with all dependencies
func NewActionProcessorFull(world *World, agents map[uuid.UUID]*Agent, worldObjects *WorldObjectManager, itemRegistry *ItemRegistry, recipeRegistry *RecipeRegistry, currentTick int) *ActionProcessor {
	return &ActionProcessor{
		world:          world,
		agents:         agents,
		worldObjects:   worldObjects,
		itemRegistry:   itemRegistry,
		recipeRegistry: recipeRegistry,
		currentTick:    currentTick,
	}
}

// Process applies an action and returns the result
func (ap *ActionProcessor) Process(action Action) ActionResult {
	agent, ok := ap.agents[action.AgentID]
	if !ok {
		return ActionResult{
			AgentID: action.AgentID,
			Action:  action.Type,
			Success: false,
			Message: "agent not found",
		}
	}

	// Dead agents can only wait
	if agent.IsDead {
		return ActionResult{
			AgentID: agent.ID,
			Action:  action.Type,
			Success: false,
			Message: "agent is dead",
		}
	}

	switch action.Type {
	case ActionMove:
		return ap.processMove(agent, action.Params.Direction)
	case ActionClaim:
		return ap.processClaim(agent)
	case ActionMessage:
		return ap.processMessage(agent, action.Params.Target, action.Params.Message)
	case ActionWait, ActionHold:
		return ActionResult{
			AgentID: action.AgentID,
			Action:  action.Type,
			Success: true,
			Message: "holding position",
		}
	case ActionFight:
		return ap.processFight(agent, action.Params.Target)
	case ActionPickup:
		return ap.processPickup(agent)
	case ActionDrop:
		return ap.processDrop(agent, action.Params.ItemID, action.Params.Quantity)
	case ActionUse:
		return ap.processUse(agent, action.Params.ItemID)
	case ActionPlace:
		return ap.processPlace(agent, action.Params.ItemID)
	case ActionCraft:
		return ap.processCraft(agent, action.Params.RecipeID)
	case ActionHarvest:
		return ap.processHarvest(agent)
	case ActionScan:
		return ap.processScan(agent)
	case ActionInteract:
		return ap.processInteract(agent, action.Params.Message)
	case ActionUpgrade:
		return ap.processUpgrade(agent, action.Params.UpgradeType)
	default:
		return ActionResult{
			AgentID: action.AgentID,
			Action:  action.Type,
			Success: false,
			Message: "unknown action type",
		}
	}
}

// processMove handles a move action
func (ap *ActionProcessor) processMove(agent *Agent, dir Direction) ActionResult {
	oldPos := agent.GetPosition()
	newPos := GetNewPosition(oldPos, dir)

	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionMove,
		OldPos:  &oldPos,
	}

	// Check if new position is valid
	if !ap.world.IsValidPosition(newPos) {
		result.Success = false
		result.Message = "cannot move out of bounds"
		return result
	}

	// Check if tile is passable (not water or mountain)
	tile := ap.world.GetTile(newPos)
	if tile.Terrain == TerrainWater || tile.Terrain == TerrainMountain {
		result.Success = false
		result.Message = "cannot move to impassable terrain"
		return result
	}

	// Check for blocking structures
	if ap.worldObjects != nil && ap.worldObjects.HasBlockingObject(newPos) {
		result.Success = false
		result.Message = "path blocked by structure"
		return result
	}

	// Check if another agent is already there (skip dead agents)
	for _, other := range ap.agents {
		if other.ID != agent.ID && !other.IsDead && other.GetPosition() == newPos {
			result.Success = false
			result.Message = "tile occupied by another agent"
			return result
		}
	}

	// Move successful
	agent.SetPosition(newPos)
	result.Success = true
	result.NewPos = &newPos
	result.Message = "moved successfully"

	// Add to memory
	agent.AddMemory("Moved " + string(dir) + " to (" + posString(newPos) + ")")

	return result
}

// posString formats a position for memory/display
func posString(pos Position) string {
	return fmt.Sprintf("%d,%d", pos.X, pos.Y)
}

// processClaim handles a claim action
func (ap *ActionProcessor) processClaim(agent *Agent) ActionResult {
	pos := agent.GetPosition()

	result := ActionResult{
		AgentID:   agent.ID,
		Action:    ActionClaim,
		ClaimedAt: &pos,
	}

	tile := ap.world.GetTile(pos)
	if tile == nil {
		result.Success = false
		result.Message = "invalid position"
		return result
	}

	// Check if already owned by this agent
	if tile.OwnerID != nil && *tile.OwnerID == agent.ID {
		result.Success = false
		result.Message = "already own this tile"
		return result
	}

	// Claim the tile
	ap.world.SetOwner(pos, &agent.ID)
	result.Success = true

	if tile.OwnerID != nil {
		result.Message = "captured tile from enemy"
		agent.AddMemory("Captured enemy tile at (" + posString(pos) + ")")
	} else {
		result.Message = "claimed new tile"
		agent.AddMemory("Claimed tile at (" + posString(pos) + ")")
	}

	return result
}

// processMessage handles a message action
func (ap *ActionProcessor) processMessage(agent *Agent, target *uuid.UUID, message string) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionMessage,
		Success: true,
	}

	if target == nil {
		result.Message = "broadcast message sent"
		agent.AddMemory("Broadcast: " + message)
	} else {
		if _, ok := ap.agents[*target]; ok {
			result.Message = "message sent"
			agent.AddMemory("Sent message to agent: " + message)
		} else {
			result.Success = false
			result.Message = "target agent not found"
		}
	}

	return result
}

// processFight handles a fight action
func (ap *ActionProcessor) processFight(agent *Agent, targetID *uuid.UUID) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionFight,
	}

	if targetID == nil {
		result.Success = false
		result.Message = "no target specified"
		return result
	}

	target, ok := ap.agents[*targetID]
	if !ok || target.IsDead {
		result.Success = false
		result.Message = "target not found or already dead"
		return result
	}

	// Check adjacency (within 1 tile)
	agentPos := agent.GetPosition()
	targetPos := target.GetPosition()
	dx := agentPos.X - targetPos.X
	dy := agentPos.Y - targetPos.Y
	if dx < -1 || dx > 1 || dy < -1 || dy > 1 || (dx == 0 && dy == 0) {
		result.Success = false
		result.Message = "target not adjacent"
		return result
	}

	// Calculate damage
	damage := agent.GetEffectiveStrength()
	result.TargetID = targetID
	result.DamageDealt = damage

	// Apply damage
	killed := target.TakeDamage(damage)
	result.Success = true

	if killed {
		target.Kill(ap.currentTick)
		// Clear target's tiles on death
		ownedTiles := ap.world.GetOwnedTiles(target.ID)
		for _, pos := range ownedTiles {
			ap.world.SetOwner(pos, nil)
		}
		// Clear inventory
		if target.Inventory != nil {
			target.Inventory.Clear()
		}
		result.Message = fmt.Sprintf("killed %s", target.Name)
		agent.AddMemory(fmt.Sprintf("Killed %s in combat", target.Name))
	} else {
		result.Message = fmt.Sprintf("dealt %d damage to %s (HP: %d/%d)", damage, target.Name, target.GetHP(), target.MaxHP)
		agent.AddMemory(fmt.Sprintf("Attacked %s for %d damage", target.Name, damage))
	}

	return result
}

// processPickup handles picking up dropped items
func (ap *ActionProcessor) processPickup(agent *Agent) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionPickup,
	}

	if ap.worldObjects == nil || agent.Inventory == nil {
		result.Success = false
		result.Message = "system not initialized"
		return result
	}

	pos := agent.GetPosition()
	droppedItems := ap.worldObjects.GetDroppedItemsAt(pos)

	if len(droppedItems) == 0 {
		result.Success = false
		result.Message = "no items to pick up"
		return result
	}

	// Pick up the first item
	obj := droppedItems[0]
	if obj.Item == nil {
		result.Success = false
		result.Message = "invalid item"
		return result
	}

	// Try to add to inventory
	remaining := agent.Inventory.AddItem(obj.Item.DefinitionID, obj.Item.Quantity)
	if remaining == obj.Item.Quantity {
		result.Success = false
		result.Message = "inventory full"
		return result
	}

	pickedUp := obj.Item.Quantity - remaining
	result.Success = true
	result.ItemID = obj.Item.DefinitionID
	result.ItemQuantity = pickedUp

	if remaining > 0 {
		obj.Item.Quantity = remaining
		result.Message = fmt.Sprintf("picked up %d %s (partial)", pickedUp, obj.Item.DefinitionID)
	} else {
		ap.worldObjects.Remove(obj.ID)
		result.Message = fmt.Sprintf("picked up %d %s", pickedUp, obj.Item.DefinitionID)
	}

	agent.AddMemory(fmt.Sprintf("Picked up %d %s", pickedUp, obj.Item.DefinitionID))
	return result
}

// processDrop handles dropping items
func (ap *ActionProcessor) processDrop(agent *Agent, itemID string, quantity int) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionDrop,
	}

	if agent.Inventory == nil || ap.worldObjects == nil {
		result.Success = false
		result.Message = "system not initialized"
		return result
	}

	if quantity <= 0 {
		quantity = 1
	}

	available := agent.Inventory.GetItemCount(itemID)
	if available == 0 {
		result.Success = false
		result.Message = "item not in inventory"
		return result
	}

	toDrop := min(quantity, available)
	agent.Inventory.RemoveItem(itemID, toDrop)

	// Create dropped item in world
	pos := agent.GetPosition()
	droppedItem := NewItemInstance(itemID, toDrop)
	obj := NewDroppedItem(droppedItem, pos, ap.currentTick, 50) // Despawn after 50 ticks
	ap.worldObjects.Add(obj)

	result.Success = true
	result.ItemID = itemID
	result.ItemQuantity = toDrop
	result.Message = fmt.Sprintf("dropped %d %s", toDrop, itemID)
	agent.AddMemory(fmt.Sprintf("Dropped %d %s", toDrop, itemID))

	return result
}

// processUse handles using consumable items
func (ap *ActionProcessor) processUse(agent *Agent, itemID string) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionUse,
	}

	if agent.Inventory == nil || ap.itemRegistry == nil {
		result.Success = false
		result.Message = "system not initialized"
		return result
	}

	if !agent.Inventory.HasItems(itemID, 1) {
		result.Success = false
		result.Message = "item not in inventory"
		return result
	}

	def := ap.itemRegistry.Get(itemID)
	if def == nil {
		result.Success = false
		result.Message = "unknown item"
		return result
	}

	if !def.Usable {
		result.Success = false
		result.Message = "item cannot be used"
		return result
	}

	// Apply item effects
	effectApplied := false

	if healAmount := def.GetPropertyInt("heal_amount", 0); healAmount > 0 {
		agent.Heal(healAmount)
		effectApplied = true
		result.Message = fmt.Sprintf("healed %d HP", healAmount)
	}

	if energyAmount := def.GetPropertyInt("energy_amount", 0); energyAmount > 0 {
		agent.AddEnergy(energyAmount)
		effectApplied = true
		result.Message = fmt.Sprintf("restored %d energy", energyAmount)
	}

	if !effectApplied {
		result.Success = false
		result.Message = "item has no effect"
		return result
	}

	// Consume the item
	if def.Consumable {
		agent.Inventory.RemoveItem(itemID, 1)
	}

	result.Success = true
	result.ItemID = itemID
	agent.AddMemory(fmt.Sprintf("Used %s", def.Name))

	return result
}

// processPlace handles placing structures
func (ap *ActionProcessor) processPlace(agent *Agent, itemID string) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionPlace,
	}

	if agent.Inventory == nil || ap.itemRegistry == nil || ap.worldObjects == nil {
		result.Success = false
		result.Message = "system not initialized"
		return result
	}

	if !agent.Inventory.HasItems(itemID, 1) {
		result.Success = false
		result.Message = "item not in inventory"
		return result
	}

	def := ap.itemRegistry.Get(itemID)
	if def == nil || !def.Placeable {
		result.Success = false
		result.Message = "item cannot be placed"
		return result
	}

	// Check energy cost
	if def.EnergyCost > 0 && agent.GetEnergy() < def.EnergyCost {
		result.Success = false
		result.Message = fmt.Sprintf("need %d energy to place", def.EnergyCost)
		return result
	}

	pos := agent.GetPosition()

	// Check if position already has a structure
	if existing := ap.worldObjects.GetStructureAt(pos); existing != nil {
		result.Success = false
		result.Message = "position already has a structure"
		return result
	}

	// Determine structure type from item
	var structureType StructureType
	switch itemID {
	case "wall":
		structureType = StructureWall
	case "beacon":
		structureType = StructureBeacon
	case "trap":
		structureType = StructureTrap
	default:
		result.Success = false
		result.Message = "unknown structure type"
		return result
	}

	// Consume item and energy
	agent.Inventory.RemoveItem(itemID, 1)
	if def.EnergyCost > 0 {
		agent.SpendEnergy(def.EnergyCost)
	}

	// Create structure
	structure := NewStructure(structureType, pos, agent.ID, ap.currentTick)
	ap.worldObjects.Add(structure)

	result.Success = true
	result.Placed = string(structureType)
	result.Message = fmt.Sprintf("placed %s", def.Name)
	agent.AddMemory(fmt.Sprintf("Placed %s at (%s)", def.Name, posString(pos)))

	return result
}

// processCraft handles crafting items
func (ap *ActionProcessor) processCraft(agent *Agent, recipeID string) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionCraft,
	}

	if agent.Inventory == nil || ap.recipeRegistry == nil {
		result.Success = false
		result.Message = "system not initialized"
		return result
	}

	recipe := ap.recipeRegistry.Get(recipeID)
	if recipe == nil {
		result.Success = false
		result.Message = "unknown recipe"
		return result
	}

	// Check if can craft
	canCraft, reason := ap.recipeRegistry.CanCraft(recipeID, agent.Inventory, agent.GetEnergy())
	if !canCraft {
		result.Success = false
		result.Message = reason
		return result
	}

	// Consume ingredients
	for itemID, qty := range recipe.Ingredients {
		agent.Inventory.RemoveItem(itemID, qty)
	}

	// Consume energy
	if recipe.EnergyCost > 0 {
		agent.SpendEnergy(recipe.EnergyCost)
	}

	// Add result item
	remaining := agent.Inventory.AddItem(recipe.Result.ItemID, recipe.Result.Quantity)
	if remaining > 0 {
		// Drop overflow
		pos := agent.GetPosition()
		droppedItem := NewItemInstance(recipe.Result.ItemID, remaining)
		obj := NewDroppedItem(droppedItem, pos, ap.currentTick, 50)
		ap.worldObjects.Add(obj)
	}

	result.Success = true
	result.Crafted = recipe.Result.ItemID
	result.ItemQuantity = recipe.Result.Quantity
	result.Message = fmt.Sprintf("crafted %d %s", recipe.Result.Quantity, recipe.Result.ItemID)
	agent.AddMemory(fmt.Sprintf("Crafted %s", recipe.Name))

	return result
}

// processHarvest handles harvesting resources
func (ap *ActionProcessor) processHarvest(agent *Agent) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionHarvest,
	}

	if agent.Inventory == nil || ap.worldObjects == nil {
		result.Success = false
		result.Message = "system not initialized"
		return result
	}

	pos := agent.GetPosition()
	resource := ap.worldObjects.GetResourceAt(pos)

	if resource == nil || resource.Remaining <= 0 {
		result.Success = false
		result.Message = "no resource to harvest"
		return result
	}

	// Harvest 1 unit
	harvested := resource.Harvest(1)
	if harvested == 0 {
		result.Success = false
		result.Message = "resource depleted"
		return result
	}

	// Map resource type to item ID
	itemID := string(resource.ResourceType)

	// Add to inventory
	remaining := agent.Inventory.AddItem(itemID, harvested)
	if remaining == harvested {
		// Inventory full, put resource back
		resource.Remaining += harvested
		result.Success = false
		result.Message = "inventory full"
		return result
	}

	actualHarvested := harvested - remaining
	result.Success = true
	result.Harvested = itemID
	result.ItemQuantity = actualHarvested
	result.Message = fmt.Sprintf("harvested %d %s", actualHarvested, itemID)
	agent.AddMemory(fmt.Sprintf("Harvested %d %s", actualHarvested, itemID))

	return result
}

// processScan handles extended vision scan
func (ap *ActionProcessor) processScan(agent *Agent) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionScan,
	}

	const scanCost = 2

	if agent.GetEnergy() < scanCost {
		result.Success = false
		result.Message = "need 2 energy to scan"
		return result
	}

	agent.SpendEnergy(scanCost)
	result.Success = true
	result.Message = "scan complete - extended vision this tick"
	agent.AddMemory("Performed area scan")

	return result
}

// processInteract handles interacting with world objects
func (ap *ActionProcessor) processInteract(agent *Agent, message string) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionInteract,
	}

	if ap.worldObjects == nil {
		result.Success = false
		result.Message = "system not initialized"
		return result
	}

	pos := agent.GetPosition()
	interactive := ap.worldObjects.GetInteractiveAt(pos)

	if interactive == nil {
		result.Success = false
		result.Message = "nothing to interact with"
		return result
	}

	switch interactive.InteractiveType {
	case InteractiveShrine:
		if !interactive.CanBeActivatedBy(agent.ID) {
			result.Success = false
			result.Message = "shrine already activated"
			return result
		}
		interactive.Activate(agent.ID)
		agent.MaxHP += interactive.HPReward
		agent.Heal(interactive.HPReward)
		result.Success = true
		result.Message = fmt.Sprintf("activated shrine - max HP increased to %d", agent.MaxHP)
		agent.AddMemory("Activated a shrine, gained +1 max HP")

	case InteractiveCache:
		if !interactive.CanBeActivatedBy(agent.ID) {
			result.Success = false
			result.Message = "cache already claimed"
			return result
		}
		interactive.Activate(agent.ID)
		agent.AddEnergy(interactive.EnergyReward)
		// Remove the cache after use
		ap.worldObjects.Remove(interactive.ID)
		result.Success = true
		result.Message = fmt.Sprintf("found cache - gained %d energy", interactive.EnergyReward)
		agent.AddMemory(fmt.Sprintf("Found a cache with %d energy", interactive.EnergyReward))

	case InteractivePortal:
		if interactive.Destination == nil {
			result.Success = false
			result.Message = "portal has no destination"
			return result
		}
		agent.SetPosition(*interactive.Destination)
		result.Success = true
		result.NewPos = interactive.Destination
		result.Message = fmt.Sprintf("teleported to (%s)", posString(*interactive.Destination))
		agent.AddMemory(fmt.Sprintf("Used portal to (%s)", posString(*interactive.Destination)))

	case InteractiveObelisk:
		if message != "" {
			// Write message
			interactive.Message = message
			result.Success = true
			result.Message = "inscribed message on obelisk"
			agent.AddMemory("Left a message on an obelisk")
		} else if interactive.Message != "" {
			// Read message
			result.Success = true
			result.Message = fmt.Sprintf("obelisk reads: %s", interactive.Message)
			agent.AddMemory(fmt.Sprintf("Read obelisk: %s", interactive.Message))
		} else {
			result.Success = true
			result.Message = "obelisk is blank"
		}

	default:
		result.Success = false
		result.Message = "unknown interactive type"
	}

	return result
}

// processUpgrade handles upgrading agent abilities
func (ap *ActionProcessor) processUpgrade(agent *Agent, upgradeType string) ActionResult {
	result := ActionResult{
		AgentID: agent.ID,
		Action:  ActionUpgrade,
	}

	canUpgrade, cost, reason := agent.CanUpgrade(upgradeType)
	if !canUpgrade {
		result.Success = false
		result.Message = reason
		return result
	}

	if !agent.ApplyUpgrade(upgradeType) {
		result.Success = false
		result.Message = "upgrade failed"
		return result
	}

	var newLevel int
	switch upgradeType {
	case "vision":
		newLevel = agent.VisionLevel
	case "memory":
		newLevel = agent.MemoryLevel
	case "strength":
		newLevel = agent.StrengthLevel
	case "storage":
		newLevel = agent.StorageLevel
	}

	result.Success = true
	result.Upgraded = upgradeType
	result.NewLevel = newLevel
	result.Message = fmt.Sprintf("upgraded %s to level %d (cost: %d energy)", upgradeType, newLevel, cost)
	agent.AddMemory(fmt.Sprintf("Upgraded %s to level %d", upgradeType, newLevel))

	return result
}

// ProcessAll applies all actions and returns results
func (ap *ActionProcessor) ProcessAll(actions []Action) []ActionResult {
	results := make([]ActionResult, len(actions))
	for i, action := range actions {
		results[i] = ap.Process(action)
	}
	return results
}
