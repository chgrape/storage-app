

-- name: GetRecord :one
SELECT * FROM records
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListRecords :many
SELECT * FROM records WHERE user_id = $1;

-- name: CreateRecord :one
INSERT INTO records (
  id, filename, mime_type, size, path, uploaded_at, user_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: DeleteRecord :exec
DELETE FROM records
WHERE id = $1 AND user_id = $2;

