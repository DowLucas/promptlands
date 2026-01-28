package worldgen

// BiomeType represents distinct biome classifications
type BiomeType string

const (
	// === CORE BIOMES (matching reference aesthetic) ===
	BiomeForest      BiomeType = "forest"       // Lush green forest with trees
	BiomeDesert      BiomeType = "desert"       // Sandy yellow desert with cacti/dunes
	BiomeVolcanic    BiomeType = "volcanic"     // Dark red volcanic with lava flows
	BiomeIce         BiomeType = "ice"          // Blue/white frozen tundra with crystals
	BiomeSavanna     BiomeType = "savanna"      // Teal/green grassland with scattered trees
	BiomeBadlands    BiomeType = "badlands"     // Orange rocky canyon terrain
	BiomeSwamp       BiomeType = "swamp"        // Dark green murky wetlands
	BiomeCrystal     BiomeType = "crystal"      // Purple/pink crystalline formations

	// === FANTASY/SCI-FI BIOMES ===
	BiomeVoid        BiomeType = "void"         // Dark purple corrupted void zone
	BiomeNeon        BiomeType = "neon"         // Bright cyan bioluminescent jungle
	BiomePlasma      BiomeType = "plasma"       // Orange/yellow energy fields
	BiomeAncient     BiomeType = "ancient"      // Golden ruined civilization

	// === WATER/BARRIER BIOMES ===
	BiomeOcean       BiomeType = "ocean"        // Deep blue water
	BiomeMountain    BiomeType = "mountain"     // Gray impassable peaks
)

// BiomeProperties defines the gameplay characteristics of a biome
type BiomeProperties struct {
	// Display properties
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"` // Hex color for rendering

	// Movement properties
	Passable     bool    `json:"passable"`
	MovementCost float64 `json:"movement_cost"` // Energy cost multiplier (1.0 = normal)

	// Resource generation
	EnergyPerTick int `json:"energy_per_tick"` // Passive energy when owned
	ClaimCost     int `json:"claim_cost"`      // Energy cost to claim this tile

	// Combat modifiers
	DefenseBonus int `json:"defense_bonus"` // Damage reduction when defending
	AttackBonus  int `json:"attack_bonus"`  // Damage bonus when attacking

	// Environmental effects
	DamagePerTick  int `json:"damage_per_tick"`  // Environmental damage per tick
	HealPerTick    int `json:"heal_per_tick"`    // Environmental healing per tick
	VisionModifier int `json:"vision_modifier"`  // Vision range modifier

	// Spawning rules
	CanSpawnAgent   bool    `json:"can_spawn_agent"`
	ResourceChance  float64 `json:"resource_chance"`  // Chance for resource nodes
	StructureChance float64 `json:"structure_chance"` // Chance for structures

	// Special effects
	SpecialEffect string `json:"special_effect,omitempty"`

	// Terrain classification (backwards compatibility)
	TerrainClass TerrainType `json:"terrain_class"`
}

// BiomeRegistry holds all biome definitions
type BiomeRegistry struct {
	Biomes map[BiomeType]BiomeProperties
}

