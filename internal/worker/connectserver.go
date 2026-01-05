package workerservice

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	coordinatorv1 "github.com/gilwong00/task-runner/internal/gen/coordinator/v1"
	"github.com/gilwong00/task-runner/internal/gen/coordinator/v1/coordinatorv1connect"
	workerv1 "github.com/gilwong00/task-runner/internal/gen/worker/v1"
	"github.com/gilwong00/task-runner/internal/gen/worker/v1/workerv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type workerService struct {
	id                    uint32
	ctx                   context.Context
	workerHost            string
	workerServerPort      string
	coordinatorHost       string
	coordinatorPort       string
	coordinatorHTTPClient *http.Client
	coordinatorClient     coordinatorv1connect.CoordinatorServiceClient
	httpServer            *http.Server
	cancel                context.CancelFunc
	wg                    sync.WaitGroup // WaitGroup to wait for all goroutines to finish
	taskQueue             chan *workerv1.ReceiveTaskRequest
	listener              net.Listener
	heartbeatInterval     time.Duration
}

func newService(
	workerServerPort string,
	coordinatorHost string,
	coordinatorPort string,
) *workerService {
	ctx, cancel := context.WithCancel(context.Background())
	return &workerService{
		workerServerPort:  workerServerPort,
		coordinatorHost:   coordinatorHost,
		coordinatorPort:   coordinatorPort,
		heartbeatInterval: 5 * time.Second,
		ctx:               ctx,
		cancel:            cancel,
		taskQueue:         make(chan *workerv1.ReceiveTaskRequest, 100), // queue to process all tasks received by the workers
	}
}

func (w *workerService) Start() error {
	w.startWorkerPool(maxWorkerPoolSize)
	if err := w.connectToCoordinator(); err != nil {
		return fmt.Errorf("failed to connect to coordinator: %w", err)
	}
	go w.emitHeartbeat()
	if err := w.startHTTPServer(); err != nil {
		return fmt.Errorf("connect server start failed: %w", err)
	}
	return w.awaitShutdown()
}

// startHTTPServer initializes and starts the HTTP server with Connect handlers
func (w *workerService) startHTTPServer() error {
	mux := http.NewServeMux()
	path, handler := workerv1connect.NewWorkerServiceHandler(w)
	mux.Handle(path, handler)
	// If no port specified, find an available one
	addr := w.workerServerPort
	if addr == "" {
		// Listen on a random available port
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			return fmt.Errorf("failed to find available port: %w", err)
		}
		addr = listener.Addr().String()
		w.workerServerPort = fmt.Sprintf(":%d", listener.Addr().(*net.TCPAddr).Port)
		listener.Close() // Close temporary listener
	}
	// Create HTTP server with h2c support
	w.httpServer = &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(mux, &http2.Server{}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting worker server on %s\n", w.workerServerPort)
	// Start server in goroutine
	go func() {
		if err := w.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Worker HTTP server failed: %v", err)
		}
	}()
	return nil
}

func (w *workerService) connectToCoordinator() error {
	log.Println("Connecting to coordinator...")
	// Create HTTP client with HTTP/2 support
	w.coordinatorHTTPClient = &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true, // Allow HTTP/2 without TLS (equivalent to insecure credentials)
			// For production with TLS, configure TLSClientConfig here
		},
		Timeout: 30 * time.Second,
	}
	// Create Connect client
	coordinatorURL := fmt.Sprintf("http://%s:%s", w.coordinatorHost, w.coordinatorPort)
	w.coordinatorClient = coordinatorv1connect.NewCoordinatorServiceClient(
		w.coordinatorHTTPClient,
		coordinatorURL,
	)
	log.Println("Connected to coordinator!")
	return nil
}

// startWorkerPool starts a pool of worker goroutines.
func (w *workerService) startWorkerPool(totalWorkers int) {
	for range totalWorkers {
		w.wg.Add(1)
		go w.worker()
	}
}

func (w *workerService) awaitShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	w.stop()
	return nil
}

// worker is the function run by each worker goroutine.
func (w *workerService) worker() {
	defer w.wg.Done() // Signal this worker is done when the function returns.

	// TODO: implement logic
}

func (w *workerService) stop() error {
	log.Println("Initiating worker shutdown...")
	// Signal all goroutines to stop
	w.cancel()
	// Shutdown HTTP server with timeout
	if w.httpServer != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := w.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}
	// Wait for all goroutines to finish
	w.wg.Wait()
	// Close HTTP client connections to coordinator
	if w.coordinatorHTTPClient != nil {
		w.coordinatorHTTPClient.CloseIdleConnections()
		log.Println("Closed coordinator connections")
	}
	// Close task queue if needed
	if w.taskQueue != nil {
		close(w.taskQueue)
	}
	log.Println("Worker node stopped")
	return nil
}

func (w *workerService) sendHeartbeat() error {
	req := connect.NewRequest(&coordinatorv1.SendHeartbeatRequest{
		WorkerId: w.id,
		Address:  w.workerHost,
	})
	// TODO: if this was a production system we should log the ack
	_, err := w.coordinatorClient.SendHeartbeat(w.ctx, req)
	return err
}

func (w *workerService) emitHeartbeat() {
	w.wg.Add(1)
	defer w.wg.Done()
	ticker := time.NewTicker(w.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.sendHeartbeat(); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
				return
			}
		case <-w.ctx.Done():
			return
		}
	}
}

// TODO: implement process task method
