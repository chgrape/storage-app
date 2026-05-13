

-- name: GetRecord :one
SELECT * FROM records
WHERE id = $1 LIMIT 1;

-- name: ListRecords :many
SELECT * FROM records;

-- name: CreateRecord :one
INSERT INTO records (
  id, filename, mime_type, size, path, uploaded_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: DeleteRecord :exec
DELETE FROM records
WHERE id = $1;

