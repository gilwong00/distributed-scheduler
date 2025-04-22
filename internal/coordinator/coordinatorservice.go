package coordinatorservice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gilwong00/task-runner/internal/taskdb"
	pb "github.com/gilwong00/task-runner/proto/gen"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"

	taskpostgres "github.com/gilwong00/task-runner/internal/taskdb/gen"
)

const (
	scanInterval     = 10 * time.Second
	defaultMaxMisses = 1
	defaultHeartbeat = 5 * time.Second
)

// CoordinatorServer manages worker registration, heartbeat tracking, and task coordination
// in a distributed scheduling system. It also handles gRPC server lifecycle and interacts
// with a backing task store and database.
type CoordinatorServer struct {
	pb.UnimplementedCoordinatorServiceServer                        // Embeds unimplemented methods for forward compatibility with gRPC
	ctx                                      context.Context        // Root context shared across all goroutines; used for graceful shutdown
	cancel                                   context.CancelFunc     // Cancel function tied to the root context
	wg                                       sync.WaitGroup         // WaitGroup used to track and wait for the completion of all spawned goroutines
	coordinatorPort                          string                 // Port on which the coordinator's gRPC server listens
	store                                    *taskdb.Store          // Task storage layer used to persist and retrieve task/job metadata
	listener                                 net.Listener           // Network listener for accepting incoming gRPC connections
	grpcServer                               *grpc.Server           // gRPC server instance handling coordination service RPCs
	WorkerPool                               map[uint32]*workerInfo // Map of registered workers, keyed by worker ID
	WorkerPoolMutex                          sync.Mutex             // Mutex protecting access to the WorkerPool map
	WorkerPoolKeys                           []uint32               // Slice of worker IDs used for consistent hashing or task distribution
	WorkerPoolKeysMutex                      sync.RWMutex           // RWMutex to allow concurrent reads of WorkerPoolKeys with safe writes
	dbPool                                   *pgxpool.Pool          // PostgreSQL connection pool for data persistence or coordination metadata
	maxMissedHeartbeats                      uint8                  // Max allowed missed heartbeats before a worker is considered inactive
	heartbeatInterval                        time.Duration          // Expected interval between heartbeat messages from each worker
	roundRobinIndex                          uint32                 // Index used for round-robin distribution of tasks to workers
}

func NewServer(coordinatorPort string, store *taskdb.Store) *CoordinatorServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoordinatorServer{
		ctx:                 ctx,
		cancel:              cancel,
		coordinatorPort:     coordinatorPort,
		store:               store,
		WorkerPool:          make(map[uint32]*workerInfo),
		maxMissedHeartbeats: defaultMaxMisses,
		heartbeatInterval:   defaultHeartbeat,
	}
}

// Start initiates the server's operations.
func (c *CoordinatorServer) Start() error {
	var err error
	// init worker pool
	go c.manageWorkerPool()

	if err = c.startGRPCServer(); err != nil {
		return fmt.Errorf("gRPC server start failed: %w", err)
	}

	// function to get task to send off to workers
	go c.getAllExecutableTask()

	return c.awaitShutdown()
}

func (c *CoordinatorServer) manageWorkerPool() {
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
func (c *CoordinatorServer) pruneWorkers() {
	c.WorkerPoolMutex.Lock()
	defer c.WorkerPoolMutex.Unlock()

	for workerID, worker := range c.WorkerPool {
		if worker.heartbeatMisses > c.maxMissedHeartbeats {
			log.Printf("Removing inactive worker: %d\n", workerID)
			worker.grpcConnection.Close()
			delete(c.WorkerPool, workerID)

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

func (c *CoordinatorServer) getAllExecutableTask() {
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
func (c *CoordinatorServer) processTasksForWorkers() {
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
		receiveTaskRequest := &pb.ReceiveTaskRequest{
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
func (c *CoordinatorServer) getNextWorker() *workerInfo {
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
func (c *CoordinatorServer) submitTaskToWorker(task *pb.ReceiveTaskRequest) error {
	worker := c.getNextWorker()
	if worker == nil {
		return errors.New("no available workers")
	}
	_, err := worker.workerServiceClient.ReceiveTask(context.Background(), task)
	return err
}
