package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"yted/internal/db"
	applog "yted/internal/log"
	"yted/internal/ytdl"
)

// addDownloadToLibrary adds a completed download to the video library.
// knownPath is yt-dlp's reported final output path; when it exists on disk
// it is used directly, otherwise we fall back to scanning outputDir.
func (a *App) addDownloadToLibrary(dl db.Download, outputDir, knownPath string) {
	logger := applog.GetLogger()

	// Extract YouTube ID from URL
	youtubeID := extractYoutubeID(dl.URL)
	if youtubeID == "" {
		logger.Warn("Download", "Could not extract YouTube ID from URL", map[string]string{"url": dl.URL})
		return
	}

	// Get video info if we have the URL and any metadata is missing
	ctx, cancel := context.WithTimeout(a.ctx, 2*time.Minute)
	defer cancel()

	var videoInfo *ytdl.VideoInfo
	var err error

	// Get video info if we have the URL and any metadata is missing.
	// This is a fallback: startDownload already persists fetched metadata
	// (see downloadNeedsMetadata/applyVideoInfoToDownload), so a download
	// that went through startDownload normally skips this second fetch.
	if libraryNeedsMetadata(&dl) {
		videoInfo, err = a.ytdl.GetInfo(ctx, cleanYouTubeURL(dl.URL))
		if err != nil {
			logger.Warn("Download", "Could not get video info for library", map[string]string{"error": err.Error()})
		} else {
			// Update download with info since we fetched it (metadata-only
			// update: never clobber status/timestamps of the completed row)
			applyVideoInfoToDownload(&dl, videoInfo)
			if updateErr := a.db.UpdateDownloadMetadata(dl.ID, videoInfo.Title, videoInfo.Channel, videoInfo.Thumbnail, videoInfo.Duration); updateErr != nil {
				logger.Warn("Download", "Failed to update download info", map[string]string{"error": updateErr.Error()})
			}
		}
	}

	// Get title
	title := getDownloadTitle(&dl, videoInfo, youtubeID)

	// Determine file extension based on format
	ext := getDownloadExtension(&dl)

	// Get format ID for finding the specific version
	formatID := ""
	if dl.FormatID != nil {
		formatID = *dl.FormatID
	}

	// Prefer yt-dlp's reported final path; it is exact, while directory
	// scanning is only a fallback for older/foreign files.
	filePath := ""
	if knownPath != "" {
		if info, err := os.Stat(knownPath); err == nil && !info.IsDir() {
			filePath = knownPath
			logger.Debug("Download", "Using yt-dlp reported output path", map[string]string{
				"path": filePath,
			})
		} else {
			logger.Warn("Download", "Reported output path missing, falling back to scan", map[string]string{
				"path": knownPath,
			})
		}
	}
	if filePath == "" {
		// Find the actual downloaded file in the output directory
		// yt-dlp sanitizes filenames, so we need to search for files with the YouTube ID and format
		filePath = findDownloadedFile(outputDir, youtubeID, formatID, ext)
	}
	if filePath == "" {
		// Fallback: try to construct the path (may not exist if yt-dlp sanitized differently)
		filename := fmt.Sprintf("%s.%s", title, ext)
		filePath = filepath.Join(outputDir, filename)
		logger.Warn("Download", "Could not find actual downloaded file, using estimated path", map[string]string{
			"path": filePath,
		})
	}

	// Get file size if file exists
	var fileSize int64
	if info, err := os.Stat(filePath); err == nil {
		fileSize = info.Size()
	}

	// Check if file is in our managed folder
	isManaged := a.fm != nil && a.fm.IsManagedFile(filePath)

	a.libraryMu.Lock()
	defer a.libraryMu.Unlock()

	// Create unique content identifier: YouTube ID + Format
	// This allows multiple versions (e.g., 720p vs 1080p) of same video
	contentHash := youtubeID
	if formatID != "" {
		contentHash = youtubeID + "_" + formatID
	}

	// Check if a video with this file_hash already exists (duplicate download)
	existingVideo, err := a.db.GetVideoByFileHash(contentHash)
	if err != nil {
		logger.Error("Download", "Failed to check for existing video by hash", err, map[string]string{
			"file_hash": contentHash,
		})
	}

	if existingVideo != nil {
		// Update existing record instead of creating duplicate
		logger.Info("Download", "Updating existing video record by hash", map[string]string{
			"id":         existingVideo.ID,
			"youtube_id": youtubeID,
			"file_hash":  contentHash,
		})

		existingVideo.FilePath = filePath
		existingVideo.FileSize = fileSize
		existingVideo.DownloadedAt = time.Now()

		if err := a.db.UpdateVideo(existingVideo); err != nil {
			logger.Error("Download", "Failed to update existing video", err, map[string]string{
				"id": existingVideo.ID,
			})
			return
		}

		runtime.EventsEmit(a.ctx, "library:updated", existingVideo)
		return
	}

	// Also check by YouTube ID to find and clean up legacy duplicates
	existingByID, err := a.db.GetVideosByYoutubeID(youtubeID)
	if err != nil {
		logger.Error("Download", "Failed to check for existing video by ID", err, map[string]string{
			"youtube_id": youtubeID,
		})
	}

	if len(existingByID) > 0 {
		// Use the most recent entry and delete the rest (cleanup duplicates)
		logger.Info("Download", "Found existing videos by YouTube ID, cleaning up duplicates", map[string]interface{}{
			"youtube_id": youtubeID,
			"count":      len(existingByID),
		})

		// Update the first (most recent) entry
		primary := existingByID[0]
		primary.FilePath = filePath
		primary.FileSize = fileSize
		primary.FileHash = contentHash
		primary.DownloadedAt = time.Now()

		if err := a.db.UpdateVideo(&primary); err != nil {
			logger.Error("Download", "Failed to update primary video", err, map[string]string{
				"id": primary.ID,
			})
			return
		}

		// Delete duplicate entries
		for i := 1; i < len(existingByID); i++ {
			if err := a.db.DeleteVideoByID(existingByID[i].ID); err != nil {
				logger.Error("Download", "Failed to delete duplicate video", err, map[string]string{
					"id": existingByID[i].ID,
				})
			} else {
				logger.Debug("Download", "Deleted duplicate video entry", map[string]string{
					"id": existingByID[i].ID,
				})
			}
		}

		runtime.EventsEmit(a.ctx, "library:updated", primary)
		return
	}

	// Create new video record
	video := createVideoRecord(&dl, videoInfo, youtubeID, title, filePath, fileSize, contentHash, isManaged)

	// Save to database
	if err := a.db.CreateVideo(video); err != nil {
		logger.Error("Download", "Failed to add video to library", err, map[string]string{
			"youtube_id": youtubeID,
			"title":      title,
		})
		return
	}

	logger.Info("Download", "Video added to library", map[string]string{
		"id":         video.ID,
		"youtube_id": youtubeID,
		"title":      title,
	})

	// Emit event to refresh library
	runtime.EventsEmit(a.ctx, "library:updated", video)
}

