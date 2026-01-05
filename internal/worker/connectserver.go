// package workerservice

// import (
// 	"fmt"
// 	"log"
// 	"net"

// 	pb "github.com/gilwong00/task-runner/proto/gen"

// 	"google.golang.org/grpc"
// )

// func (w *WorkerService) startGRPCServer() error {
// 	var err error
// 	if w.workerServerPort == "" {
// 		// Find a free port using a temporary socket
// 		// Bind to any available port
// 		w.listener, err = net.Listen("tcp", ":0")
// 		// Get the assigned port
// 		w.workerServerPort = fmt.Sprintf(":%d", w.listener.Addr().(*net.TCPAddr).Port)
// 	} else {
// 		w.listener, err = net.Listen("tcp", w.workerServerPort)
// 	}
// 	if err != nil {
// 		return fmt.Errorf("failed to listen on port: %s: %w", w.workerServerPort, err)
// 	}

// 	log.Printf("Starting worker server on %s\n", w.workerServerPort)
// 	w.grpcServer = grpc.NewServer()
// 	pb.RegisterWorkerServiceServer(w.grpcServer, w)

// 	go func() {
// 		if err := w.grpcServer.Serve(w.listener); err != nil {
// 			log.Fatalf("worker gRPC server failed: %v", err)
// 		}
// 	}()
// 	return nil
// }

// func (w *WorkerService) closeGRPCConnection() {
// 	if w.grpcServer != nil {
// 		w.grpcServer.GracefulStop()
// 	}
// 	if w.listener != nil {
// 		if err := w.listener.Close(); err != nil {
// 			log.Printf("Error while closing the listener: %v", err)
// 		}
// 	}
// 	// terminate connection with coordinator node
// 	if err := w.coordinatorClient.Close(); err != nil {
// 		log.Printf("Error while closing client connection with coordinator: %v", err)
// 	}
// }
