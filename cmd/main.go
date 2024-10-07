package main

import (
	"context"
	"fmt"

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
	fmt.Println(">>>> store", store)
}
