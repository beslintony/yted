package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindDownloadedFile(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	
	// Test case 1: File with [youtubeID][formatID] pattern
	youtubeID := "dQw4w9WgXcQ"
	formatID := "18"
	
	// Create test files
	expectedFile := filepath.Join(tempDir, "Video Title ["+youtubeID+"]["+formatID+"].mp4")
	err := os.WriteFile(expectedFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Find the file
	result := findDownloadedFile(tempDir, youtubeID, formatID, "mp4")
	if result != expectedFile {
		t.Errorf("findDownloadedFile() = %q, want %q", result, expectedFile)
	}
	
	// Test case 2: File with just [youtubeID] pattern (backward compat)
	youtubeID2 := "abc123"
	expectedFile2 := filepath.Join(tempDir, "Another Video ["+youtubeID2+"].mp4")
	err = os.WriteFile(expectedFile2, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	result2 := findDownloadedFile(tempDir, youtubeID2, "", "mp4")
	if result2 != expectedFile2 {
		t.Errorf("findDownloadedFile() = %q, want %q", result2, expectedFile2)
	}
	
	// Test case 3: Non-existent file
	result3 := findDownloadedFile(tempDir, "nonexistent", "99", "mp4")
	if result3 != "" {
		t.Errorf("findDownloadedFile() = %q, want empty string", result3)
	}
}

func TestFindDownloadedFileTypeMatching(t *testing.T) {
	tempDir := t.TempDir()
	youtubeID := "test123"
	
	// Create an audio file
	audioFile := filepath.Join(tempDir, "Audio ["+youtubeID+"].mp3")
	err := os.WriteFile(audioFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create audio file: %v", err)
	}
	
	// Create a video file
	videoFile := filepath.Join(tempDir, "Video ["+youtubeID+"].mp4")
	err = os.WriteFile(videoFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	
	// When looking for video (mp4), should prefer video file
	result := findDownloadedFile(tempDir, youtubeID, "", "mp4")
	if result != videoFile {
		t.Errorf("findDownloadedFile() for video = %q, want %q", result, videoFile)
	}
	
	// When looking for audio (mp3), should prefer audio file
	result2 := findDownloadedFile(tempDir, youtubeID, "", "mp3")
	if result2 != audioFile {
		t.Errorf("findDownloadedFile() for audio = %q, want %q", result2, audioFile)
	}
}

func TestFindDownloadedFileRecentFallback(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create an old file
	oldFile := filepath.Join(tempDir, "old.mp4")
	err := os.WriteFile(oldFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	// Set modification time to 1 minute ago
	oldTime := time.Now().Add(-1 * time.Minute)
	os.Chtimes(oldFile, oldTime, oldTime)
	
	// Create a recent file
	recentFile := filepath.Join(tempDir, "recent.mp4")
	err = os.WriteFile(recentFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create recent file: %v", err)
	}
	
	// Should find the recent file when no ID match
	result := findDownloadedFile(tempDir, "nonexistent", "99", "mp4")
	if result != recentFile {
		t.Errorf("findDownloadedFile() fallback = %q, want %q", result, recentFile)
	}
}

func TestFindDownloadedFileEmptyDir(t *testing.T) {
	tempDir := t.TempDir()
	
	result := findDownloadedFile(tempDir, "anyid", "anyformat", "mp4")
	if result != "" {
		t.Errorf("findDownloadedFile() = %q, want empty string for empty dir", result)
	}
}
