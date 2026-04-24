package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	"github.com/1kyryll/onetube/server/internal/common/queue"
	s3client "github.com/1kyryll/onetube/server/internal/common/s3"
	"github.com/1kyryll/onetube/server/internal/video/types"
)

var (
	ErrForbidden = errors.New("Forbidden")
	ErrNotReady  = errors.New("Video not ready")
)

const (
	uploadUrlTTL   = 2 * time.Hour
	playbackUrlTTL = 24 * time.Hour
)

type VideoServiceImpl struct {
	queries   *gen.Queries
	s3        *s3client.Client
	publisher *queue.Publisher
}

func NewVideoService(queries *gen.Queries, s3 *s3client.Client, publisher *queue.Publisher) *VideoServiceImpl {
	return &VideoServiceImpl{
		queries:   queries,
		s3:        s3,
		publisher: publisher,
	}
}

func (s *VideoServiceImpl) CreateUploadIntent(ctx context.Context, uploaderID uuid.UUID, req types.CreateVideoRequest) (*types.CreateVideoResponse, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(req.Filename)), ".")
	if ext == "" {
		ext = "mp4"
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = "public"
	}
	if visibility != "public" && visibility != "unlisted" && visibility != "private" {
		return nil, errors.New("Invalid visibility")
	}

	videoID := uuid.New()
	if _, err := s.queries.CreateVideo(ctx, gen.CreateVideoParams{
		ID:          videoID,
		UploaderID:  uploaderID,
		Title:       req.Title,
		Description: req.Description,
		Visibility:  visibility,
		Status:      "uploading",
	}); err != nil {
		return nil, fmt.Errorf("Failed to create video record: %w", err)
	}

	key := s3client.RawKey(videoID, ext)
	url, err := s.s3.PresignPut(ctx, key, req.ContentType, uploadUrlTTL)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate presigned URL: %w", err)
	}

	return &types.CreateVideoResponse{
		VideoID:    videoID,
		UploadURL:  url,
		StorageKey: key,
		ExpiresIn:  int(uploadUrlTTL.Seconds()),
	}, nil
}

func (s *VideoServiceImpl) MarkUploaded(ctx context.Context, uploaderID, videoID uuid.UUID, extension string) error {
	v, err := s.queries.GetVideoByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("Failed to fetch video: %w", err)
	}

	if v.UploaderID != uploaderID {
		return ErrForbidden
	}

	if v.Status != "uploading" {
		// idempotent
		return nil
	}

	if err := s.queries.SetVideoStatus(ctx, gen.SetVideoStatusParams{
		ID:     videoID,
		Status: "processing",
	}); err != nil {
		return fmt.Errorf("Failed to update video status: %w", err)
	}

	ext := strings.TrimPrefix(strings.ToLower(extension), ".")
	if ext == "" {
		ext = "mp4"
	}

	return s.publisher.PublishTranscode(ctx, queue.TranscodeJob{
		VideoID:   videoID,
		RawKey:    s3client.RawKey(videoID, ext),
		Extension: ext,
	})
}

func (s *VideoServiceImpl) GetForWatch(ctx context.Context, videoID uuid.UUID) (*types.WatchResponse, error) {
	v, err := s.queries.GetVideoByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch video: %w", err)
	}

	if v.Status != "ready" {
		return nil, ErrNotReady
	}

	u, err := s.queries.GetUserByID(ctx, v.UploaderID)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch uploader: %w", err)
	}

	masterURL, err := s.s3.PresignGet(ctx, s3client.MasterPlaylistKey(videoID), playbackUrlTTL)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate presigned URL: %w", err)
	}

	dto := s.toDTO(v, u.Username)
	if v.ThumbnailKey.Valid {
		thumbURL, err := s.s3.PresignGet(ctx, v.ThumbnailKey.String, playbackUrlTTL)
		if err != nil {
			dto.ThumbnailURL = thumbURL
		}
	}

	return &types.WatchResponse{
		Video:             dto,
		MasterPlaylistURL: masterURL,
	}, nil
}

