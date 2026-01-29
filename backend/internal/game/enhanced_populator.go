package game

import (
	"math/rand"

	"github.com/lucas/promptlands/internal/game/worldgen"
)

// EnhancedWorldPopulator handles placing objects using the new biome and loot table system
type EnhancedWorldPopulator struct {
	rng           *rand.Rand
	world         *World
	worldObjects  *WorldObjectManager
	config        *worldgen.MapConfig
	lootTables    *worldgen.LootTableRegistry
	biomeRegistry *worldgen.BiomeRegistry
	biomeLoot     map[worldgen.BiomeType]worldgen.BiomeLootTable
	tiles         [][]worldgen.EnhancedTileData
}

// NewEnhancedWorldPopulator creates a new enhanced world populator
func NewEnhancedWorldPopulator(
	seed int64,
	world *World,
	worldObjects *WorldObjectManager,
	config *worldgen.MapConfig,
	tiles [][]worldgen.EnhancedTileData,
) *EnhancedWorldPopulator {
	return &EnhancedWorldPopulator{
		rng:           rand.New(rand.NewSource(seed + 2000)), // Offset seed
		world:         world,
		worldObjects:  worldObjects,
		config:        config,
		lootTables:    worldgen.NewLootTableRegistry(seed + 3000),
		biomeRegistry: worldgen.DefaultBiomeRegistry(),
		biomeLoot:     worldgen.GetBiomeLootTables(),
		tiles:         tiles,
	}
}

// PopulateWorld adds all world objects using biome-aware loot tables
func (wp *EnhancedWorldPopulator) PopulateWorld() {
	wp.spawnBiomeResources()
	wp.PopulateInteractives()
}

// PopulateInteractives adds all interactive objects (no resources)
func (wp *EnhancedWorldPopulator) PopulateInteractives() {
	wp.spawnShrines()
	wp.spawnCaches()
	wp.spawnPortals()
	wp.spawnObelisks()
	wp.spawnDungeons()
	wp.spawnVillages()
	wp.spawnRuins()
}

// spawnBiomeResources places resources using biome-specific loot tables
func (wp *EnhancedWorldPopulator) spawnBiomeResources() {
	size := wp.config.GetActualSize()

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			tile := wp.tiles[y][x]

			// Get biome properties
			biomeProps, ok := wp.biomeRegistry.GetBiome(tile.Biome)
			if !ok || biomeProps.ResourceChance <= 0 {
				continue
			}

			// Roll for resource spawn
			if wp.rng.Float64() > biomeProps.ResourceChance*wp.config.ResourceDensity {
				continue
			}

			// Get biome loot table
			biomeLootConfig, hasLoot := wp.biomeLoot[tile.Biome]
			if !hasLoot {
				continue
			}

			// Roll loot table
			lootResults := wp.lootTables.Roll(biomeLootConfig.ResourceTable, 1.0)
			if len(lootResults) == 0 {
				continue
			}

			// Use first result to determine resource type
			result := lootResults[0]
			resourceType := wp.itemToResourceType(result.ItemID)
			if resourceType == "" {
				continue
			}

			pos := Position{X: x, Y: y}
			node := NewResourceNode(resourceType, pos, result.Quantity)
			wp.worldObjects.Add(node)
		}
	}
}

// itemToResourceType converts an item ID to a ResourceType
func (wp *EnhancedWorldPopulator) itemToResourceType(itemID string) ResourceType {
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
		// For special biome resources, map to crystal for now
		// In the future, add new resource types
		return ResourceCrystal
	}
}

// spawnShrines places shrines based on config
func (wp *EnhancedWorldPopulator) spawnShrines() {
	chunkCount := wp.config.GetChunkCount()
	shrinesPerChunk := wp.config.Structures.ShrinesPerChunk

	totalShrines := int(float64(chunkCount*chunkCount) * shrinesPerChunk)
	wp.spawnInteractiveObjects(totalShrines, InteractiveShrine)
}

// spawnCaches places energy caches based on config
func (wp *EnhancedWorldPopulator) spawnCaches() {
	chunkCount := wp.config.GetChunkCount()
	cachesPerChunk := wp.config.Structures.CachesPerChunk

	totalCaches := int(float64(chunkCount*chunkCount) * cachesPerChunk)
	wp.spawnCacheObjects(totalCaches)
}

