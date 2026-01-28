package game

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

// ItemCategory represents the category of an item
type ItemCategory string

const (
	CategoryTool       ItemCategory = "tool"
	CategoryConsumable ItemCategory = "consumable"
	CategoryMaterial   ItemCategory = "material"
	CategoryPlaceable  ItemCategory = "placeable"
	CategoryEquipment  ItemCategory = "equipment"
)

// ItemRarity represents the rarity level of an item
type ItemRarity string

const (
	RarityCommon    ItemRarity = "common"
	RarityUncommon  ItemRarity = "uncommon"
	RarityRare      ItemRarity = "rare"
	RarityEpic      ItemRarity = "epic"
	RarityLegendary ItemRarity = "legendary"
)

// ItemDefinition defines a template for items loaded from JSON
type ItemDefinition struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    ItemCategory      `json:"category"`
	Rarity      ItemRarity        `json:"rarity"`
	MaxStack    int               `json:"max_stack"`
	EnergyCost  int               `json:"energy_cost"`
	CraftCost   map[string]int    `json:"craft_cost,omitempty"`
	Properties  map[string]any    `json:"properties,omitempty"`
	Consumable  bool              `json:"consumable"`
	Placeable   bool              `json:"placeable"`
	Usable      bool              `json:"usable"`
	Equippable  bool              `json:"equippable"`
}

// ItemInstance represents an actual item in inventory or world
type ItemInstance struct {
	ID           uuid.UUID      `json:"id"`
	DefinitionID string         `json:"definition_id"`
	Quantity     int            `json:"quantity"`
	Durability   *int           `json:"durability,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// NewItemInstance creates a new item instance from a definition
func NewItemInstance(defID string, quantity int) *ItemInstance {
	return &ItemInstance{
		ID:           uuid.New(),
		DefinitionID: defID,
		Quantity:     quantity,
		Metadata:     make(map[string]any),
	}
}

// NewItemInstanceWithDurability creates an item instance with durability
func NewItemInstanceWithDurability(defID string, quantity int, durability int) *ItemInstance {
	item := NewItemInstance(defID, quantity)
	item.Durability = &durability
	return item
}

// Clone creates a copy of the item instance
func (i *ItemInstance) Clone() *ItemInstance {
	clone := &ItemInstance{
		ID:           uuid.New(),
		DefinitionID: i.DefinitionID,
		Quantity:     i.Quantity,
	}
	if i.Durability != nil {
		dur := *i.Durability
		clone.Durability = &dur
	}
	if i.Metadata != nil {
		clone.Metadata = make(map[string]any)
		for k, v := range i.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}

// ItemInstanceSnapshot is a serializable representation of an item instance
type ItemInstanceSnapshot struct {
	ID           uuid.UUID      `json:"id"`
	DefinitionID string         `json:"definition_id"`
	Quantity     int            `json:"quantity"`
	Durability   *int           `json:"durability,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// Snapshot creates a serializable copy
func (i *ItemInstance) Snapshot() ItemInstanceSnapshot {
	return ItemInstanceSnapshot{
		ID:           i.ID,
		DefinitionID: i.DefinitionID,
		Quantity:     i.Quantity,
		Durability:   i.Durability,
		Metadata:     i.Metadata,
	}
}

// ItemRegistry holds all item definitions
type ItemRegistry struct {
	mu    sync.RWMutex
	items map[string]*ItemDefinition
}

// NewItemRegistry creates a new item registry
func NewItemRegistry() *ItemRegistry {
	return &ItemRegistry{
		items: make(map[string]*ItemDefinition),
	}
}

// Register adds an item definition to the registry
func (r *ItemRegistry) Register(def *ItemDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[def.ID] = def
}

// Get retrieves an item definition by ID
func (r *ItemRegistry) Get(id string) *ItemDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.items[id]
}

// GetAll returns all item definitions
func (r *ItemRegistry) GetAll() []*ItemDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]*ItemDefinition, 0, len(r.items))
	for _, def := range r.items {
		defs = append(defs, def)
	}
	return defs
}

