package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres manages PostgreSQL connections
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a new PostgreSQL connection pool
func NewPostgres(connString string) (*Postgres, error) {
	if connString == "" {
		return &Postgres{}, nil
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	log.Println("Connected to PostgreSQL")
	return &Postgres{pool: pool}, nil
}

// Close closes the connection pool
func (p *Postgres) Close() {
	if p != nil && p.pool != nil {
		p.pool.Close()
	}
}

// Pool returns the underlying connection pool
func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

// IsConnected returns true if the database is connected
func (p *Postgres) IsConnected() bool {
	return p.pool != nil
}

// TODO: Add game persistence methods
// - SaveGame(game *game.Engine) error
// - LoadGame(gameID uuid.UUID) (*game.Engine, error)
// - SaveGameEvent(event *game.GameEvent) error
// - GetGameEvents(gameID uuid.UUID, fromTick, toTick int) ([]game.GameEvent, error)
