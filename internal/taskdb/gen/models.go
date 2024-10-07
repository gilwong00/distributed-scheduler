// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package taskpostgres

import (
	"database/sql"
	"time"

	uuid "github.com/gofrs/uuid/v5"
)

type Task struct {
	ID          uuid.UUID
	Command     string
	CreatedAt   time.Time
	ScheduledAt time.Time
	PickedAt    sql.NullTime
	StartedAt   sql.NullTime
	CompletedAt sql.NullTime
	FailedAt    sql.NullTime
}
