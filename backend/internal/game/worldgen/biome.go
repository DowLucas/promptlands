package worldgen

// TerrainType represents the type of terrain on a tile (matches game.TerrainType)
type TerrainType string

const (
	TerrainPlains   TerrainType = "plains"
	TerrainForest   TerrainType = "forest"
	TerrainMountain TerrainType = "mountain"
	TerrainWater    TerrainType = "water"
)

// BiomeThresholds defines the thresholds for biome classification
type BiomeThresholds struct {
	WaterMax     float64 // Below this elevation = water
	MountainMin  float64 // Above this elevation = mountain
	ForestMinMoi float64 // Above this moisture (with mid elevation) = forest
}

// DefaultThresholds returns the default biome thresholds
func DefaultThresholds() BiomeThresholds {
	return BiomeThresholds{
		WaterMax:     0.35,
		MountainMin:  0.65,
		ForestMinMoi: 0.55,
	}
}

// DetermineBiome classifies terrain based on elevation and moisture values
func DetermineBiome(elevation, moisture float64, thresholds BiomeThresholds) TerrainType {
	// Low elevation = water
	if elevation < thresholds.WaterMax {
		return TerrainWater
	}

	// High elevation = mountain
	if elevation > thresholds.MountainMin {
		return TerrainMountain
	}

	// Mid elevation with high moisture = forest
	if moisture > thresholds.ForestMinMoi {
		return TerrainForest
	}

	// Default = plains
	return TerrainPlains
}

// IsPassable returns true if the terrain type allows movement/spawning
func IsPassable(terrain TerrainType) bool {
	switch terrain {
	case TerrainPlains, TerrainForest:
		return true
	case TerrainWater, TerrainMountain:
		return false
	default:
		return true
	}
}

// IsPassableString checks passability using string terrain type (for use with game.TerrainType)
func IsPassableString(terrain string) bool {
	return IsPassable(TerrainType(terrain))
}