// DefaultBiomeRegistry returns the standard biome definitions
func DefaultBiomeRegistry() *BiomeRegistry {
	return &BiomeRegistry{
		Biomes: map[BiomeType]BiomeProperties{
			// === FOREST - Lush green woodland ===
			BiomeForest: {
				Name:            "Enchanted Forest",
				Description:     "Dense woodland where ancient trees whisper secrets and spirits dwell among the roots",
				Color:           "#2D5A27", // Dark forest green
				Passable:        true,
				MovementCost:    1.2,
				EnergyPerTick:   2,
				ClaimCost:       6,
				DefenseBonus:    2,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  -1,
				CanSpawnAgent:   true,
				ResourceChance:  0.10,
				StructureChance: 0.03,
				SpecialEffect:   "nature_blessing",
				TerrainClass:    TerrainForest,
			},

			// === DESERT - Sandy dunes with cacti ===
			BiomeDesert: {
				Name:            "Scorching Dunes",
				Description:     "Vast sandy expanse dotted with ancient ruins and hardy cacti, hiding treasures beneath the sand",
				Color:           "#E8C86B", // Sandy yellow
				Passable:        true,
				MovementCost:    1.3,
				EnergyPerTick:   1,
				ClaimCost:       4,
				DefenseBonus:    0,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  2,
				CanSpawnAgent:   true,
				ResourceChance:  0.04,
				StructureChance: 0.05,
				SpecialEffect:   "mirage",
				TerrainClass:    TerrainPlains,
			},

			// === VOLCANIC - Lava and fire ===
			BiomeVolcanic: {
				Name:            "Infernal Caldera",
				Description:     "Scorched earth where rivers of lava flow and fire elementals roam the blackened rock",
				Color:           "#4A1515", // Dark volcanic red
				Passable:        true,
				MovementCost:    1.5,
				EnergyPerTick:   3,
				ClaimCost:       10,
				DefenseBonus:    0,
				AttackBonus:     2,
				DamagePerTick:   1,
				HealPerTick:     0,
				VisionModifier:  1,
				CanSpawnAgent:   false,
				ResourceChance:  0.08,
				StructureChance: 0.02,
				SpecialEffect:   "burning_ground",
				TerrainClass:    TerrainMountain,
			},

			// === ICE - Frozen tundra ===
			BiomeIce: {
				Name:            "Frostbound Wastes",
				Description:     "Eternal winter realm where crystalline ice formations hold frozen memories and ancient power",
				Color:           "#1E3A5F", // Deep icy blue
				Passable:        true,
				MovementCost:    1.4,
				EnergyPerTick:   1,
				ClaimCost:       7,
				DefenseBonus:    1,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  0,
				CanSpawnAgent:   true,
				ResourceChance:  0.06,
				StructureChance: 0.03,
				SpecialEffect:   "frozen_time",
				TerrainClass:    TerrainPlains,
			},

			// === SAVANNA - Grassland with scattered trees ===
			BiomeSavanna: {
				Name:            "Mystic Savanna",
				Description:     "Rolling grasslands with scattered ancient trees, where ley lines converge beneath the soil",
				Color:           "#3D7A6B", // Teal-green
				Passable:        true,
				MovementCost:    1.0,
				EnergyPerTick:   2,
				ClaimCost:       5,
				DefenseBonus:    0,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  1,
				CanSpawnAgent:   true,
				ResourceChance:  0.07,
				StructureChance: 0.04,
				SpecialEffect:   "ley_convergence",
				TerrainClass:    TerrainPlains,
			},

			// === BADLANDS - Orange rocky terrain ===
			BiomeBadlands: {
				Name:            "Ember Badlands",
				Description:     "Scorched orange canyons and mesas where volcanic activity once shaped the land",
				Color:           "#C46B2E", // Burnt orange
				Passable:        true,
				MovementCost:    1.3,
				EnergyPerTick:   1,
				ClaimCost:       5,
				DefenseBonus:    1,
				AttackBonus:     1,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  0,
				CanSpawnAgent:   true,
				ResourceChance:  0.06,
				StructureChance: 0.03,
				SpecialEffect:   "mineral_rich",
				TerrainClass:    TerrainPlains,
			},

			// === SWAMP - Dark murky wetlands ===
			BiomeSwamp: {
				Name:            "Shadowmire",
				Description:     "Dark mist-shrouded wetlands where shadows move with purpose and secrets lurk beneath murky waters",
				Color:           "#2A4A3A", // Dark swamp green
				Passable:        true,
				MovementCost:    1.6,
				EnergyPerTick:   2,
				ClaimCost:       6,
				DefenseBonus:    2,
				AttackBonus:     1,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  -2,
				CanSpawnAgent:   false,
				ResourceChance:  0.08,
				StructureChance: 0.03,
				SpecialEffect:   "shadow_veil",
				TerrainClass:    TerrainForest,
			},

			// === CRYSTAL - Crystalline formations ===
			BiomeCrystal: {
				Name:            "Crystal Spires",
				Description:     "Towering crystalline formations that hum with arcane energy and refract light into rainbows",
				Color:           "#9B4DCA", // Purple crystal
				Passable:        true,
				MovementCost:    1.1,
				EnergyPerTick:   4,
				ClaimCost:       12,
				DefenseBonus:    0,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  1,
				CanSpawnAgent:   false,
				ResourceChance:  0.15,
				StructureChance: 0.05,
				SpecialEffect:   "energy_resonance",
				TerrainClass:    TerrainMountain,
			},

			// === VOID - Corrupted dimension ===
			BiomeVoid: {
				Name:            "Void Rift",
				Description:     "Reality tears apart here, revealing glimpses of the space between dimensions",
				Color:           "#1A0A2E", // Deep void purple
				Passable:        true,
				MovementCost:    1.4,
				EnergyPerTick:   5,
				ClaimCost:       15,
				DefenseBonus:    0,
				AttackBonus:     0,
				DamagePerTick:   1,
				HealPerTick:     0,
				VisionModifier:  -2,
				CanSpawnAgent:   false,
				ResourceChance:  0.12,
				StructureChance: 0.06,
				SpecialEffect:   "reality_warp",
				TerrainClass:    TerrainPlains,
			},

			// === NEON - Bioluminescent jungle ===
			BiomeNeon: {
				Name:            "Neon Wilds",
				Description:     "Alien flora glows with bioluminescence, a remnant of ancient terraforming technology",
				Color:           "#00CED1", // Bright cyan
				Passable:        true,
				MovementCost:    1.2,
				EnergyPerTick:   3,
				ClaimCost:       9,
				DefenseBonus:    1,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     1,
				VisionModifier:  0,
				CanSpawnAgent:   true,
				ResourceChance:  0.10,
				StructureChance: 0.04,
				SpecialEffect:   "bio_regen",
				TerrainClass:    TerrainForest,
			},

			// === PLASMA - Energy fields ===
			BiomePlasma: {
				Name:            "Plasma Fields",
				Description:     "Crackling energy pylons harvest power from the atmosphere in endless fields of light",
				Color:           "#FF6B35", // Bright orange
				Passable:        true,
				MovementCost:    1.0,
				EnergyPerTick:   4,
				ClaimCost:       11,
				DefenseBonus:    0,
				AttackBonus:     2,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  2,
				CanSpawnAgent:   true,
				ResourceChance:  0.08,
				StructureChance: 0.02,
				SpecialEffect:   "energy_surge",
				TerrainClass:    TerrainPlains,
			},

			// === ANCIENT - Ruined civilization ===
			BiomeAncient: {
				Name:            "Ancient Ruins",
				Description:     "Crumbling temples and monuments of a forgotten civilization, their runes still pulsing with power",
				Color:           "#DAA520", // Golden
				Passable:        true,
				MovementCost:    1.1,
				EnergyPerTick:   2,
				ClaimCost:       8,
				DefenseBonus:    1,
				AttackBonus:     1,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  0,
				CanSpawnAgent:   true,
				ResourceChance:  0.10,
				StructureChance: 0.10,
				SpecialEffect:   "ancient_knowledge",
				TerrainClass:    TerrainPlains,
			},

			// === OCEAN - Deep water (impassable) ===
			BiomeOcean: {
				Name:            "Abyssal Deep",
				Description:     "Dark waters where ancient leviathans slumber and no land-dweller can tread",
				Color:           "#0A1628", // Deep ocean blue
				Passable:        false,
				MovementCost:    0,
				EnergyPerTick:   0,
				ClaimCost:       0,
				DefenseBonus:    0,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  0,
				CanSpawnAgent:   false,
				ResourceChance:  0,
				StructureChance: 0,
				SpecialEffect:   "",
				TerrainClass:    TerrainWater,
			},

			// === MOUNTAIN - Impassable peaks ===
			BiomeMountain: {
				Name:            "Skyward Peaks",
				Description:     "Towering mountains that pierce the clouds, impassable to all but the mightiest",
				Color:           "#4A4A4A", // Gray stone
				Passable:        false,
				MovementCost:    0,
				EnergyPerTick:   0,
				ClaimCost:       0,
				DefenseBonus:    0,
				AttackBonus:     0,
				DamagePerTick:   0,
				HealPerTick:     0,
				VisionModifier:  0,
				CanSpawnAgent:   false,
				ResourceChance:  0,
				StructureChance: 0,
				SpecialEffect:   "",
				TerrainClass:    TerrainMountain,
			},
		},
	}
}

