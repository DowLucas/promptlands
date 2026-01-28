package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Game     GameConfig     `yaml:"game"`
	LLM      LLMConfig      `yaml:"llm"`
	Database DatabaseConfig `yaml:"database"`
	Dev      DevConfig      `yaml:"dev"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type GameConfig struct {
	TickDuration   time.Duration `yaml:"tick_duration"`
	MapSize        int           `yaml:"map_size"`
	MaxPlayers     int           `yaml:"max_players"`
	VisionRadius   int           `yaml:"vision_radius"`
	MaxMemoryItems int           `yaml:"max_memory_items"`
	WinAfterTicks  int           `yaml:"win_after_ticks"`
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
	Enabled bool `yaml:"enabled"`
	MockLLM bool `yaml:"mock_llm"`
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
			TickDuration:   10 * time.Second,
			MapSize:        20,
			MaxPlayers:     4,
			VisionRadius:   3,
			MaxMemoryItems: 10,
			WinAfterTicks:  100,
		},
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
