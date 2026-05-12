

-- name: GetRecord :one
SELECT * FROM records
WHERE id = $1 LIMIT 1;

-- name: ListRecords :many
SELECT * FROM records
ORDER BY created_at;

-- name: CreateRecord :exec
INSERT INTO records (
  filename, mime_type, size, path, created_at
) VALUES (
  $1, $2, $3, $4, $5
);

-- name: DeleteRecord :exec
DELETE FROM records
WHERE id = $1;

