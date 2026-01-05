// Package workerservice executes distributed tasks as part of a coordinated worker pool.
package workerservice

import (
	"time"
)

const (
	// taskProcessTime is how long each task takes to process.
	taskProcessTime = 5 * time.Second

	// maxWorkerPoolSize is the number of concurrent task processors.
	maxWorkerPoolSize = 5
)

// WorkerService manages the worker lifecycle.
// It registers with a coordinator, receives tasks, and processes them concurrently.
type WorkerService interface {
	// Start initializes the worker and blocks until shutdown.
	// It connects to the coordinator, registers, and begins processing tasks.
	Start() error
}

// NewService creates a new worker that listens on workerServerPort
// and connects to the coordinator at coordinatorPort.
// If workerServerPort is empty, a random available port is used.
func NewService(
	workerServerPort string,
	coordinatorHost string,
	coordinatorPort string,
) WorkerService {
	return newService(workerServerPort, coordinatorHost, coordinatorPort)
}
