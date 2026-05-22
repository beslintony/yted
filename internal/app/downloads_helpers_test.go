package app

import (
	"testing"

	"yted/internal/db"
	"yted/internal/ytdl"
)

func TestCleanYouTubeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard youtube.com URL",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:     "youtube.com URL with playlist param",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLsomeplaylist",
			expected: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:     "youtube.com URL with multiple params",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=120s&ab_channel=Test",
			expected: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:     "youtu.be short URL",
			input:    "https://youtu.be/dQw4w9WgXcQ",
			expected: "https://youtu.be/dQw4w9WgXcQ",
		},
		{
			name:     "youtu.be short URL with params",
			input:    "https://youtu.be/dQw4w9WgXcQ?t=120",
			expected: "https://youtu.be/dQw4w9WgXcQ",
		},
		{
			name:     "URL with whitespace",
			input:    "  https://www.youtube.com/watch?v=dQw4w9WgXcQ  ",
			expected: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:     "invalid URL returns manual extraction",
			input:    "not-a-url-but-has-v=dQw4w9WgXcQ",
			expected: "not-a-url-but-has-v=dQw4w9WgXcQ",
		},
		{
			name:     "youtube shorts URL",
			input:    "https://youtube.com/shorts/abc123",
			expected: "https://youtube.com/shorts/abc123",
		},
		{
			name:     "mobile youtube URL",
			input:    "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanYouTubeURL(tt.input)
			if result != tt.expected {
				t.Errorf("cleanYouTubeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractVideoIDManual(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "v= parameter",
			input:    "https://example.com/watch?v=dQw4w9WgXcQ&other=param",
			expected: "https://example.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:     "v= parameter at end",
			input:    "https://example.com/watch?v=dQw4w9WgXcQ",
			expected: "https://example.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:     "youtu.be format",
			input:    "https://youtu.be/dQw4w9WgXcQ?extra=params",
			expected: "https://youtu.be/dQw4w9WgXcQ",
		},
		{
			name:     "no video ID found",
			input:    "https://example.com/no-id-here",
			expected: "https://example.com/no-id-here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVideoIDManual(tt.input)
			if result != tt.expected {
				t.Errorf("extractVideoIDManual(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractYoutubeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard youtube.com URL",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "youtu.be short URL",
			input:    "https://youtu.be/dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "youtube shorts URL",
			input:    "https://youtube.com/shorts/abc123def",
			expected: "abc123def",
		},
		{
			name:     "URL with extra path",
			input:    "https://youtu.be/dQw4w9WgXcQ/extra",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "invalid URL",
			input:    "not-a-valid-url",
			expected: "",
		},
		{
			name:     "empty URL",
			input:    "",
			expected: "",
		},
		{
			name:     "youtube.com with multiple params",
			input:    "https://www.youtube.com/watch?v=abc123&list=PL123",
			expected: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractYoutubeID(tt.input)
			if result != tt.expected {
				t.Errorf("extractYoutubeID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetDownloadExtension(t *testing.T) {
	tests := []struct {
		name     string
		quality  *string
		formatID *string
		expected string
	}{
		{
			name:     "audio quality",
			quality:  strPtr("audio"),
			formatID: nil,
			expected: "mp3",
		},
		{
			name:     "audio format ID",
			quality:  nil,
			formatID: strPtr("audio-only"),
			expected: "mp3",
		},
		{
			name:     "video quality",
			quality:  strPtr("720p"),
			formatID: nil,
			expected: "mp4",
		},
		{
			name:     "video format ID",
			quality:  nil,
			formatID: strPtr("18"),
			expected: "mp4",
		},
		{
			name:     "nil values",
			quality:  nil,
			formatID: nil,
			expected: "mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dl := &db.Download{
				Quality:  tt.quality,
				FormatID: tt.formatID,
			}
			result := getDownloadExtension(dl)
			if result != tt.expected {
				t.Errorf("getDownloadExtension() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetDownloadTitle(t *testing.T) {
	tests := []struct {
		name      string
		dlTitle   *string
		infoTitle string
		youtubeID string
		expected  string
	}{
		{
			name:      "download title available",
			dlTitle:   strPtr("My Video"),
			infoTitle: "Other Title",
			youtubeID: "abc123",
			expected:  "My Video",
		},
		{
			name:      "video info title fallback",
			dlTitle:   nil,
			infoTitle: "Info Title",
			youtubeID: "abc123",
			expected:  "Info Title",
		},
		{
			name:      "youtube ID fallback",
			dlTitle:   nil,
			infoTitle: "",
			youtubeID: "abc123",
			expected:  "abc123",
		},
		{
			name:      "empty download title uses info",
			dlTitle:   strPtr(""),
			infoTitle: "Info Title",
			youtubeID: "abc123",
			expected:  "Info Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dl := &db.Download{Title: tt.dlTitle}
			var info *ytdl.VideoInfo
			if tt.infoTitle != "" {
				info = &ytdl.VideoInfo{Title: tt.infoTitle}
			}
			result := getDownloadTitle(dl, info, tt.youtubeID)
			if result != tt.expected {
				t.Errorf("getDownloadTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractDownloadInfo(t *testing.T) {
	title := "Test Title"
	channel := "Test Channel"
	thumbnail := "https://example.com/thumb.jpg"

	dl := &db.Download{
		Title:        &title,
		Channel:      &channel,
		ThumbnailURL: &thumbnail,
	}

	gotTitle, gotChannel, gotThumbnail := extractDownloadInfo(dl)

	if gotTitle != title {
		t.Errorf("extractDownloadInfo() title = %q, want %q", gotTitle, title)
	}
	if gotChannel != channel {
		t.Errorf("extractDownloadInfo() channel = %q, want %q", gotChannel, channel)
	}
	if gotThumbnail != thumbnail {
		t.Errorf("extractDownloadInfo() thumbnail = %q, want %q", gotThumbnail, thumbnail)
	}
}

func TestExtractDownloadInfoNilPointers(t *testing.T) {
	dl := &db.Download{
		Title:        nil,
		Channel:      nil,
		ThumbnailURL: nil,
	}

	gotTitle, gotChannel, gotThumbnail := extractDownloadInfo(dl)

	if gotTitle != "" {
		t.Errorf("extractDownloadInfo() title = %q, want empty", gotTitle)
	}
	if gotChannel != "" {
		t.Errorf("extractDownloadInfo() channel = %q, want empty", gotChannel)
	}
	if gotThumbnail != "" {
		t.Errorf("extractDownloadInfo() thumbnail = %q, want empty", gotThumbnail)
	}
}

func TestCreateVideoRecord(t *testing.T) {
	title := "Test Video"
	channel := "Test Channel"
	thumbnail := "https://example.com/thumb.jpg"
	quality := "720p"
	formatID := "18"
	duration := 120

	dl := &db.Download{
		Title:        &title,
		Channel:      &channel,
		ThumbnailURL: &thumbnail,
		Quality:      &quality,
		FormatID:     &formatID,
		Duration:     &duration,
	}

	info := &ytdl.VideoInfo{
		Duration:    120,
		Description: "Test description",
		ChannelID:   "UC123",
	}

	video := createVideoRecord(dl, info, "youtube123", title, "/path/to/video.mp4", 1024, "hash123", true)

	if video.Title != title {
		t.Errorf("createVideoRecord() Title = %q, want %q", video.Title, title)
	}
	if video.Channel != channel {
		t.Errorf("createVideoRecord() Channel = %q, want %q", video.Channel, channel)
	}
	if video.FilePath != "/path/to/video.mp4" {
		t.Errorf("createVideoRecord() FilePath = %q, want %q", video.FilePath, "/path/to/video.mp4")
	}
	if video.FileSize != 1024 {
		t.Errorf("createVideoRecord() FileSize = %d, want 1024", video.FileSize)
	}
	if video.FileHash != "hash123" {
		t.Errorf("createVideoRecord() FileHash = %q, want %q", video.FileHash, "hash123")
	}
	if !video.IsManaged {
		t.Error("createVideoRecord() IsManaged should be true")
	}
	if video.Duration != 120 {
		t.Errorf("createVideoRecord() Duration = %d, want 120", video.Duration)
	}
	if video.Description != "Test description" {
		t.Errorf("createVideoRecord() Description = %q, want %q", video.Description, "Test description")
	}
	if video.ChannelID != "UC123" {
		t.Errorf("createVideoRecord() ChannelID = %q, want %q", video.ChannelID, "UC123")
	}
	if video.Quality != quality {
		t.Errorf("createVideoRecord() Quality = %q, want %q", video.Quality, quality)
	}
	if video.Format != formatID {
		t.Errorf("createVideoRecord() Format = %q, want %q", video.Format, formatID)
	}
}

func TestCreateVideoRecordFallbackDuration(t *testing.T) {
	title := "Test Video"
	duration := 300

	dl := &db.Download{
		Title:    &title,
		Duration: &duration,
	}

	// No video info provided
	video := createVideoRecord(dl, nil, "youtube123", title, "/path/to/video.mp4", 1024, "hash", false)

	if video.Duration != 300 {
		t.Errorf("createVideoRecord() Duration = %d, want 300 (from dl.Duration)", video.Duration)
	}
}

func TestConvertDownloadToResult(t *testing.T) {
	title := "Test Video"
	channel := "Test Channel"
	thumbnail := "https://example.com/thumb.jpg"
	formatID := "18"
	quality := "720p"
	errMsg := "Network error"

	dl := db.Download{
		ID:           "dl-123",
		URL:          "https://youtube.com/watch?v=abc123",
		Status:       "error",
		Progress:     45.5,
		Title:        &title,
		Channel:      &channel,
		ThumbnailURL: &thumbnail,
		FormatID:     &formatID,
		Quality:      &quality,
		ErrorMessage: &errMsg,
	}

	result := convertDownloadToResult(dl)

	if result.ID != "dl-123" {
		t.Errorf("convertDownloadToResult() ID = %q, want %q", result.ID, "dl-123")
	}
	if result.URL != "https://youtube.com/watch?v=abc123" {
		t.Errorf("convertDownloadToResult() URL = %q, want %q", result.URL, "https://youtube.com/watch?v=abc123")
	}
	if result.Status != "error" {
		t.Errorf("convertDownloadToResult() Status = %q, want %q", result.Status, "error")
	}
	if result.Progress != 45.5 {
		t.Errorf("convertDownloadToResult() Progress = %f, want 45.5", result.Progress)
	}
	if result.Title != title {
		t.Errorf("convertDownloadToResult() Title = %q, want %q", result.Title, title)
	}
	if result.Channel != channel {
		t.Errorf("convertDownloadToResult() Channel = %q, want %q", result.Channel, channel)
	}
	if result.ThumbnailURL != thumbnail {
		t.Errorf("convertDownloadToResult() ThumbnailURL = %q, want %q", result.ThumbnailURL, thumbnail)
	}
	if result.FormatID != formatID {
		t.Errorf("convertDownloadToResult() FormatID = %q, want %q", result.FormatID, formatID)
	}
	if result.Quality != quality {
		t.Errorf("convertDownloadToResult() Quality = %q, want %q", result.Quality, quality)
	}
	if result.ErrorMessage != errMsg {
		t.Errorf("convertDownloadToResult() ErrorMessage = %q, want %q", result.ErrorMessage, errMsg)
	}
	if result.YoutubeID != "abc123" {
		t.Errorf("convertDownloadToResult() YoutubeID = %q, want %q", result.YoutubeID, "abc123")
	}
}

func TestConvertDownloadToResultNilPointers(t *testing.T) {
	dl := db.Download{
		ID:     "dl-123",
		URL:    "https://youtube.com/watch?v=abc123",
		Status: "pending",
	}

	result := convertDownloadToResult(dl)

	if result.Title != "" {
		t.Errorf("convertDownloadToResult() Title = %q, want empty", result.Title)
	}
	if result.Channel != "" {
		t.Errorf("convertDownloadToResult() Channel = %q, want empty", result.Channel)
	}
	if result.ThumbnailURL != "" {
		t.Errorf("convertDownloadToResult() ThumbnailURL = %q, want empty", result.ThumbnailURL)
	}
	if result.FormatID != "" {
		t.Errorf("convertDownloadToResult() FormatID = %q, want empty", result.FormatID)
	}
	if result.Quality != "" {
		t.Errorf("convertDownloadToResult() Quality = %q, want empty", result.Quality)
	}
	if result.ErrorMessage != "" {
		t.Errorf("convertDownloadToResult() ErrorMessage = %q, want empty", result.ErrorMessage)
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
