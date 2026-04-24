package types

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/1kyryll/onetube/server/internal/common/gen"
)

type CreateVideoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

type CreateVideoResponse struct {
	VideoID    uuid.UUID `json:"video_id"`
	UploadURL  string    `json:"upload_url"`
	StorageKey string    `json:"storage_key"`
	ExpiresIn  int       `json:"expires_in_seconds"`
}

type CompleteUploadRequest struct {
	Extension string `json:"extension"`
}

type VideoDTO struct {
	ID              uuid.UUID `json:"id"`
	UploaderID      uuid.UUID `json:"uploader_id"`
	UploaderName    string    `json:"uploader_name"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationSeconds int32     `json:"duration_seconds,omitempty"`
	Visibility      string    `json:"visibility"`
	Status          string    `json:"status"`
	ViewCount       int       `json:"view_count"`
	ThumbnailURL    string    `json:"thumbnail_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at,omitempty"`
}

type WatchResponse struct {
	Video             VideoDTO `json:"video"`
	MasterPlaylistURL string   `json:"master_playlist_url"`
}

type FeedPage struct {
	Items      []VideoDTO  `json:"items"`
	NextCursor *FeedCursor `json:"next_cursor,omitempty"`
}

type FeedCursor struct {
	PublishedAt time.Time `json:"published_at"`
	ID          uuid.UUID `json:"id"`
}

type VideoService interface {
	CreateUploadIntent(ctx context.Context, uploaderID uuid.UUID, req CreateVideoRequest) (*CreateVideoResponse, error)
	MarkUploaded(ctx context.Context, uploaderID, videoID uuid.UUID, extension string) error
	GetForWatch(ctx context.Context, videoID uuid.UUID) (*WatchResponse, error)
	ListFeed(ctx context.Context, limit int32, cursor *FeedCursor) (*FeedPage, error)
	ListByUploader(ctx context.Context, uploaderID uuid.UUID, limit int32, cursor *FeedCursor) (*FeedPage, error)
	SoftDelete(ctx context.Context, uploaderID, videoID uuid.UUID) error
	IncrementViewCount(ctx context.Context, videoID uuid.UUID) error
}

// compile-time assertion helper so callers can import gen types from one place
var __ = gen.Video{}
