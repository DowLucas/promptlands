package worldgen

// GeneratorConfig holds the configuration for world generation
type GeneratorConfig struct {
	// Elevation noise parameters
	ElevationOctaves     int
	ElevationFrequency   float64
	ElevationPersistence float64

	// Moisture noise parameters
	MoistureOctaves     int
	MoistureFrequency   float64
	MoisturePersistence float64

	// Biome thresholds
	Thresholds BiomeThresholds
}

// DefaultConfig returns the default generator configuration
func DefaultConfig() GeneratorConfig {
	return GeneratorConfig{
		ElevationOctaves:     4,
		ElevationFrequency:   0.02,
		ElevationPersistence: 0.5,

		MoistureOctaves:     3,
		MoistureFrequency:   0.03,
		MoisturePersistence: 0.5,

		Thresholds: DefaultThresholds(),
	}
}

// TileData holds the generated terrain data for a tile
type TileData struct {
	X       int
	Y       int
	Terrain TerrainType
}

// WorldGenerator generates procedural terrain using multi-layered noise
type WorldGenerator struct {
	seed           int64
	elevationNoise *NoiseGenerator
	moistureNoise  *NoiseGenerator
	config         GeneratorConfig
}

// NewWorldGenerator creates a new world generator with the given seed
func NewWorldGenerator(seed int64) *WorldGenerator {
	return NewWorldGeneratorWithConfig(seed, DefaultConfig())
}

// NewWorldGeneratorWithConfig creates a new world generator with custom config
func NewWorldGeneratorWithConfig(seed int64, config GeneratorConfig) *WorldGenerator {
	// Use offset seeds for independent noise layers
	return &WorldGenerator{
		seed:           seed,
		elevationNoise: NewNoiseGenerator(seed),
		moistureNoise:  NewNoiseGenerator(seed + 1000),
		config:         config,
	}
}

// Generate creates a 2D tile data array for a world of the given size
func (g *WorldGenerator) Generate(size int) [][]TileData {
	tiles := make([][]TileData, size)

	for y := 0; y < size; y++ {
		tiles[y] = make([]TileData, size)
		for x := 0; x < size; x++ {
			elevation := g.elevationNoise.Octave2D(
				float64(x),
				float64(y),
				g.config.ElevationOctaves,
				g.config.ElevationFrequency,
				g.config.ElevationPersistence,
			)

			moisture := g.moistureNoise.Octave2D(
				float64(x),
				float64(y),
				g.config.MoistureOctaves,
				g.config.MoistureFrequency,
				g.config.MoisturePersistence,
			)

			terrain := DetermineBiome(elevation, moisture, g.config.Thresholds)

			tiles[y][x] = TileData{
				X:       x,
				Y:       y,
				Terrain: terrain,
			}
		}
	}

	return tiles
}

// Seed returns the generator's seed
func (g *WorldGenerator) Seed() int64 {
	return g.seed
}

// GetElevation returns the elevation value at a position (useful for debugging)
func (g *WorldGenerator) GetElevation(x, y int) float64 {
	return g.elevationNoise.Octave2D(
		float64(x),
		float64(y),
		g.config.ElevationOctaves,
		g.config.ElevationFrequency,
		g.config.ElevationPersistence,
	)
}

// GetMoisture returns the moisture value at a position (useful for debugging)
func (g *WorldGenerator) GetMoisture(x, y int) float64 {
	return g.moistureNoise.Octave2D(
		float64(x),
		float64(y),
		g.config.MoistureOctaves,
		g.config.MoistureFrequency,
		g.config.MoisturePersistence,
	)
}
