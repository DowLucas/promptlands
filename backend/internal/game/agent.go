package game

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/lucas/promptlands/internal/config"
)

// Agent represents a player's AI agent in the game
type Agent struct {
	mu           sync.RWMutex
	ID           uuid.UUID  `json:"id"`
	GameID       uuid.UUID  `json:"game_id"`
	PlayerID     *uuid.UUID `json:"player_id,omitempty"`
	Name         string     `json:"name"`
	SystemPrompt string     `json:"system_prompt"`
	Position     Position   `json:"position"`
	Memory       []string   `json:"memory"`
	MaxMemory    int        `json:"-"`
	IsAdversary  bool       `json:"is_adversary"`
	AdversaryType string    `json:"adversary_type,omitempty"`

	// Combat & Resource fields
	HP        int `json:"hp"`
	MaxHP     int `json:"max_hp"`
	Energy    int `json:"energy"`
	MaxEnergy int `json:"max_energy"`

	// Currency
	Coins int `json:"coins"`

	// Inventory
	Inventory *Inventory `json:"-"`

	// Upgrade levels (1-5 for vision/memory, 1-3 for strength/storage/speed/claim)
	VisionLevel   int `json:"vision_level"`
	MemoryLevel   int `json:"memory_level"`
	StrengthLevel int `json:"strength_level"`
	StorageLevel  int `json:"storage_level"`
	SpeedLevel    int `json:"speed_level"`
	ClaimLevel    int `json:"claim_level"`

	// Death & Respawn
	IsDead      bool `json:"is_dead"`
	DeathTick   int  `json:"death_tick,omitempty"`
	RespawnTick int  `json:"respawn_tick,omitempty"`

	// Fog of War - tiles the agent has seen (for persistence)
	// MULTIPLAYER: Already per-agent, scales naturally
	ExploredTiles map[string]bool `json:"-"` // "x,y" -> true, not serialized to client
}

// Default values for new agents (used when no balance config provided)
const (
	DefaultHP        = 3
	DefaultMaxHP     = 3
	DefaultEnergy    = 0
	DefaultMaxEnergy = 100
	RespawnTicks     = 5
)

// DefaultInventorySlots is defined in inventory.go

// NewAgent creates a new agent with default balance values
func NewAgent(gameID uuid.UUID, name, systemPrompt string, startPos Position, maxMemory int) *Agent {
	return NewAgentWithBalance(gameID, name, systemPrompt, startPos, maxMemory, nil)
}

// NewAgentWithBalance creates a new agent with specific balance configuration
func NewAgentWithBalance(gameID uuid.UUID, name, systemPrompt string, startPos Position, maxMemory int, balance *config.BalanceConfig) *Agent {
	hp := DefaultHP
	maxHP := DefaultMaxHP
	energy := DefaultEnergy
	maxEnergy := DefaultMaxEnergy

	if balance != nil {
		hp = balance.Agent.DefaultHP
		maxHP = balance.Agent.DefaultMaxHP
		energy = balance.Agent.DefaultEnergy
		maxEnergy = balance.Agent.DefaultMaxEnergy
	}

	agent := &Agent{
		ID:            uuid.New(),
		GameID:        gameID,
		Name:          name,
		SystemPrompt:  systemPrompt,
		Position:      startPos,
		Memory:        make([]string, 0, maxMemory),
		MaxMemory:     maxMemory,
		IsAdversary:   false,
		HP:            hp,
		MaxHP:         maxHP,
		Energy:        energy,
		MaxEnergy:     maxEnergy,
		VisionLevel:   1,
		MemoryLevel:   1,
		StrengthLevel: 1,
		StorageLevel:  1,
		SpeedLevel:    1,
		ClaimLevel:    1,
	}
	return agent
}

// InitInventory initializes the agent's inventory with the given registry
func (a *Agent) InitInventory(registry *ItemRegistry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Inventory = NewInventory(a.ID, DefaultInventorySlots+(a.StorageLevel-1)*5, registry)
}

