package game

import (
	"sync"

	"github.com/google/uuid"
)

// WorldObjectType represents the type of world object
type WorldObjectType string

const (
	ObjectDroppedItem  WorldObjectType = "dropped_item"
	ObjectStructure    WorldObjectType = "structure"
	ObjectResource     WorldObjectType = "resource"
	ObjectInteractive  WorldObjectType = "interactive"
)

// StructureType represents types of buildable structures
type StructureType string

const (
	StructureWall   StructureType = "wall"
	StructureBeacon StructureType = "beacon"
	StructureTrap   StructureType = "trap"
)

// ResourceType represents types of harvestable resources
type ResourceType string

const (
	ResourceWood    ResourceType = "wood"
	ResourceStone   ResourceType = "stone"
	ResourceCrystal ResourceType = "crystal"
	ResourceHerb    ResourceType = "herb"
)

// InteractiveType represents types of interactive objects
type InteractiveType string

const (
	InteractiveShrine  InteractiveType = "shrine"
	InteractiveCache   InteractiveType = "cache"
	InteractivePortal  InteractiveType = "portal"
	InteractiveObelisk InteractiveType = "obelisk"
)

// WorldObject represents an object in the game world
type WorldObject struct {
	ID              uuid.UUID       `json:"id"`
	Type            WorldObjectType `json:"type"`
	Position        Position        `json:"position"`
	OwnerID         *uuid.UUID      `json:"owner_id,omitempty"`
	CreatedTick     int             `json:"created_tick"`

	// Structure fields
	StructureType   StructureType `json:"structure_type,omitempty"`
	HP              int           `json:"hp,omitempty"`
	MaxHP           int           `json:"max_hp,omitempty"`
	Hidden          bool          `json:"hidden,omitempty"`
	BlocksMovement  bool          `json:"blocks_movement,omitempty"`
	VisionBonus     int           `json:"vision_bonus,omitempty"`
	Damage          int           `json:"damage,omitempty"`

	// Resource fields
	ResourceType ResourceType `json:"resource_type,omitempty"`
	Remaining    int          `json:"remaining,omitempty"`

	// Interactive fields
	InteractiveType InteractiveType `json:"interactive_type,omitempty"`
	Destination     *Position       `json:"destination,omitempty"`
	Message         string          `json:"message,omitempty"`
	Activated       bool            `json:"activated,omitempty"`
	ActivatedBy     map[uuid.UUID]bool `json:"activated_by,omitempty"`
	EnergyReward    int             `json:"energy_reward,omitempty"`
	HPReward        int             `json:"hp_reward,omitempty"`

	// Dropped item fields
	Item *ItemInstance `json:"item,omitempty"`
	DespawnTick int    `json:"despawn_tick,omitempty"`
}

// NewStructure creates a new structure world object
func NewStructure(structureType StructureType, pos Position, ownerID uuid.UUID, tick int) *WorldObject {
	obj := &WorldObject{
		ID:            uuid.New(),
		Type:          ObjectStructure,
		Position:      pos,
		OwnerID:       &ownerID,
		CreatedTick:   tick,
		StructureType: structureType,
	}

	switch structureType {
	case StructureWall:
		obj.HP = 3
		obj.MaxHP = 3
		obj.BlocksMovement = true
	case StructureBeacon:
		obj.HP = 2
		obj.MaxHP = 2
		obj.VisionBonus = 2
	case StructureTrap:
		obj.HP = 1
		obj.MaxHP = 1
		obj.Hidden = true
		obj.Damage = 1
	}

	return obj
}

// NewResourceNode creates a new resource node
func NewResourceNode(resourceType ResourceType, pos Position, remaining int) *WorldObject {
	return &WorldObject{
		ID:           uuid.New(),
		Type:         ObjectResource,
		Position:     pos,
		ResourceType: resourceType,
		Remaining:    remaining,
	}
}

// NewDroppedItem creates a dropped item object
func NewDroppedItem(item *ItemInstance, pos Position, tick int, despawnAfter int) *WorldObject {
	return &WorldObject{
		ID:          uuid.New(),
		Type:        ObjectDroppedItem,
		Position:    pos,
		Item:        item,
		CreatedTick: tick,
		DespawnTick: tick + despawnAfter,
	}
}

