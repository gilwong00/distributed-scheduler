package coordinatorservice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/gilwong00/task-runner/internal/gen/proto/coordinator/v1/coordinatorv1connect"
	workerv1 "github.com/gilwong00/task-runner/internal/gen/proto/worker/v1"
	"github.com/gilwong00/task-runner/internal/gen/proto/worker/v1/workerv1connect"
	"github.com/gilwong00/task-runner/internal/taskdb"
	taskpostgres "github.com/gilwong00/task-runner/internal/taskdb/gen"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type workerInfo struct {
	heartbeatMisses     uint8
	address             string
	httpClient          *http.Client
	workerServiceClient workerv1connect.WorkerServiceClient
}

// CoordinatorServer manages worker registration, heartbeat tracking, and task coordination
// in a distributed scheduling system. It also handles gRPC server lifecycle and interacts
// with a backing task store and database.
type coordinatorService struct {
	ctx                 context.Context        // Root context shared across all goroutines; used for graceful shutdown
	cancel              context.CancelFunc     // Cancel function tied to the root context
	wg                  sync.WaitGroup         // WaitGroup used to track and wait for the completion of all spawned goroutines
	coordinatorPort     string                 // Port on which the coordinator's gRPC server listens
	store               *taskdb.Store          // Task storage layer used to persist and retrieve task/job metadata
	httpServer          *http.Server           // gRPC server instance handling coordination service RPCs
	WorkerPool          map[uint32]*workerInfo // Map of registered workers, keyed by worker ID
	WorkerPoolMutex     sync.Mutex             // Mutex protecting access to the WorkerPool map
	WorkerPoolKeys      []uint32               // Slice of worker IDs used for consistent hashing or task distribution
	WorkerPoolKeysMutex sync.RWMutex           // RWMutex to allow concurrent reads of WorkerPoolKeys with safe writes
	dbPool              *pgxpool.Pool          // PostgreSQL connection pool for data persistence or coordination metadata
	maxMissedHeartbeats uint8                  // Max allowed missed heartbeats before a worker is considered inactive
	heartbeatInterval   time.Duration          // Expected interval between heartbeat messages from each worker
	roundRobinIndex     uint32                 // Index used for round-robin distribution of tasks to workers
}

func newService(coordinatorPort string, store *taskdb.Store) *coordinatorService {
	ctx, cancel := context.WithCancel(context.Background())
	return &coordinatorService{
		ctx:                 ctx,
		cancel:              cancel,
		coordinatorPort:     coordinatorPort,
		store:               store,
		WorkerPool:          make(map[uint32]*workerInfo),
		maxMissedHeartbeats: defaultMaxMisses,
		heartbeatInterval:   defaultHeartbeat,
	}
}

func (c *coordinatorService) Start() error {
	// Start HTTP server with Connect handlers
	if err := c.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	// Init worker pool
	go c.manageWorkerPool()
	// Function to get task to send off to workers
	go c.getAllExecutableTask()
	return c.awaitShutdown()
}

func (c *coordinatorService) startHTTPServer() error {
	mux := http.NewServeMux()
	path, handler := coordinatorv1connect.NewCoordinatorServiceHandler(c)
	mux.Handle(path, handler)
	// Create HTTP server with h2c support
	c.httpServer = &http.Server{
		Addr:         c.coordinatorPort,
		Handler:      h2c.NewHandler(mux, &http2.Server{}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	errChan := make(chan error, 1)
	go func() {
		log.Printf("ConnectRPC server listening on %s", c.coordinatorPort)
		if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("listen failed: %w", err)
		}
	}()
	// Wait for cancellation signal on the context.
	<-c.ctx.Done()
	log.Printf("ConnectRPC server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Attempt graceful shutdown within timeout.
	if err := c.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	// Return any listen error received.
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// awaitShutdown waits for a shutdown signal and then calls Stop.
func (c *coordinatorService) awaitShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutdown signal received")
	return c.Stop()
}

// Stop gracefully shuts down the server.
// Can be called directly for programmatic shutdown or via awaitShutdown for signal handling.
func (c *coordinatorService) Stop() error {
	log.Println("Initiating graceful shutdown...")

	// Signal all goroutines to stop
	c.cancel()

	// Shutdown HTTP server with timeout
	if c.httpServer != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := c.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}

	// Wait for all goroutines to finish
	c.wg.Wait()

	// Close all worker connections
	c.WorkerPoolMutex.Lock()
	for workerID, worker := range c.WorkerPool {
		if worker.httpClient != nil {
			worker.httpClient.CloseIdleConnections()
			log.Printf("Closed connections for worker %d", workerID)
		}
	}
	c.WorkerPoolMutex.Unlock()

	// Close database pool
	if c.dbPool != nil {
		c.dbPool.Close()
		log.Println("Database pool closed")
	}

	log.Println("Shutdown complete")
	return nil
}

func (c *coordinatorService) manageWorkerPool() {
	c.wg.Add(1)
	defer c.wg.Done()

	ticker := time.NewTicker(time.Duration(c.maxMissedHeartbeats) * c.heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.pruneWorkers()
		case <-c.ctx.Done():
			return
		}
	}
}

