package main

import (
	"context"
	"flag"

	coordinatorservice "github.com/gilwong00/task-runner/internal/coordinator"
	"github.com/gilwong00/task-runner/internal/taskdb"
)

var (
	coordinatorPort = flag.String("coordinator_port", ":8080", "Port on which the Coordinator listens for incoming requests.")
)

func main() {
	flag.Parse()
	ctx := context.Background()
	config := taskdb.NewConfig("", "", "", 0, "task", 0)
	postgresConnection, err := taskdb.NewDB(ctx, config)
	if err != nil {
		panic(err)
	}
	defer postgresConnection.Pool.Close()
	store := taskdb.NewStore(
		postgresConnection.NewDB(),
	)
	coordinator := coordinatorservice.NewService(*coordinatorPort, store)
	if err := coordinator.Start(); err != nil {
		panic(err)
	}
}
