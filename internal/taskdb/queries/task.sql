-- name: CreateTask :one
INSERT INTO tasks (command, scheduled_at)
VALUES ($1, $2) RETURNING *;

-- name: UpdateTaskPickedAtByID :exec
UPDATE tasks SET picked_at = NOW() WHERE id = $1;