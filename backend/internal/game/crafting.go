package game

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Recipe represents a crafting recipe
type Recipe struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Result      RecipeResult   `json:"result"`
	Ingredients map[string]int `json:"ingredients"`
	EnergyCost  int            `json:"energy_cost"`
}

// RecipeResult defines what a recipe produces
type RecipeResult struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// RecipeRegistry holds all crafting recipes
type RecipeRegistry struct {
	mu      sync.RWMutex
	recipes map[string]*Recipe
}

// NewRecipeRegistry creates a new recipe registry
func NewRecipeRegistry() *RecipeRegistry {
	return &RecipeRegistry{
		recipes: make(map[string]*Recipe),
	}
}

// Register adds a recipe to the registry
func (r *RecipeRegistry) Register(recipe *Recipe) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recipes[recipe.ID] = recipe
}

// Get retrieves a recipe by ID
func (r *RecipeRegistry) Get(id string) *Recipe {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.recipes[id]
}

// GetAll returns all recipes
func (r *RecipeRegistry) GetAll() []*Recipe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	recipes := make([]*Recipe, 0, len(r.recipes))
	for _, recipe := range r.recipes {
		recipes = append(recipes, recipe)
	}
	return recipes
}

// LoadFromFile loads recipes from a JSON file
func (r *RecipeRegistry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read recipes file: %w", err)
	}

	var fileData struct {
		Recipes []*Recipe `json:"recipes"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to parse recipes file: %w", err)
	}

	for _, recipe := range fileData.Recipes {
		if recipe.Result.Quantity == 0 {
			recipe.Result.Quantity = 1
		}
		r.Register(recipe)
	}

	return nil
}

// LoadFromJSON loads recipes from a JSON byte slice
func (r *RecipeRegistry) LoadFromJSON(data []byte) error {
	var fileData struct {
		Recipes []*Recipe `json:"recipes"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to parse recipes JSON: %w", err)
	}

	for _, recipe := range fileData.Recipes {
		if recipe.Result.Quantity == 0 {
			recipe.Result.Quantity = 1
		}
		r.Register(recipe)
	}

	return nil
}

// CanCraft checks if an inventory has all ingredients for a recipe
func (r *RecipeRegistry) CanCraft(recipeID string, inv *Inventory, energy int) (bool, string) {
	recipe := r.Get(recipeID)
	if recipe == nil {
		return false, "recipe not found"
	}

	if energy < recipe.EnergyCost {
		return false, "not enough energy"
	}

	for itemID, qty := range recipe.Ingredients {
		if !inv.HasItems(itemID, qty) {
			return false, fmt.Sprintf("missing %d %s", qty, itemID)
		}
	}

	return true, ""
}

// GetAvailableRecipes returns recipes that can be crafted with current inventory
func (r *RecipeRegistry) GetAvailableRecipes(inv *Inventory, energy int) []*Recipe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var available []*Recipe
	for _, recipe := range r.recipes {
		if can, _ := r.CanCraft(recipe.ID, inv, energy); can {
			available = append(available, recipe)
		}
	}
	return available
}

// DefaultRecipeRegistry creates a registry with default recipes
func DefaultRecipeRegistry() *RecipeRegistry {
	r := NewRecipeRegistry()

	// Structure recipes
	r.Register(&Recipe{
		ID:   "craft_wall",
		Name: "Craft Stone Wall",
		Result: RecipeResult{
			ItemID:   "wall",
			Quantity: 1,
		},
		Ingredients: map[string]int{
			"stone": 3,
		},
		EnergyCost: 2,
	})

	r.Register(&Recipe{
		ID:   "craft_trap",
		Name: "Craft Hidden Trap",
		Result: RecipeResult{
			ItemID:   "trap",
			Quantity: 1,
		},
		Ingredients: map[string]int{
			"wood":  2,
			"stone": 1,
		},
		EnergyCost: 2,
	})

	r.Register(&Recipe{
		ID:   "craft_beacon",
		Name: "Craft Vision Beacon",
		Result: RecipeResult{
			ItemID:   "beacon",
			Quantity: 1,
		},
		Ingredients: map[string]int{
			"stone":   2,
			"crystal": 1,
		},
		EnergyCost: 3,
	})

	// Consumable recipes
	r.Register(&Recipe{
		ID:   "craft_health_potion",
		Name: "Craft Health Potion",
		Result: RecipeResult{
			ItemID:   "health_potion",
			Quantity: 1,
		},
		Ingredients: map[string]int{
			"herb": 2,
		},
		EnergyCost: 1,
	})

	r.Register(&Recipe{
		ID:   "craft_energy_potion",
		Name: "Craft Energy Potion",
		Result: RecipeResult{
			ItemID:   "energy_potion",
			Quantity: 1,
		},
		Ingredients: map[string]int{
			"herb":    1,
			"crystal": 1,
		},
		EnergyCost: 2,
	})

	return r
}

// RecipeSnapshot is a serializable representation of a recipe
type RecipeSnapshot struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	ResultItem  string         `json:"result_item"`
	ResultQty   int            `json:"result_qty"`
	Ingredients map[string]int `json:"ingredients"`
	EnergyCost  int            `json:"energy_cost"`
	CanCraft    bool           `json:"can_craft"`
}

// Snapshot creates a serializable copy with craftability status
func (recipe *Recipe) Snapshot(inv *Inventory, energy int, registry *RecipeRegistry) RecipeSnapshot {
	canCraft := false
	if registry != nil {
		canCraft, _ = registry.CanCraft(recipe.ID, inv, energy)
	}

	return RecipeSnapshot{
		ID:          recipe.ID,
		Name:        recipe.Name,
		Description: recipe.Description,
		ResultItem:  recipe.Result.ItemID,
		ResultQty:   recipe.Result.Quantity,
		Ingredients: recipe.Ingredients,
		EnergyCost:  recipe.EnergyCost,
		CanCraft:    canCraft,
	}
}

// SnapshotAll creates snapshots of all recipes
func (r *RecipeRegistry) SnapshotAll(inv *Inventory, energy int) []RecipeSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshots := make([]RecipeSnapshot, 0, len(r.recipes))
	for _, recipe := range r.recipes {
		snapshots = append(snapshots, recipe.Snapshot(inv, energy, r))
	}
	return snapshots
}
