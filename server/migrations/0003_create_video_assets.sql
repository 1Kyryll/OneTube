-- +goose Up
CREATE TABLE video_assets (
    id              UUID PRIMARY KEY,
    video_id        UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    quality         TEXT NOT NULL,
    codec           TEXT NOT NULL,
    container       TEXT NOT NULL,
    storage_key     TEXT NOT NULL,
    file_size_bytes BIGINT,
    bitrate         INT,
    width           INT,
    height          INT
);

CREATE INDEX video_assets_video_idx ON video_assets (video_id);

-- +goose Down
DROP TABLE video_assets;