// LoadFromFile loads item definitions from a JSON file
func (r *ItemRegistry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read items file: %w", err)
	}

	var fileData struct {
		Items []*ItemDefinition `json:"items"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to parse items file: %w", err)
	}

	for _, def := range fileData.Items {
		// Set default values
		if def.MaxStack == 0 {
			def.MaxStack = 1
		}
		if def.Rarity == "" {
			def.Rarity = RarityCommon
		}
		r.Register(def)
	}

	return nil
}

// LoadFromJSON loads item definitions from a JSON byte slice
func (r *ItemRegistry) LoadFromJSON(data []byte) error {
	var fileData struct {
		Items []*ItemDefinition `json:"items"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to parse items JSON: %w", err)
	}

	for _, def := range fileData.Items {
		if def.MaxStack == 0 {
			def.MaxStack = 1
		}
		if def.Rarity == "" {
			def.Rarity = RarityCommon
		}
		r.Register(def)
	}

	return nil
}

// GetProperty returns a property value from an item definition
func (d *ItemDefinition) GetProperty(key string) (any, bool) {
	if d.Properties == nil {
		return nil, false
	}
	val, ok := d.Properties[key]
	return val, ok
}

// GetPropertyInt returns an integer property value
func (d *ItemDefinition) GetPropertyInt(key string, defaultVal int) int {
	val, ok := d.GetProperty(key)
	if !ok {
		return defaultVal
	}
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return defaultVal
	}
}

// GetPropertyBool returns a boolean property value
func (d *ItemDefinition) GetPropertyBool(key string, defaultVal bool) bool {
	val, ok := d.GetProperty(key)
	if !ok {
		return defaultVal
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultVal
}

// GetPropertyString returns a string property value
func (d *ItemDefinition) GetPropertyString(key string, defaultVal string) string {
	val, ok := d.GetProperty(key)
	if !ok {
		return defaultVal
	}
	if s, ok := val.(string); ok {
		return s
	}
	return defaultVal
}

// DefaultItemRegistry creates a registry with default items
func DefaultItemRegistry() *ItemRegistry {
	r := NewItemRegistry()

	// Materials
	r.Register(&ItemDefinition{
		ID:       "wood",
		Name:     "Wood",
		Category: CategoryMaterial,
		Rarity:   RarityCommon,
		MaxStack: 64,
	})
	r.Register(&ItemDefinition{
		ID:       "stone",
		Name:     "Stone",
		Category: CategoryMaterial,
		Rarity:   RarityCommon,
		MaxStack: 64,
	})
	r.Register(&ItemDefinition{
		ID:       "crystal",
		Name:     "Crystal",
		Category: CategoryMaterial,
		Rarity:   RarityRare,
		MaxStack: 32,
	})
	r.Register(&ItemDefinition{
		ID:       "herb",
		Name:     "Herb",
		Category: CategoryMaterial,
		Rarity:   RarityCommon,
		MaxStack: 32,
	})

	// Placeables
	r.Register(&ItemDefinition{
		ID:         "wall",
		Name:       "Stone Wall",
		Category:   CategoryPlaceable,
		Rarity:     RarityCommon,
		MaxStack:   10,
		EnergyCost: 5,
		Placeable:  true,
		Properties: map[string]any{
			"hp":              3,
			"blocks_movement": true,
		},
	})
	r.Register(&ItemDefinition{
		ID:         "beacon",
		Name:       "Vision Beacon",
		Category:   CategoryPlaceable,
		Rarity:     RarityUncommon,
		MaxStack:   5,
		EnergyCost: 8,
		Placeable:  true,
		Properties: map[string]any{
			"vision_bonus": 2,
		},
	})
	r.Register(&ItemDefinition{
		ID:         "trap",
		Name:       "Hidden Trap",
		Category:   CategoryPlaceable,
		Rarity:     RarityUncommon,
		MaxStack:   5,
		EnergyCost: 4,
		Placeable:  true,
		Properties: map[string]any{
			"damage": 1,
			"hidden": true,
		},
	})

	// Consumables
	r.Register(&ItemDefinition{
		ID:         "health_potion",
		Name:       "Health Potion",
		Category:   CategoryConsumable,
		Rarity:     RarityCommon,
		MaxStack:   5,
		Usable:     true,
		Consumable: true,
		Properties: map[string]any{
			"heal_amount": 1,
		},
	})
	r.Register(&ItemDefinition{
		ID:         "energy_potion",
		Name:       "Energy Potion",
		Category:   CategoryConsumable,
		Rarity:     RarityCommon,
		MaxStack:   5,
		Usable:     true,
		Consumable: true,
		Properties: map[string]any{
			"energy_amount": 20,
		},
	})

	return r
}
