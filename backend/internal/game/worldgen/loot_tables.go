package worldgen

import (
	"encoding/json"
	"math/rand"
	"os"
)

// LootRarity defines item rarity tiers
type LootRarity string

const (
	RarityCommon    LootRarity = "common"
	RarityUncommon  LootRarity = "uncommon"
	RarityRare      LootRarity = "rare"
	RarityEpic      LootRarity = "epic"
	RarityLegendary LootRarity = "legendary"
	RarityMythic    LootRarity = "mythic"
)

// LootEntry represents a single item in a loot table
type LootEntry struct {
	ItemID      string     `json:"item_id"`
	Weight      int        `json:"weight"`        // Higher = more common
	MinQuantity int        `json:"min_quantity"`
	MaxQuantity int        `json:"max_quantity"`
	Rarity      LootRarity `json:"rarity"`
}

// LootPool represents a pool of loot entries
type LootPool struct {
	Name       string      `json:"name"`
	Rolls      RollRange   `json:"rolls"`
	BonusRolls RollRange   `json:"bonus_rolls"`
	Entries    []LootEntry `json:"entries"`
}

// RollRange defines a min/max range for rolls
type RollRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// LootTable represents a complete loot table
type LootTable struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Pools       []LootPool `json:"pools"`
}

// LootTableRegistry holds all loot tables
type LootTableRegistry struct {
	Tables map[string]*LootTable
	rng    *rand.Rand
}

// NewLootTableRegistry creates a new loot table registry
func NewLootTableRegistry(seed int64) *LootTableRegistry {
	registry := &LootTableRegistry{
		Tables: make(map[string]*LootTable),
		rng:    rand.New(rand.NewSource(seed)),
	}
	registry.registerDefaultTables()
	return registry
}

// Register adds a loot table
func (r *LootTableRegistry) Register(table *LootTable) {
	r.Tables[table.ID] = table
}

// Get retrieves a loot table by ID
func (r *LootTableRegistry) Get(id string) (*LootTable, bool) {
	table, ok := r.Tables[id]
	return table, ok
}

// LootResult represents the result of a loot roll
type LootResult struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// Roll performs a loot roll on a table
func (r *LootTableRegistry) Roll(tableID string, luck float64) []LootResult {
	table, ok := r.Tables[tableID]
	if !ok {
		return nil
	}
	return r.RollTable(table, luck)
}

// RollTable performs a loot roll on a specific table
func (r *LootTableRegistry) RollTable(table *LootTable, luck float64) []LootResult {
	var results []LootResult

	for _, pool := range table.Pools {
		rolls := r.rollRange(pool.Rolls)
		bonusRolls := int(float64(r.rollRange(pool.BonusRolls)) * luck)
		totalRolls := rolls + bonusRolls

		for i := 0; i < totalRolls; i++ {
			if entry := r.rollPool(pool); entry != nil {
				quantity := r.rng.Intn(entry.MaxQuantity-entry.MinQuantity+1) + entry.MinQuantity
				results = append(results, LootResult{
					ItemID:   entry.ItemID,
					Quantity: quantity,
				})
			}
		}
	}

	return results
}

func (r *LootTableRegistry) rollRange(rr RollRange) int {
	if rr.Max <= rr.Min {
		return rr.Min
	}
	return r.rng.Intn(rr.Max-rr.Min+1) + rr.Min
}

func (r *LootTableRegistry) rollPool(pool LootPool) *LootEntry {
	if len(pool.Entries) == 0 {
		return nil
	}

	totalWeight := 0
	for _, entry := range pool.Entries {
		totalWeight += entry.Weight
	}

	if totalWeight <= 0 {
		return nil
	}

	roll := r.rng.Intn(totalWeight)
	cumulative := 0
	for i := range pool.Entries {
		cumulative += pool.Entries[i].Weight
		if roll < cumulative {
			return &pool.Entries[i]
		}
	}

	return &pool.Entries[len(pool.Entries)-1]
}

