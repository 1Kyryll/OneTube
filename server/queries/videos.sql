-- name: GetVideoByID :one
SELECT * FROM videos WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateVideo :one
INSERT INTO videos (id, uploader_id, title, description, visibility, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: SetVideoStatus :exec
UPDATE videos SET status = $2 WHERE id = $1;

-- name: IncrementViewCount :exec
UPDATE videos SET view_count = view_count + 1 WHERE id = $1;
