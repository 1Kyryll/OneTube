package types

import (
	"context"

	"github.com/1kyryll/onetube/server/internal/common/queue"
)

type Rendition struct {
	Quality string // 360p, 720p, 1080p
	Height  int
	Bitrate int // kbps
}

var DefaultRenditions = []Rendition{
	{
		Quality: "360",
		Height:  360,
		Bitrate: 800,
	},
	{
		Quality: "720",
		Height:  720,
		Bitrate: 2500,
	},
	{
		Quality: "1080",
		Height:  1080,
		Bitrate: 5000,
	},
}

type Transcoder interface {
	Process(ctx context.Context, job queue.TranscodeJob) error
}

type ProbeResult struct {
	DurationSeconds int32
	Width           int
	Height          int
}
