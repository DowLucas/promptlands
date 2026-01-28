package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

const (
	DefaultInventorySlots = 10
	MaxInventorySlots     = 30
)

// InventorySlot represents a single slot in the inventory
type InventorySlot struct {
	Item *ItemInstance `json:"item,omitempty"`
}

// EquipmentSlot represents an equipment slot type
type EquipmentSlot string

const (
	SlotWeapon  EquipmentSlot = "weapon"
	SlotArmor   EquipmentSlot = "armor"
	SlotTrinket EquipmentSlot = "trinket"
)

// Inventory represents an agent's inventory
type Inventory struct {
	mu       sync.RWMutex
	OwnerID  uuid.UUID             `json:"owner_id"`
	Slots    []*InventorySlot      `json:"slots"`
	MaxSlots int                   `json:"max_slots"`
	Weapon   *ItemInstance         `json:"weapon,omitempty"`
	Armor    *ItemInstance         `json:"armor,omitempty"`
	Trinket  *ItemInstance         `json:"trinket,omitempty"`
	registry *ItemRegistry
}

// NewInventory creates a new inventory for an agent
func NewInventory(ownerID uuid.UUID, maxSlots int, registry *ItemRegistry) *Inventory {
	if maxSlots < 1 {
		maxSlots = DefaultInventorySlots
	}
	if maxSlots > MaxInventorySlots {
		maxSlots = MaxInventorySlots
	}

	slots := make([]*InventorySlot, maxSlots)
	for i := range slots {
		slots[i] = &InventorySlot{}
	}

	return &Inventory{
		OwnerID:  ownerID,
		Slots:    slots,
		MaxSlots: maxSlots,
		registry: registry,
	}
}

// SetRegistry sets the item registry for the inventory
func (inv *Inventory) SetRegistry(registry *ItemRegistry) {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	inv.registry = registry
}

// AddItem adds an item to the inventory, stacking if possible
// Returns the quantity that couldn't be added (0 if all added)
func (inv *Inventory) AddItem(defID string, quantity int) int {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if inv.registry == nil || quantity <= 0 {
		return quantity
	}

	def := inv.registry.Get(defID)
	if def == nil {
		return quantity
	}

	remaining := quantity

	// First, try to stack with existing items
	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.DefinitionID == defID {
			canAdd := def.MaxStack - slot.Item.Quantity
			if canAdd > 0 {
				toAdd := min(canAdd, remaining)
				slot.Item.Quantity += toAdd
				remaining -= toAdd
				if remaining == 0 {
					return 0
				}
			}
		}
	}

	// Then, find empty slots for remaining items
	for _, slot := range inv.Slots {
		if slot.Item == nil {
			toAdd := min(def.MaxStack, remaining)
			slot.Item = NewItemInstance(defID, toAdd)
			remaining -= toAdd
			if remaining == 0 {
				return 0
			}
		}
	}

	return remaining
}

// AddItemInstance adds a specific item instance to the inventory
// Returns true if added successfully
func (inv *Inventory) AddItemInstance(item *ItemInstance) bool {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if item == nil {
		return false
	}

	// Find empty slot
	for _, slot := range inv.Slots {
		if slot.Item == nil {
			slot.Item = item
			return true
		}
	}

	return false
}

// RemoveItem removes a quantity of items by definition ID
// Returns the quantity actually removed
func (inv *Inventory) RemoveItem(defID string, quantity int) int {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	removed := 0

	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.DefinitionID == defID {
			toRemove := min(slot.Item.Quantity, quantity-removed)
			slot.Item.Quantity -= toRemove
			removed += toRemove

			if slot.Item.Quantity <= 0 {
				slot.Item = nil
			}

			if removed >= quantity {
				return removed
			}
		}
	}

	return removed
}

// RemoveItemByID removes an item by its instance ID
// Returns the removed item or nil
func (inv *Inventory) RemoveItemByID(itemID uuid.UUID) *ItemInstance {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.ID == itemID {
			item := slot.Item
			slot.Item = nil
			return item
		}
	}

	return nil
}

