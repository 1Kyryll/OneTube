package s3

import (
	"fmt"

	"github.com/google/uuid"
)

func RawKey(videoID uuid.UUID, ext string) string {
	return fmt.Sprintf("raw/%s/original.%s", videoID, ext)
}

func MasterPlaylistKey(videoID uuid.UUID) string {
	return fmt.Sprintf("hls/%s/master.m3u8", videoID)
}

func RenditionPlaylistKey(videoID uuid.UUID, quality string) string {
	return fmt.Sprintf("hls/%s/%s/playlist.m3u8", videoID, quality)
}

func SegmentKey(videoID uuid.UUID, quality string, index int) string {
	return fmt.Sprintf("hls/%s/%s/seg_%03d.ts", videoID, quality, index)
}

func RenditionPrefix(videoID uuid.UUID, quality string) string {
	return fmt.Sprintf("hls/%s/%s/", videoID, quality)
}

func HLSPrefix(videoID uuid.UUID) string {
	return fmt.Sprintf("hls/%s/", videoID)
}

func RawPrefix(videoID uuid.UUID) string {
	return fmt.Sprintf("raw/%s/", videoID)
}

func ThumbnailKey(videoID uuid.UUID) string {
	return fmt.Sprintf("thumbnails/%s/default.jpg", videoID)
}

func AvatarKey(userID uuid.UUID) string {
	return fmt.Sprintf("avatars/%s/avatar.jpg", userID)
}
