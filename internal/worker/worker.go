package worker

import (
	"context"
	"sync"
	"time"
)

const (
	taskProcessTime   = 5 * time.Second
	maxWorkerPoolSize = 5 // Number of workers in the pool
)

type WorkerServer struct {
	ctx             context.Context
	workerPort      string
	coordinatorPort string
	cancel          context.CancelFunc
	wg              sync.WaitGroup // WaitGroup to wait for all goroutines to finish
}

func NewServer(workerPort string, coordinatorPort string) *WorkerServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerServer{
		workerPort:      workerPort,
		coordinatorPort: coordinatorPort,
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (s *WorkerServer) Start() error {
	s.startWorkerPool(maxWorkerPoolSize)
	return nil
}

// startWorkerPool starts a pool of worker goroutines.
func (w *WorkerServer) startWorkerPool(totalWorkers int) {
	for range totalWorkers {
		w.wg.Add(1)
		go w.worker()
	}
}

// worker is the function run by each worker goroutine.
func (w *WorkerServer) worker() {
	defer w.wg.Done() // Signal this worker is done when the function returns.
}
