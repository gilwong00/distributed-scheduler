// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package taskpostgres

import (
	"context"
)

type Querier interface {
	GetTasks(ctx context.Context) ([]Task, error)
}

var _ Querier = (*Queries)(nil)
