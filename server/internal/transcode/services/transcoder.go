package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"

	"github.com/1kyryll/onetube/server/internal/common/gen"
	"github.com/1kyryll/onetube/server/internal/common/queue"
	s3keys "github.com/1kyryll/onetube/server/internal/common/s3"
	"github.com/1kyryll/onetube/server/internal/transcode/types"
)

type Transcoder struct {
	queries *gen.Queries
	s3      *s3keys.Client
}

func NewTranscoder(queries *gen.Queries, s3 *s3keys.Client) *Transcoder {
	return &Transcoder{queries: queries, s3: s3}
}

func (t *Transcoder) Process(ctx context.Context, job queue.TranscodeJob) error {
	log.Printf("transcode: start video=%s", job.VideoID)

	workDir, err := os.MkdirTemp("", "transcode-"+job.VideoID.String()+"-")
	if err != nil {
		return fmt.Errorf("mkdir temp: %w", err)
	}
	defer os.RemoveAll(workDir)

	rawPath := filepath.Join(workDir, "original."+job.Extension)
	if err := t.downloadObject(ctx, job.RawKey, rawPath); err != nil {
		return t.failVideo(ctx, job.VideoID, fmt.Errorf("download raw: %w", err))
	}

	probe, err := Probe(ctx, rawPath)
	if err != nil {
		return t.failVideo(ctx, job.VideoID, err)
	}

	// Thumbnail
	thumbLocal := filepath.Join(workDir, "thumbnail.jpg")
	if err := Thumbnail(ctx, rawPath, thumbLocal); err != nil {
		return t.failVideo(ctx, job.VideoID, err)
	}
	thumbKey := s3keys.ThumbnailKey(job.VideoID)
	if err := t.uploadObject(ctx, thumbLocal, thumbKey, "image/jpeg"); err != nil {
		return t.failVideo(ctx, job.VideoID, fmt.Errorf("upload thumb: %w", err))
	}

	// Renditions
	var produced []types.Rendition
	for _, r := range types.DefaultRenditions {
		if probe.Height > 0 && r.Height > probe.Height {
			// Do not upscale. Skip renditions taller than the source.
			continue
		}
		outDir := filepath.Join(workDir, r.Quality)
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return t.failVideo(ctx, job.VideoID, err)
		}
		if err := TranscodeRendition(ctx, rawPath, outDir, r); err != nil {
			return t.failVideo(ctx, job.VideoID, err)
		}
		if err := t.uploadRendition(ctx, job.VideoID, r, outDir); err != nil {
			return t.failVideo(ctx, job.VideoID, err)
		}
		if err := t.upsertAsset(ctx, job.VideoID, r, probe); err != nil {
			return t.failVideo(ctx, job.VideoID, err)
		}
		produced = append(produced, r)
	}
	if len(produced) == 0 {
		return t.failVideo(ctx, job.VideoID, fmt.Errorf("no renditions produced"))
	}

	// Master playlist
	master := buildMasterPlaylist(produced)
	masterKey := s3keys.MasterPlaylistKey(job.VideoID)
	if err := t.putBytes(ctx, masterKey, []byte(master), "application/vnd.apple.mpegurl"); err != nil {
		return t.failVideo(ctx, job.VideoID, fmt.Errorf("upload master: %w", err))
	}

	duration := pgtype.Int4{Valid: false}
	if probe.DurationSeconds > 0 {
		duration = pgtype.Int4{Int32: probe.DurationSeconds, Valid: true}
	}
	thumb := pgtype.Text{String: thumbKey, Valid: true}
	if err := t.queries.MarkVideoReady(ctx, gen.MarkVideoReadyParams{
		ID:              job.VideoID,
		DurationSeconds: duration,
		ThumbnailKey:    thumb,
	}); err != nil {
		return fmt.Errorf("mark ready: %w", err)
	}
	log.Printf("transcode: done  video=%s renditions=%d", job.VideoID, len(produced))
	return nil
}

