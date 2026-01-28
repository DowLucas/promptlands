package worldgen

// GetMapPresets returns all built-in map presets
func GetMapPresets() []*MapConfig {
	return []*MapConfig{
		DefaultMapConfig(),
		InfernalRealmsPreset(),
		FrozenWastesPreset(),
		AncientWorldPreset(),
		NeonWildernessPreset(),
		VoidIncursionPreset(),
		CrystallineExpansePreset(),
	}
}

// InfernalRealmsPreset creates a volcanic-dominated world
func InfernalRealmsPreset() *MapConfig {
	return &MapConfig{
		ID:          "infernal_realms",
		Name:        "Infernal Realms",
		Description: "A world dominated by volcanic activity, lava flows, and badlands. High risk, high reward.",
		Theme:       "volcanic",
		Size:        MapSizeHuge, // 2048x2048
		ChunkSize:   32,

		// Noise scaled for larger maps
		ElevationNoise: NoiseLayerConfig{
			Octaves:     4,
			Frequency:   0.0018,
			Persistence: 0.55,
			Amplitude:   1.2,
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.0025,
			Persistence: 0.4,
			Amplitude:   0.7, // Less moisture overall
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0015,
			Persistence: 0.5,
			Amplitude:   1.3, // Hotter overall
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.01,
			Persistence: 0.4,
			Amplitude:   0.2,
			SeedOffset:  3000,
		},

		BiomeDistributions: []BiomeDistribution{
			{ElevationMin: 0.0, ElevationMax: 0.15, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeOcean, Priority: 100},
			{ElevationMin: 0.88, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeMountain, Priority: 100},
			// Heavy volcanic presence
			{ElevationMin: 0.40, ElevationMax: 0.88, MoistureMin: 0.0, MoistureMax: 0.60, TemperatureMin: 0.55, TemperatureMax: 1.0, Biome: BiomeVolcanic, Priority: 90},
			{ElevationMin: 0.25, ElevationMax: 0.65, MoistureMin: 0.0, MoistureMax: 0.45, TemperatureMin: 0.45, TemperatureMax: 0.75, Biome: BiomeBadlands, Priority: 85},
			{ElevationMin: 0.20, ElevationMax: 0.55, MoistureMin: 0.0, MoistureMax: 0.35, TemperatureMin: 0.50, TemperatureMax: 0.90, Biome: BiomeDesert, Priority: 80},
			{ElevationMin: 0.30, ElevationMax: 0.60, MoistureMin: 0.25, MoistureMax: 0.55, TemperatureMin: 0.60, TemperatureMax: 0.90, Biome: BiomePlasma, Priority: 75},
			{ElevationMin: 0.0, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeSavanna, Priority: 0},
		},

		OceanBorder:      true,
		OceanBorderWidth: 32,
		RiverCount:       4,
		LakeChance:       0.005,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.08,
			CachesPerChunk:    0.12,
			PortalPairsPerMap: 20,
			ObelisksPerChunk:  0.05,
			DungeonsPerMap:    16,
			VillagesPerMap:    4,
			RuinsPerChunk:     0.06,
		},

		ResourceDensity:      1.2,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 1.5,
	}
}

