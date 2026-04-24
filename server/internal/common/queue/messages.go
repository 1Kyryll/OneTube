package queue

import "github.com/google/uuid"

type TranscodeJob struct {
	VideoID   uuid.UUID `json:"video_id"`
	RawKey    string    `json:"raw_key"`
	Extension string    `json:"extension"`
}
