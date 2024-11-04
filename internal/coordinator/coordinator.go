package coordinator

import (
	"context"
	"sync"
)

type CoordinatorServer struct {
	ctx             context.Context    // The root context for all goroutines
	cancel          context.CancelFunc // Function to cancel the context
	wg              sync.WaitGroup     // WaitGroup to wait for all goroutines to finish
	coordinatorPort string
}

func NewServer(coordinatorPort string) *CoordinatorServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoordinatorServer{
		ctx:             ctx,
		cancel:          cancel,
		coordinatorPort: coordinatorPort,
	}
}