// BiomeLootTable maps biomes to their loot table IDs
type BiomeLootTable struct {
	BiomeType     BiomeType `json:"biome_type"`
	ResourceTable string    `json:"resource_table"`
	ChestTable    string    `json:"chest_table"`
}

// GetBiomeLootTables returns the loot table mappings for all biomes
func GetBiomeLootTables() map[BiomeType]BiomeLootTable {
	return map[BiomeType]BiomeLootTable{
		BiomeForest:   {BiomeType: BiomeForest, ResourceTable: "loot_forest_resource", ChestTable: "loot_forest_chest"},
		BiomeDesert:   {BiomeType: BiomeDesert, ResourceTable: "loot_desert_resource", ChestTable: "loot_desert_chest"},
		BiomeVolcanic: {BiomeType: BiomeVolcanic, ResourceTable: "loot_volcanic_resource", ChestTable: "loot_volcanic_chest"},
		BiomeIce:      {BiomeType: BiomeIce, ResourceTable: "loot_ice_resource", ChestTable: "loot_ice_chest"},
		BiomeSavanna:  {BiomeType: BiomeSavanna, ResourceTable: "loot_savanna_resource", ChestTable: "loot_savanna_chest"},
		BiomeBadlands: {BiomeType: BiomeBadlands, ResourceTable: "loot_badlands_resource", ChestTable: "loot_badlands_chest"},
		BiomeSwamp:    {BiomeType: BiomeSwamp, ResourceTable: "loot_swamp_resource", ChestTable: "loot_swamp_chest"},
		BiomeCrystal:  {BiomeType: BiomeCrystal, ResourceTable: "loot_crystal_resource", ChestTable: "loot_crystal_chest"},
		BiomeVoid:     {BiomeType: BiomeVoid, ResourceTable: "loot_void_resource", ChestTable: "loot_void_chest"},
		BiomeNeon:     {BiomeType: BiomeNeon, ResourceTable: "loot_neon_resource", ChestTable: "loot_neon_chest"},
		BiomePlasma:   {BiomeType: BiomePlasma, ResourceTable: "loot_plasma_resource", ChestTable: "loot_plasma_chest"},
		BiomeAncient:  {BiomeType: BiomeAncient, ResourceTable: "loot_ancient_resource", ChestTable: "loot_ancient_chest"},
	}
}

