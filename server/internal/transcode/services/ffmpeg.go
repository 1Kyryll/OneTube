package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/1kyryll/onetube/server/internal/transcode/types"
)

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func Probe(ctx context.Context, path string) (*types.ProbeResult, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe command failed: %w (%s)", err, stderr.String())
	}

	var out ffprobeOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("Failed to parse ffprobe: %w", err)
	}

	r := &types.ProbeResult{}
	for _, s := range out.Streams {
		if s.CodecType == "video" {
			r.Width = s.Width
			r.Height = s.Height
			break
		}
	}

	if dur, err := strconv.ParseFloat(out.Format.Duration, 64); err == nil {
		r.DurationSeconds = int32(math.Round(dur))
	}

	return r, nil
}

// Transcode runs ffmpeg to produce fragmented-MP4 HLS in `outdir/<quality>`
// with `playlist.m3u8` + `seg_NNN.ts` files
func TranscodeRendition(ctx context.Context, srcPath, outDir string, r types.Rendition) error {
	segmentPath := filepath.Join(outDir, "seg_%03d.ts")
	playlistPath := filepath.Join(outDir, "playlist.m3u8")
	scale := fmt.Sprintf("scale=-2:%d", r.Height)
	bitrate := fmt.Sprintf("%dk", r.Bitrate)
	maxrate := fmt.Sprintf("%dk", int(float64(r.Bitrate)*1.07))
	bufsize := fmt.Sprintf("%dk", r.Bitrate*2)

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-i", srcPath,
		"-vf", scale,
		"-c:v", "libx264",
		"-profile:v", "main",
		"-preset", "veryfast",
		"-b:v", bitrate,
		"-maxrate", maxrate,
		"-bufsize", bufsize,
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",
		"-hls_time", "4",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segmentPath,
		"-f", "hls",
		playlistPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg command failed %s: %w (%s)", r.Quality, err, stderr.String())
	}

	return nil
}

func Thumbnail(ctx context.Context, srcPath, outPath string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-ss", "00:00:01",
		"-i", srcPath,
		"-frames:v", "1",
		"-q:v", "2",
		outPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg thumbnail command failed: %w (%s)", err, stderr.String())
	}

	return nil
}