// pruneWorkers removes workers from the pool that have exceeded the maximum
// number of allowed missed heartbeats.
//
// It closes the worker's gRPC connection, deletes the worker from the pool,
// and refreshes the internal list of worker keys used for routing or scheduling.
//
// Active workers that have missed a heartbeat will have their miss count incremented.
func (c *coordinatorService) pruneWorkers() {
	c.WorkerPoolMutex.Lock()
	defer c.WorkerPoolMutex.Unlock()

	for workerID, worker := range c.WorkerPool {
		if worker.heartbeatMisses > c.maxMissedHeartbeats {
			log.Printf("Removing inactive worker: %d\n", workerID)

			// Close idle connections in the HTTP client
			if worker.httpClient != nil {
				worker.httpClient.CloseIdleConnections()
			}
			// Remove worker from pool
			delete(c.WorkerPool, workerID)
			// Update worker keys for round-robin
			c.WorkerPoolKeysMutex.Lock()
			workerCount := len(c.WorkerPool)
			c.WorkerPoolKeys = make([]uint32, 0, workerCount)
			for k := range c.WorkerPool {
				c.WorkerPoolKeys = append(c.WorkerPoolKeys, k)
			}
			c.WorkerPoolKeysMutex.Unlock()
		} else {
			worker.heartbeatMisses++
		}
	}
}

func (c *coordinatorService) getAllExecutableTask() {
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go c.processTasksForWorkers()
		case <-c.ctx.Done():
			log.Println("Finish db scan.")
			return
		}
	}
}

// processTasksForWorkers scans for tasks that are eligible to be executed, submits them to available workers,
// and updates their status in the database once they are successfully picked.
//
// It performs the following steps for each task:
//  1. Queries the task store for tasks that are ready to be executed.
//  2. Submits the task to an available worker using `submitTaskToWorker()`.
//  3. If the task is successfully submitted to a worker, it updates the task's `picked_at` timestamp
//     in the database using `UpdateTaskPickedAtByID()`.
//
// The function uses a 30-second timeout context to ensure database transactions complete in a timely manner.
// If any task fails to be submitted or updated, the error is logged and the function proceeds to the next task.
//
// This method should be invoked periodically to process new tasks and assign them to workers.
func (c *coordinatorService) processTasksForWorkers() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var tasks []taskpostgres.Task
	if err := c.store.ReadTx(ctx, func(tx *taskpostgres.Queries) error {
		var err error
		tasks, err = tx.GetAllExecutableTask(ctx)
		return err
	}); err != nil {
		log.Printf("Error executing query: %v\n", err)
	}

	for _, task := range tasks {
		// submit task to a worker
		receiveTaskRequest := &workerv1.ReceiveTaskRequest{
			TaskId:  task.ID.String(),
			Payload: task.Command,
		}
		if err := c.submitTaskToWorker(receiveTaskRequest); err != nil {
			log.Printf("Failed to submit task %s: %v\n", task.ID, err)
			continue
		}
		// if success update task `is_picked` in the db
		if err := c.store.ReadWriteTx(ctx, func(tx *taskpostgres.Queries) error {
			return tx.UpdateTaskPickedAtByID(ctx, task.ID)
		}); err != nil {
			log.Printf("Failed to update task %s: %v\n", task.ID, err)
			continue
		}
	}
}

// getNextWorker retrieves the next available worker in a round-robin fashion.
// It reads from the `WorkerPoolKeys` slice to determine the next worker based on
// the current `roundRobinIndex`. The worker is then returned for task assignment.
//
// It uses a read lock on `WorkerPoolKeysMutex` to ensure safe concurrent access
// to the list of worker keys. If no workers are available (i.e., `WorkerPoolKeys` is empty),
// it returns `nil`.
func (c *coordinatorService) getNextWorker() *workerInfo {
	c.WorkerPoolKeysMutex.RLock()
	defer c.WorkerPoolKeysMutex.RUnlock()

	workerCount := len(c.WorkerPoolKeys)
	if workerCount == 0 {
		return nil
	}
	worker := c.WorkerPool[c.WorkerPoolKeys[c.roundRobinIndex%uint32(workerCount)]]
	c.roundRobinIndex++
	return worker
}

// submitTaskToWorker submits a given task to the next available worker.
//
// This function first calls `getNextWorker()` to get the next worker from the
// `WorkerPool` using a round-robin mechanism. If no worker is available, it returns
// an error with the message "no available workers". If a worker is found, the task
// is submitted to that worker using the `ReceiveTask` RPC.
//
// It handles the task submission by sending a `ReceiveTaskRequest` to the worker's
// `workerServiceClient`. If the task is successfully submitted, the worker will begin
// processing the task; otherwise, an error is returned.
func (c *coordinatorService) submitTaskToWorker(task *workerv1.ReceiveTaskRequest) error {
	worker := c.getNextWorker()
	if worker == nil {
		return errors.New("no available workers")
	}
	taskRequest := connect.NewRequest(task)
	_, err := worker.workerServiceClient.ReceiveTask(context.Background(), taskRequest)
	return err
}