// FrozenWastesPreset creates an ice-dominated world
func FrozenWastesPreset() *MapConfig {
	return &MapConfig{
		ID:          "frozen_wastes",
		Name:        "Frozen Wastes",
		Description: "A frozen world where ice dominates and survival is challenging. Crystal formations hold ancient power.",
		Theme:       "ice",
		Size:        MapSizeHuge, // 2048x2048
		ChunkSize:   32,

		ElevationNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0015,
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.002,
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.0012,
			Persistence: 0.5,
			Amplitude:   0.5, // Much colder overall
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.008,
			Persistence: 0.4,
			Amplitude:   0.1,
			SeedOffset:  3000,
		},

		BiomeDistributions: []BiomeDistribution{
			{ElevationMin: 0.0, ElevationMax: 0.18, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeOcean, Priority: 100},
			{ElevationMin: 0.85, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeMountain, Priority: 100},
			// Heavy ice presence
			{ElevationMin: 0.18, ElevationMax: 0.85, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 0.45, Biome: BiomeIce, Priority: 90},
			{ElevationMin: 0.50, ElevationMax: 0.85, MoistureMin: 0.30, MoistureMax: 0.80, TemperatureMin: 0.30, TemperatureMax: 0.55, Biome: BiomeCrystal, Priority: 85},
			{ElevationMin: 0.25, ElevationMax: 0.60, MoistureMin: 0.50, MoistureMax: 1.0, TemperatureMin: 0.40, TemperatureMax: 0.60, Biome: BiomeForest, Priority: 70},
			{ElevationMin: 0.0, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeSavanna, Priority: 0},
		},

		OceanBorder:      true,
		OceanBorderWidth: 40,
		RiverCount:       6,
		LakeChance:       0.02,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.06,
			CachesPerChunk:    0.08,
			PortalPairsPerMap: 12,
			ObelisksPerChunk:  0.06,
			DungeonsPerMap:    10,
			VillagesPerMap:    6,
			RuinsPerChunk:     0.04,
		},

		ResourceDensity:      0.9,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 1.3,
	}
}

// AncientWorldPreset creates a world with prominent ruins and forests
func AncientWorldPreset() *MapConfig {
	return &MapConfig{
		ID:          "ancient_world",
		Name:        "Ancient World",
		Description: "A world of ancient civilizations, mystical forests, and forgotten ruins. History awaits discovery.",
		Theme:       "ancient",
		Size:        MapSizeMassive, // 4096x4096 - epic exploration
		ChunkSize:   32,

		ElevationNoise: NoiseLayerConfig{
			Octaves:     4,
			Frequency:   0.0008,
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.001,
			Persistence: 0.55,
			Amplitude:   1.1,
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.0008,
			Persistence: 0.5,
			Amplitude:   0.9,
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.004,
			Persistence: 0.4,
			Amplitude:   0.15,
			SeedOffset:  3000,
		},

		BiomeDistributions: []BiomeDistribution{
			{ElevationMin: 0.0, ElevationMax: 0.20, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeOcean, Priority: 100},
			{ElevationMin: 0.88, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeMountain, Priority: 100},
			// Ancient and forest focus
			{ElevationMin: 0.35, ElevationMax: 0.70, MoistureMin: 0.25, MoistureMax: 0.70, TemperatureMin: 0.35, TemperatureMax: 0.70, Biome: BiomeAncient, Priority: 85},
			{ElevationMin: 0.20, ElevationMax: 0.65, MoistureMin: 0.50, MoistureMax: 1.0, TemperatureMin: 0.30, TemperatureMax: 0.65, Biome: BiomeForest, Priority: 80},
			{ElevationMin: 0.20, ElevationMax: 0.45, MoistureMin: 0.65, MoistureMax: 1.0, TemperatureMin: 0.40, TemperatureMax: 0.60, Biome: BiomeSwamp, Priority: 75},
			{ElevationMin: 0.20, ElevationMax: 0.55, MoistureMin: 0.0, MoistureMax: 0.35, TemperatureMin: 0.55, TemperatureMax: 0.85, Biome: BiomeDesert, Priority: 70},
			{ElevationMin: 0.0, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeSavanna, Priority: 0},
		},

		OceanBorder:      true,
		OceanBorderWidth: 64,
		RiverCount:       12,
		LakeChance:       0.03,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.08,
			CachesPerChunk:    0.08,
			PortalPairsPerMap: 32,
			ObelisksPerChunk:  0.08,
			DungeonsPerMap:    24,
			VillagesPerMap:    16,
			RuinsPerChunk:     0.10,
		},

		ResourceDensity:      1.1,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 1.0,
	}
}

