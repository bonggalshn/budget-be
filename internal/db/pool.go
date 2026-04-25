package db

import (
	"context"
	"fmt"

	"github.com/bonggalshn/budget-be/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps a pgxpool.Pool for database connections.
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a new database connection pool.
// Validates connectivity by pinging the database.
func NewPool(ctx context.Context, cfg config.DBConfig) (*Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// Close releases all database connections in the pool.
func (p *Pool) Close() {
	p.Pool.Close()
}
