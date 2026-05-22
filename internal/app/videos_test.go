package app

import (
	"testing"
	"time"

	"yted/internal/db"
)

func TestVideoToResult(t *testing.T) {
	now := time.Now()

	video := db.Video{
		ID:            "vid-123",
		YoutubeID:     "youtube456",
		Title:         "Test Video",
		Channel:       "Test Channel",
		ChannelID:     "UC123",
		Duration:      300,
		Description:   "Test description",
		ThumbnailURL:  "https://example.com/thumb.jpg",
		FilePath:      "/path/to/video.mp4",
		FileSize:      1024000,
		Format:        "18",
		Quality:       "720p",
		DownloadedAt:  now,
		WatchPosition: 45,
		WatchCount:    3,
	}

	result := videoToResult(video)

	if result.ID != "vid-123" {
		t.Errorf("videoToResult() ID = %q, want %q", result.ID, "vid-123")
	}
	if result.YoutubeID != "youtube456" {
		t.Errorf("videoToResult() YoutubeID = %q, want %q", result.YoutubeID, "youtube456")
	}
	if result.Title != "Test Video" {
		t.Errorf("videoToResult() Title = %q, want %q", result.Title, "Test Video")
	}
	if result.Channel != "Test Channel" {
		t.Errorf("videoToResult() Channel = %q, want %q", result.Channel, "Test Channel")
	}
	if result.ChannelID != "UC123" {
		t.Errorf("videoToResult() ChannelID = %q, want %q", result.ChannelID, "UC123")
	}
	if result.Duration != 300 {
		t.Errorf("videoToResult() Duration = %d, want 300", result.Duration)
	}
	if result.Description != "Test description" {
		t.Errorf("videoToResult() Description = %q, want %q", result.Description, "Test description")
	}
	if result.ThumbnailURL != "https://example.com/thumb.jpg" {
		t.Errorf("videoToResult() ThumbnailURL = %q, want %q", result.ThumbnailURL, "https://example.com/thumb.jpg")
	}
	if result.FilePath != "/path/to/video.mp4" {
		t.Errorf("videoToResult() FilePath = %q, want %q", result.FilePath, "/path/to/video.mp4")
	}
	if result.FileSize != 1024000 {
		t.Errorf("videoToResult() FileSize = %d, want 1024000", result.FileSize)
	}
	if result.Format != "18" {
		t.Errorf("videoToResult() Format = %q, want %q", result.Format, "18")
	}
	if result.Quality != "720p" {
		t.Errorf("videoToResult() Quality = %q, want %q", result.Quality, "720p")
	}
	if result.DownloadedAt != now.Unix() {
		t.Errorf("videoToResult() DownloadedAt = %d, want %d", result.DownloadedAt, now.Unix())
	}
	if result.WatchPosition != 45 {
		t.Errorf("videoToResult() WatchPosition = %d, want 45", result.WatchPosition)
	}
	if result.WatchCount != 3 {
		t.Errorf("videoToResult() WatchCount = %d, want 3", result.WatchCount)
	}
}

func TestVideoToResultZeroValues(t *testing.T) {
	video := db.Video{
		ID:        "vid-empty",
		YoutubeID: "",
		Title:     "",
	}

	result := videoToResult(video)

	if result.ID != "vid-empty" {
		t.Errorf("videoToResult() ID = %q, want %q", result.ID, "vid-empty")
	}
	if result.YoutubeID != "" {
		t.Errorf("videoToResult() YoutubeID = %q, want empty", result.YoutubeID)
	}
	if result.Title != "" {
		t.Errorf("videoToResult() Title = %q, want empty", result.Title)
	}
	if result.Duration != 0 {
		t.Errorf("videoToResult() Duration = %d, want 0", result.Duration)
	}
	if result.FileSize != 0 {
		t.Errorf("videoToResult() FileSize = %d, want 0", result.FileSize)
	}
	// Zero time Unix() returns a negative value, which is expected behavior
	if result.DownloadedAt >= 0 {
		t.Errorf("videoToResult() DownloadedAt = %d, want negative value for zero time", result.DownloadedAt)
	}
	if result.WatchPosition != 0 {
		t.Errorf("videoToResult() WatchPosition = %d, want 0", result.WatchPosition)
	}
	if result.WatchCount != 0 {
		t.Errorf("videoToResult() WatchCount = %d, want 0", result.WatchCount)
	}
}
