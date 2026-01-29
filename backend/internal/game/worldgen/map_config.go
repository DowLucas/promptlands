package worldgen

import (
	"encoding/json"
	"os"
)

// MapSize represents predefined map sizes for open-world feel
// Larger maps create a more pronounced pixelated effect when zoomed out
type MapSize string

const (
	MapSizeTiny    MapSize = "tiny"    // 128x128 - Quick games
	MapSizeSmall   MapSize = "small"   // 256x256 - Short games
	MapSizeMedium  MapSize = "medium"  // 512x512 - Standard games
	MapSizeLarge   MapSize = "large"   // 1024x1024 - Long games
	MapSizeHuge    MapSize = "huge"    // 2048x2048 - Epic games (pixelated aesthetic)
	MapSizeMassive MapSize = "massive" // 4096x4096 - Massive open world
)

// GetMapSizeValue returns the actual size value
func GetMapSizeValue(size MapSize) int {
	switch size {
	case MapSizeTiny:
		return 128
	case MapSizeSmall:
		return 256
	case MapSizeMedium:
		return 512
	case MapSizeLarge:
		return 1024
	case MapSizeHuge:
		return 2048
	case MapSizeMassive:
		return 4096
	default:
		return 512
	}
}

// BiomeWeight defines the spawn weight for a biome in a region
type BiomeWeight struct {
	Biome  BiomeType `json:"biome"`
	Weight float64   `json:"weight"` // Higher = more common
}

// BiomeDistribution defines how biomes are distributed based on climate factors
type BiomeDistribution struct {
	// Climate ranges (0.0 to 1.0)
	TemperatureMin float64 `json:"temperature_min"`
	TemperatureMax float64 `json:"temperature_max"`
	MoistureMin    float64 `json:"moisture_min"`
	MoistureMax    float64 `json:"moisture_max"`
	ElevationMin   float64 `json:"elevation_min"`
	ElevationMax   float64 `json:"elevation_max"`

	// Biome to assign
	Biome BiomeType `json:"biome"`

	// Priority (higher = checked first)
	Priority int `json:"priority"`
}

// NoiseLayerConfig configures a noise layer for terrain generation
type NoiseLayerConfig struct {
	Octaves     int     `json:"octaves"`
	Frequency   float64 `json:"frequency"`
	Persistence float64 `json:"persistence"`
	Amplitude   float64 `json:"amplitude"`
	SeedOffset  int64   `json:"seed_offset"`
}

// FrequencyScaling controls automatic noise frequency adjustment based on map size
type FrequencyScaling struct {
	Enabled       bool `json:"enabled"`
	ReferenceSize int  `json:"reference_size"` // Base size frequencies are tuned for (default: 512)
}

// StructureSpawnConfig defines how structures spawn in a map
type StructureSpawnConfig struct {
	ShrinesPerChunk   float64 `json:"shrines_per_chunk"`
	CachesPerChunk    float64 `json:"caches_per_chunk"`
	PortalPairsPerMap int     `json:"portal_pairs_per_map"`
	ObelisksPerChunk  float64 `json:"obelisks_per_chunk"`
	DungeonsPerMap    int     `json:"dungeons_per_map"`
	VillagesPerMap    int     `json:"villages_per_map"`
	RuinsPerChunk     float64 `json:"ruins_per_chunk"`
}

// MapConfig defines the complete configuration for a procedural map
type MapConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Theme       string `json:"theme"`

	Size       MapSize `json:"size"`
	CustomSize int     `json:"custom_size,omitempty"`

	Seed      int64 `json:"seed,omitempty"`
	ChunkSize int   `json:"chunk_size"`

	// Noise configuration - LOW frequency = LARGE biome regions
	ElevationNoise   NoiseLayerConfig `json:"elevation_noise"`
	MoistureNoise    NoiseLayerConfig `json:"moisture_noise"`
	TemperatureNoise NoiseLayerConfig `json:"temperature_noise"`
	VariationNoise   NoiseLayerConfig `json:"variation_noise"`

	BiomeDistributions []BiomeDistribution `json:"biome_distributions"`

	OceanBorder      bool    `json:"ocean_border"`
	OceanBorderWidth int     `json:"ocean_border_width"`
	RiverCount       int     `json:"river_count"`
	LakeChance       float64 `json:"lake_chance"`

	Structures StructureSpawnConfig `json:"structures"`

	ResourceDensity      float64 `json:"resource_density"`
	WinCondition         string  `json:"win_condition,omitempty"`
	WinThreshold         int     `json:"win_threshold,omitempty"`
	MaxTicks             int     `json:"max_ticks,omitempty"`
	RespawnEnabled       bool    `json:"respawn_enabled"`
	FogOfWar             bool    `json:"fog_of_war"`
	DifficultyMultiplier float64 `json:"difficulty_multiplier"`

	FrequencyScaling FrequencyScaling `json:"frequency_scaling,omitempty"`
}