// NewAdversaryAgent creates an AI adversary agent with default balance values
func NewAdversaryAgent(gameID uuid.UUID, adversaryType string, startPos Position, maxMemory int) *Agent {
	return NewAdversaryAgentWithBalance(gameID, adversaryType, startPos, maxMemory, nil)
}

// NewAdversaryAgentWithBalance creates an AI adversary agent with specific balance configuration
func NewAdversaryAgentWithBalance(gameID uuid.UUID, adversaryType string, startPos Position, maxMemory int, balance *config.BalanceConfig) *Agent {
	advConfig, ok := Adversaries[adversaryType]
	if !ok {
		advConfig = Adversaries["chaotic"] // Default to chaotic if unknown
	}

	hp := DefaultHP
	maxHP := DefaultMaxHP
	energy := DefaultEnergy
	maxEnergy := DefaultMaxEnergy

	if balance != nil {
		hp = balance.Agent.DefaultHP
		maxHP = balance.Agent.DefaultMaxHP
		energy = balance.Agent.DefaultEnergy
		maxEnergy = balance.Agent.DefaultMaxEnergy
	}

	return &Agent{
		ID:            uuid.New(),
		GameID:        gameID,
		Name:          advConfig.Name,
		SystemPrompt:  advConfig.Prompt,
		Position:      startPos,
		Memory:        make([]string, 0, maxMemory),
		MaxMemory:     maxMemory,
		IsAdversary:   true,
		AdversaryType: adversaryType,
		HP:            hp,
		MaxHP:         maxHP,
		Energy:        energy,
		MaxEnergy:     maxEnergy,
		VisionLevel:   1,
		MemoryLevel:   1,
		StrengthLevel: 1,
		StorageLevel:  1,
		SpeedLevel:    1,
		ClaimLevel:    1,
	}
}

// SetPlayerID assigns a player to this agent
func (a *Agent) SetPlayerID(playerID uuid.UUID) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.PlayerID = &playerID
}

// GetPosition returns the agent's current position
func (a *Agent) GetPosition() Position {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Position
}

// SetPosition updates the agent's position
func (a *Agent) SetPosition(pos Position) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Position = pos
}

// AddMemory adds an event to the agent's memory
func (a *Agent) AddMemory(event string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.Memory = append(a.Memory, event)

	// Trim oldest memories if over limit
	if len(a.Memory) > a.MaxMemory {
		a.Memory = a.Memory[len(a.Memory)-a.MaxMemory:]
	}
}

// GetMemory returns a copy of the agent's memory
func (a *Agent) GetMemory() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	memory := make([]string, len(a.Memory))
	copy(memory, a.Memory)
	return memory
}

// ClearMemory clears the agent's memory
func (a *Agent) ClearMemory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Memory = make([]string, 0, a.MaxMemory)
}

// Snapshot creates a serializable copy of the agent state
func (a *Agent) Snapshot() AgentSnapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return AgentSnapshot{
		ID:            a.ID,
		Name:          a.Name,
		Position:      a.Position,
		IsAdversary:   a.IsAdversary,
		AdversaryType: a.AdversaryType,
		HP:            a.HP,
		MaxHP:         a.MaxHP,
		Energy:        a.Energy,
		MaxEnergy:     a.MaxEnergy,
		Coins:         a.Coins,
		VisionLevel:   a.VisionLevel,
		MemoryLevel:   a.MemoryLevel,
		StrengthLevel: a.StrengthLevel,
		StorageLevel:  a.StorageLevel,
		SpeedLevel:    a.SpeedLevel,
		ClaimLevel:    a.ClaimLevel,
		IsDead:        a.IsDead,
	}
}