// GetBiome returns the properties for a biome type
func (r *BiomeRegistry) GetBiome(biomeType BiomeType) (BiomeProperties, bool) {
	props, ok := r.Biomes[biomeType]
	return props, ok
}

// GetPassableBiomes returns all passable biome types
func (r *BiomeRegistry) GetPassableBiomes() []BiomeType {
	var passable []BiomeType
	for biomeType, props := range r.Biomes {
		if props.Passable {
			passable = append(passable, biomeType)
		}
	}
	return passable
}

// GetSpawnableBiomes returns all biomes where agents can spawn
func (r *BiomeRegistry) GetSpawnableBiomes() []BiomeType {
	var spawnable []BiomeType
	for biomeType, props := range r.Biomes {
		if props.CanSpawnAgent {
			spawnable = append(spawnable, biomeType)
		}
	}
	return spawnable
}

// IsPassable checks if a biome allows movement
func (r *BiomeRegistry) IsPassable(biomeType BiomeType) bool {
	props, ok := r.Biomes[biomeType]
	if !ok {
		return false
	}
	return props.Passable
}

// GetTerrainClass maps a biome to its base terrain type
func (r *BiomeRegistry) GetTerrainClass(biomeType BiomeType) TerrainType {
	props, ok := r.Biomes[biomeType]
	if !ok {
		return TerrainPlains
	}
	return props.TerrainClass
}

// GetAllBiomeTypes returns all defined biome types
func (r *BiomeRegistry) GetAllBiomeTypes() []BiomeType {
	types := make([]BiomeType, 0, len(r.Biomes))
	for biomeType := range r.Biomes {
		types = append(types, biomeType)
	}
	return types
}

// GetBiomeColor returns the hex color for a biome
func (r *BiomeRegistry) GetBiomeColor(biomeType BiomeType) string {
	props, ok := r.Biomes[biomeType]
	if !ok {
		return "#7CCD7C" // Default green
	}
	return props.Color
}
