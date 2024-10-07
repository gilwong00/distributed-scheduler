package taskdb

import (
	"context"
	"database/sql"
)

type Config struct {
	MaxConnections int
	User           string
	Password       string
	Host           string
	Port           uint16
}

const (
	// DefaultPort is the default port used when connection to postgres.
	DefaultPort = 5432
	// DefaultHost is the default host used when connecting to postgres.
	DefaultHost = "0.0.0.0"
)

func NewConfig(
	user string,
	password string,
	host string,
	port uint16,
	maxConnections int,
) (*Config, error) {
	return &Config{
		User:           user,
		Password:       password,
		Host:           host,
		Port:           port,
		MaxConnections: maxConnections,
	}, nil
}

func NewStore(db *sql.DB) *Store {
	return newStore(db)
}

func NewDB(ctx context.Context, config *Config) (*Connection, error) {
	return NewPoolConnection(ctx, config.Host, config.Port)
}