// NeonWildernessPreset creates a bioluminescent sci-fi world
func NeonWildernessPreset() *MapConfig {
	return &MapConfig{
		ID:          "neon_wilderness",
		Name:        "Neon Wilderness",
		Description: "Alien terraforming left this world glowing with bioluminescent life. Strange and beautiful.",
		Theme:       "scifi",
		Size:        MapSizeHuge, // 2048x2048
		ChunkSize:   32,

		ElevationNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0018,
			Persistence: 0.5,
			Amplitude:   0.9,
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0015,
			Persistence: 0.6,
			Amplitude:   1.2, // Wetter world
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0012,
			Persistence: 0.55,
			Amplitude:   1.1, // Warmer
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.01,
			Persistence: 0.4,
			Amplitude:   0.2,
			SeedOffset:  3000,
		},

		BiomeDistributions: []BiomeDistribution{
			{ElevationMin: 0.0, ElevationMax: 0.18, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeOcean, Priority: 100},
			{ElevationMin: 0.85, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeMountain, Priority: 100},
			// Neon and plasma dominance
			{ElevationMin: 0.20, ElevationMax: 0.65, MoistureMin: 0.55, MoistureMax: 1.0, TemperatureMin: 0.55, TemperatureMax: 1.0, Biome: BiomeNeon, Priority: 90},
			{ElevationMin: 0.30, ElevationMax: 0.65, MoistureMin: 0.20, MoistureMax: 0.60, TemperatureMin: 0.60, TemperatureMax: 0.95, Biome: BiomePlasma, Priority: 85},
			{ElevationMin: 0.55, ElevationMax: 0.85, MoistureMin: 0.30, MoistureMax: 0.70, TemperatureMin: 0.30, TemperatureMax: 0.60, Biome: BiomeCrystal, Priority: 80},
			{ElevationMin: 0.20, ElevationMax: 0.50, MoistureMin: 0.60, MoistureMax: 1.0, TemperatureMin: 0.35, TemperatureMax: 0.55, Biome: BiomeSwamp, Priority: 75},
			{ElevationMin: 0.0, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeSavanna, Priority: 0},
		},

		OceanBorder:      true,
		OceanBorderWidth: 40,
		RiverCount:       10,
		LakeChance:       0.03,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.06,
			CachesPerChunk:    0.12,
			PortalPairsPerMap: 20,
			ObelisksPerChunk:  0.04,
			DungeonsPerMap:    12,
			VillagesPerMap:    8,
			RuinsPerChunk:     0.05,
		},

		ResourceDensity:      1.3,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 1.1,
	}
}

// VoidIncursionPreset creates a world being consumed by void corruption
func VoidIncursionPreset() *MapConfig {
	return &MapConfig{
		ID:          "void_incursion",
		Name:        "Void Incursion",
		Description: "Reality is collapsing as the Void spreads. A desperate battle for survival against cosmic corruption.",
		Theme:       "horror",
		Size:        MapSizeLarge, // 1024x1024 - more intense
		ChunkSize:   32,

		ElevationNoise: NoiseLayerConfig{
			Octaves:     4,
			Frequency:   0.003,
			Persistence: 0.5,
			Amplitude:   1.1,
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0035,
			Persistence: 0.45,
			Amplitude:   0.9,
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0025,
			Persistence: 0.55,
			Amplitude:   1.0,
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.015,
			Persistence: 0.5,
			Amplitude:   0.25,
			SeedOffset:  3000,
		},

		BiomeDistributions: []BiomeDistribution{
			{ElevationMin: 0.0, ElevationMax: 0.18, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeOcean, Priority: 100},
			{ElevationMin: 0.88, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeMountain, Priority: 100},
			// Void spreads from specific conditions
			{ElevationMin: 0.40, ElevationMax: 0.80, MoistureMin: 0.0, MoistureMax: 0.35, TemperatureMin: 0.20, TemperatureMax: 0.50, Biome: BiomeVoid, Priority: 92},
			{ElevationMin: 0.50, ElevationMax: 0.88, MoistureMin: 0.0, MoistureMax: 0.50, TemperatureMin: 0.65, TemperatureMax: 1.0, Biome: BiomeVolcanic, Priority: 85},
			{ElevationMin: 0.20, ElevationMax: 0.50, MoistureMin: 0.60, MoistureMax: 1.0, TemperatureMin: 0.30, TemperatureMax: 0.55, Biome: BiomeSwamp, Priority: 80},
			{ElevationMin: 0.55, ElevationMax: 0.88, MoistureMin: 0.35, MoistureMax: 0.75, TemperatureMin: 0.30, TemperatureMax: 0.55, Biome: BiomeCrystal, Priority: 75},
			{ElevationMin: 0.25, ElevationMax: 0.55, MoistureMin: 0.45, MoistureMax: 0.85, TemperatureMin: 0.35, TemperatureMax: 0.60, Biome: BiomeForest, Priority: 65},
			{ElevationMin: 0.0, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeSavanna, Priority: 0},
		},

		OceanBorder:      true,
		OceanBorderWidth: 24,
		RiverCount:       4,
		LakeChance:       0.01,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.05,
			CachesPerChunk:    0.06,
			PortalPairsPerMap: 24,
			ObelisksPerChunk:  0.10,
			DungeonsPerMap:    16,
			VillagesPerMap:    2,
			RuinsPerChunk:     0.12,
		},

		ResourceDensity:      0.8,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 2.0,
	}
}

