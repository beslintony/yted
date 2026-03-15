package app

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// findDownloadedFile searches for the downloaded file in the output directory
// Looks for files containing the YouTube ID AND format (for multiple versions support)
// If formatID is empty, falls back to just matching YouTube ID
func findDownloadedFile(outputDir, youtubeID, formatID, ext string) string {
	// Read directory contents
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return ""
	}

	// Build the expected format suffix for matching
	// New format: %(title)s [%(id)s][%(format_id)s].%(ext)s
	formatSuffix := ""
	if formatID != "" {
		formatSuffix = "[" + formatID + "]"
	}

	// First pass: Look for files with BOTH YouTube ID AND format ID
	if formatSuffix != "" {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Check if filename contains both the YouTube ID and format ID
			if strings.Contains(name, youtubeID) && strings.Contains(name, formatSuffix) {
				// Check extension (video or audio)
				lowerName := strings.ToLower(name)
				if strings.HasSuffix(lowerName, ".mp4") ||
					strings.HasSuffix(lowerName, ".webm") ||
					strings.HasSuffix(lowerName, ".mkv") ||
					strings.HasSuffix(lowerName, ".mp3") ||
					strings.HasSuffix(lowerName, ".m4a") ||
					strings.HasSuffix(lowerName, ".ogg") {
					return filepath.Join(outputDir, name)
				}
			}
		}
	}

	// Second pass: Look for files with just the YouTube ID (backward compatibility)
	// This handles older downloads without format in filename
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Check if filename contains the YouTube ID
		if strings.Contains(name, youtubeID) {
			// Check extension (video or audio)
			lowerName := strings.ToLower(name)
			if strings.HasSuffix(lowerName, ".mp4") ||
				strings.HasSuffix(lowerName, ".webm") ||
				strings.HasSuffix(lowerName, ".mkv") ||
				strings.HasSuffix(lowerName, ".mp3") ||
				strings.HasSuffix(lowerName, ".m4a") ||
				strings.HasSuffix(lowerName, ".ogg") {
				return filepath.Join(outputDir, name)
			}
		}
	}

	// Second pass: Look for files modified in the last 2 minutes
	// This handles cases where the template doesn't include the ID
	cutoff := time.Now().Add(-2 * time.Minute)
	var mostRecent os.DirEntry
	var mostRecentTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if file is a video/audio file and was modified recently
		lowerName := strings.ToLower(entry.Name())
		isMedia := strings.HasSuffix(lowerName, ".mp4") ||
			strings.HasSuffix(lowerName, ".webm") ||
			strings.HasSuffix(lowerName, ".mkv") ||
			strings.HasSuffix(lowerName, ".mp3") ||
			strings.HasSuffix(lowerName, ".m4a") ||
			strings.HasSuffix(lowerName, ".ogg") ||
			strings.HasSuffix(lowerName, ".part") // yt-dlp partial download

		if isMedia && info.ModTime().After(cutoff) {
			if info.ModTime().After(mostRecentTime) {
				mostRecent = entry
				mostRecentTime = info.ModTime()
			}
		}
	}

	if mostRecent != nil {
		path := filepath.Join(outputDir, mostRecent.Name())
		// If it's a .part file, return without the extension
		if strings.HasSuffix(path, ".part") {
			return path[:len(path)-5]
		}
		return path
	}

	return ""
}