// registerDefaultTables adds all default loot tables
func (r *LootTableRegistry) registerDefaultTables() {
	// === FOREST BIOME ===
	r.Register(&LootTable{
		ID:          "loot_forest_resource",
		Type:        "resource",
		Description: "Resources found in Enchanted Forest",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "wood", Weight: 40, MinQuantity: 3, MaxQuantity: 6, Rarity: RarityCommon},
				{ItemID: "herb", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityCommon},
				{ItemID: "ancient_wood", Weight: 20, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityUncommon},
				{ItemID: "spirit_essence", Weight: 10, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityRare},
			},
		}},
	})

	// === DESERT BIOME ===
	r.Register(&LootTable{
		ID:          "loot_desert_resource",
		Type:        "resource",
		Description: "Resources found in Scorching Dunes",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 1},
			Entries: []LootEntry{
				{ItemID: "stone", Weight: 35, MinQuantity: 2, MaxQuantity: 5, Rarity: RarityCommon},
				{ItemID: "sand_glass", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "cactus_fruit", Weight: 20, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityUncommon},
				{ItemID: "sun_crystal", Weight: 15, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityRare},
			},
		}},
	})

	// === VOLCANIC BIOME ===
	r.Register(&LootTable{
		ID:          "loot_volcanic_resource",
		Type:        "resource",
		Description: "Resources found in Infernal Caldera",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "obsite", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "ember_ore", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "fire_crystal", Weight: 25, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityRare},
				{ItemID: "magma_core", Weight: 10, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				{ItemID: "phoenix_feather", Weight: 5, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityLegendary},
			},
		}},
	})

	// === ICE BIOME ===
	r.Register(&LootTable{
		ID:          "loot_ice_resource",
		Type:        "resource",
		Description: "Resources found in Frostbound Wastes",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 1},
			Entries: []LootEntry{
				{ItemID: "frost_stone", Weight: 35, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "ice_crystal", Weight: 30, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityRare},
				{ItemID: "frozen_herb", Weight: 20, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityUncommon},
				{ItemID: "permafrost_core", Weight: 10, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				{ItemID: "temporal_ice", Weight: 5, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityLegendary},
			},
		}},
	})

	// === SAVANNA BIOME ===
	r.Register(&LootTable{
		ID:          "loot_savanna_resource",
		Type:        "resource",
		Description: "Resources found in Mystic Savanna",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "herb", Weight: 35, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityCommon},
				{ItemID: "stone", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityCommon},
				{ItemID: "wood", Weight: 20, MinQuantity: 2, MaxQuantity: 3, Rarity: RarityCommon},
				{ItemID: "ley_essence", Weight: 15, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityRare},
			},
		}},
	})

	// === BADLANDS BIOME ===
	r.Register(&LootTable{
		ID:          "loot_badlands_resource",
		Type:        "resource",
		Description: "Resources found in Ember Badlands",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "stone", Weight: 30, MinQuantity: 3, MaxQuantity: 5, Rarity: RarityCommon},
				{ItemID: "copper_ore", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "iron_ore", Weight: 25, MinQuantity: 2, MaxQuantity: 3, Rarity: RarityUncommon},
				{ItemID: "gold_nugget", Weight: 15, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityRare},
			},
		}},
	})

	// === SWAMP BIOME ===
	r.Register(&LootTable{
		ID:          "loot_swamp_resource",
		Type:        "resource",
		Description: "Resources found in Shadowmire",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "shadow_moss", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "dark_herb", Weight: 30, MinQuantity: 2, MaxQuantity: 3, Rarity: RarityUncommon},
				{ItemID: "nightmare_essence", Weight: 25, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityRare},
				{ItemID: "umbral_crystal", Weight: 10, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				{ItemID: "shadow_heart", Weight: 5, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityLegendary},
			},
		}},
	})

	// === CRYSTAL BIOME ===
	r.Register(&LootTable{
		ID:          "loot_crystal_resource",
		Type:        "resource",
		Description: "Resources found in Crystal Spires",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "crystal", Weight: 35, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityRare},
				{ItemID: "resonant_crystal", Weight: 25, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityEpic},
				{ItemID: "crystal_shard", Weight: 25, MinQuantity: 3, MaxQuantity: 5, Rarity: RarityUncommon},
				{ItemID: "harmonic_gem", Weight: 10, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityLegendary},
				{ItemID: "void_crystal", Weight: 5, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityMythic},
			},
		}},
	})

	// === VOID BIOME ===
	r.Register(&LootTable{
		ID:          "loot_void_resource",
		Type:        "resource",
		Description: "Resources found in Void Rift",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "void_shard", Weight: 30, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityRare},
				{ItemID: "null_stone", Weight: 25, MinQuantity: 2, MaxQuantity: 3, Rarity: RarityUncommon},
				{ItemID: "reality_fragment", Weight: 20, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityEpic},
				{ItemID: "entropy_essence", Weight: 15, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				{ItemID: "void_heart", Weight: 10, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityMythic},
			},
		}},
	})

	// === NEON BIOME ===
	r.Register(&LootTable{
		ID:          "loot_neon_resource",
		Type:        "resource",
		Description: "Resources found in Neon Wilds",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "bio_luminescent", Weight: 35, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "neon_sap", Weight: 25, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "alien_spore", Weight: 20, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityRare},
				{ItemID: "xenoflora_core", Weight: 15, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				{ItemID: "primal_genome", Weight: 5, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityLegendary},
			},
		}},
	})

	// === PLASMA BIOME ===
	r.Register(&LootTable{
		ID:          "loot_plasma_resource",
		Type:        "resource",
		Description: "Resources found in Plasma Fields",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "plasma_cell", Weight: 35, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "energy_crystal", Weight: 25, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityRare},
				{ItemID: "conduit_shard", Weight: 20, MinQuantity: 2, MaxQuantity: 3, Rarity: RarityUncommon},
				{ItemID: "fusion_core", Weight: 15, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				{ItemID: "antimatter_cell", Weight: 5, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityLegendary},
			},
		}},
	})

	// === ANCIENT BIOME ===
	r.Register(&LootTable{
		ID:          "loot_ancient_resource",
		Type:        "resource",
		Description: "Resources found in Ancient Ruins",
		Pools: []LootPool{{
			Name: "main", Rolls: RollRange{Min: 1, Max: 2},
			Entries: []LootEntry{
				{ItemID: "rune_stone", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "arcane_dust", Weight: 30, MinQuantity: 2, MaxQuantity: 4, Rarity: RarityUncommon},
				{ItemID: "glyph_fragment", Weight: 20, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityRare},
				{ItemID: "ancient_rune", Weight: 15, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				{ItemID: "primordial_sigil", Weight: 5, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityLegendary},
			},
		}},
	})

	// === CHEST LOOT TABLES ===
	r.registerChestLootTables()
}