// CrystallineExpansePreset creates a crystal-dominated mystical world
func CrystallineExpansePreset() *MapConfig {
	return &MapConfig{
		ID:          "crystalline_expanse",
		Name:        "Crystalline Expanse",
		Description: "A world where crystalline formations dominate the landscape, humming with arcane energy.",
		Theme:       "crystal",
		Size:        MapSizeHuge, // 2048x2048
		ChunkSize:   32,

		ElevationNoise: NoiseLayerConfig{
			Octaves:     4,
			Frequency:   0.0015,
			Persistence: 0.55,
			Amplitude:   1.2, // More elevation variation
			SeedOffset:  0,
		},
		MoistureNoise: NoiseLayerConfig{
			Octaves:     3,
			Frequency:   0.0018,
			Persistence: 0.5,
			Amplitude:   1.0,
			SeedOffset:  1000,
		},
		TemperatureNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.0015,
			Persistence: 0.5,
			Amplitude:   0.85, // Cooler
			SeedOffset:  2000,
		},
		VariationNoise: NoiseLayerConfig{
			Octaves:     2,
			Frequency:   0.008,
			Persistence: 0.4,
			Amplitude:   0.15,
			SeedOffset:  3000,
		},

		BiomeDistributions: []BiomeDistribution{
			{ElevationMin: 0.0, ElevationMax: 0.18, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeOcean, Priority: 100},
			{ElevationMin: 0.88, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeMountain, Priority: 100},
			// Crystal dominance
			{ElevationMin: 0.45, ElevationMax: 0.88, MoistureMin: 0.20, MoistureMax: 0.80, TemperatureMin: 0.20, TemperatureMax: 0.65, Biome: BiomeCrystal, Priority: 90},
			{ElevationMin: 0.20, ElevationMax: 0.60, MoistureMin: 0.0, MoistureMax: 0.50, TemperatureMin: 0.0, TemperatureMax: 0.35, Biome: BiomeIce, Priority: 85},
			{ElevationMin: 0.25, ElevationMax: 0.55, MoistureMin: 0.50, MoistureMax: 1.0, TemperatureMin: 0.35, TemperatureMax: 0.60, Biome: BiomeForest, Priority: 75},
			{ElevationMin: 0.35, ElevationMax: 0.65, MoistureMin: 0.25, MoistureMax: 0.60, TemperatureMin: 0.40, TemperatureMax: 0.65, Biome: BiomeAncient, Priority: 70},
			{ElevationMin: 0.0, ElevationMax: 1.0, MoistureMin: 0.0, MoistureMax: 1.0, TemperatureMin: 0.0, TemperatureMax: 1.0, Biome: BiomeSavanna, Priority: 0},
		},

		OceanBorder:      true,
		OceanBorderWidth: 48,
		RiverCount:       8,
		LakeChance:       0.02,

		Structures: StructureSpawnConfig{
			ShrinesPerChunk:   0.10,
			CachesPerChunk:    0.12,
			PortalPairsPerMap: 16,
			ObelisksPerChunk:  0.06,
			DungeonsPerMap:    12,
			VillagesPerMap:    6,
			RuinsPerChunk:     0.05,
		},

		ResourceDensity:      1.4,
		RespawnEnabled:       true,
		FogOfWar:             true,
		DifficultyMultiplier: 1.2,
	}
}