// NewShrine creates a shrine interactive object
func NewShrine(pos Position) *WorldObject {
	return &WorldObject{
		ID:              uuid.New(),
		Type:            ObjectInteractive,
		Position:        pos,
		InteractiveType: InteractiveShrine,
		HPReward:        1,
		ActivatedBy:     make(map[uuid.UUID]bool),
	}
}

// NewCache creates a cache interactive object
func NewCache(pos Position, energy int) *WorldObject {
	return &WorldObject{
		ID:              uuid.New(),
		Type:            ObjectInteractive,
		Position:        pos,
		InteractiveType: InteractiveCache,
		EnergyReward:    energy,
	}
}

// NewPortal creates a portal interactive object
func NewPortal(pos Position, destination Position) *WorldObject {
	return &WorldObject{
		ID:              uuid.New(),
		Type:            ObjectInteractive,
		Position:        pos,
		InteractiveType: InteractivePortal,
		Destination:     &destination,
	}
}

// NewObelisk creates an obelisk interactive object
func NewObelisk(pos Position) *WorldObject {
	return &WorldObject{
		ID:              uuid.New(),
		Type:            ObjectInteractive,
		Position:        pos,
		InteractiveType: InteractiveObelisk,
		Message:         "",
	}
}

// TakeDamage applies damage to a structure and returns true if destroyed
func (o *WorldObject) TakeDamage(damage int) bool {
	if o.Type != ObjectStructure {
		return false
	}
	o.HP -= damage
	return o.HP <= 0
}

// Harvest removes resources from a resource node and returns the amount harvested
func (o *WorldObject) Harvest(amount int) int {
	if o.Type != ObjectResource || o.Remaining <= 0 {
		return 0
	}
	harvested := min(amount, o.Remaining)
	o.Remaining -= harvested
	return harvested
}

// CanBeActivatedBy checks if an agent can activate this interactive object
func (o *WorldObject) CanBeActivatedBy(agentID uuid.UUID) bool {
	if o.Type != ObjectInteractive {
		return false
	}

	switch o.InteractiveType {
	case InteractiveShrine:
		// Shrines can only be activated once per agent
		return !o.ActivatedBy[agentID]
	case InteractiveCache:
		// Caches can only be activated once (by anyone)
		return !o.Activated
	case InteractivePortal, InteractiveObelisk:
		// Portals and obelisks can be used multiple times
		return true
	default:
		return false
	}
}

// Activate marks the object as activated by an agent
func (o *WorldObject) Activate(agentID uuid.UUID) {
	if o.Type != ObjectInteractive {
		return
	}

	switch o.InteractiveType {
	case InteractiveShrine:
		if o.ActivatedBy == nil {
			o.ActivatedBy = make(map[uuid.UUID]bool)
		}
		o.ActivatedBy[agentID] = true
	case InteractiveCache:
		o.Activated = true
	}
}

// WorldObjectSnapshot is a serializable representation of a world object
type WorldObjectSnapshot struct {
	ID              uuid.UUID       `json:"id"`
	Type            WorldObjectType `json:"type"`
	Position        Position        `json:"position"`
	OwnerID         *uuid.UUID      `json:"owner_id,omitempty"`

	// Structure fields
	StructureType   StructureType `json:"structure_type,omitempty"`
	HP              int           `json:"hp,omitempty"`
	MaxHP           int           `json:"max_hp,omitempty"`
	Hidden          bool          `json:"hidden,omitempty"`
	BlocksMovement  bool          `json:"blocks_movement,omitempty"`

	// Resource fields
	ResourceType ResourceType `json:"resource_type,omitempty"`
	Remaining    int          `json:"remaining,omitempty"`

	// Interactive fields
	InteractiveType InteractiveType `json:"interactive_type,omitempty"`
	Message         string          `json:"message,omitempty"`
	Activated       bool            `json:"activated,omitempty"`

	// Dropped item fields
	Item *ItemInstanceSnapshot `json:"item,omitempty"`
}

