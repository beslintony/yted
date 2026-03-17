package app

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// findDownloadedFile searches for the downloaded file in the output directory
// Uses strict matching to avoid linking wrong files when multiple versions exist
// Priority order:
//  1. Files with [youtubeID][formatID] pattern (exact format match)
//  2. Files with [youtubeID][any_number] pattern (format was resolved)
//  3. Files with [youtubeID] pattern (backward compat, strict bracket matching)
//  4. Most recently modified media file (within 30s window, last resort)
func findDownloadedFile(outputDir, youtubeID, formatID, ext string) string {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return ""
	}

	// Determine expected media type based on quality/format
	isAudioFormat := ext == "mp3" || ext == "m4a" || ext == "ogg" ||
		strings.Contains(formatID, "audio")

	// Helper to check if file is expected media type
	isExpectedType := func(filename string) bool {
		lower := strings.ToLower(filename)
		if isAudioFormat {
			// For audio downloads, prioritize audio extensions
			return strings.HasSuffix(lower, ".mp3") ||
				strings.HasSuffix(lower, ".m4a") ||
				strings.HasSuffix(lower, ".ogg") ||
				strings.HasSuffix(lower, ".mp4") // mp3 can be in mp4 container
		}
		// For video downloads, look for video extensions
		return strings.HasSuffix(lower, ".mp4") ||
			strings.HasSuffix(lower, ".webm") ||
			strings.HasSuffix(lower, ".mkv") ||
			strings.HasSuffix(lower, ".avi")
	}

	// Helper to check if file is any media type
	isMediaFile := func(filename string) bool {
		lower := strings.ToLower(filename)
		return strings.HasSuffix(lower, ".mp4") ||
			strings.HasSuffix(lower, ".webm") ||
			strings.HasSuffix(lower, ".mkv") ||
			strings.HasSuffix(lower, ".avi") ||
			strings.HasSuffix(lower, ".mp3") ||
			strings.HasSuffix(lower, ".m4a") ||
			strings.HasSuffix(lower, ".ogg")
	}

	// PASS 1: Exact format match [youtubeID][formatID]
	if formatID != "" {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Look for exact pattern: [youtubeID][formatID]
			if strings.Contains(name, "["+youtubeID+"]["+formatID+"]") &&
				isExpectedType(name) {
				return filepath.Join(outputDir, name)
			}
		}
	}

	// PASS 2: Format resolved to different code [youtubeID][digits]
	if formatID != "" {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Match pattern: [youtubeID][number] - specific format was used
			if strings.Contains(name, "["+youtubeID+"][") &&
				!strings.Contains(name, "["+youtubeID+"]["+formatID+"]") &&
				isExpectedType(name) {
				return filepath.Join(outputDir, name)
			}
		}
	}

	// PASS 3: Backward compatibility [youtubeID] only (no format)
	// STRICT: Must be surrounded by brackets to avoid matching IDs in titles
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match pattern: [youtubeID].ext or [youtubeID]_something.ext
		// The ID must be in brackets to be valid
		if strings.Contains(name, "["+youtubeID+"]") &&
			!strings.Contains(name, "["+youtubeID+"][") && // Exclude format-specific files
			isExpectedType(name) {
			return filepath.Join(outputDir, name)
		}
	}

	// PASS 4: Last resort - most recently modified media file within 30s window
	// This handles edge cases but with strict time limit to avoid wrong matches
	cutoff := time.Now().Add(-30 * time.Second)
	var mostRecent os.DirEntry
	var mostRecentTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !isMediaFile(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Only consider files modified very recently
		if info.ModTime().After(cutoff) && info.ModTime().After(mostRecentTime) {
			// For audio downloads, prefer audio files; for video, prefer video
			if isExpectedType(name) {
				mostRecent = entry
				mostRecentTime = info.ModTime()
			} else if mostRecent == nil {
				// Only use non-preferred type if nothing else found
				mostRecent = entry
				mostRecentTime = info.ModTime()
			}
		}
	}

	if mostRecent != nil {
		return filepath.Join(outputDir, mostRecent.Name())
	}

	return ""
}