// spawnCacheObjects places caches with biome-appropriate energy rewards
func (wp *EnhancedWorldPopulator) spawnCacheObjects(count int) {
	size := wp.config.GetActualSize()
	placed := 0
	attempts := 0
	maxAttempts := count * 50

	for placed < count && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		// Get biome for energy scaling
		tile := wp.tiles[pos.Y][pos.X]
		biomeProps, ok := wp.biomeRegistry.GetBiome(tile.Biome)

		// Base energy + biome bonus
		baseEnergy := 10 + wp.rng.Intn(11) // 10-20
		if ok {
			baseEnergy += biomeProps.EnergyPerTick * 2 // Bonus based on biome value
		}

		cache := NewCache(pos, baseEnergy)
		wp.worldObjects.Add(cache)
		placed++
	}
}

// spawnInteractiveObjects places generic interactive objects
func (wp *EnhancedWorldPopulator) spawnInteractiveObjects(count int, interactiveType InteractiveType) {
	size := wp.config.GetActualSize()
	placed := 0
	attempts := 0
	maxAttempts := count * 50

	for placed < count && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		var obj *WorldObject
		switch interactiveType {
		case InteractiveShrine:
			obj = NewShrine(pos)
		case InteractiveObelisk:
			obj = NewObelisk(pos)
		default:
			continue
		}

		wp.worldObjects.Add(obj)
		placed++
	}
}

// spawnPortals places linked portal pairs based on config
func (wp *EnhancedWorldPopulator) spawnPortals() {
	portalPairs := wp.config.Structures.PortalPairsPerMap
	size := wp.config.GetActualSize()
	minDistance := size / 4 // Portals should be reasonably far apart

	for i := 0; i < portalPairs; i++ {
		var pos1, pos2 Position
		found1, found2 := false, false

		// Find first portal position
		for attempts := 0; attempts < 100 && !found1; attempts++ {
			pos1 = Position{
				X: wp.rng.Intn(size),
				Y: wp.rng.Intn(size),
			}
			if wp.isValidPlacement(pos1) {
				found1 = true
			}
		}

		// Find second portal position (with distance requirement)
		for attempts := 0; attempts < 100 && !found2; attempts++ {
			pos2 = Position{
				X: wp.rng.Intn(size),
				Y: wp.rng.Intn(size),
			}
			dx := pos1.X - pos2.X
			dy := pos1.Y - pos2.Y
			dist := dx*dx + dy*dy
			if wp.isValidPlacement(pos2) && dist > minDistance*minDistance {
				found2 = true
			}
		}

		if found1 && found2 {
			portal1 := NewPortal(pos1, pos2)
			portal2 := NewPortal(pos2, pos1)
			wp.worldObjects.Add(portal1)
			wp.worldObjects.Add(portal2)
		}
	}
}

// spawnObelisks places message obelisks based on config
func (wp *EnhancedWorldPopulator) spawnObelisks() {
	chunkCount := wp.config.GetChunkCount()
	obelisksPerChunk := wp.config.Structures.ObelisksPerChunk

	totalObelisks := int(float64(chunkCount*chunkCount) * obelisksPerChunk)
	wp.spawnInteractiveObjects(totalObelisks, InteractiveObelisk)
}

// spawnDungeons places dungeon entrances
func (wp *EnhancedWorldPopulator) spawnDungeons() {
	dungeonCount := wp.config.Structures.DungeonsPerMap
	size := wp.config.GetActualSize()

	placed := 0
	attempts := 0
	maxAttempts := dungeonCount * 100

	for placed < dungeonCount && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		// Place dungeon as a special obelisk for now
		// Future: Add DungeonEntrance type
		dungeon := NewObelisk(pos)
		dungeon.Message = "DUNGEON_ENTRANCE"
		wp.worldObjects.Add(dungeon)
		placed++
	}
}

// spawnVillages places village centers
func (wp *EnhancedWorldPopulator) spawnVillages() {
	villageCount := wp.config.Structures.VillagesPerMap
	size := wp.config.GetActualSize()
	minDistance := size / 3 // Villages should be spread out

	var villagePositions []Position
	attempts := 0
	maxAttempts := villageCount * 100

	for len(villagePositions) < villageCount && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		// Check distance from other villages
		tooClose := false
		for _, vp := range villagePositions {
			dx := pos.X - vp.X
			dy := pos.Y - vp.Y
			if dx*dx+dy*dy < minDistance*minDistance {
				tooClose = true
				break
			}
		}

		if tooClose {
			continue
		}

		// Place village center as beacon
		village := NewObelisk(pos)
		village.Message = "VILLAGE_CENTER"
		wp.worldObjects.Add(village)
		villagePositions = append(villagePositions, pos)

		// Spawn some resources around village
		wp.spawnVillageResources(pos)
	}
}

