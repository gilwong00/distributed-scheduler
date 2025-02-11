package coordinator

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gilwong00/task-runner/internal/taskdb"
	pb "github.com/gilwong00/task-runner/proto/gen"
	"google.golang.org/grpc"
)

type CoordinatorServer struct {
	pb.UnimplementedCoordinatorServiceServer
	ctx             context.Context    // The root context for all goroutines
	cancel          context.CancelFunc // Function to cancel the context
	wg              sync.WaitGroup     // WaitGroup to wait for all goroutines to finish
	coordinatorPort string
	store           *taskdb.Store
	listener        net.Listener
	grpcServer      *grpc.Server
}

func NewServer(coordinatorPort string, store *taskdb.Store) *CoordinatorServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CoordinatorServer{
		ctx:             ctx,
		cancel:          cancel,
		coordinatorPort: coordinatorPort,
		store:           store,
	}
}

// Start initiates the server's operations.
func (s *CoordinatorServer) Start() error {
	var err error
	if err = s.startGRPCServer(); err != nil {
		return fmt.Errorf("gRPC server start failed: %w", err)
	}
	return s.awaitShutdown()
}

func (s *CoordinatorServer) startGRPCServer() error {
	var err error
	s.listener, err = net.Listen("tcp", s.coordinatorPort)
	if err != nil {
		return err
	}
	log.Printf("Starting gRPC server on %s\n", s.coordinatorPort)
	s.grpcServer = grpc.NewServer()
	pb.RegisterCoordinatorServiceServer(s.grpcServer, s)

	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()
	return nil
}

func (s *CoordinatorServer) awaitShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// TODO implement a stop method that closes the grpc service, db connection and
	// flushes all go routines
	return nil
}
