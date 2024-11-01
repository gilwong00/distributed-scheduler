package main

import (
	"context"
	"log"

	schedulerservice "github.com/gilwong00/task-runner/internal/scheduler"
	"github.com/gilwong00/task-runner/internal/taskdb"
)

func main() {
	// TODO: get configs
	ctx := context.Background()
	postgresConnection, err := taskdb.NewDB(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer postgresConnection.Pool.Close()
	store := taskdb.NewStore(
		postgresConnection.NewDB(),
	)
	schedulerservice := schedulerservice.NewSchedulerService(5000, store)
	if err := schedulerservice.Start(); err != nil {
		log.Fatalf("Error while starting server: %+v", err)
	}
}
