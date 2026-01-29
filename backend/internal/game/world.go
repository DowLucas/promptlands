package game

import (
	"sync"

	"github.com/google/uuid"
)

// TerrainType represents the type of terrain on a tile
type TerrainType string

const (
	TerrainPlains   TerrainType = "plains"
	TerrainForest   TerrainType = "forest"
	TerrainMountain TerrainType = "mountain"
	TerrainWater    TerrainType = "water"
)

// Position represents a coordinate on the map
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Tile represents a single cell in the game world
type Tile struct {
	Position Position    `json:"position"`
	OwnerID  *uuid.UUID  `json:"owner_id,omitempty"`
	Terrain  TerrainType `json:"terrain"`
	Biome    string      `json:"biome,omitempty"`
}

// World manages the game map state
type World struct {
	mu       sync.RWMutex
	size     int
	seed     int64
	tiles    [][]*Tile
	ownerMap map[uuid.UUID][]Position // Quick lookup of tiles owned by each agent
}

// NewWorld creates a new game world of the given size with default terrain
func NewWorld(size int) *World {
	tiles := make([][]*Tile, size)
	for y := 0; y < size; y++ {
		tiles[y] = make([]*Tile, size)
		for x := 0; x < size; x++ {
			tiles[y][x] = &Tile{
				Position: Position{X: x, Y: y},
				OwnerID:  nil,
				Terrain:  TerrainPlains,
				Biome:    "savanna",
			}
		}
	}

	return &World{
		size:     size,
		seed:     0,
		tiles:    tiles,
		ownerMap: make(map[uuid.UUID][]Position),
	}
}

// NewWorldWithSeed creates a new game world with procedurally generated terrain
func NewWorldWithSeed(size int, seed int64, tiles [][]*Tile) *World {
	return &World{
		size:     size,
		seed:     seed,
		tiles:    tiles,
		ownerMap: make(map[uuid.UUID][]Position),
	}
}

// Seed returns the world's generation seed
func (w *World) Seed() int64 {
	return w.seed
}

// Size returns the world dimensions
func (w *World) Size() int {
	return w.size
}

// GetTile returns the tile at the given position
func (w *World) GetTile(pos Position) *Tile {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.isValidPosition(pos) {
		return nil
	}
	return w.tiles[pos.Y][pos.X]
}

// SetOwner claims a tile for an agent
func (w *World) SetOwner(pos Position, ownerID *uuid.UUID) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isValidPosition(pos) {
		return false
	}

	tile := w.tiles[pos.Y][pos.X]

	// Remove from previous owner's list
	if tile.OwnerID != nil {
		w.removeFromOwnerMap(*tile.OwnerID, pos)
	}

	// Set new owner
	tile.OwnerID = ownerID

	// Add to new owner's list
	if ownerID != nil {
		w.ownerMap[*ownerID] = append(w.ownerMap[*ownerID], pos)
	}

	return true
}

// GetVisibleTiles returns all tiles visible from a position within the given radius (circular).
func (w *World) GetVisibleTiles(center Position, radius int) []*Tile {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var visible []*Tile
	r2 := radius * radius

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > r2 {
				continue
			}
			pos := Position{X: center.X + dx, Y: center.Y + dy}
			if w.isValidPosition(pos) {
				visible = append(visible, w.tiles[pos.Y][pos.X])
			}
		}
	}

	return visible
}

// GetOwnedTiles returns all tiles owned by an agent
func (w *World) GetOwnedTiles(ownerID uuid.UUID) []Position {
	w.mu.RLock()
	defer w.mu.RUnlock()

	positions := w.ownerMap[ownerID]
	result := make([]Position, len(positions))
	copy(result, positions)
	return result
}

// CountOwnedTiles returns the number of tiles owned by an agent
func (w *World) CountOwnedTiles(ownerID uuid.UUID) int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.ownerMap[ownerID])
}

// GetAllTiles returns all tiles in the world (for full state sync)
func (w *World) GetAllTiles() []*Tile {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var all []*Tile
	for y := 0; y < w.size; y++ {
		for x := 0; x < w.size; x++ {
			all = append(all, w.tiles[y][x])
		}
	}
	return all
}

// GetOwnershipMap returns a map of agent IDs to tile counts
func (w *World) GetOwnershipMap() map[uuid.UUID]int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	result := make(map[uuid.UUID]int)
	for id, positions := range w.ownerMap {
		result[id] = len(positions)
	}
	return result
}

// isValidPosition checks if a position is within world bounds
func (w *World) isValidPosition(pos Position) bool {
	return pos.X >= 0 && pos.X < w.size && pos.Y >= 0 && pos.Y < w.size
}

// IsValidPosition checks if a position is within world bounds (exported)
func (w *World) IsValidPosition(pos Position) bool {
	return w.isValidPosition(pos)
}

// removeFromOwnerMap removes a position from an owner's list
func (w *World) removeFromOwnerMap(ownerID uuid.UUID, pos Position) {
	positions := w.ownerMap[ownerID]
	for i, p := range positions {
		if p.X == pos.X && p.Y == pos.Y {
			w.ownerMap[ownerID] = append(positions[:i], positions[i+1:]...)
			break
		}
	}
	if len(w.ownerMap[ownerID]) == 0 {
		delete(w.ownerMap, ownerID)
	}
}

// Snapshot creates a copy of the world state for serialization
func (w *World) Snapshot() WorldSnapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	tiles := make([]TileSnapshot, 0, w.size*w.size)
	for y := 0; y < w.size; y++ {
		for x := 0; x < w.size; x++ {
			t := w.tiles[y][x]
			tiles = append(tiles, TileSnapshot{
				X:       t.Position.X,
				Y:       t.Position.Y,
				OwnerID: t.OwnerID,
				Terrain: t.Terrain,
				Biome:   t.Biome,
			})
		}
	}

	return WorldSnapshot{
		Size:  w.size,
		Seed:  w.seed,
		Tiles: tiles,
	}
}

// WorldSnapshot is a serializable representation of the world
type WorldSnapshot struct {
	Size  int            `json:"size"`
	Seed  int64          `json:"seed"`
	Tiles []TileSnapshot `json:"tiles"`
}

// TileSnapshot is a serializable representation of a tile
type TileSnapshot struct {
	X       int         `json:"x"`
	Y       int         `json:"y"`
	OwnerID *uuid.UUID  `json:"owner_id,omitempty"`
	Terrain TerrainType `json:"terrain"`
	Biome   string      `json:"biome,omitempty"`
}