func (s *VideoServiceImpl) ListFeed(ctx context.Context, limit int32, cursor *types.FeedCursor) (*types.FeedPage, error) {
	params := gen.ListFeedParams{RowLimit: limit}
	if cursor != nil {
		params.HasCursor = true
		params.CursorPublishedAt = pgtype.Timestamptz{Time: cursor.PublishedAt, Valid: true}
		params.CursorID = cursor.ID
	}

	rows, err := s.queries.ListFeed(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Failed to list feed: %w", err)
	}

	out := make([]types.VideoDTO, 0, len(rows))
	for _, r := range rows {
		dto := types.VideoDTO{
			ID:           r.ID,
			UploaderID:   r.UploaderID,
			UploaderName: r.UploaderUsername,
			Title:        r.Title,
			Description:  r.Description,
			Visibility:   r.Visibility,
			Status:       r.Status,
			ViewCount:    int(r.ViewCount),
			CreatedAt:    r.CreatedAt.Time,
		}

		if r.DurationSeconds.Valid {
			v := r.DurationSeconds.Int32
			dto.DurationSeconds = v
		}
		if r.PublishedAt.Valid {
			t := r.PublishedAt.Time
			dto.PublishedAt = t
		}
		if r.ThumbnailKey.Valid {
			if u, err := s.s3.PresignGet(ctx, r.ThumbnailKey.String, playbackUrlTTL); err == nil {
				dto.ThumbnailURL = u
			}
		}
		out = append(out, dto)
	}

	page := &types.FeedPage{Items: out}
	if int32(len(out)) == limit && len(out) > 0 {
		last := out[len(out)-1]
		if !last.PublishedAt.IsZero() {
			page.NextCursor = &types.FeedCursor{
				PublishedAt: last.PublishedAt,
				ID:          last.ID,
			}
		}
	}
	return page, nil
}

func (s *VideoServiceImpl) ListByUploader(ctx context.Context, uploaderID uuid.UUID, limit int32, cursor *types.FeedCursor) (*types.FeedPage, error) {
	params := gen.ListByUploaderParams{
		UploaderID: uploaderID,
		RowLimit:   limit,
	}

	if cursor != nil {
		params.HasCursor = true
		params.CursorCreatedAt = pgtype.Timestamptz{Time: cursor.PublishedAt, Valid: true}
		params.CursorID = cursor.ID
	}

	rows, err := s.queries.ListByUploader(ctx, params)
	if err != nil {
		return nil, err
	}

	u, err := s.queries.GetUserByID(ctx, uploaderID)
	if err != nil {
		return nil, err
	}

	out := make([]types.VideoDTO, 0, len(rows))
	for _, v := range rows {
		out = append(out, s.toDTO(v, u.Username))
	}

	page := &types.FeedPage{Items: out}
	if int32(len(out)) == limit && len(out) > 0 {
		last := rows[len(rows)-1]
		page.NextCursor = &types.FeedCursor{
			PublishedAt: last.CreatedAt.Time,
			ID:          last.ID,
		}
	}

	return page, nil
}

func (s *VideoServiceImpl) SoftDelete(ctx context.Context, uploaderID, videoID uuid.UUID) error {
	return s.queries.SoftDeleteVideo(ctx, gen.SoftDeleteVideoParams{
		ID:         videoID,
		UploaderID: uploaderID,
	})
}

func (s *VideoServiceImpl) IncrementView(ctx context.Context, videoID uuid.UUID) error {
	return s.queries.IncrementViewCount(ctx, videoID)
}

func (s *VideoServiceImpl) toDTO(v gen.Video, uploaderName string) types.VideoDTO {
	dto := types.VideoDTO{
		ID:           v.ID,
		UploaderID:   v.UploaderID,
		UploaderName: uploaderName,
		Title:        v.Title,
		Description:  v.Description,
		Visibility:   v.Visibility,
		Status:       v.Status,
		ViewCount:    int(v.ViewCount),
		CreatedAt:    v.CreatedAt.Time,
	}
	if v.DurationSeconds.Valid {
		d := v.DurationSeconds.Int32
		dto.DurationSeconds = d
	}
	if v.PublishedAt.Valid {
		t := v.PublishedAt.Time
		dto.PublishedAt = t
	}
	return dto
}