// AgentSnapshot is a serializable representation of an agent (public info only)
type AgentSnapshot struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Position      Position  `json:"position"`
	IsAdversary   bool      `json:"is_adversary"`
	AdversaryType string    `json:"adversary_type,omitempty"`
	HP            int       `json:"hp"`
	MaxHP         int       `json:"max_hp"`
	Energy        int       `json:"energy"`
	MaxEnergy     int       `json:"max_energy"`
	Coins         int       `json:"coins"`
	VisionLevel   int       `json:"vision_level"`
	MemoryLevel   int       `json:"memory_level"`
	StrengthLevel int       `json:"strength_level"`
	StorageLevel  int       `json:"storage_level"`
	SpeedLevel    int       `json:"speed_level"`
	ClaimLevel    int       `json:"claim_level"`
	IsDead        bool      `json:"is_dead"`
}

// GetHP returns the agent's current HP (thread-safe)
func (a *Agent) GetHP() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.HP
}

// SetHP sets the agent's HP (thread-safe)
func (a *Agent) SetHP(hp int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.HP = hp
	if a.HP > a.MaxHP {
		a.HP = a.MaxHP
	}
	if a.HP < 0 {
		a.HP = 0
	}
}

// GetEnergy returns the agent's current energy (thread-safe)
func (a *Agent) GetEnergy() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Energy
}

// SetEnergy sets the agent's energy (thread-safe)
func (a *Agent) SetEnergy(energy int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Energy = energy
	if a.Energy > a.MaxEnergy {
		a.Energy = a.MaxEnergy
	}
	if a.Energy < 0 {
		a.Energy = 0
	}
}

// AddEnergy adds energy to the agent (thread-safe)
func (a *Agent) AddEnergy(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Energy += amount
	if a.Energy > a.MaxEnergy {
		a.Energy = a.MaxEnergy
	}
}

// SpendEnergy attempts to spend energy, returns true if successful
func (a *Agent) SpendEnergy(amount int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Energy < amount {
		return false
	}
	a.Energy -= amount
	return true
}

// GetCoins returns the agent's coin count (thread-safe)
func (a *Agent) GetCoins() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Coins
}

// AddCoins adds coins to the agent (thread-safe, no cap)
func (a *Agent) AddCoins(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Coins += amount
}

// SpendCoins attempts to spend coins, returns true if successful
func (a *Agent) SpendCoins(amount int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Coins < amount {
		return false
	}
	a.Coins -= amount
	return true
}

// TakeDamage applies damage to the agent, returns true if killed
func (a *Agent) TakeDamage(damage int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.HP -= damage
	if a.HP <= 0 {
		a.HP = 0
		return true
	}
	return false
}

// Heal restores HP to the agent
func (a *Agent) Heal(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.HP += amount
	if a.HP > a.MaxHP {
		a.HP = a.MaxHP
	}
}

// Kill marks the agent as dead and sets respawn timer
func (a *Agent) Kill(currentTick int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.IsDead = true
	a.HP = 0
	a.Coins = 0
	a.DeathTick = currentTick
	a.RespawnTick = currentTick + RespawnTicks
}

// Respawn resets the agent to alive state at a new position
func (a *Agent) Respawn(pos Position) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.IsDead = false
	a.HP = a.MaxHP
	a.Position = pos
	a.DeathTick = 0
	a.RespawnTick = 0
}

// ShouldRespawn checks if the agent should respawn at the given tick
func (a *Agent) ShouldRespawn(currentTick int) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.IsDead && currentTick >= a.RespawnTick
}

// GetEffectiveVision returns vision radius with upgrades
func (a *Agent) GetEffectiveVision(baseVision int) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return baseVision + (a.VisionLevel - 1)
}

// GetEffectiveMemory returns memory slots with upgrades
func (a *Agent) GetEffectiveMemory() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.MaxMemory + (a.MemoryLevel-1)*2
}

// GetEffectiveStrength returns attack damage with upgrades
func (a *Agent) GetEffectiveStrength() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 1 + (a.StrengthLevel - 1)
}

// GetEffectiveMoveSpeed returns move speed (tiles per MOVE) with upgrades
func (a *Agent) GetEffectiveMoveSpeed(baseMoveSpeed int) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return baseMoveSpeed + (a.SpeedLevel - 1)
}

