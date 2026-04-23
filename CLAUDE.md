# CLAUDE.md

Demo YouTube-like app: signup/login, upload, watch with quality selection (360p/720p/1080p), public feed, profile picture.

## Stack

- Backend: Go (`pgx`/`pgxpool`, `golang-migrate`)
- Frontend: Next.js (App Router)
- DB: Postgres
- Object storage: MinIO S3
- Queue: RabbitMQ
- Streaming: HLS (H.264 in fragmented MP4), played via hls.js
- Auth: JWT (HS256) in `HttpOnly; Secure; SameSite=Lax` cookie, 24h expiry

## Flow

1. Client gets JWT cookie from login
2. Client requests pre-signed S3 PUT URL from API, uploads raw file **directly to S3**
3. API inserts Video (`status=processing`), publishes transcode job to RabbitMQ
4. Worker runs ffmpeg → HLS renditions to S3 → inserts VideoAssets → sets `status=ready`
5. Client polls status; on ready, plays master `.m3u8` via hls.js

## Models

**User** — `id`, `email` (unique), `password_hash` (argon2id), `username` (unique), `display_name`, `avatar_key`, `created_at`

**Video** — `id`, `uploader_id` (FK User), `title`, `description`, `duration_seconds`, `thumbnail_key`, `visibility` (public/unlisted/private), `status` (uploading/processing/ready/failed), `view_count`, `created_at`, `published_at`, `deleted_at`

**VideoAsset** — `id`, `video_id` (FK Video), `quality`, `codec`, `container`, `storage_key`, `file_size_bytes`, `bitrate`, `width`, `height`

## Indexes

- `videos (uploader_id)`
- `videos (status, visibility, published_at DESC)`
- `video_assets (video_id)`
- `users (email)` unique
- `users (username)` unique

## S3 Layout

```
raw/{video_id}/original.{ext}             # lifecycle-deleted 7d after transcode
hls/{video_id}/master.m3u8
hls/{video_id}/{quality}/playlist.m3u8
hls/{video_id}/{quality}/seg_NNN.ts
thumbnails/{video_id}/default.jpg
avatars/{user_id}/avatar.jpg
```

Bucket is private. All client reads go through pre-signed GET URLs (1–4h expiry).

## Rules

- **API never touches video bytes.** Uploads and downloads go client ↔ S3 via pre-signed URLs.
- **Ownership checks on every mutation:** `WHERE id = ? AND uploader_id = ?`. No role-based auth.
- **Cursor pagination** on `(published_at, id)` or `(created_at, id)`. Never OFFSET.
- **JOIN to avoid N+1** — feed joins Videos with Users for uploader username.
- **Worker is idempotent** — reprocessing a job must be safe (overwrite S3 keys, upsert VideoAssets).
- **Soft delete videos** (`deleted_at`); a cleanup job removes S3 objects later.
- **`view_count` is denormalized** on Video; incremented on view.
- **JWT never in localStorage.** Always HttpOnly cookie.
- **`context.Context` is the first arg** on any Go function doing I/O.
