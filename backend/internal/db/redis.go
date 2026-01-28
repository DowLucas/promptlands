package db

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// Redis manages Redis connections
type Redis struct {
	client *redis.Client
}

// NewRedis creates a new Redis client
func NewRedis(addr string) (*Redis, error) {
	if addr == "" {
		return &Redis{}, nil
	}

	opts, err := redis.ParseURL(addr)
	if err != nil {
		// Try as plain address
		opts = &redis.Options{
			Addr: addr,
		}
	}

	client := redis.NewClient(opts)

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	log.Println("Connected to Redis")
	return &Redis{client: client}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	if r != nil && r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Client returns the underlying Redis client
func (r *Redis) Client() *redis.Client {
	return r.client
}

// IsConnected returns true if Redis is connected
func (r *Redis) IsConnected() bool {
	return r.client != nil
}

// TODO: Add game state caching methods
// - SetGameState(gameID uuid.UUID, state []byte) error
// - GetGameState(gameID uuid.UUID) ([]byte, error)
// - PublishTick(gameID uuid.UUID, tick int, data []byte) error
// - SubscribeToGame(ctx context.Context, gameID uuid.UUID) (<-chan []byte, error)
