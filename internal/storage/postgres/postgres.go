package postgres

import (
	"context"
	"fmt"

	"github.com/gabrielrondon/zapiki/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a PostgreSQL connection pool
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new PostgreSQL store
func New(cfg *config.DatabaseConfig) (*Store, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

// Close closes the database connection pool
func (s *Store) Close() {
	s.pool.Close()
}

// Pool returns the underlying connection pool
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// Health checks if the database is healthy
func (s *Store) Health(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
