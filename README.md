# OneTube

A small YouTube-style demo: signup, login, upload a video, transcode it to multi-bitrate HLS in the background, and watch it from a public feed with a quality selector.

The point of this project is the architecture — direct-to-S3 uploads, async transcoding via a queue, and HLS playback against a private bucket — not a polished product.

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.25 (`net/http`, `pgx`/`pgxpool`, `sqlc`-generated queries, `golang-migrate`) |
| Worker  | Go + `ffmpeg` / `ffprobe` (separate binary, same module) |
| Frontend | Next.js (App Router), React 18, hls.js |
| Database | Postgres 16 |
| Object storage | MinIO (S3-compatible, private bucket) |
| Queue | RabbitMQ |
| Auth | JWT (HS256) in an `HttpOnly; Secure; SameSite=Lax` cookie, 24h |
| Streaming | HLS — H.264 in MPEG-TS segments, played via hls.js |

## How it works

```
              ┌──────────────────────────────────────────────────────┐
              │                       Browser                         │
              └───────┬────────────────────────────────┬──────────────┘
                      │                                │
   1. POST /api/videos│                       6. GET master.m3u8 →
   2. PUT raw → S3────┼───────────────┐       rendition.m3u8 →
   3. POST .../complete                │      .ts segments
                      │                ▼
              ┌───────▼─────────┐  ┌───────────┐
              │   Go API (8080) │  │  MinIO    │
              │                 │◄─┤  (9000)   │  ← presigned PUT/GET
              │  publishes job  │  └─────┬─────┘
              └────────┬────────┘        ▲
                       │                 │ 5. uploads thumb,
                       ▼                 │    rendition .m3u8 + .ts,
                  ┌──────────┐           │    master.m3u8
                  │ RabbitMQ │           │
                  └────┬─────┘           │
                       │                 │
                       ▼                 │
              ┌──────────────────┐       │
              │   Go Worker      │───────┘
              │   (ffmpeg)       │
              │ 4. download raw, │
              │    transcode,    │
              │    upload HLS    │
              └──────────────────┘
```

### Upload

1. Client requests an upload intent — server inserts a `Video` row with `status=uploading` and returns a presigned PUT URL.
2. Client uploads the raw file **directly to MinIO**.
3. Client calls `complete`. Server flips status to `processing` and publishes a transcode job to RabbitMQ.

### Transcode (worker)

4. Worker pulls the job, downloads the raw file, runs `ffprobe` for source dimensions, generates a JPEG thumbnail.
5. For each rendition (360p / 720p / 1080p, skipping any taller than the source) it runs `ffmpeg` to produce an HLS playlist + `.ts` segments, uploads them to S3, upserts a `VideoAsset` row.
6. Builds `master.m3u8`, uploads it, marks the video `ready`.

The worker is **idempotent** — re-running a job overwrites the same S3 keys and upserts the same asset rows.

### Playback

The bucket is private. To make HLS work without leaking object access, the API serves the playlist files itself and signs every segment URL on the fly:

1. `GET /api/videos/{id}` returns `master_playlist_url` pointing to the API.
2. hls.js fetches `GET /api/videos/{id}/hls/master.m3u8` — server reads from S3, returns body unchanged. Variant entries stay relative.
3. hls.js resolves variants to `GET /api/videos/{id}/hls/{quality}/playlist.m3u8` — server reads from S3 and **rewrites every `.ts` line into a presigned MinIO URL** (4h TTL).
4. hls.js fetches `.ts` segments **directly from MinIO** with valid signatures.

The API never touches video bytes — only tiny text manifests pass through it.

## Project layout

```
.
├── client/                       # Next.js app
│   ├── app/                      # App Router pages: /, /login, /signup, /upload, /watch/[id]
│   └── features/                 # Server actions, components, types per feature
├── server/
│   ├── cmd/
│   │   ├── api/                  # HTTP API entrypoint
│   │   └── worker/               # Transcode worker entrypoint
│   ├── internal/
│   │   ├── auth/                 # JWT middleware, cookies, password hashing
│   │   ├── common/
│   │   │   ├── gen/              # sqlc-generated queries
│   │   │   ├── queue/            # RabbitMQ publisher / consumer
│   │   │   └── s3/               # MinIO client + key helpers + presigning
│   │   ├── config/
│   │   ├── transcode/            # ffmpeg pipeline, master playlist builder
│   │   ├── user/                 # signup, login, profile
│   │   └── video/                # upload intent, watch, feed, playlists
│   ├── migrations/               # golang-migrate SQL files
│   └── queries/                  # sqlc input
├── docs/plans/                   # Design notes per feature
├── docker-compose.yaml           # Postgres, MinIO, RabbitMQ
└── README.md
```

## Data model

**users** — `id`, `email` (unique), `password_hash` (argon2id), `username` (unique), `display_name`, `avatar_key`, `created_at`

**videos** — `id`, `uploader_id`, `title`, `description`, `duration_seconds`, `thumbnail_key`, `visibility` (`public`/`unlisted`/`private`), `status` (`uploading`/`processing`/`ready`/`failed`), `view_count`, `created_at`, `published_at`, `deleted_at`

