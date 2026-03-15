package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal.txt", "normal.txt"},
		{"file/with/slashes.txt", "file-with-slashes.txt"},
		{"file:with:colons.txt", "file-with-colons.txt"},
		{"file*with*stars.txt", "file-with-stars.txt"},
		{"file?with?questions.txt", "file-with-questions.txt"},
		{`file"with"quotes.txt`, "file'with'quotes.txt"},
		{"file<with>brackets.txt", "file-with-brackets.txt"},
		{"file|with|pipes.txt", "file-with-pipes.txt"},
		{"file\\with\\backslashes.txt", "file-with-backslashes.txt"},
	}

	for _, tt := range tests {
		result := SanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeFilenameMaxLength(t *testing.T) {
	// Test that long filenames are truncated
	longName := ""
	for i := 0; i < 200; i++ {
		longName += "a"
	}
	result := SanitizeFilename(longName)
	if len(result) > 100 {
		t.Errorf("SanitizeFilename length = %d, want <= 100", len(result))
	}
}

func TestFileManagerIsManagedFile(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()
	
	cfg := &mockConfig{
		downloadPath: tempDir,
	}
	
	fm := NewFileManager(cfg)
	
	// Test managed file (inside YTed folder)
	managedFile := filepath.Join(tempDir, "video.mp4")
	if !fm.IsManagedFile(managedFile) {
		t.Errorf("IsManagedFile(%q) = false, want true for managed file", managedFile)
	}
	
	// Test unmanaged file (outside YTed folder)
	unmanagedFile := "/tmp/video.mp4"
	if fm.IsManagedFile(unmanagedFile) {
		t.Errorf("IsManagedFile(%q) = true, want false for unmanaged file", unmanagedFile)
	}
	
	// Test empty path
	if fm.IsManagedFile("") {
		t.Errorf("IsManagedFile(\"\") = true, want false for empty path")
	}
}

// mockConfig implements a minimal config interface for testing
type mockConfig struct {
	downloadPath string
}

func (m *mockConfig) Get() *mockConfigData {
	return &mockConfigData{
		DownloadPath: m.downloadPath,
	}
}

type mockConfigData struct {
	DownloadPath string
}
