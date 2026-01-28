package worldgen

import (
	"math/rand"
	"sort"
)

// EnhancedTileData holds the generated terrain data with full biome info
type EnhancedTileData struct {
	X             int
	Y             int
	Biome         BiomeType
	Terrain       TerrainType // For backwards compatibility
	Elevation     float64
	Moisture      float64
	Temperature   float64
	Variation     float64
}

// EnhancedWorldGenerator generates procedural terrain using the new biome system
type EnhancedWorldGenerator struct {
	seed             int64
	config           *MapConfig
	biomeRegistry    *BiomeRegistry
	elevationNoise   *NoiseGenerator
	moistureNoise    *NoiseGenerator
	temperatureNoise *NoiseGenerator
	variationNoise   *NoiseGenerator
	rng              *rand.Rand
}

// NewEnhancedWorldGenerator creates a new enhanced world generator
func NewEnhancedWorldGenerator(seed int64, config *MapConfig) *EnhancedWorldGenerator {
	if config == nil {
		config = DefaultMapConfig()
	}

	// Use provided seed or generate random one
	actualSeed := seed
	if actualSeed == 0 {
		actualSeed = rand.Int63()
	}

	return &EnhancedWorldGenerator{
		seed:             actualSeed,
		config:           config,
		biomeRegistry:    DefaultBiomeRegistry(),
		elevationNoise:   NewNoiseGenerator(actualSeed + config.ElevationNoise.SeedOffset),
		moistureNoise:    NewNoiseGenerator(actualSeed + config.MoistureNoise.SeedOffset),
		temperatureNoise: NewNoiseGenerator(actualSeed + config.TemperatureNoise.SeedOffset),
		variationNoise:   NewNoiseGenerator(actualSeed + config.VariationNoise.SeedOffset),
		rng:              rand.New(rand.NewSource(actualSeed)),
	}
}

// Generate creates the world terrain data
func (g *EnhancedWorldGenerator) Generate() [][]EnhancedTileData {
	size := g.config.GetActualSize()
	tiles := make([][]EnhancedTileData, size)

	for y := 0; y < size; y++ {
		tiles[y] = make([]EnhancedTileData, size)
		for x := 0; x < size; x++ {
			tiles[y][x] = g.generateTile(x, y, size)
		}
	}

	return tiles
}

// generateTile creates data for a single tile
func (g *EnhancedWorldGenerator) generateTile(x, y, size int) EnhancedTileData {
	// Calculate base noise values
	elevation := g.sampleNoise(g.elevationNoise, x, y, g.config.ElevationNoise)
	moisture := g.sampleNoise(g.moistureNoise, x, y, g.config.MoistureNoise)
	temperature := g.sampleNoise(g.temperatureNoise, x, y, g.config.TemperatureNoise)
	variation := g.sampleNoise(g.variationNoise, x, y, g.config.VariationNoise)

	// Apply ocean border if configured
	if g.config.OceanBorder {
		borderWidth := g.config.OceanBorderWidth
		distToEdge := min(x, y, size-1-x, size-1-y)
		if distToEdge < borderWidth {
			// Fade elevation to create ocean
			fade := float64(distToEdge) / float64(borderWidth)
			elevation = elevation * fade * 0.3
		}
	}

	// Add variation noise to create more natural biome boundaries
	elevation += variation * 0.1
	moisture += variation * 0.05
	temperature += variation * 0.05

	// Clamp values to 0-1
	elevation = clamp(elevation, 0, 1)
	moisture = clamp(moisture, 0, 1)
	temperature = clamp(temperature, 0, 1)

	// Determine biome
	biome := g.determineBiome(elevation, moisture, temperature)

	// Get terrain class for backwards compatibility
	terrain := g.biomeRegistry.GetTerrainClass(biome)

	return EnhancedTileData{
		X:           x,
		Y:           y,
		Biome:       biome,
		Terrain:     terrain,
		Elevation:   elevation,
		Moisture:    moisture,
		Temperature: temperature,
		Variation:   variation,
	}
}

// sampleNoise samples noise with the given configuration
func (g *EnhancedWorldGenerator) sampleNoise(noise *NoiseGenerator, x, y int, config NoiseLayerConfig) float64 {
	value := noise.Octave2D(
		float64(x),
		float64(y),
		config.Octaves,
		config.Frequency,
		config.Persistence,
	)
	return value * config.Amplitude
}

