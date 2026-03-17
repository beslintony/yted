package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"yted/internal/db"
	applog "yted/internal/log"
)

// CacheInfo provides information about cache status
type CacheInfo struct {
	DownloadCount      int   `json:"download_count"`
	CompletedCount     int   `json:"completed_count"`
	PendingCount       int   `json:"pending_count"`
	VideoCount         int   `json:"video_count"`
	TotalLibrarySize   int64 `json:"total_library_size"`
	OrphanedFilesCount int   `json:"orphaned_files_count"`
	OrphanedFilesSize  int64 `json:"orphaned_files_size"`
}

// GetCacheInfo returns comprehensive cache information
func (a *App) GetCacheInfo() (*CacheInfo, error) {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	info := &CacheInfo{}

	// Get download counts
	downloads, err := a.db.ListDownloads()
	if err != nil {
		logger.Error("Cache", "Failed to list downloads", err)
		return nil, err
	}
	info.DownloadCount = len(downloads)

	for _, dl := range downloads {
		switch dl.Status {
		case "completed":
			info.CompletedCount++
		case "pending", "downloading":
			info.PendingCount++
		}
	}

	// Get video library stats
	totalVideos, totalSize, err := a.db.GetStats()
	if err != nil {
		logger.Error("Cache", "Failed to get library stats", err)
		return nil, err
	}
	info.VideoCount = totalVideos
	info.TotalLibrarySize = totalSize

	// Check for orphaned files (files in YTed folder not in DB)
	if a.fm != nil {
		orphanedCount, orphanedSize, err := a.findOrphanedFiles()
		if err != nil {
			logger.Warn("Cache", "Failed to find orphaned files", map[string]string{"error": err.Error()})
		} else {
			info.OrphanedFilesCount = orphanedCount
			info.OrphanedFilesSize = orphanedSize
		}
	}

	return info, nil
}

// findOrphanedFiles finds files in the YTed folder that are not tracked in the database
func (a *App) findOrphanedFiles() (int, int64, error) {
	if a.fm == nil {
		return 0, 0, nil
	}

	downloadPath := a.fm.GetDownloadPath()
	if downloadPath == "" {
		return 0, 0, nil
	}

	// Get all tracked file paths from database
	trackedFiles := make(map[string]bool)
	videos, err := a.db.ListVideosWithHash(db.ListVideosOptions{
		Limit:  10000,
		Offset: 0,
	})
	if err != nil {
		return 0, 0, err
	}

	for _, v := range videos {
		if v.FilePath != "" {
			trackedFiles[strings.ToLower(v.FilePath)] = true
		}
	}

	// Scan YTed folder for files
	entries, err := os.ReadDir(downloadPath)
	if err != nil {
		return 0, 0, err
	}

	var orphanedCount int
	var orphanedSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Check if it's a media file
		lowerName := strings.ToLower(name)
		if !isMediaFile(lowerName) {
			continue
		}

		fullPath := filepath.Join(downloadPath, name)

		// Check if this file is tracked
		if !trackedFiles[strings.ToLower(fullPath)] {
			info, err := entry.Info()
			if err == nil {
				orphanedCount++
				orphanedSize += info.Size()
			}
		}
	}

	return orphanedCount, orphanedSize, nil
}

// isMediaFile checks if filename has a media extension
func isMediaFile(filename string) bool {
	return strings.HasSuffix(filename, ".mp4") ||
		strings.HasSuffix(filename, ".webm") ||
		strings.HasSuffix(filename, ".mkv") ||
		strings.HasSuffix(filename, ".mp3") ||
		strings.HasSuffix(filename, ".m4a") ||
		strings.HasSuffix(filename, ".ogg")
}