// spawnVillageResources places resources around a village center
func (wp *EnhancedWorldPopulator) spawnVillageResources(center Position) {
	radius := 5

	for i := 0; i < 3; i++ {
		dx := wp.rng.Intn(radius*2+1) - radius
		dy := wp.rng.Intn(radius*2+1) - radius

		pos := Position{
			X: center.X + dx,
			Y: center.Y + dy,
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		// Random resource type
		resourceTypes := []ResourceType{ResourceWood, ResourceStone, ResourceHerb}
		resourceType := resourceTypes[wp.rng.Intn(len(resourceTypes))]
		amount := 3 + wp.rng.Intn(4) // 3-6

		node := NewResourceNode(resourceType, pos, amount)
		wp.worldObjects.Add(node)
	}
}

// spawnRuins places ancient ruins
func (wp *EnhancedWorldPopulator) spawnRuins() {
	chunkCount := wp.config.GetChunkCount()
	ruinsPerChunk := wp.config.Structures.RuinsPerChunk

	totalRuins := int(float64(chunkCount*chunkCount) * ruinsPerChunk)
	size := wp.config.GetActualSize()

	placed := 0
	attempts := 0
	maxAttempts := totalRuins * 50

	for placed < totalRuins && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		// Place ruin as cache with higher reward
		energy := 15 + wp.rng.Intn(16) // 15-30
		ruin := NewCache(pos, energy)
		wp.worldObjects.Add(ruin)
		placed++
	}
}

// isValidPlacement checks if a position is suitable for placing an object
func (wp *EnhancedWorldPopulator) isValidPlacement(pos Position) bool {
	size := wp.config.GetActualSize()

	// Check bounds
	if pos.X < 0 || pos.Y < 0 || pos.X >= size || pos.Y >= size {
		return false
	}

	// Get tile data
	tile := wp.tiles[pos.Y][pos.X]

	// Check if biome allows placement
	biomeProps, ok := wp.biomeRegistry.GetBiome(tile.Biome)
	if !ok || !biomeProps.Passable {
		return false
	}

	// Don't place where there's already an interactive object
	existing := wp.worldObjects.GetAt(pos)
	for _, obj := range existing {
		if obj.Type == ObjectInteractive {
			return false
		}
	}

	return true
}

// GetBiomeAt returns the biome type at a position
func (wp *EnhancedWorldPopulator) GetBiomeAt(pos Position) worldgen.BiomeType {
	size := wp.config.GetActualSize()
	if pos.X < 0 || pos.Y < 0 || pos.X >= size || pos.Y >= size {
		return worldgen.BiomeOcean
	}
	return wp.tiles[pos.Y][pos.X].Biome
}

// GetResourceCountByType returns the count of resource nodes by type
func (wp *EnhancedWorldPopulator) GetResourceCountByType() map[ResourceType]int {
	counts := make(map[ResourceType]int)
	for _, obj := range wp.worldObjects.GetAll() {
		if obj.Type == ObjectResource {
			counts[obj.ResourceType]++
		}
	}
	return counts
}

// GetInteractiveCount returns the count of interactive objects by type
func (wp *EnhancedWorldPopulator) GetInteractiveCount() map[InteractiveType]int {
	counts := make(map[InteractiveType]int)
	for _, obj := range wp.worldObjects.GetAll() {
		if obj.Type == ObjectInteractive {
			counts[obj.InteractiveType]++
		}
	}
	return counts
}

// RollLoot rolls loot from a table and returns results
func (wp *EnhancedWorldPopulator) RollLoot(tableID string, luck float64) []worldgen.LootResult {
	return wp.lootTables.Roll(tableID, luck)
}

// RollBiomeLoot rolls resource loot for a specific biome
func (wp *EnhancedWorldPopulator) RollBiomeLoot(biome worldgen.BiomeType, luck float64) []worldgen.LootResult {
	biomeLoot, ok := wp.biomeLoot[biome]
	if !ok {
		return nil
	}
	return wp.lootTables.Roll(biomeLoot.ResourceTable, luck)
}

// RollBiomeChestLoot rolls chest loot for a specific biome
func (wp *EnhancedWorldPopulator) RollBiomeChestLoot(biome worldgen.BiomeType, luck float64) []worldgen.LootResult {
	biomeLoot, ok := wp.biomeLoot[biome]
	if !ok {
		return nil
	}
	return wp.lootTables.Roll(biomeLoot.ChestTable, luck)
}
