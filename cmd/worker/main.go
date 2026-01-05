package main

import (
	"flag"

	workerservice "github.com/gilwong00/task-runner/internal/worker"
)

var (
	workerPort      = flag.String("worker_port", "", "Port on which worker process the task.")
	coordinatorHost = flag.String("coordinator_host", "", "Coordinator host")
	coordinatorPort = flag.String("coordinator_port", ":8080", "Network address of the coordinator node.")
)

func main() {
	flag.Parse()
	workerService := workerservice.NewService(
		*workerPort,
		*coordinatorHost,
		*coordinatorPort,
	)
	if err := workerService.Start(); err != nil {
		panic(err)
	}
}