// getDownloadTitle gets the title from download info
func getDownloadTitle(dl *db.Download, videoInfo *ytdl.VideoInfo, youtubeID string) string {
	if dl.Title != nil && *dl.Title != "" {
		return *dl.Title
	}
	if videoInfo != nil && videoInfo.Title != "" {
		return videoInfo.Title
	}
	return youtubeID
}

// getDownloadExtension determines the file extension based on format
func getDownloadExtension(dl *db.Download) string {
	if dl.Quality != nil && *dl.Quality == "audio" {
		return "mp3"
	}
	// Only an exclusively-audio format is mp3; combined selectors like
	// "bestvideo+bestaudio" produce a video file
	if dl.FormatID != nil && strings.Contains(*dl.FormatID, "audio") && !strings.Contains(*dl.FormatID, "video") {
		return "mp3"
	}
	return "mp4"
}

// createVideoRecord creates a new video record from download info
func createVideoRecord(dl *db.Download, videoInfo *ytdl.VideoInfo, youtubeID, title, filePath string, fileSize int64, contentHash string, isManaged bool) *db.Video {
	video := &db.Video{
		ID:           uuid.New().String(),
		YoutubeID:    youtubeID,
		Title:        title,
		FilePath:     filePath,
		FileSize:     fileSize,
		FileHash:     contentHash,
		IsManaged:    isManaged,
		DownloadedAt: time.Now(),
	}

	// Add optional fields
	if dl.Channel != nil {
		video.Channel = *dl.Channel
	}
	if dl.ThumbnailURL != nil {
		video.ThumbnailURL = *dl.ThumbnailURL
	}
	if dl.Quality != nil {
		video.Quality = *dl.Quality
	}
	if dl.FormatID != nil {
		video.Format = *dl.FormatID
	}

	// Get duration and description from video info if available
	if videoInfo != nil {
		video.Duration = videoInfo.Duration
		video.Description = videoInfo.Description
		video.ChannelID = videoInfo.ChannelID
		if videoInfo.Channel != "" && video.Channel == "" {
			video.Channel = videoInfo.Channel
		}
		if videoInfo.Thumbnail != "" && video.ThumbnailURL == "" {
			video.ThumbnailURL = videoInfo.Thumbnail
		}
	}

	// Fallback to the saved download.Duration if we didn't fetch it just now
	if video.Duration == 0 && dl.Duration != nil {
		video.Duration = *dl.Duration
	}

	return video
}
