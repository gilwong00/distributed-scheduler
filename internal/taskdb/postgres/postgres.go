package postgres

import "time"

// Postgres defaults
const (
	//
	DefaultMaxOpenConnections = 10
	//
	DefaultMaxIdleConnections = DefaultMaxOpenConnections
	//
	DefaultMaxConnectionLifeTime = 5 * time.Minute
)
