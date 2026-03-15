package ytdl

import (
	"testing"
	"time"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1536, "1.5 KB"},
		{1024 * 1536, "1.5 MB"},
	}

	for _, tt := range tests {
		result := FormatFileSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatFileSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestFormatETA(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, ""},
		{-1 * time.Second, ""},
		{5 * time.Second, "5s"},
		{65 * time.Second, "1m 5s"},
		{90 * time.Second, "1m 30s"},
		{3665 * time.Second, "1h 1m 5s"},
		{48 * time.Hour, "2d 0h 0m"},
		{100 * 24 * time.Hour, ">99d"},
	}

	for _, tt := range tests {
		result := FormatETA(tt.duration)
		if result != tt.expected {
			t.Errorf("FormatETA(%v) = %s, want %s", tt.duration, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{0, "0:00"},
		{30, "0:30"},
		{60, "1:00"},
		{90, "1:30"},
		{3600, "1:00:00"},
		{3665, "1:01:05"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.seconds)
		if result != tt.expected {
			t.Errorf("FormatDuration(%d) = %s, want %s", tt.seconds, result, tt.expected)
		}
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://youtu.be/dQw4w9WgXcQ", true},
		{"https://www.youtube.com/shorts/abc123", true},
		{"https://youtube.com/shorts/abc123", true},
		{"https://www.youtube.com/playlist?list=PL123", true},
		{"https://youtube.com/playlist?list=PL123", true},
		{"https://vimeo.com/123456", false},
		{"not-a-url", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidURL(tt.url)
		if result != tt.expected {
			t.Errorf("IsValidURL(%q) = %v, want %v", tt.url, result, tt.expected)
		}
	}
}
