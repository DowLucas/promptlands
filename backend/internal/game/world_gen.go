package game

import (
	"math/rand"
)

// WorldPopulator handles placing objects in the generated world
type WorldPopulator struct {
	rng            *rand.Rand
	world          *World
	worldObjects   *WorldObjectManager
}

// NewWorldPopulator creates a new world populator
func NewWorldPopulator(seed int64, world *World, worldObjects *WorldObjectManager) *WorldPopulator {
	return &WorldPopulator{
		rng:          rand.New(rand.NewSource(seed + 1000)), // Offset seed
		world:        world,
		worldObjects: worldObjects,
	}
}

// PopulateWorld adds resources, shrines, caches, portals, and obelisks to the world
func (wp *WorldPopulator) PopulateWorld() {
	wp.spawnResources()
	wp.PopulateInteractives()
}

// PopulateInteractives adds shrines, caches, portals, and obelisks (no resources)
func (wp *WorldPopulator) PopulateInteractives() {
	wp.spawnShrines()
	wp.spawnCaches()
	wp.spawnPortals()
	wp.spawnObelisks()
}

// spawnResources places resource nodes throughout the world
func (wp *WorldPopulator) spawnResources() {
	size := wp.world.Size()
	resourceChance := 0.05 // 5% of tiles

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if wp.rng.Float64() > resourceChance {
				continue
			}

			pos := Position{X: x, Y: y}
			tile := wp.world.GetTile(pos)
			if tile == nil {
				continue
			}

			var resourceType ResourceType
			var amount int

			switch tile.Terrain {
			case TerrainForest:
				// Forests have wood and herbs
				if wp.rng.Float64() < 0.7 {
					resourceType = ResourceWood
					amount = 3 + wp.rng.Intn(3) // 3-5
				} else {
					resourceType = ResourceHerb
					amount = 2 + wp.rng.Intn(2) // 2-3
				}
			case TerrainPlains:
				// Plains have stone and occasionally herbs
				if wp.rng.Float64() < 0.6 {
					resourceType = ResourceStone
					amount = 3 + wp.rng.Intn(3) // 3-5
				} else {
					resourceType = ResourceHerb
					amount = 1 + wp.rng.Intn(2) // 1-2
				}
			case TerrainMountain:
				// Mountains border areas have stone and rare crystals
				if wp.rng.Float64() < 0.8 {
					resourceType = ResourceStone
					amount = 4 + wp.rng.Intn(3) // 4-6
				} else {
					resourceType = ResourceCrystal
					amount = 1 + wp.rng.Intn(2) // 1-2
				}
			default:
				continue // Skip water
			}

			node := NewResourceNode(resourceType, pos, amount)
			wp.worldObjects.Add(node)
		}
	}
}

// spawnShrines places shrines that grant +1 max HP
func (wp *WorldPopulator) spawnShrines() {
	size := wp.world.Size()
	shrineCount := 2 + wp.rng.Intn(2) // 2-3 shrines

	placed := 0
	attempts := 0
	maxAttempts := shrineCount * 20

	for placed < shrineCount && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		shrine := NewShrine(pos)
		wp.worldObjects.Add(shrine)
		placed++
	}
}

// spawnCaches places energy caches
func (wp *WorldPopulator) spawnCaches() {
	size := wp.world.Size()
	cacheCount := 3 + wp.rng.Intn(2) // 3-4 caches

	placed := 0
	attempts := 0
	maxAttempts := cacheCount * 20

	for placed < cacheCount && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		energy := 10 + wp.rng.Intn(11) // 10-20 energy
		cache := NewCache(pos, energy)
		wp.worldObjects.Add(cache)
		placed++
	}
}

// spawnPortals places linked portal pairs
func (wp *WorldPopulator) spawnPortals() {
	size := wp.world.Size()
	portalPairs := 1 + wp.rng.Intn(2) // 1-2 portal pairs

	for i := 0; i < portalPairs; i++ {
		// Find two valid positions
		var pos1, pos2 Position
		found1, found2 := false, false

		for attempts := 0; attempts < 50 && !found1; attempts++ {
			pos1 = Position{
				X: wp.rng.Intn(size),
				Y: wp.rng.Intn(size),
			}
			if wp.isValidPlacement(pos1) {
				found1 = true
			}
		}

		for attempts := 0; attempts < 50 && !found2; attempts++ {
			pos2 = Position{
				X: wp.rng.Intn(size),
				Y: wp.rng.Intn(size),
			}
			// Ensure some distance between portals
			dx := pos1.X - pos2.X
			dy := pos1.Y - pos2.Y
			dist := dx*dx + dy*dy
			if wp.isValidPlacement(pos2) && dist > 25 { // At least 5 tiles apart
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

// spawnObelisks places message obelisks
func (wp *WorldPopulator) spawnObelisks() {
	size := wp.world.Size()
	obeliskCount := 2 + wp.rng.Intn(2) // 2-3 obelisks

	placed := 0
	attempts := 0
	maxAttempts := obeliskCount * 20

	for placed < obeliskCount && attempts < maxAttempts {
		attempts++

		pos := Position{
			X: wp.rng.Intn(size),
			Y: wp.rng.Intn(size),
		}

		if !wp.isValidPlacement(pos) {
			continue
		}

		obelisk := NewObelisk(pos)
		wp.worldObjects.Add(obelisk)
		placed++
	}
}

// isValidPlacement checks if a position is suitable for placing an object
func (wp *WorldPopulator) isValidPlacement(pos Position) bool {
	tile := wp.world.GetTile(pos)
	if tile == nil {
		return false
	}

	// Don't place on water or mountains
	if tile.Terrain == TerrainWater || tile.Terrain == TerrainMountain {
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

// GetResourceCountByType returns the count of resource nodes by type
func (wp *WorldPopulator) GetResourceCountByType() map[ResourceType]int {
	counts := make(map[ResourceType]int)
	for _, obj := range wp.worldObjects.GetAll() {
		if obj.Type == ObjectResource {
			counts[obj.ResourceType]++
		}
	}
	return counts
}

// GetInteractiveCount returns the count of interactive objects by type
func (wp *WorldPopulator) GetInteractiveCount() map[InteractiveType]int {
	counts := make(map[InteractiveType]int)
	for _, obj := range wp.worldObjects.GetAll() {
		if obj.Type == ObjectInteractive {
			counts[obj.InteractiveType]++
		}
	}
	return counts
}
