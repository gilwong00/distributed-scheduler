-- name: GetTasks :many
SELECT * FROM tasks;

-- name: GetTaskByID :one
SELECT * from tasks
WHERE id = $1;