// CleanupOrphanedFiles removes files from YTed folder that are not in the database
func (a *App) CleanupOrphanedFiles(deleteFiles bool) (map[string]interface{}, error) {
	logger := applog.GetLogger()

	if a.fm == nil {
		return nil, fmt.Errorf("file manager not initialized")
	}

	downloadPath := a.fm.GetDownloadPath()
	if downloadPath == "" {
		return nil, fmt.Errorf("download path not configured")
	}

	// Get orphaned files
	orphanedCount, orphanedSize, err := a.findOrphanedFiles()
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"found":      orphanedCount,
		"size":       orphanedSize,
		"deleted":    0,
		"failed":     0,
		"deleteMode": deleteFiles,
	}

	if orphanedCount == 0 {
		return result, nil
	}

	// If not in delete mode, just return info
	if !deleteFiles {
		return result, nil
	}

	// Actually delete the orphaned files
	entries, err := os.ReadDir(downloadPath)
	if err != nil {
		return nil, err
	}

	// Get tracked files again for comparison
	trackedFiles := make(map[string]bool)
	videos, err := a.db.ListVideosWithHash(db.ListVideosOptions{
		Limit:  10000,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	for _, v := range videos {
		if v.FilePath != "" {
			trackedFiles[strings.ToLower(v.FilePath)] = true
		}
	}

	var deletedCount, failedCount int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		lowerName := strings.ToLower(name)
		if !isMediaFile(lowerName) {
			continue
		}

		fullPath := filepath.Join(downloadPath, name)
		if !trackedFiles[strings.ToLower(fullPath)] {
			if err := os.Remove(fullPath); err != nil {
				logger.Error("Cache", "Failed to delete orphaned file", err, map[string]string{
					"path": fullPath,
				})
				failedCount++
			} else {
				logger.Info("Cache", "Deleted orphaned file", map[string]string{
					"path": fullPath,
				})
				deletedCount++
			}
		}
	}

	result["deleted"] = deletedCount
	result["failed"] = failedCount

	return result, nil
}

// ClearTempFiles removes temporary/partial download files
func (a *App) ClearTempFiles() (map[string]interface{}, error) {
	logger := applog.GetLogger()

	if a.fm == nil {
		return nil, fmt.Errorf("file manager not initialized")
	}

	downloadPath := a.fm.GetDownloadPath()
	if downloadPath == "" {
		return nil, fmt.Errorf("download path not configured")
	}

	entries, err := os.ReadDir(downloadPath)
	if err != nil {
		return nil, err
	}

	var foundCount, deletedCount, failedCount int
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// yt-dlp temporary files
		if strings.HasSuffix(name, ".part") ||
			strings.HasSuffix(name, ".ytdl") ||
			strings.HasSuffix(name, ".temp") ||
			strings.Contains(name, ".part-") {
			foundCount++
			fullPath := filepath.Join(downloadPath, name)
			info, err := entry.Info()
			if err == nil {
				totalSize += info.Size()
			}

			if err := os.Remove(fullPath); err != nil {
				logger.Error("Cache", "Failed to delete temp file", err, map[string]string{
					"path": fullPath,
				})
				failedCount++
			} else {
				logger.Info("Cache", "Deleted temp file", map[string]string{
					"path": fullPath,
				})
				deletedCount++
			}
		}
	}

	return map[string]interface{}{
		"found":   foundCount,
		"deleted": deletedCount,
		"failed":  failedCount,
		"size":    totalSize,
	}, nil
}

// RepairLibrary scans library and fixes inconsistencies
func (a *App) RepairLibrary() (map[string]interface{}, error) {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	result := map[string]interface{}{
		"checked":      0,
		"missingFiles": 0,
		"fixed":        0,
		"errors":       0,
	}

	// Get all videos
	videos, err := a.db.ListVideosWithHash(db.ListVideosOptions{
		Limit:  10000,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	checked := 0
	missingFiles := 0
	fixed := 0
	errors := 0

	for _, video := range videos {
		checked++

		if video.FilePath == "" {
			continue
		}

		// Check if file exists
		_, err := os.Stat(video.FilePath)
		if os.IsNotExist(err) {
			missingFiles++
			logger.Warn("Library", "Video file missing", map[string]string{
				"video_id": video.ID,
				"path":     video.FilePath,
			})
			// Optionally: mark as missing in database
		} else if err != nil {
			errors++
			logger.Error("Library", "Error checking file", err, map[string]string{
				"video_id": video.ID,
				"path":     video.FilePath,
			})
		}
	}

	result["checked"] = checked
	result["missingFiles"] = missingFiles
	result["fixed"] = fixed
	result["errors"] = errors

	logger.Info("Library", "Library repair scan completed", result)

	return result, nil
}