// GetActualSize returns the actual map size in tiles
func (c *MapConfig) GetActualSize() int {
	if c.CustomSize > 0 {
		return c.CustomSize
	}
	return GetMapSizeValue(c.Size)
}

// GetChunkCount returns the number of chunks in each dimension
func (c *MapConfig) GetChunkCount() int {
	size := c.GetActualSize()
	chunkSize := c.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 32
	}
	return (size + chunkSize - 1) / chunkSize
}

// GetEffectiveFrequency returns noise frequency adjusted for map size
// This ensures biomes appear consistently sized across different map sizes
func (c *MapConfig) GetEffectiveFrequency(baseFreq float64) float64 {
	if !c.FrequencyScaling.Enabled {
		return baseFreq
	}
	refSize := c.FrequencyScaling.ReferenceSize
	if refSize <= 0 {
		refSize = 512
	}
	actualSize := c.GetActualSize()
	return baseFreq * (float64(refSize) / float64(actualSize))
}

// DefaultMapConfig returns a balanced default map configuration
// Uses LOW frequency noise to create LARGE biome regions like the reference image
// Large map size (2048x2048) creates pixelated aesthetic when zoomed out
func DefaultMapConfig() *MapConfig {
	return &MapConfig{
		ID:          "default",
		Name:        "Standard World",
		Description: "A balanced procedurally generated world with large distinct biomes",
		Theme:       "mixed",
		Size:        MapSizeHuge, // 2048x2048 for pixelated open world feel
		ChunkSize:   32,

		// Moderate frequency = visible distinct regions without fragmentation
		// NOT scaled - these frequencies work well at any map size
		ElevationNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.002, // Creates visible continent-scale regions
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.0025, // Moisture zones
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.0015, // Temperature bands
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     1,
			Frequency:   0.001,
			Persistence: 0.4,
			Amplitude:   0.0, // Disabled - not used
			SeedOffset:  3000,
		},

		BiomeDistributions: DefaultBiomeDistributions(),

		OceanBorder:      true,
		OceanBorderWidth: 48, // Proportionally larger for bigger map
		RiverCount:       8,
		LakeChance:       0.01,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.06,
			CachesPerChunk:    0.10,
			PortalPairsPerMap: 16,
			ObelisksPerChunk:  0.04,
			DungeonsPerMap:    12,
			VillagesPerMap:    8,
			RuinsPerChunk:     0.03,
		},

		ResourceDensity:      1.0,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 1.0,

		FrequencyScaling: FrequencyScaling{
			Enabled:       false, // Disabled - manual frequencies work at any map size
			ReferenceSize: 512,
		},
	}
}

