-- name: GetVideoByID :one
SELECT * FROM videos WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateVideo :one
INSERT INTO videos (id, uploader_id, title, description, visibility, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: SetVideoStatus :exec
UPDATE videos SET status = $2 WHERE id = $1;

-- name: MarkVideoReady :exec
UPDATE videos
SET status = 'ready',
    published_at = COALESCE(published_at, now()),
    duration_seconds = $2,
    thumbnail_key = $3
WHERE id = $1;

-- name: IncrementViewCount :exec
UPDATE videos SET view_count = view_count + 1 WHERE id = $1;

-- name: SoftDeleteVideo :exec
UPDATE videos SET deleted_at = now()
WHERE id = $1 AND uploader_id = $2 AND deleted_at IS NULL;

-- name: ListFeed :many
SELECT v.id, v.uploader_id, v.title, v.description, v.duration_seconds,
       v.thumbnail_key, v.visibility, v.status, v.view_count,
       v.created_at, v.published_at, v.deleted_at,
       u.username AS uploader_username,
       u.display_name AS uploader_display_name,
       u.avatar_key AS uploader_avatar_key
FROM videos v
JOIN users u ON u.id = v.uploader_id
WHERE v.status = 'ready'
  AND v.visibility = 'public'
  AND v.deleted_at IS NULL
  AND (
    @has_cursor::bool = false
    OR v.published_at < @cursor_published_at::timestamptz
    OR (v.published_at = @cursor_published_at::timestamptz AND v.id < @cursor_id::uuid)
  )
ORDER BY v.published_at DESC, v.id DESC
LIMIT @row_limit::int;

-- name: ListByUploader :many
SELECT * FROM videos
WHERE uploader_id = @uploader_id::uuid
  AND deleted_at IS NULL
  AND (
    @has_cursor::bool = false
    OR created_at < @cursor_created_at::timestamptz
    OR (created_at = @cursor_created_at::timestamptz AND id < @cursor_id::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT @row_limit::int;
