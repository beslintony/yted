package app

import (
	"testing"

	"yted/internal/db"
	"yted/internal/ytdl"
)

func TestDownloadNeedsMetadata(t *testing.T) {
	title := "Some Title"
	empty := ""

	tests := []struct {
		name     string
		dl       *db.Download
		expected bool
	}{
		{"nil title needs fetch", &db.Download{Title: nil}, true},
		{"empty title needs fetch", &db.Download{Title: &empty}, true},
		{"present title skips fetch", &db.Download{Title: &title}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := downloadNeedsMetadata(tt.dl); got != tt.expected {
				t.Errorf("downloadNeedsMetadata() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLibraryNeedsMetadata(t *testing.T) {
	title := "Some Title"
	empty := ""
	duration := 120
	zeroDuration := 0

	tests := []struct {
		name     string
		dl       *db.Download
		expected bool
	}{
		{"nil title needs fetch", &db.Download{Title: nil, Duration: &duration}, true},
		{"empty title needs fetch", &db.Download{Title: &empty, Duration: &duration}, true},
		{"nil duration needs fetch (never fetched)", &db.Download{Title: &title, Duration: nil}, true},
		{"zero duration skips fetch (fetched, unknown)", &db.Download{Title: &title, Duration: &zeroDuration}, false},
		{"complete metadata skips fetch", &db.Download{Title: &title, Duration: &duration}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := libraryNeedsMetadata(tt.dl); got != tt.expected {
				t.Errorf("libraryNeedsMetadata() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestApplyVideoInfoToDownload(t *testing.T) {
	dl := &db.Download{}
	info := &ytdl.VideoInfo{
		Title:     "Fetched Title",
		Channel:   "Fetched Channel",
		Thumbnail: "https://example.com/thumb.jpg",
		Duration:  300,
	}

	applyVideoInfoToDownload(dl, info)

	if dl.Title == nil || *dl.Title != "Fetched Title" {
		t.Errorf("Title = %v, want Fetched Title", dl.Title)
	}
	if dl.Channel == nil || *dl.Channel != "Fetched Channel" {
		t.Errorf("Channel = %v, want Fetched Channel", dl.Channel)
	}
	if dl.ThumbnailURL == nil || *dl.ThumbnailURL != "https://example.com/thumb.jpg" {
		t.Errorf("ThumbnailURL = %v, want thumbnail URL", dl.ThumbnailURL)
	}
	if dl.Duration == nil || *dl.Duration != 300 {
		t.Errorf("Duration = %v, want 300", dl.Duration)
	}

	// After applying, both predicates must report "no fetch needed",
	// even when duration is unknown (0): this is what kills the double
	// GetInfo subprocess.
	if downloadNeedsMetadata(dl) {
		t.Error("downloadNeedsMetadata() = true after apply, want false")
	}
	if libraryNeedsMetadata(dl) {
		t.Error("libraryNeedsMetadata() = true after apply, want false")
	}
}

func TestApplyVideoInfoZeroDurationSkipsLibraryFetch(t *testing.T) {
	dl := &db.Download{}
	applyVideoInfoToDownload(dl, &ytdl.VideoInfo{
		Title:     "Live Stream",
		Channel:   "Channel",
		Thumbnail: "https://example.com/t.jpg",
		Duration:  0,
	})

	if dl.Duration == nil {
		t.Fatal("Duration should be non-nil after apply (fetched-but-unknown)")
	}
	if *dl.Duration != 0 {
		t.Errorf("Duration = %d, want 0", *dl.Duration)
	}
	if libraryNeedsMetadata(dl) {
		t.Error("libraryNeedsMetadata() = true for fetched zero duration, want false (no refetch)")
	}
}
