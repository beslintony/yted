package app

import (
	"testing"
)

func TestIsMediaFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "mp4 video",
			filename: "video.mp4",
			expected: true,
		},
		{
			name:     "webm video",
			filename: "video.webm",
			expected: true,
		},
		{
			name:     "mkv video",
			filename: "video.mkv",
			expected: true,
		},
		{
			name:     "mp3 audio",
			filename: "audio.mp3",
			expected: true,
		},
		{
			name:     "m4a audio",
			filename: "audio.m4a",
			expected: true,
		},
		{
			name:     "ogg audio",
			filename: "audio.ogg",
			expected: true,
		},
		{
			name:     "txt file",
			filename: "notes.txt",
			expected: false,
		},
		{
			name:     "jpg image",
			filename: "image.jpg",
			expected: false,
		},
		{
			name:     "png image",
			filename: "image.png",
			expected: false,
		},
		{
			name:     "pdf document",
			filename: "doc.pdf",
			expected: false,
		},
		{
			name:     "avi video (not supported)",
			filename: "video.avi",
			expected: false,
		},
		{
			name:     "uppercase extension",
			filename: "video.MP4",
			expected: false, // isMediaFile uses lowercase comparison
		},
		{
			name:     "mixed case extension",
			filename: "video.Mp4",
			expected: false,
		},
		{
			name:     "empty filename",
			filename: "",
			expected: false,
		},
		{
			name:     "no extension",
			filename: "video",
			expected: false,
		},
		{
			name:     "partial match",
			filename: "mp4.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMediaFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isMediaFile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}
