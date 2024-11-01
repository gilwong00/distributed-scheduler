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
	config := taskdb.NewConfig("", "", "", 0, "task", 0)
	postgresConnection, err := taskdb.NewDB(ctx, config)
	if err != nil {
		panic(err)
	}
	defer postgresConnection.Pool.Close()
	store := taskdb.NewStore(
		postgresConnection.NewDB(),
	)
	schedulerservice := schedulerservice.NewSchedulerService(5001, store)
	if err := schedulerservice.StartHttpServer(); err != nil {
		log.Fatalf("Error while starting server: %+v", err)
	}
}
