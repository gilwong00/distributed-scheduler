package workerservice

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/gilwong00/task-runner/proto/gen"

	"google.golang.org/grpc"
)

const (
	taskProcessTime   = 5 * time.Second
	maxWorkerPoolSize = 5 // Number of workers in the pool
)

type WorkerService struct {
	pb.UnimplementedWorkerServiceServer
	ctx               context.Context
	workerServerPort  string
	coordinatorPort   string
	coordinatorClient *grpc.ClientConn
	listener          net.Listener
	grpcServer        *grpc.Server
	cancel            context.CancelFunc
	wg                sync.WaitGroup // WaitGroup to wait for all goroutines to finish
}

func NewService(workerServerPort string, coordinatorPort string) *WorkerService {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerService{
		workerServerPort: workerServerPort,
		coordinatorPort:  coordinatorPort,
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (w *WorkerService) Start() error {
	w.startWorkerPool(maxWorkerPoolSize)

	if err := w.startGRPCServer(); err != nil {
		return fmt.Errorf("gRPC server start failed: %w", err)
	}

	return w.awaitShutdown()
}

// startWorkerPool starts a pool of worker goroutines.
func (w *WorkerService) startWorkerPool(totalWorkers int) {
	for range totalWorkers {
		w.wg.Add(1)
		go w.worker()
	}
}

func (w *WorkerService) awaitShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	w.stop()
	return nil
}

// worker is the function run by each worker goroutine.
func (w *WorkerService) worker() {
	defer w.wg.Done() // Signal this worker is done when the function returns.
}

func (w *WorkerService) stop() {
	// Signal all goroutines to stop
	w.cancel()
	// Wait for all goroutines to finish
	w.wg.Wait()
	// close GRPC connection
	w.closeGRPCConnection()
	log.Println("Worker server stopped")
}
