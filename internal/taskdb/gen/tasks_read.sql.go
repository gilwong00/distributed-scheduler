// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: tasks_read.sql

package taskpostgres

import (
	"context"

	uuid "github.com/gofrs/uuid/v5"
)

const getTaskByID = `-- name: GetTaskByID :one
SELECT id, command, created_at, scheduled_at, picked_at, started_at, completed_at, failed_at from tasks
WHERE id = $1
`

func (q *Queries) GetTaskByID(ctx context.Context, id uuid.UUID) (Task, error) {
	row := q.db.QueryRowContext(ctx, getTaskByID, id)
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

const getTasks = `-- name: GetTasks :many
SELECT id, command, created_at, scheduled_at, picked_at, started_at, completed_at, failed_at FROM tasks
`

func (q *Queries) GetTasks(ctx context.Context) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, getTasks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Task{}
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.Command,
			&i.CreatedAt,
			&i.ScheduledAt,
			&i.PickedAt,
			&i.StartedAt,
			&i.CompletedAt,
			&i.FailedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
