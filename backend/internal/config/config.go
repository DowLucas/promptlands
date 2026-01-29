package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Game     GameConfig     `yaml:"game"`
	Balance  BalanceConfig  `yaml:"balance"`
	LLM      LLMConfig      `yaml:"llm"`
	Database DatabaseConfig `yaml:"database"`
	Dev      DevConfig      `yaml:"dev"`
}

// BalanceConfig centralizes game balance values for easy tuning
type BalanceConfig struct {
	Agent    AgentBalance    `yaml:"agent"`
	Combat   CombatBalance   `yaml:"combat"`
	Upgrades UpgradeBalance  `yaml:"upgrades"`
	Actions  ActionsBalance  `yaml:"actions"`
}

// AgentBalance contains agent-related balance values
type AgentBalance struct {
	DefaultHP          int `yaml:"default_hp"`
	DefaultMaxHP       int `yaml:"default_max_hp"`
	DefaultEnergy      int `yaml:"default_energy"`
	DefaultMaxEnergy   int `yaml:"default_max_energy"`
	RespawnTicks       int `yaml:"respawn_ticks"`
	DefaultVision      int `yaml:"default_vision"`
	DefaultMemory      int `yaml:"default_memory"`
	DefaultMoveSpeed   int `yaml:"default_move_speed"`
	DefaultClaimRadius int `yaml:"default_claim_radius"`
}

// CombatBalance contains combat-related balance values
type CombatBalance struct {
	BaseDamage       int `yaml:"base_damage"`
	TrapDamage       int `yaml:"trap_damage"`
	WallHP           int `yaml:"wall_hp"`
	BeaconHP         int `yaml:"beacon_hp"`
	TrapHP           int `yaml:"trap_hp"`
	BeaconVisionBonus int `yaml:"beacon_vision_bonus"`
}

// UpgradeBalance contains upgrade costs and limits
type UpgradeBalance struct {
	VisionMaxLevel   int   `yaml:"vision_max_level"`
	MemoryMaxLevel   int   `yaml:"memory_max_level"`
	StrengthMaxLevel int   `yaml:"strength_max_level"`
	StorageMaxLevel  int   `yaml:"storage_max_level"`
	SpeedMaxLevel    int   `yaml:"speed_max_level"`
	ClaimMaxLevel    int   `yaml:"claim_max_level"`
	UpgradeCosts     []int `yaml:"upgrade_costs"`
	MemoryPerLevel   int   `yaml:"memory_per_level"`
	StoragePerLevel  int   `yaml:"storage_per_level"`
	SpeedPerLevel    int   `yaml:"speed_per_level"`
	ClaimPerLevel    int   `yaml:"claim_per_level"`
}

