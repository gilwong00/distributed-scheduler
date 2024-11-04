-- name: CreateTask :one
INSERT INTO tasks (command, scheduled_at)
VALUES ($1, $2) RETURNING *;
