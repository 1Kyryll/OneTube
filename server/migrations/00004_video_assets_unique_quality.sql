-- +goose Up
CREATE UNIQUE INDEX video_assets_video_quality_uidx ON video_assets (video_id, quality); 

-- +goose Down
DROP INDEX video_assets_video_quality_uidx;
