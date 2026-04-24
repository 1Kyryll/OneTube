-- name: ListAssetsByVideo :many
SELECT * FROM video_assets WHERE video_id = $1;

-- name: UpsertVideoAsset :exec
INSERT INTO video_assets (id, video_id, quality, codec, container, storage_key, file_size_bytes, bitrate, width, height)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    storage_key = EXCLUDED.storage_key,
    file_size_bytes = EXCLUDED.file_size_bytes,
    bitrate = EXCLUDED.bitrate,
    width = EXCLUDED.width,
    height = EXCLUDED.height;

-- name: DeleteAssetsByVideo :exec
DELETE FROM video_assets WHERE video_id = $1;