func (t *Transcoder) downloadObject(ctx context.Context, key, dst string) error {
	obj, err := t.s3.API.GetObject(ctx, t.s3.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer obj.Close()
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, obj); err != nil {
		return err
	}
	return nil
}

func (t *Transcoder) uploadObject(ctx context.Context, srcPath, key, contentType string) error {
	_, err := t.s3.API.FPutObject(ctx, t.s3.Bucket, key, srcPath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (t *Transcoder) putBytes(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := t.s3.API.PutObject(ctx, t.s3.Bucket, key, bytesReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (t *Transcoder) uploadRendition(ctx context.Context, videoID uuid.UUID, r types.Rendition, localDir string) error {
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		local := filepath.Join(localDir, e.Name())
		var key, ctype string
		switch {
		case strings.HasSuffix(e.Name(), ".m3u8"):
			key = s3keys.RenditionPlaylistKey(videoID, r.Quality)
			ctype = "application/vnd.apple.mpegurl"
			if err := t.rewritePlaylistAndUpload(ctx, local, key, videoID, r.Quality); err != nil {
				return err
			}
			continue
		case strings.HasSuffix(e.Name(), ".ts"):
			key = s3keys.RenditionPrefix(videoID, r.Quality) + e.Name()
			ctype = "video/mp2t"
		default:
			continue
		}
		if err := t.uploadObject(ctx, local, key, ctype); err != nil {
			return err
		}
	}

	return nil
}

// rewritePlaylistAndUpload ensures segment lines in the rendition playlist are
// bare filenames (ffmpeg already emits them that way, but we also normalize any
// absolute paths that crept in). We upload the file as-is.
func (t *Transcoder) rewritePlaylistAndUpload(ctx context.Context, src, key string, videoID uuid.UUID, quality string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf strings.Builder
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasSuffix(line, ".ts") {
			line = filepath.Base(line)
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return t.putBytes(ctx, key, []byte(buf.String()), "application/vnd.apple.mpegurl")
}

func (t *Transcoder) upsertAsset(ctx context.Context, videoID uuid.UUID, r types.Rendition, probe *types.ProbeResult) error {
	// Sum of segment sizes would be ideal; keep nil for now.
	return t.queries.UpsertVideoAsset(ctx, gen.UpsertVideoAssetParams{
		ID:         uuid.New(),
		VideoID:    videoID,
		Quality:    r.Quality,
		Codec:      "h264",
		Container:  "hls",
		StorageKey: s3keys.RenditionPlaylistKey(videoID, r.Quality),
		Bitrate:    pgtype.Int4{Int32: int32(r.Bitrate * 1000), Valid: true},
		Width:      pgtype.Int4{Int32: int32(probe.Width * r.Height / max1(probe.Height)), Valid: probe.Height > 0},
		Height:     pgtype.Int4{Int32: int32(r.Height), Valid: true},
	})
}

func (t *Transcoder) failVideo(ctx context.Context, videoID uuid.UUID, cause error) error {
	log.Printf("transcode: FAIL video=%s err=%v", videoID, cause)
	_ = t.queries.SetVideoStatus(ctx, gen.SetVideoStatusParams{ID: videoID, Status: "failed"})
	return cause
}

func buildMasterPlaylist(rs []types.Rendition) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:3\n")
	for _, r := range rs {
		bandwidth := r.Bitrate * 1100 // add 10% for audio + overhead
		width := r.Height * 16 / 9
		b.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n", bandwidth*1000, width, r.Height))
		b.WriteString(fmt.Sprintf("%s/playlist.m3u8\n", r.Quality))
	}
	return b.String()
}

func max1(n int) int {
	if n <= 0 {
		return 1
	}
	return n
}

// bytesReader avoids pulling in bytes.NewReader signature at call sites.
func bytesReader(b []byte) *bytesReaderImpl {
	return &bytesReaderImpl{b: b}
}

type bytesReaderImpl struct {
	b []byte
	i int
}

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
