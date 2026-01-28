package worldgen

import (
	"encoding/json"
	"os"
)

// MapSize represents predefined map sizes for open-world feel
type MapSize string

const (
	MapSizeTiny    MapSize = "tiny"    // 32x32 - Quick games
	MapSizeSmall   MapSize = "small"   // 64x64 - Short games
	MapSizeMedium  MapSize = "medium"  // 128x128 - Standard games
	MapSizeLarge   MapSize = "large"   // 256x256 - Long games
	MapSizeHuge    MapSize = "huge"    // 512x512 - Epic games
	MapSizeMassive MapSize = "massive" // 1024x1024 - Open world
)

// GetMapSizeValue returns the actual size value
func GetMapSizeValue(size MapSize) int {
	switch size {
	case MapSizeTiny:
		return 32
	case MapSizeSmall:
		return 64
	case MapSizeMedium:
		return 128
	case MapSizeLarge:
		return 256
	case MapSizeHuge:
		return 512
	case MapSizeMassive:
		return 1024
	default:
		return 128
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

// DefaultMapConfig returns a balanced default map configuration
// Uses LOW frequency noise to create LARGE biome regions like the reference image
func DefaultMapConfig() *MapConfig {
	return &MapConfig{
		ID:          "default",
		Name:        "Standard World",
		Description: "A balanced procedurally generated world with large distinct biomes",
		Theme:       "mixed",
		Size:        MapSizeLarge, // 256x256 for open world feel
		ChunkSize:   32,

		// LOW frequency = LARGE biome regions (like reference image)
		ElevationNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.006, // Very low = huge elevation regions
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.008, // Low = large moisture zones
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.005, // Very low = huge temperature bands
			Persistence: 0.6,
			Amplitude:   1.0,
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.03, // Slight variation for organic edges
			Persistence: 0.4,
			Amplitude:   0.15,
			SeedOffset:  3000,
		},

		BiomeDistributions: DefaultBiomeDistributions(),

		OceanBorder:      true,
		OceanBorderWidth: 12,
		RiverCount:       4,
		LakeChance:       0.01,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.08,
			CachesPerChunk:    0.12,
			PortalPairsPerMap: 6,
			ObelisksPerChunk:  0.06,
			DungeonsPerMap:    4,
			VillagesPerMap:    3,
			RuinsPerChunk:     0.04,
		},

		ResourceDensity:      1.0,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 1.0,
	}
}

// DefaultBiomeDistributions returns biome distribution rules for large contiguous regions
// Designed to create the distinct biome zones seen in the reference image
func DefaultBiomeDistributions() []BiomeDistribution {
	return []BiomeDistribution{
		// === WATER (lowest elevation) ===
		{
			ElevationMin: 0.0, ElevationMax: 0.20,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.0, TemperatureMax: 1.0,
			Biome: BiomeOcean, Priority: 100,
		},

		// === MOUNTAINS (highest elevation) ===
		{
			ElevationMin: 0.85, ElevationMax: 1.0,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.0, TemperatureMax: 1.0,
			Biome: BiomeMountain, Priority: 100,
		},

		// === ICE BIOME (cold temperature) ===
		{
			ElevationMin: 0.20, ElevationMax: 0.85,
			MoistureMin: 0.0, MoistureMax: 1.0,
			TemperatureMin: 0.0, TemperatureMax: 0.20,
			Biome: BiomeIce, Priority: 90,
		},

		// === VOLCANIC (hot + high elevation) ===
		{
			ElevationMin: 0.55, ElevationMax: 0.85,
			MoistureMin: 0.0, MoistureMax: 0.50,
			TemperatureMin: 0.75, TemperatureMax: 1.0,
			Biome: BiomeVolcanic, Priority: 88,
		},

		// === DESERT (hot + dry) ===
		{
			ElevationMin: 0.20, ElevationMax: 0.65,
			MoistureMin: 0.0, MoistureMax: 0.30,
			TemperatureMin: 0.60, TemperatureMax: 1.0,
			Biome: BiomeDesert, Priority: 85,
		},

		// === BADLANDS (warm + dry + mid elevation) ===
		{
			ElevationMin: 0.35, ElevationMax: 0.70,
			MoistureMin: 0.0, MoistureMax: 0.35,
			TemperatureMin: 0.50, TemperatureMax: 0.75,
			Biome: BiomeBadlands, Priority: 82,
		},

		// === CRYSTAL (high elevation + moderate temp) ===
		{
			ElevationMin: 0.60, ElevationMax: 0.85,
			MoistureMin: 0.30, MoistureMax: 0.70,
			TemperatureMin: 0.25, TemperatureMax: 0.55,
			Biome: BiomeCrystal, Priority: 80,
		},

		// === VOID (rare - extreme conditions) ===
		{
			ElevationMin: 0.50, ElevationMax: 0.75,
			MoistureMin: 0.0, MoistureMax: 0.20,
			TemperatureMin: 0.20, TemperatureMax: 0.40,
			Biome: BiomeVoid, Priority: 78,
		},

		// === PLASMA (hot + mid moisture) ===
		{
			ElevationMin: 0.30, ElevationMax: 0.60,
			MoistureMin: 0.20, MoistureMax: 0.50,
			TemperatureMin: 0.70, TemperatureMax: 0.95,
			Biome: BiomePlasma, Priority: 75,
		},

		// === NEON (hot + wet) ===
		{
			ElevationMin: 0.20, ElevationMax: 0.55,
			MoistureMin: 0.65, MoistureMax: 1.0,
			TemperatureMin: 0.65, TemperatureMax: 1.0,
			Biome: BiomeNeon, Priority: 75,
		},

		// === SWAMP (low elevation + wet) ===
		{
			ElevationMin: 0.20, ElevationMax: 0.40,
			MoistureMin: 0.70, MoistureMax: 1.0,
			TemperatureMin: 0.35, TemperatureMax: 0.65,
			Biome: BiomeSwamp, Priority: 72,
		},

		// === ANCIENT RUINS (mid temp + mid moisture + mid elevation) ===
		{
			ElevationMin: 0.40, ElevationMax: 0.65,
			MoistureMin: 0.30, MoistureMax: 0.60,
			TemperatureMin: 0.40, TemperatureMax: 0.65,
			Biome: BiomeAncient, Priority: 70,
		},

		// === FOREST (wet + moderate temp) ===
		{
			ElevationMin: 0.20, ElevationMax: 0.70,
			MoistureMin: 0.50, MoistureMax: 1.0,
			TemperatureMin: 0.30, TemperatureMax: 0.65,
			Biome: BiomeForest, Priority: 65,
		},

		// === SAVANNA (moderate everything - fills remaining space) ===
		{
			ElevationMin: 0.20, ElevationMax: 0.70,
			MoistureMin: 0.25, MoistureMax: 0.65,
			TemperatureMin: 0.35, TemperatureMax: 0.70,
			Biome: BiomeSavanna, Priority: 50,
		},

		// === FALLBACK (savanna for anything not matched) ===
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