// DefaultBiomeDistributions returns biome distribution rules for large contiguous regions
// Uses non-overlapping "core" ranges to prevent biome fragmentation from noise variations
// Elevation is primary separator, then temperature, then moisture
func DefaultBiomeDistributions() []BiomeDistribution {
	return []BiomeDistribution{
		// === EXTREME ELEVATION (non-overlapping) ===
		// Ocean: lowest elevation - exclusive
		{
			ElevationMin: 0.0, ElevationMax: 0.20,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.0, TemperatureMax: 1.0,
			Biome: BiomeOcean, Priority: 100,
		},
		// Mountain: highest elevation - exclusive
		{
			ElevationMin: 0.85, ElevationMax: 1.0,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.0, TemperatureMax: 1.0,
			Biome: BiomeMountain, Priority: 100,
		},

		// === TEMPERATURE-BASED PRIMARY BIOMES (wide ranges) ===
		// Ice: cold temperatures dominate
		{
			ElevationMin: 0.20, ElevationMax: 0.85,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.0, TemperatureMax: 0.25,
			Biome: BiomeIce, Priority: 95,
		},
		// Volcanic: hot + high elevation
		{
			ElevationMin: 0.60, ElevationMax: 0.85,
			MoistureMin: 0.0, MoistureMax: 0.40,
			TemperatureMin: 0.75, TemperatureMax: 1.0,
			Biome: BiomeVolcanic, Priority: 90,
		},
		// Desert: hot + dry + lower elevation
		{
			ElevationMin: 0.20, ElevationMax: 0.50,
			MoistureMin: 0.0, MoistureMax: 0.35,
			TemperatureMin: 0.65, TemperatureMax: 1.0,
			Biome: BiomeDesert, Priority: 85,
		},

		// === HIGH ELEVATION BIOMES ===
		// Crystal: high elevation + cool/moderate temp
		{
			ElevationMin: 0.65, ElevationMax: 0.85,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.25, TemperatureMax: 0.55,
			Biome: BiomeCrystal, Priority: 80,
		},

		// === WET BIOMES ===
		// Swamp: low elevation + wet + moderate temp
		{
			ElevationMin: 0.20, ElevationMax: 0.35,
			MoistureMin: 0.70, MoistureMax: 1.0,
			TemperatureMin: 0.35, TemperatureMax: 0.65,
			Biome: BiomeSwamp, Priority: 75,
		},
		// Forest: mid elevation + wet + moderate temp
		{
			ElevationMin: 0.35, ElevationMax: 0.65,
			MoistureMin: 0.55, MoistureMax: 1.0,
			TemperatureMin: 0.30, TemperatureMax: 0.60,
			Biome: BiomeForest, Priority: 70,
		},

		// === DRY BIOMES ===
		// Badlands: mid elevation + dry + warm
		{
			ElevationMin: 0.35, ElevationMax: 0.60,
			MoistureMin: 0.0, MoistureMax: 0.40,
			TemperatureMin: 0.55, TemperatureMax: 0.75,
			Biome: BiomeBadlands, Priority: 65,
		},

		// === SPECIAL BIOMES (more specific conditions) ===
		// Ancient: centered mid-range conditions
		{
			ElevationMin: 0.45, ElevationMax: 0.65,
			MoistureMin: 0.35, MoistureMax: 0.55,
			TemperatureMin: 0.40, TemperatureMax: 0.60,
			Biome: BiomeAncient, Priority: 60,
		},
		// Neon: hot + wet
		{
			ElevationMin: 0.20, ElevationMax: 0.45,
			MoistureMin: 0.60, MoistureMax: 1.0,
			TemperatureMin: 0.70, TemperatureMax: 1.0,
			Biome: BiomeNeon, Priority: 55,
		},
		// Plasma: hot + mid moisture
		{
			ElevationMin: 0.35, ElevationMax: 0.55,
			MoistureMin: 0.25, MoistureMax: 0.55,
			TemperatureMin: 0.65, TemperatureMax: 0.90,
			Biome: BiomePlasma, Priority: 50,
		},
		// Void: cold + dry + mid-high elevation (rare)
		{
			ElevationMin: 0.50, ElevationMax: 0.70,
			MoistureMin: 0.0, MoistureMax: 0.25,
			TemperatureMin: 0.15, TemperatureMax: 0.35,
			Biome: BiomeVoid, Priority: 45,
		},

		// === FALLBACK (fills remaining space) ===
		{
			ElevationMin: 0.20, ElevationMax: 0.70,
			MoistureMin: 0.25, MoistureMax: 0.55,
			TemperatureMin: 0.35, TemperatureMax: 0.70,
			Biome: BiomeSavanna, Priority: 10,
		},
		// Ultimate fallback
		{
			ElevationMin: 0.0, ElevationMax: 1.0,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.0, TemperatureMax: 1.0,
			Biome: BiomeSavanna, Priority: 0,
		},
	}
}

// LoadMapConfig loads a map configuration from a JSON file
func LoadMapConfig(path string) (*MapConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config MapConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveMapConfig saves a map configuration to a JSON file
func SaveMapConfig(config *MapConfig, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// MapConfigRegistry holds all available map configurations
type MapConfigRegistry struct {
	Configs map[string]*MapConfig
}

// NewMapConfigRegistry creates a new map config registry
func NewMapConfigRegistry() *MapConfigRegistry {
	registry := &MapConfigRegistry{
		Configs: make(map[string]*MapConfig),
	}

	// Register built-in presets
	for _, preset := range GetMapPresets() {
		registry.Register(preset)
	}

	return registry
}

// Register adds a map configuration to the registry
func (r *MapConfigRegistry) Register(config *MapConfig) {
	r.Configs[config.ID] = config
}

// Get retrieves a map configuration by ID
func (r *MapConfigRegistry) Get(id string) (*MapConfig, bool) {
	config, ok := r.Configs[id]
	return config, ok
}

// List returns all available map configuration IDs
func (r *MapConfigRegistry) List() []string {
	ids := make([]string, 0, len(r.Configs))
	for id := range r.Configs {
		ids = append(ids, id)
	}
	return ids
}