// RemoveFromSlot removes the item from a specific slot index
// Returns the removed item or nil
func (inv *Inventory) RemoveFromSlot(slotIndex int) *ItemInstance {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if slotIndex < 0 || slotIndex >= len(inv.Slots) {
		return nil
	}

	item := inv.Slots[slotIndex].Item
	inv.Slots[slotIndex].Item = nil
	return item
}

// GetItemCount returns the total count of a specific item
func (inv *Inventory) GetItemCount(defID string) int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	count := 0
	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.DefinitionID == defID {
			count += slot.Item.Quantity
		}
	}
	return count
}

// HasItems checks if the inventory has at least the specified quantity of an item
func (inv *Inventory) HasItems(defID string, quantity int) bool {
	return inv.GetItemCount(defID) >= quantity
}

// HasAllItems checks if the inventory has all items in the map
func (inv *Inventory) HasAllItems(items map[string]int) bool {
	for defID, qty := range items {
		if !inv.HasItems(defID, qty) {
			return false
		}
	}
	return true
}

// GetItemByID finds an item by its instance ID
func (inv *Inventory) GetItemByID(itemID uuid.UUID) *ItemInstance {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.ID == itemID {
			return slot.Item
		}
	}
	return nil
}

// GetItemByDefID finds the first item matching the definition ID
func (inv *Inventory) GetItemByDefID(defID string) *ItemInstance {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for _, slot := range inv.Slots {
		if slot.Item != nil && slot.Item.DefinitionID == defID {
			return slot.Item
		}
	}
	return nil
}

// GetSlotItem returns the item at a specific slot index
func (inv *Inventory) GetSlotItem(slotIndex int) *ItemInstance {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	if slotIndex < 0 || slotIndex >= len(inv.Slots) {
		return nil
	}
	return inv.Slots[slotIndex].Item
}

// IsFull checks if the inventory has no empty slots
func (inv *Inventory) IsFull() bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for _, slot := range inv.Slots {
		if slot.Item == nil {
			return false
		}
	}
	return true
}

// IsEmpty checks if the inventory has no items
func (inv *Inventory) IsEmpty() bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for _, slot := range inv.Slots {
		if slot.Item != nil {
			return false
		}
	}
	return true
}

// EmptySlotCount returns the number of empty slots
func (inv *Inventory) EmptySlotCount() int {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	count := 0
	for _, slot := range inv.Slots {
		if slot.Item == nil {
			count++
		}
	}
	return count
}

// ExpandSlots adds more slots to the inventory
func (inv *Inventory) ExpandSlots(additionalSlots int) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	newMax := inv.MaxSlots + additionalSlots
	if newMax > MaxInventorySlots {
		newMax = MaxInventorySlots
	}

	for i := inv.MaxSlots; i < newMax; i++ {
		inv.Slots = append(inv.Slots, &InventorySlot{})
	}
	inv.MaxSlots = newMax
}

// Equip equips an item to the specified slot
func (inv *Inventory) Equip(itemID uuid.UUID, slot EquipmentSlot) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	// Find the item in inventory
	var itemSlotIndex int = -1
	var item *ItemInstance
	for i, s := range inv.Slots {
		if s.Item != nil && s.Item.ID == itemID {
			itemSlotIndex = i
			item = s.Item
			break
		}
	}

	if item == nil {
		return fmt.Errorf("item not found in inventory")
	}

	// Check if item is equippable
	if inv.registry != nil {
		def := inv.registry.Get(item.DefinitionID)
		if def == nil || !def.Equippable {
			return fmt.Errorf("item cannot be equipped")
		}
	}

	// Unequip current item in that slot and swap
	var currentEquipped *ItemInstance
	switch slot {
	case SlotWeapon:
		currentEquipped = inv.Weapon
		inv.Weapon = item
	case SlotArmor:
		currentEquipped = inv.Armor
		inv.Armor = item
	case SlotTrinket:
		currentEquipped = inv.Trinket
		inv.Trinket = item
	default:
		return fmt.Errorf("invalid equipment slot")
	}

	// Put the previously equipped item in the freed inventory slot
	inv.Slots[itemSlotIndex].Item = currentEquipped

	return nil
}

