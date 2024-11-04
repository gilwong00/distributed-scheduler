package main

import (
	"flag"

	workerservice "github.com/gilwong00/task-runner/internal/worker"
)

var (
	workerPort      = flag.String("worker_port", "", "Port on which worker process the task.")
	coordinatorPort = flag.String("coordinator", ":8080", "Network address of the coordinator node.")
)

func main() {
	flag.Parse()
	workerService := workerservice.NewService(*workerPort, *coordinatorPort)
	workerService.Start()
}
