package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"yted/internal/db"
)

func TestFindDownloadedFile(t *testing.T) {
	// Test case 1: File with [youtubeID][formatID] pattern
	t.Run("exact format match", func(t *testing.T) {
		tempDir := t.TempDir()
		youtubeID := "dQw4w9WgXcQ"
		formatID := "18"

		expectedFile := filepath.Join(tempDir, "Video Title ["+youtubeID+"]["+formatID+"].mp4")
		err := os.WriteFile(expectedFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := findDownloadedFile(tempDir, youtubeID, formatID, "mp4")
		if result != expectedFile {
			t.Errorf("findDownloadedFile() = %q, want %q", result, expectedFile)
		}
	})

	// Test case 2: File with just [youtubeID] pattern (backward compat)
	t.Run("backward compat", func(t *testing.T) {
		tempDir := t.TempDir()
		youtubeID := "abc123"
		expectedFile := filepath.Join(tempDir, "Another Video ["+youtubeID+"].mp4")
		err := os.WriteFile(expectedFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result := findDownloadedFile(tempDir, youtubeID, "", "mp4")
		if result != expectedFile {
			t.Errorf("findDownloadedFile() = %q, want %q", result, expectedFile)
		}
	})

	// Test case 3: Non-existent file
	t.Run("non-existent", func(t *testing.T) {
		tempDir := t.TempDir()
		result := findDownloadedFile(tempDir, "nonexistent", "99", "mp4")
		if result != "" {
			t.Errorf("findDownloadedFile() = %q, want empty string", result)
		}
	})
}

// TestFindDownloadedFileRejectsFragments ensures yt-dlp fragment files
// (unmerged .fNNN.ext parts) are never matched as the final video
func TestFindDownloadedFileRejectsFragments(t *testing.T) {
	tempDir := t.TempDir()
	youtubeID := "TVXWJROgUv8"
	formatID := "bestvideo+bestaudio/best"

	// Simulate an unmerged download: audio fragment + video fragment
	audioFrag := filepath.Join(tempDir, "Some Video ["+youtubeID+"][399+140].f140.m4a")
	videoFrag := filepath.Join(tempDir, "Some Video ["+youtubeID+"][399+140].f399.mp4")
	for _, f := range []string{audioFrag, videoFrag} {
		if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	result := findDownloadedFile(tempDir, youtubeID, formatID, "mp4")
	if result != "" {
		t.Errorf("findDownloadedFile() matched fragment %q, want empty string", result)
	}
}

// TestFindDownloadedFileCombinedFormatIsVideo ensures combined selectors
// like "bestvideo+bestaudio" are treated as video, not audio
func TestFindDownloadedFileCombinedFormatIsVideo(t *testing.T) {
	tempDir := t.TempDir()
	youtubeID := "abc123xyz00"
	formatID := "bestvideo+bestaudio/best"

	// Merged output exists alongside a stray audio file
	merged := filepath.Join(tempDir, "Video ["+youtubeID+"][399+140].mp4")
	strayAudio := filepath.Join(tempDir, "Video ["+youtubeID+"][399+140].m4a")
	for _, f := range []string{merged, strayAudio} {
		if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	result := findDownloadedFile(tempDir, youtubeID, formatID, "mp4")
	if result != merged {
		t.Errorf("findDownloadedFile() = %q, want %q", result, merged)
	}
}

func TestGetDownloadExtensionCombinedFormat(t *testing.T) {
	formatID := "bestvideo+bestaudio/best"
	quality := "best"
	dl := &db.Download{FormatID: &formatID, Quality: &quality}
	if ext := getDownloadExtension(dl); ext != "mp4" {
		t.Errorf("getDownloadExtension() = %q, want mp4 for combined format", ext)
	}

	audioFormatID := "bestaudio"
	dl = &db.Download{FormatID: &audioFormatID}
	if ext := getDownloadExtension(dl); ext != "mp3" {
		t.Errorf("getDownloadExtension() = %q, want mp3 for audio-only format", ext)
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
