// Package coordinatorservice manages distributed task execution across worker nodes.
// It handles worker registration, health monitoring, and task distribution.
package coordinatorservice

import (
	"time"

	"github.com/gilwong00/task-runner/internal/taskdb"
)

const (
	// scanInterval is how often the coordinator scans for new tasks.
	scanInterval = 10 * time.Second
	// defaultMaxMisses is the max consecutive missed heartbeats before
	// a worker is removed from the pool.
	defaultMaxMisses = 1
	// defaultHeartbeat is the expected interval between worker heartbeats.
	defaultHeartbeat = 5 * time.Second
)

// CoordinatorService manages the coordinator lifecycle.
// It distributes tasks to workers and monitors their health via heartbeats.
type CoordinatorService interface {
	// Start initializes the coordinator and blocks until shutdown.
	// It starts the HTTP server, worker pool management, and task distribution.
	Start() error
}

// NewService creates a new coordinator that listens on the given port.
// The store is used for task persistence and retrieval.
func NewService(coordinatorPort string, store *taskdb.Store) CoordinatorService {
	return newService(coordinatorPort, store)
}