**video_assets** — `id`, `video_id`, `quality`, `codec`, `container`, `storage_key`, `file_size_bytes`, `bitrate`, `width`, `height`

Indexes:
- `videos (uploader_id)`
- `videos (status, visibility, published_at DESC)` — feed
- `video_assets (video_id)`
- `users (email)` unique, `users (username)` unique

## S3 layout

```
raw/{video_id}/original.{ext}             # lifecycle-deleted 7d after transcode
hls/{video_id}/master.m3u8
hls/{video_id}/{quality}/playlist.m3u8
hls/{video_id}/{quality}/seg_NNN.ts
thumbnails/{video_id}/default.jpg
avatars/{user_id}/avatar.jpg
```

## API

All routes under `/api`. Auth = JWT cookie `onetube_session`.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/auth/signup` | — | Create user, set cookie |
| POST | `/auth/login` | — | Verify credentials, set cookie |
| POST | `/auth/logout` | — | Clear cookie |
| GET  | `/auth/me` | ✓ | Return current user |
| POST | `/videos` | ✓ | Create upload intent (returns presigned PUT URL) |
| POST | `/videos/{id}/complete` | ✓ | Mark uploaded → enqueue transcode |
| GET  | `/videos/{id}` | — | Watch info (or `{status: "processing"}`) |
| GET  | `/videos/{id}/hls/master.m3u8` | — | Master playlist |
| GET  | `/videos/{id}/hls/{quality}/playlist.m3u8` | — | Rendition playlist with presigned segments |
| POST | `/videos/{id}/view` | — | Increment view count |
| GET  | `/videos/feed` | — | Public feed, cursor-paginated |
| GET  | `/videos/mine` | ✓ | Caller's uploads |
| DELETE | `/videos/{id}` | ✓ | Soft delete (ownership-checked) |

## Running locally

### Prerequisites

- Go 1.25+
- Node 20+
- Docker
- `ffmpeg` / `ffprobe` on PATH (worker shells out to them)
- `migrate` CLI (`brew install golang-migrate` or equivalent) and `sqlc` if regenerating queries

### 1. Bring up infra

```bash
cp .env.example .env   # fill in DB_*, JWT_SECRET, etc.
docker compose up -d
```

That gives you Postgres on `:5433`, MinIO on `:9000` (console `:9001`, login `minioadmin` / `minioadmin`), RabbitMQ on `:5672` (UI `:15672`).

Create the bucket once (MinIO console → Buckets → Create → name `onetube`, keep it private).

### 2. Run migrations

```bash
cd server
migrate -path migrations -database "$DATABASE_URL" up
```

### 3. Start the API

```bash
cd server
go run ./cmd/api
```

Listens on `:8080`. Required env:

| Var | Example |
|---|---|
| `DATABASE_URL` | `postgres://onetube:onetube@localhost:5433/onetube?sslmode=disable` |
| `JWT_SECRET` | any 32+ char string |
| `S3_ENDPOINT` | `localhost:9000` |
| `S3_PUBLIC_ENDPOINT` | `localhost:9000` (what the **browser** uses; differs in prod) |
| `S3_ACCESS_KEY` / `S3_SECRET_KEY` | `minioadmin` / `minioadmin` |
| `S3_BUCKET` | `onetube` |
| `S3_USE_SSL` | `false` |
| `API_PUBLIC_BASE_URL` | `http://localhost:8080` (used to build master-playlist URLs) |
| `RABBIT_URL` | `amqp://guest:guest@localhost:5672/` |
| `TRANSCODE_QUEUE` | `transcode.jobs` |
| `PORT` | `8080` |

### 4. Start the worker

In another terminal, with the same env:

```bash
cd server
go run ./cmd/worker
```

### 5. Start the frontend

```bash
cd client
npm install
npm run dev
```

Open `http://localhost:3000`.

## Design rules

These are non-negotiable for any change in this repo:

- **The API never touches video bytes.** Uploads go client → S3, segments go S3 → client, both via presigned URLs. Only tiny text manifests pass through the API.
- **Ownership checked on every mutation:** `WHERE id = ? AND uploader_id = ?`. No role-based auth.
- **Cursor pagination** on `(published_at, id)` or `(created_at, id)`. Never `OFFSET`.
- **JOIN to avoid N+1.** The feed query joins `videos` with `users` for the uploader name.
- **The worker is idempotent.** Reprocessing a job must be safe.
- **Soft delete** videos via `deleted_at`. A separate cleanup job removes S3 objects later.
- **`view_count` is denormalized** on the video row; incremented on view.
- **JWT lives in an HttpOnly cookie.** Never in localStorage.
- **`context.Context` is the first arg** on any Go function doing I/O.

## Known gaps

This is a demo, not production. In particular:

- No rate limiting, captcha, or email verification.
- `unlisted` and `private` visibility are stored but not enforced on the playlist endpoints — any URL with a valid `{id}` can fetch the master.
- View counts are incremented on first play with no deduplication.
- MinIO CORS for `Range` headers may need configuring if you proxy MinIO behind a non-default origin.
- No tests yet.