// Snapshot creates a serializable copy of the world object
func (o *WorldObject) Snapshot() WorldObjectSnapshot {
	snap := WorldObjectSnapshot{
		ID:              o.ID,
		Type:            o.Type,
		Position:        o.Position,
		OwnerID:         o.OwnerID,
		StructureType:   o.StructureType,
		HP:              o.HP,
		MaxHP:           o.MaxHP,
		Hidden:          o.Hidden,
		BlocksMovement:  o.BlocksMovement,
		ResourceType:    o.ResourceType,
		Remaining:       o.Remaining,
		InteractiveType: o.InteractiveType,
		Message:         o.Message,
		Activated:       o.Activated,
	}

	if o.Item != nil {
		itemSnap := o.Item.Snapshot()
		snap.Item = &itemSnap
	}

	return snap
}

// WorldObjectManager manages all world objects
type WorldObjectManager struct {
	mu      sync.RWMutex
	objects map[uuid.UUID]*WorldObject
	byPos   map[Position][]*WorldObject
}

// NewWorldObjectManager creates a new world object manager
func NewWorldObjectManager() *WorldObjectManager {
	return &WorldObjectManager{
		objects: make(map[uuid.UUID]*WorldObject),
		byPos:   make(map[Position][]*WorldObject),
	}
}

// Add adds a world object to the manager
func (m *WorldObjectManager) Add(obj *WorldObject) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.objects[obj.ID] = obj
	m.byPos[obj.Position] = append(m.byPos[obj.Position], obj)
}

// Remove removes a world object from the manager
func (m *WorldObjectManager) Remove(id uuid.UUID) *WorldObject {
	m.mu.Lock()
	defer m.mu.Unlock()

	obj, ok := m.objects[id]
	if !ok {
		return nil
	}

	delete(m.objects, id)

	// Remove from position map
	objs := m.byPos[obj.Position]
	for i, o := range objs {
		if o.ID == id {
			m.byPos[obj.Position] = append(objs[:i], objs[i+1:]...)
			break
		}
	}
	if len(m.byPos[obj.Position]) == 0 {
		delete(m.byPos, obj.Position)
	}

	return obj
}

// Get retrieves a world object by ID
func (m *WorldObjectManager) Get(id uuid.UUID) *WorldObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.objects[id]
}

// GetAt returns all objects at a position
func (m *WorldObjectManager) GetAt(pos Position) []*WorldObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	objs := m.byPos[pos]
	result := make([]*WorldObject, len(objs))
	copy(result, objs)
	return result
}

// GetAtOfType returns objects of a specific type at a position
func (m *WorldObjectManager) GetAtOfType(pos Position, objType WorldObjectType) []*WorldObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*WorldObject
	for _, obj := range m.byPos[pos] {
		if obj.Type == objType {
			result = append(result, obj)
		}
	}
	return result
}

// GetResourceAt returns the resource node at a position
func (m *WorldObjectManager) GetResourceAt(pos Position) *WorldObject {
	objs := m.GetAtOfType(pos, ObjectResource)
	if len(objs) > 0 {
		return objs[0]
	}
	return nil
}

// GetStructureAt returns the structure at a position
func (m *WorldObjectManager) GetStructureAt(pos Position) *WorldObject {
	objs := m.GetAtOfType(pos, ObjectStructure)
	if len(objs) > 0 {
		return objs[0]
	}
	return nil
}

// GetInteractiveAt returns the interactive object at a position
func (m *WorldObjectManager) GetInteractiveAt(pos Position) *WorldObject {
	objs := m.GetAtOfType(pos, ObjectInteractive)
	if len(objs) > 0 {
		return objs[0]
	}
	return nil
}

// GetDroppedItemsAt returns dropped items at a position
func (m *WorldObjectManager) GetDroppedItemsAt(pos Position) []*WorldObject {
	return m.GetAtOfType(pos, ObjectDroppedItem)
}

// GetAll returns all world objects
func (m *WorldObjectManager) GetAll() []*WorldObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*WorldObject, 0, len(m.objects))
	for _, obj := range m.objects {
		result = append(result, obj)
	}
	return result
}

