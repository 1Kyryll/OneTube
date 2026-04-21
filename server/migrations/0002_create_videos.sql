-- +goose Up
CREATE TABLE videos (
    id               UUID PRIMARY KEY,
    uploader_id      UUID NOT NULL REFERENCES users(id),
    title            TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    duration_seconds INT,
    thumbnail_key    TEXT,
    visibility       TEXT NOT NULL DEFAULT 'public',
    status           TEXT NOT NULL DEFAULT 'uploading',
    view_count       BIGINT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at     TIMESTAMPTZ,
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX videos_uploader_idx ON videos (uploader_id);
CREATE INDEX videos_feed_idx ON videos (status, visibility, published_at DESC);

-- +goose Down
DROP TABLE videos;