// GetEffectiveClaimRadius returns claim radius with upgrades
func (a *Agent) GetEffectiveClaimRadius(baseClaimRadius int) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return baseClaimRadius + (a.ClaimLevel - 1)
}

// GetUpgradeCost returns the cost to upgrade to the next level
func GetUpgradeCost(currentLevel int) int {
	switch currentLevel {
	case 1:
		return 10
	case 2:
		return 20
	case 3:
		return 35
	case 4:
		return 55
	default:
		return 0 // Max level
	}
}

// CanUpgrade checks if the agent can upgrade a specific stat
func (a *Agent) CanUpgrade(upgradeType string) (bool, int, string) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var currentLevel, maxLevel int
	switch upgradeType {
	case "vision":
		currentLevel = a.VisionLevel
		maxLevel = 5
	case "memory":
		currentLevel = a.MemoryLevel
		maxLevel = 5
	case "strength":
		currentLevel = a.StrengthLevel
		maxLevel = 3
	case "storage":
		currentLevel = a.StorageLevel
		maxLevel = 3
	case "speed":
		currentLevel = a.SpeedLevel
		maxLevel = 3
	case "claim":
		currentLevel = a.ClaimLevel
		maxLevel = 3
	default:
		return false, 0, "invalid upgrade type"
	}

	if currentLevel >= maxLevel {
		return false, 0, "already at max level"
	}

	cost := GetUpgradeCost(currentLevel)
	if a.Energy < cost {
		return false, cost, "not enough energy"
	}

	return true, cost, ""
}

// ApplyUpgrade applies an upgrade to the agent
func (a *Agent) ApplyUpgrade(upgradeType string) bool {
	can, cost, _ := a.CanUpgrade(upgradeType)
	if !can {
		return false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.Energy -= cost

	switch upgradeType {
	case "vision":
		a.VisionLevel++
	case "memory":
		a.MemoryLevel++
		// Expand memory capacity
		a.MaxMemory += 2
	case "strength":
		a.StrengthLevel++
	case "storage":
		a.StorageLevel++
		// Expand inventory
		if a.Inventory != nil {
			a.Inventory.ExpandSlots(5)
		}
	case "speed":
		a.SpeedLevel++
	case "claim":
		a.ClaimLevel++
	}

	return true
}

// UpdateExploredTiles marks tiles as explored for fog of war
func (a *Agent) UpdateExploredTiles(visibleTiles []*Tile) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ExploredTiles == nil {
		a.ExploredTiles = make(map[string]bool)
	}
	for _, tile := range visibleTiles {
		key := fmt.Sprintf("%d,%d", tile.Position.X, tile.Position.Y)
		a.ExploredTiles[key] = true
	}
}

// GetExploredTiles returns a copy of the explored tiles map
func (a *Agent) GetExploredTiles() map[string]bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.ExploredTiles == nil {
		return make(map[string]bool)
	}
	explored := make(map[string]bool, len(a.ExploredTiles))
	for k, v := range a.ExploredTiles {
		explored[k] = v
	}
	return explored
}

// AgentContext contains all information needed for an agent to make a decision
type AgentContext struct {
	Agent            *Agent
	VisibleTiles     []*Tile
	VisibleObjects   []*WorldObject
	VisibleAgents    []*AgentSnapshot
	OwnedCount       int
	Messages         []IncomingMessage
	CurrentTick      int
	WorldSize        int
	CurrentTileOwned bool
	CurrentTileEnemy bool
	EnergyPerTick    int    // Passive income from owned tiles
	MoveSpeed        int    // Effective tiles per MOVE action
	ClaimRadius      int    // Effective claim radius
	CurrentBiome     string // Biome type at agent's current position
}

// IncomingMessage represents a message received by an agent
type IncomingMessage struct {
	FromAgentID   uuid.UUID `json:"from_agent_id"`
	FromAgentName string    `json:"from_agent_name"`
	Content       string    `json:"content"`
	IsBroadcast   bool      `json:"is_broadcast"`
}