// ActionsBalance contains action-related balance values
type ActionsBalance struct {
	ScanEnergyCost    int `yaml:"scan_energy_cost"`
	ItemDespawnTicks  int `yaml:"item_despawn_ticks"`
	DefaultInventory  int `yaml:"default_inventory_slots"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type GameConfig struct {
	TickDuration      time.Duration `yaml:"tick_duration"`
	MapSize           int           `yaml:"map_size"`
	MaxPlayers        int           `yaml:"max_players"`
	VisionRadius      int           `yaml:"vision_radius"`
	MaxMemoryItems    int           `yaml:"max_memory_items"`
	WinAfterTicks     int           `yaml:"win_after_ticks"`
	ResourceSpawnRate float64       `yaml:"resource_spawn_rate"`
	Map               MapYAMLConfig `yaml:"map"`
}

// MapYAMLConfig holds the nested map configuration from YAML
type MapYAMLConfig struct {
	Preset             string  `yaml:"preset"`
	Size               string  `yaml:"size"`
	CustomSize         int     `yaml:"custom_size"`
	Seed               int64   `yaml:"seed"`
	ChunkSize          int     `yaml:"chunk_size"`
	FogOfWar           bool    `yaml:"fog_of_war"`
	RespawnEnabled     bool    `yaml:"respawn_enabled"`
	ResourceDensity    float64 `yaml:"resource_density"`
	DifficultyMultiplier float64 `yaml:"difficulty_multiplier"`
}

// GetMapSize returns the effective map size from config
// Priority: 1. MapSize if set, 2. Map.CustomSize if > 0, 3. Map.Size string, 4. default 512
func (g *GameConfig) GetMapSize() int {
	// Direct map_size takes precedence (for backwards compatibility)
	if g.MapSize > 0 {
		return g.MapSize
	}
	// Custom size overrides preset
	if g.Map.CustomSize > 0 {
		return g.Map.CustomSize
	}
	// Parse size string preset
	switch g.Map.Size {
	case "tiny":
		return 128
	case "small":
		return 256
	case "medium":
		return 512
	case "large":
		return 1024
	case "huge":
		return 2048
	case "massive":
		return 4096
	default:
		return 512 // Default to medium
	}
}

type LLMConfig struct {
	Provider  string        `yaml:"provider"`
	Model     string        `yaml:"model"`
	Timeout   time.Duration `yaml:"timeout"`
	MaxTokens int           `yaml:"max_tokens"`
	APIKey    string        `yaml:"-"` // From environment
}

type DatabaseConfig struct {
	PostgresURL string `yaml:"postgres_url"`
	RedisURL    string `yaml:"redis_url"`
}

type DevConfig struct {
	Enabled   bool `yaml:"enabled"`
	MockLLM   bool `yaml:"mock_llm"`
	PauseTick bool `yaml:"pause_tick"` // Start games paused (no tick loop, no LLM calls)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Load API key from environment
	cfg.LLM.APIKey = os.Getenv("GEMINI_API_KEY")

	return &cfg, nil
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Game: GameConfig{
			TickDuration:      10 * time.Second,
			MapSize:           0, // Let Map.Size take precedence
			MaxPlayers:        4,
			VisionRadius:      3,
			MaxMemoryItems:    10,
			WinAfterTicks:     100,
			ResourceSpawnRate: 1.0,
			Map: MapYAMLConfig{
				Preset:             "default",
				Size:               "medium",
				CustomSize:         0,
				Seed:               0,
				ChunkSize:          32,
				FogOfWar:           true,
				RespawnEnabled:     true,
				ResourceDensity:    1.0,
				DifficultyMultiplier: 1.0,
			},
		},
		Balance: DefaultBalanceConfig(),
		LLM: LLMConfig{
			Provider:  "gemini",
			Model:     "gemini-2.5-flash-lite",
			Timeout:   8 * time.Second,
			MaxTokens: 256,
		},
		Database: DatabaseConfig{
			PostgresURL: "postgres://promptlands:promptlands@localhost:5432/promptlands?sslmode=disable",
			RedisURL:    "redis://localhost:6379",
		},
		Dev: DevConfig{
			Enabled: false,
			MockLLM: false,
		},
	}
}

// DefaultBalanceConfig returns the default balance configuration
func DefaultBalanceConfig() BalanceConfig {
	return BalanceConfig{
		Agent: AgentBalance{
			DefaultHP:          3,
			DefaultMaxHP:       3,
			DefaultEnergy:      0,
			DefaultMaxEnergy:   100,
			RespawnTicks:       5,
			DefaultVision:      3,
			DefaultMemory:      10,
			DefaultMoveSpeed:   3,
			DefaultClaimRadius: 4,
		},
		Combat: CombatBalance{
			BaseDamage:        1,
			TrapDamage:        1,
			WallHP:            3,
			BeaconHP:          2,
			TrapHP:            1,
			BeaconVisionBonus: 2,
		},
		Upgrades: UpgradeBalance{
			VisionMaxLevel:   5,
			MemoryMaxLevel:   5,
			StrengthMaxLevel: 3,
			StorageMaxLevel:  3,
			SpeedMaxLevel:    3,
			ClaimMaxLevel:    3,
			UpgradeCosts:     []int{10, 20, 35, 55}, // Cost to go from level N to N+1
			MemoryPerLevel:   2,
			StoragePerLevel:  5,
			SpeedPerLevel:    1,
			ClaimPerLevel:    1,
		},
		Actions: ActionsBalance{
			ScanEnergyCost:   2,
			ItemDespawnTicks: 50,
			DefaultInventory: 10,
		},
	}
}

// GetUpgradeCost returns the cost to upgrade from a given level
func (u *UpgradeBalance) GetUpgradeCost(currentLevel int) int {
	idx := currentLevel - 1
	if idx < 0 || idx >= len(u.UpgradeCosts) {
		return 0 // Max level or invalid
	}
	return u.UpgradeCosts[idx]
}