// Unequip removes an item from an equipment slot and puts it in inventory
func (inv *Inventory) Unequip(slot EquipmentSlot) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	var item *ItemInstance
	switch slot {
	case SlotWeapon:
		item = inv.Weapon
		inv.Weapon = nil
	case SlotArmor:
		item = inv.Armor
		inv.Armor = nil
	case SlotTrinket:
		item = inv.Trinket
		inv.Trinket = nil
	default:
		return fmt.Errorf("invalid equipment slot")
	}

	if item == nil {
		return fmt.Errorf("no item equipped in that slot")
	}

	// Find empty slot
	for _, s := range inv.Slots {
		if s.Item == nil {
			s.Item = item
			return nil
		}
	}

	// Re-equip if no space
	switch slot {
	case SlotWeapon:
		inv.Weapon = item
	case SlotArmor:
		inv.Armor = item
	case SlotTrinket:
		inv.Trinket = item
	}

	return fmt.Errorf("no empty inventory slots")
}

// GetEquipped returns the item in an equipment slot
func (inv *Inventory) GetEquipped(slot EquipmentSlot) *ItemInstance {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	switch slot {
	case SlotWeapon:
		return inv.Weapon
	case SlotArmor:
		return inv.Armor
	case SlotTrinket:
		return inv.Trinket
	default:
		return nil
	}
}

// Clear removes all items from the inventory
func (inv *Inventory) Clear() {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	for _, slot := range inv.Slots {
		slot.Item = nil
	}
	inv.Weapon = nil
	inv.Armor = nil
	inv.Trinket = nil
}

// GetAllItems returns a list of all items in the inventory
func (inv *Inventory) GetAllItems() []*ItemInstance {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	items := make([]*ItemInstance, 0)
	for _, slot := range inv.Slots {
		if slot.Item != nil {
			items = append(items, slot.Item)
		}
	}
	return items
}

// InventorySnapshot is a serializable representation of the inventory
type InventorySnapshot struct {
	OwnerID  uuid.UUID              `json:"owner_id"`
	Slots    []InventorySlotSnapshot `json:"slots"`
	MaxSlots int                    `json:"max_slots"`
	Weapon   *ItemInstanceSnapshot  `json:"weapon,omitempty"`
	Armor    *ItemInstanceSnapshot  `json:"armor,omitempty"`
	Trinket  *ItemInstanceSnapshot  `json:"trinket,omitempty"`
}

// InventorySlotSnapshot is a serializable representation of an inventory slot
type InventorySlotSnapshot struct {
	Item *ItemInstanceSnapshot `json:"item,omitempty"`
}

// Snapshot creates a serializable copy of the inventory
func (inv *Inventory) Snapshot() InventorySnapshot {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	slots := make([]InventorySlotSnapshot, len(inv.Slots))
	for i, slot := range inv.Slots {
		if slot.Item != nil {
			snap := slot.Item.Snapshot()
			slots[i] = InventorySlotSnapshot{Item: &snap}
		} else {
			slots[i] = InventorySlotSnapshot{}
		}
	}

	snapshot := InventorySnapshot{
		OwnerID:  inv.OwnerID,
		Slots:    slots,
		MaxSlots: inv.MaxSlots,
	}

	if inv.Weapon != nil {
		snap := inv.Weapon.Snapshot()
		snapshot.Weapon = &snap
	}
	if inv.Armor != nil {
		snap := inv.Armor.Snapshot()
		snapshot.Armor = &snap
	}
	if inv.Trinket != nil {
		snap := inv.Trinket.Snapshot()
		snapshot.Trinket = &snap
	}

	return snapshot
}

// ItemSummary represents a summarized view of items for display
type ItemSummary struct {
	DefinitionID string `json:"definition_id"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
}

// GetItemSummary returns a summarized list of items with totals
func (inv *Inventory) GetItemSummary() []ItemSummary {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	counts := make(map[string]int)
	for _, slot := range inv.Slots {
		if slot.Item != nil {
			counts[slot.Item.DefinitionID] += slot.Item.Quantity
		}
	}

	summaries := make([]ItemSummary, 0, len(counts))
	for defID, qty := range counts {
		name := defID
		if inv.registry != nil {
			if def := inv.registry.Get(defID); def != nil {
				name = def.Name
			}
		}
		summaries = append(summaries, ItemSummary{
			DefinitionID: defID,
			Name:         name,
			Quantity:     qty,
		})
	}

	return summaries
}
