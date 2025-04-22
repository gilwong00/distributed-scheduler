package coordinatorservice

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/gilwong00/task-runner/proto/gen"

	"google.golang.org/grpc"
)

type workerInfo struct {
	heartbeatMisses     uint8
	address             string
	grpcConnection      *grpc.ClientConn
	workerServiceClient pb.WorkerServiceClient
}

func (c *CoordinatorServer) startGRPCServer() error {
	var err error
	c.listener, err = net.Listen("tcp", c.coordinatorPort)
	if err != nil {
		return err
	}
	log.Printf("Starting gRPC server on %s\n", c.coordinatorPort)
	c.grpcServer = grpc.NewServer()
	pb.RegisterCoordinatorServiceServer(c.grpcServer, c)

	go func() {
		if err := c.grpcServer.Serve(c.listener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()
	return nil
}

func (c *CoordinatorServer) awaitShutdown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	return c.Stop()
}

// Stop gracefully shuts down the server.
func (c *CoordinatorServer) Stop() error {
	// Signal all goroutines to stop
	c.cancel()
	// Wait for all goroutines to finish
	c.wg.Wait()
	c.WorkerPoolMutex.Lock()
	defer c.WorkerPoolMutex.Unlock()

	for _, worker := range c.WorkerPool {
		if worker.grpcConnection != nil {
			worker.grpcConnection.Close()
		}
	}
	if c.grpcServer != nil {
		c.grpcServer.GracefulStop()
	}
	if c.listener != nil {
		return c.listener.Close()
	}
	// might not need this
	c.dbPool.Close()
	return nil
}
