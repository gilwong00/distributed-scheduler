package taskdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Connection struct {
	Pool *pgxpool.Pool
}

type PoolOptions func(*pgxpool.Config)

// NewDB returns a *sql.DB
//
// This function will set the max number of idle connections to 0 (connections must be managed by pgxpool)
// Do not alter the max number of idle connections on the returned db.
// This is only intered for code or workflows that requires *sql.DB
func (c *Connection) NewDB() *sql.DB {
	db := stdlib.OpenDBFromPool(c.Pool)
	db.SetMaxOpenConns(int(c.Pool.Config().MaxConns))
	return db
}

// NewPoolConnection returns a new Connection. This will be used for the main database connection pool
// also safe to use concurrently.
func NewPoolConnection(ctx context.Context, host string, port uint16, options ...PoolOptions) (*Connection, error) {
	config, err := pgxpool.ParseConfig(
		fmt.Sprintf("postgres://%s:%d", host, port),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: invalid host or port: %w", err)
	}
	for _, opt := range options {
		opt(config)
	}
	// Set defaults
	if config.MaxConns == 0 {
		config.MaxConns = 10
	}
	if config.MaxConnLifetime == 0 {
		config.MaxConnLifetime = 5 * time.Minute
	}
	if config.MaxConnIdleTime == 0 {
		config.MaxConnLifetime = 5 * time.Minute
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to created postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &Connection{
		Pool: pool,
	}, nil
}