// GetByOwner returns all objects owned by an agent
func (m *WorldObjectManager) GetByOwner(ownerID uuid.UUID) []*WorldObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*WorldObject
	for _, obj := range m.objects {
		if obj.OwnerID != nil && *obj.OwnerID == ownerID {
			result = append(result, obj)
		}
	}
	return result
}

// HasBlockingObject checks if there's a movement-blocking object at a position
func (m *WorldObjectManager) HasBlockingObject(pos Position) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, obj := range m.byPos[pos] {
		if obj.Type == ObjectStructure && obj.BlocksMovement {
			return true
		}
	}
	return false
}

// GetVisibleObjects returns all visible objects in a radius (excluding hidden traps not owned by viewer)
func (m *WorldObjectManager) GetVisibleObjects(center Position, radius int, viewerID *uuid.UUID) []*WorldObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*WorldObject
	for _, obj := range m.objects {
		// Check if within radius
		dx := obj.Position.X - center.X
		dy := obj.Position.Y - center.Y
		if dx < -radius || dx > radius || dy < -radius || dy > radius {
			continue
		}

		// Hide traps from non-owners
		if obj.Type == ObjectStructure && obj.StructureType == StructureTrap && obj.Hidden {
			if viewerID == nil || obj.OwnerID == nil || *obj.OwnerID != *viewerID {
				continue
			}
		}

		result = append(result, obj)
	}
	return result
}

// GetTrapsAt returns hidden traps at a position (for triggering)
func (m *WorldObjectManager) GetTrapsAt(pos Position) []*WorldObject {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*WorldObject
	for _, obj := range m.byPos[pos] {
		if obj.Type == ObjectStructure && obj.StructureType == StructureTrap {
			result = append(result, obj)
		}
	}
	return result
}

// ProcessDespawns removes objects that should despawn at the given tick
func (m *WorldObjectManager) ProcessDespawns(tick int) []uuid.UUID {
	m.mu.Lock()
	defer m.mu.Unlock()

	var removed []uuid.UUID
	for id, obj := range m.objects {
		if obj.Type == ObjectDroppedItem && obj.DespawnTick > 0 && tick >= obj.DespawnTick {
			delete(m.objects, id)

			// Remove from position map
			objs := m.byPos[obj.Position]
			for i, o := range objs {
				if o.ID == id {
					m.byPos[obj.Position] = append(objs[:i], objs[i+1:]...)
					break
				}
			}
			if len(m.byPos[obj.Position]) == 0 {
				delete(m.byPos, obj.Position)
			}

			removed = append(removed, id)
		}
	}
	return removed
}

// ProcessDepletedResources removes depleted resource nodes
func (m *WorldObjectManager) ProcessDepletedResources() []uuid.UUID {
	m.mu.Lock()
	defer m.mu.Unlock()

	var removed []uuid.UUID
	for id, obj := range m.objects {
		if obj.Type == ObjectResource && obj.Remaining <= 0 {
			delete(m.objects, id)

			objs := m.byPos[obj.Position]
			for i, o := range objs {
				if o.ID == id {
					m.byPos[obj.Position] = append(objs[:i], objs[i+1:]...)
					break
				}
			}
			if len(m.byPos[obj.Position]) == 0 {
				delete(m.byPos, obj.Position)
			}

			removed = append(removed, id)
		}
	}
	return removed
}

// Snapshot creates snapshots of all objects
func (m *WorldObjectManager) Snapshot() []WorldObjectSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make([]WorldObjectSnapshot, 0, len(m.objects))
	for _, obj := range m.objects {
		snapshots = append(snapshots, obj.Snapshot())
	}
	return snapshots
}

// SnapshotVisible creates snapshots of visible objects (excluding hidden traps)
func (m *WorldObjectManager) SnapshotVisible(center Position, radius int, viewerID *uuid.UUID) []WorldObjectSnapshot {
	objs := m.GetVisibleObjects(center, radius, viewerID)
	snapshots := make([]WorldObjectSnapshot, len(objs))
	for i, obj := range objs {
		snapshots[i] = obj.Snapshot()
	}
	return snapshots
}

// Clear removes all objects
func (m *WorldObjectManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects = make(map[uuid.UUID]*WorldObject)
	m.byPos = make(map[Position][]*WorldObject)
}
