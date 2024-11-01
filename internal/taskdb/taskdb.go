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
	DBName         string
}

const (
	// DefaultUser is the default user used when connecting to postgres.
	DefaultUser = "postgres"
	// DefaultPassword is the default password used when connecting to postgres.
	DefaultPassword = "postgres"
	// DefaultPort is the default port used when connecting to postgres.
	DefaultPort = 5432
	// DefaultHost is the default host used when connecting to postgres.
	DefaultHost = "0.0.0.0"
)

func NewConfig(
	user string,
	password string,
	host string,
	port uint16,
	dbName string,
	maxConnections int,
) *Config {
	config := Config{
		User:           user,
		Password:       password,
		Host:           host,
		Port:           port,
		DBName:         dbName,
		MaxConnections: maxConnections,
	}
	if config.User == "" {
		config.User = DefaultUser
	}
	if config.Password == "" {
		config.Password = DefaultPassword
	}
	if config.Port == 0 {
		config.Port = DefaultPort
	}
	if config.Host == "" {
		config.Host = DefaultHost
	}
	return &config
}

func NewStore(db *sql.DB) *Store {
	return newStore(db)
}

func NewDB(ctx context.Context, config *Config) (*Connection, error) {
	return NewPoolConnection(ctx, config.User, config.Password, config.Host, config.Port)
}