func (r *LootTableRegistry) registerChestLootTables() {
	// Generic chest for all biomes
	r.Register(&LootTable{
		ID:          "loot_generic_chest",
		Type:        "chest",
		Description: "Common cache contents",
		Pools: []LootPool{
			{
				Name: "resources", Rolls: RollRange{Min: 2, Max: 4},
				Entries: []LootEntry{
					{ItemID: "wood", Weight: 25, MinQuantity: 3, MaxQuantity: 8, Rarity: RarityCommon},
					{ItemID: "stone", Weight: 25, MinQuantity: 3, MaxQuantity: 8, Rarity: RarityCommon},
					{ItemID: "herb", Weight: 25, MinQuantity: 2, MaxQuantity: 5, Rarity: RarityCommon},
					{ItemID: "crystal", Weight: 15, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityRare},
					{ItemID: "ancient_rune", Weight: 10, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityEpic},
				},
			},
			{
				Name: "consumables", Rolls: RollRange{Min: 0, Max: 2},
				Entries: []LootEntry{
					{ItemID: "health_potion", Weight: 40, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityCommon},
					{ItemID: "energy_potion", Weight: 40, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityCommon},
					{ItemID: "speed_potion", Weight: 20, MinQuantity: 1, MaxQuantity: 1, Rarity: RarityRare},
				},
			},
		},
	})

	// Biome-specific chests
	biomes := []BiomeType{
		BiomeForest, BiomeDesert, BiomeVolcanic, BiomeIce,
		BiomeSavanna, BiomeBadlands, BiomeSwamp, BiomeCrystal,
		BiomeVoid, BiomeNeon, BiomePlasma, BiomeAncient,
	}

	for _, biome := range biomes {
		tableID := "loot_" + string(biome) + "_chest"
		r.Register(&LootTable{
			ID:          tableID,
			Type:        "chest",
			Description: "Chest contents for " + string(biome),
			Pools: []LootPool{{
				Name: "base", Rolls: RollRange{Min: 2, Max: 3},
				Entries: []LootEntry{
					{ItemID: "crystal", Weight: 30, MinQuantity: 1, MaxQuantity: 3, Rarity: RarityRare},
					{ItemID: "health_potion", Weight: 35, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityCommon},
					{ItemID: "energy_potion", Weight: 35, MinQuantity: 1, MaxQuantity: 2, Rarity: RarityCommon},
				},
			}},
		})
	}
}

// LoadLootTables loads loot tables from a JSON file
func LoadLootTables(path string) ([]*LootTable, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tables struct {
		LootTables []*LootTable `json:"loot_tables"`
	}
	if err := json.Unmarshal(data, &tables); err != nil {
		return nil, err
	}

	return tables.LootTables, nil
}

// SaveLootTables saves loot tables to a JSON file
func SaveLootTables(tables []*LootTable, path string) error {
	data := struct {
		LootTables []*LootTable `json:"loot_tables"`
	}{
		LootTables: tables,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}
