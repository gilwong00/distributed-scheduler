// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: task.sql

package taskpostgres

import (
	"context"
	"time"
)

const createTask = `-- name: CreateTask :one
INSERT INTO tasks (command, scheduled_at)
VALUES ($1, $2) RETURNING id, command, created_at, scheduled_at, picked_at, started_at, completed_at, failed_at
`

type CreateTaskParams struct {
	Command     string
	ScheduledAt time.Time
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	row := q.db.QueryRowContext(ctx, createTask, arg.Command, arg.ScheduledAt)
	var i Task
	err := row.Scan(
		&i.ID,
		&i.Command,
		&i.CreatedAt,
		&i.ScheduledAt,
		&i.PickedAt,
		&i.StartedAt,
		&i.CompletedAt,
		&i.FailedAt,
	)
	return i, err
}
