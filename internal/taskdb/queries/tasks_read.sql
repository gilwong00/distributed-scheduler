-- name: GetTasks :many
SELECT * FROM tasks;

-- name: GetTaskByID :one
SELECT * from tasks
WHERE id = $1;

-- name: GetAllExecutableTask :many
SELECT * FROM tasks
WHERE scheduled_at < (NOW() + INTERVAL '30 seconds') AND picked_at IS NULL
ORDER BY scheduled_at
FOR UPDATE SKIP LOCKED;