// determineBiome selects the appropriate biome based on climate factors
func (g *EnhancedWorldGenerator) determineBiome(elevation, moisture, temperature float64) BiomeType {
	// Sort distributions by priority (highest first)
	distributions := make([]BiomeDistribution, len(g.config.BiomeDistributions))
	copy(distributions, g.config.BiomeDistributions)
	sort.Slice(distributions, func(i, j int) bool {
		return distributions[i].Priority > distributions[j].Priority
	})

	// Find matching biome
	for _, dist := range distributions {
		if elevation >= dist.ElevationMin && elevation <= dist.ElevationMax &&
			moisture >= dist.MoistureMin && moisture <= dist.MoistureMax &&
			temperature >= dist.TemperatureMin && temperature <= dist.TemperatureMax {
			return dist.Biome
		}
	}

	// Fallback to savanna
	return BiomeSavanna
}

// Seed returns the generator's seed
func (g *EnhancedWorldGenerator) Seed() int64 {
	return g.seed
}

// Config returns the generator's configuration
func (g *EnhancedWorldGenerator) Config() *MapConfig {
	return g.config
}

// BiomeRegistry returns the biome registry
func (g *EnhancedWorldGenerator) BiomeRegistry() *BiomeRegistry {
	return g.biomeRegistry
}

// GetBiomeProperties returns properties for a biome type
func (g *EnhancedWorldGenerator) GetBiomeProperties(biome BiomeType) (BiomeProperties, bool) {
	return g.biomeRegistry.GetBiome(biome)
}

// GenerateToLegacyFormat converts enhanced tiles to legacy TileData format
func (g *EnhancedWorldGenerator) GenerateToLegacyFormat() [][]TileData {
	enhanced := g.Generate()
	size := g.config.GetActualSize()

	tiles := make([][]TileData, size)
	for y := 0; y < size; y++ {
		tiles[y] = make([]TileData, size)
		for x := 0; x < size; x++ {
			tiles[y][x] = TileData{
				X:       enhanced[y][x].X,
				Y:       enhanced[y][x].Y,
				Terrain: enhanced[y][x].Terrain,
			}
		}
	}

	return tiles
}

// GetPassablePositions returns all positions that are passable
func (g *EnhancedWorldGenerator) GetPassablePositions(tiles [][]EnhancedTileData) [][2]int {
	var positions [][2]int
	for y := range tiles {
		for x := range tiles[y] {
			if g.biomeRegistry.IsPassable(tiles[y][x].Biome) {
				positions = append(positions, [2]int{x, y})
			}
		}
	}
	return positions
}

// GetSpawnablePositions returns all positions where agents can spawn
func (g *EnhancedWorldGenerator) GetSpawnablePositions(tiles [][]EnhancedTileData) [][2]int {
	var positions [][2]int
	for y := range tiles {
		for x := range tiles[y] {
			props, ok := g.biomeRegistry.GetBiome(tiles[y][x].Biome)
			if ok && props.CanSpawnAgent {
				positions = append(positions, [2]int{x, y})
			}
		}
	}
	return positions
}

// GetTilesInChunk returns all tiles within a specific chunk
func (g *EnhancedWorldGenerator) GetTilesInChunk(tiles [][]EnhancedTileData, chunkX, chunkY int) []EnhancedTileData {
	var result []EnhancedTileData
	chunkSize := g.config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 32
	}

	startX := chunkX * chunkSize
	startY := chunkY * chunkSize
	endX := min(startX+chunkSize, len(tiles[0]))
	endY := min(startY+chunkSize, len(tiles))

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			result = append(result, tiles[y][x])
		}
	}

	return result
}

// BiomeStats holds statistics about biome distribution
type BiomeStats struct {
	TotalTiles    int
	BiomeCounts   map[BiomeType]int
	BiomePercents map[BiomeType]float64
}

// CalculateBiomeStats calculates biome distribution statistics
func (g *EnhancedWorldGenerator) CalculateBiomeStats(tiles [][]EnhancedTileData) BiomeStats {
	stats := BiomeStats{
		BiomeCounts:   make(map[BiomeType]int),
		BiomePercents: make(map[BiomeType]float64),
	}

	for y := range tiles {
		for x := range tiles[y] {
			stats.BiomeCounts[tiles[y][x].Biome]++
			stats.TotalTiles++
		}
	}

	for biome, count := range stats.BiomeCounts {
		stats.BiomePercents[biome] = float64(count) / float64(stats.TotalTiles) * 100
	}

	return stats
}

// clamp constrains a value between min and max
func clamp(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// min returns the minimum of the given integers
func min(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
