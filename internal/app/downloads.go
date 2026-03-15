package app

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"yted/internal/db"
	applog "yted/internal/log"
	"yted/internal/ytdl"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var libraryMutex sync.Mutex

// VideoInfoResult is exposed to frontend
type VideoInfoResult struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Channel     string            `json:"channel"`
	ChannelID   string            `json:"channel_id"`
	Duration    int               `json:"duration"`
	Description string            `json:"description"`
	Thumbnail   string            `json:"thumbnail"`
	Formats     []ytdl.FormatInfo `json:"formats"`
}

// GetVideoInfo extracts video information from URL
func (a *App) GetVideoInfo(url string) (*VideoInfoResult, error) {
	logger := applog.GetLogger()

	logger.Info("Download", "Fetching video info", map[string]string{"url": url})

	if a.ytdl == nil {
		err := fmt.Errorf("ytdl client not initialized")
		logger.Error("Download", "ytdl client not ready", err)
		return nil, err
	}

	if !ytdl.IsValidURL(url) {
		err := fmt.Errorf("invalid URL")
		logger.Warn("Download", "Invalid URL provided", map[string]string{"url": url})
		return nil, err
	}

	ctx := context.Background()
	info, err := a.ytdl.GetInfo(ctx, url)
	if err != nil {
		logger.Error("Download", "Failed to get video info", err, map[string]string{"url": url})
		return nil, err
	}

	logger.Info("Download", "Video info fetched successfully", map[string]string{
		"id":    info.ID,
		"title": info.Title,
	})

	return &VideoInfoResult{
		ID:          info.ID,
		Title:       info.Title,
		Channel:     info.Channel,
		ChannelID:   info.ChannelID,
		Duration:    info.Duration,
		Description: info.Description,
		Thumbnail:   info.Thumbnail,
		Formats:     info.Formats,
	}, nil
}

// AddDownload adds a new download to the queue
func (a *App) AddDownload(url string, formatID string, quality string) (string, error) {
	logger := applog.GetLogger()

	logger.Info("Download", "Adding download", map[string]string{
		"url":      url,
		"formatID": formatID,
		"quality":  quality,
	})

	if a.db == nil {
		err := fmt.Errorf("database not initialized")
		logger.Error("Download", "Database not ready", err)
		return "", err
	}

	download := &db.Download{
		ID:       uuid.New().String(),
		URL:      url,
		Status:   "pending",
		Progress: 0,
		FormatID: &formatID,
		Quality:  &quality,
	}

	if err := a.db.CreateDownload(download); err != nil {
		logger.Error("Download", "Failed to create download record", err)
		return "", err
	}

	logger.Info("Download", "Download added to queue", map[string]string{
		"id": download.ID,
	})

	// Emit event to notify frontend
	runtime.EventsEmit(a.ctx, "download:added", download)

	// Try to start the download if under limit
	go a.processDownloads()

	return download.ID, nil
}

// GetDownloads returns all downloads
func (a *App) GetDownloads() ([]db.Download, error) {
	logger := applog.GetLogger()

	if a.db == nil {
		logger.Warn("Download", "Database not initialized when getting downloads")
		return nil, nil
	}

	downloads, err := a.db.ListDownloads()
	if err != nil {
		logger.Error("Download", "Failed to list downloads", err)
		return nil, err
	}

	logger.Debug("Download", "Retrieved downloads", map[string]int{
		"count": len(downloads),
	})

	return downloads, nil
}

// GetDownloadsByStatus returns downloads filtered by status
func (a *App) GetDownloadsByStatus(status string) ([]db.Download, error) {
	if a.db == nil {
		return nil, nil
	}
	return a.db.ListDownloads(status)
}

// PauseDownload pauses a download
func (a *App) PauseDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil
	}

	if err := a.db.UpdateDownloadStatus(id, "paused"); err != nil {
		logger.Error("Download", "Failed to pause download", err, map[string]string{"id": id})
		return err
	}

	logger.Info("Download", "Download paused", map[string]string{"id": id})
	runtime.EventsEmit(a.ctx, "download:paused", id)
	return nil
}

// ResumeDownload resumes a paused download
func (a *App) ResumeDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil
	}

	if err := a.db.UpdateDownloadStatus(id, "pending"); err != nil {
		logger.Error("Download", "Failed to resume download", err, map[string]string{"id": id})
		return err
	}

	logger.Info("Download", "Download resumed", map[string]string{"id": id})
	runtime.EventsEmit(a.ctx, "download:resumed", id)

	// Try to process downloads
	go a.processDownloads()

	return nil
}

// RetryDownload retries a failed download
func (a *App) RetryDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil
	}

	if err := a.db.UpdateDownloadStatus(id, "pending"); err != nil {
		logger.Error("Download", "Failed to retry download", err, map[string]string{"id": id})
		return err
	}

	if err := a.db.UpdateDownloadProgress(id, 0); err != nil {
		logger.Error("Download", "Failed to reset download progress", err, map[string]string{"id": id})
		return err
	}

	logger.Info("Download", "Download retry initiated", map[string]string{"id": id})
	runtime.EventsEmit(a.ctx, "download:retried", id)

	// Try to process downloads
	go a.processDownloads()

	return nil
}

// CancelDownload cancels and removes a download
func (a *App) CancelDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil
	}

	if err := a.db.DeleteDownload(id); err != nil {
		logger.Error("Download", "Failed to cancel download", err, map[string]string{"id": id})
		return err
	}

	logger.Info("Download", "Download cancelled", map[string]string{"id": id})
	runtime.EventsEmit(a.ctx, "download:cancelled", id)
	return nil
}

// ClearCompletedDownloads removes all completed downloads
func (a *App) ClearCompletedDownloads() error {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil
	}

	if err := a.db.DeleteCompletedDownloads(); err != nil {
		logger.Error("Download", "Failed to clear completed downloads", err)
		return err
	}

	logger.Info("Download", "Completed downloads cleared")
	runtime.EventsEmit(a.ctx, "downloads:cleared", nil)
	return nil
}

// processDownloads starts pending downloads up to the concurrent limit
func (a *App) processDownloads() {
	logger := applog.GetLogger()
	
	// Prevent concurrent execution to avoid race conditions
	a.downloadMu.Lock()
	defer a.downloadMu.Unlock()

	if a.db == nil || a.config == nil || a.ytdl == nil {
		logger.Warn("Download", "Cannot process downloads - dependencies not ready")
		return
	}

	maxConcurrent := a.config.Get().MaxConcurrentDownloads
	if maxConcurrent < 1 {
		maxConcurrent = 3
	}

	// Count active downloads
	activeCount, err := a.db.CountActiveDownloads()
	if err != nil {
		logger.Error("Download", "Failed to count active downloads", err)
		return
	}

	// Calculate how many we can start
	slotsAvailable := maxConcurrent - activeCount
	if slotsAvailable <= 0 {
		logger.Debug("Download", "No download slots available", map[string]int{
			"active": activeCount,
			"max":    maxConcurrent,
		})
		return
	}

	// Get pending downloads
	pending, err := a.db.GetPendingDownloads(slotsAvailable)
	if err != nil {
		logger.Error("Download", "Failed to get pending downloads", err)
		return
	}

	if len(pending) == 0 {
		return
	}

	logger.Info("Download", "Starting downloads", map[string]interface{}{
		"count":          len(pending),
		"slotsAvailable": slotsAvailable,
	})

	// Start each download - mark as started SYNCHRONOUSLY before spawning goroutine
	// This prevents race conditions if processDownloads is called again quickly
	for _, dl := range pending {
		// Mark as started in DB first (synchronously)
		if err := a.db.StartDownload(dl.ID); err != nil {
			logger.Error("Download", "Failed to mark download as started, skipping", err, map[string]string{
				"id": dl.ID,
			})
			continue
		}
		
		// Now spawn the goroutine for the actual download
		go a.startDownload(dl)
	}
}

// startDownload starts a single download
// Note: The download must already be marked as 'started' in the database
// before calling this function (done synchronously in processDownloads)
func (a *App) startDownload(dl db.Download) {
	logger := applog.GetLogger()
	ctx := context.Background()

	logger.Info("Download", "Download worker starting", map[string]string{
		"id":  dl.ID,
		"url": dl.URL,
	})

	// Verify download is still in 'downloading' state (could have been cancelled/paused)
	currentDownload, err := a.db.GetDownload(dl.ID)
	if err != nil {
		logger.Error("Download", "Failed to verify download status", err, map[string]string{"id": dl.ID})
		return
	}
	if currentDownload == nil || currentDownload.Status != "downloading" {
		logger.Info("Download", "Download no longer active, aborting", map[string]string{
			"id":     dl.ID,
			"status": currentDownload.Status,
		})
		return
	}

	// Get video info if not already have title
	var title, channel, thumbnail string
	if dl.Title != nil {
		title = *dl.Title
	}
	if dl.Channel != nil {
		channel = *dl.Channel
	}
	if dl.ThumbnailURL != nil {
		thumbnail = *dl.ThumbnailURL
	}

	if title == "" {
		info, err := a.ytdl.GetInfo(ctx, dl.URL)
		if err == nil {
			title = info.Title
			channel = info.Channel
			thumbnail = info.Thumbnail

			// Update download with info
			dl.Title = &title
			dl.Channel = &channel
			dl.ThumbnailURL = &thumbnail
			dl.Duration = &info.Duration
			a.db.UpdateDownload(&dl)

			logger.Info("Download", "Video info retrieved", map[string]string{
				"id":    dl.ID,
				"title": title,
			})
		} else {
			logger.Warn("Download", "Could not get video info", map[string]string{
				"id":  dl.ID,
				"url": dl.URL,
			})
		}
	}

	// Emit started event
	runtime.EventsEmit(a.ctx, "download:started", dl)

	// Download options
	var format string
	if dl.FormatID != nil {
		format = *dl.FormatID
	}

	var quality string
	if dl.Quality != nil {
		quality = *dl.Quality
	}

	opts := ytdl.DownloadOptions{
		Format:    format,
		Quality:   quality,
		OutputDir: a.config.Get().DownloadPath,
		ProxyURL:  a.config.Get().ProxyURL,
	}

	// Progress callback
	progressCallback := func(progress ytdl.DownloadProgress) {
		a.db.UpdateDownloadProgress(dl.ID, progress.Percent)
		runtime.EventsEmit(a.ctx, "download:progress", map[string]interface{}{
			"id":       dl.ID,
			"progress": progress.Percent,
			"speed":    progress.Speed,
			"eta":      progress.ETA,
		})

		logger.Debug("Download", "Progress update", map[string]interface{}{
			"id":       dl.ID,
			"progress": progress.Percent,
		})
	}

	// Perform download
	err = a.ytdl.Download(ctx, dl.URL, opts, progressCallback)
	if err != nil {
		logger.Error("Download", "Download failed", err, map[string]string{"id": dl.ID})
		a.db.FailDownload(dl.ID, err.Error())
		runtime.EventsEmit(a.ctx, "download:error", map[string]interface{}{
			"id":    dl.ID,
			"error": err.Error(),
		})
		return
	}

	// Mark as completed
	if err := a.db.CompleteDownload(dl.ID); err != nil {
		logger.Error("Download", "Failed to mark download as completed", err, map[string]string{"id": dl.ID})
	}

	// Add to library - construct the expected file path
	// The file was downloaded to opts.OutputDir with the filename template
	// We need to add this to the videos table
	go a.addDownloadToLibrary(dl, opts.OutputDir)

	logger.Info("Download", "Download completed successfully", map[string]string{"id": dl.ID})
	runtime.EventsEmit(a.ctx, "download:completed", dl.ID)

	// Process more downloads
	a.processDownloads()
}

// ValidateURL checks if a URL is valid
func (a *App) ValidateURL(url string) bool {
	return ytdl.IsValidURL(url)
}

// addDownloadToLibrary adds a completed download to the video library
func (a *App) addDownloadToLibrary(dl db.Download, outputDir string) {
	logger := applog.GetLogger()

	// Extract YouTube ID from URL
	youtubeID := extractYoutubeID(dl.URL)
	if youtubeID == "" {
		logger.Warn("Download", "Could not extract YouTube ID from URL", map[string]string{"url": dl.URL})
		return
	}

	// Get video info if we have the URL and any metadata is missing
	ctx := context.Background()
	var videoInfo *ytdl.VideoInfo
	var err error

	if dl.Title == nil || *dl.Title == "" || dl.Duration == nil || *dl.Duration == 0 {
		videoInfo, err = a.ytdl.GetInfo(ctx, dl.URL)
		if err != nil {
			logger.Warn("Download", "Could not get video info for library", map[string]string{"error": err.Error()})
		} else {
			// Update download with info since we fetched it
			dl.Title = &videoInfo.Title
			dl.Channel = &videoInfo.Channel
			dl.ThumbnailURL = &videoInfo.Thumbnail
			dl.Duration = &videoInfo.Duration
			a.db.UpdateDownload(&dl)
		}
	}

	// Get title
	var title string
	if dl.Title != nil {
		title = *dl.Title
	} else if videoInfo != nil {
		title = videoInfo.Title
	} else {
		title = youtubeID
	}

	// Determine file extension based on format
	ext := "mp4"
	if dl.Quality != nil && *dl.Quality == "audio" {
		ext = "mp3"
	} else if dl.FormatID != nil && strings.Contains(*dl.FormatID, "audio") {
		ext = "mp3"
	}

	// Get format ID for finding the specific version
	formatID := ""
	if dl.FormatID != nil {
		formatID = *dl.FormatID
	}

	// Find the actual downloaded file in the output directory
	// yt-dlp sanitizes filenames, so we need to search for files with the YouTube ID and format
	filePath := findDownloadedFile(outputDir, youtubeID, formatID, ext)
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
	
	libraryMutex.Lock()
	defer libraryMutex.Unlock()

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
	video := &db.Video{
		ID:            uuid.New().String(),
		YoutubeID:     youtubeID,
		Title:         title,
		FilePath:      filePath,
		FileSize:      fileSize,
		FileHash:      contentHash, // Unique ID for this specific version
		IsManaged:     isManaged,
		DownloadedAt:  time.Now(),
		WatchPosition: 0,
		WatchCount:    0,
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

// extractYoutubeID extracts the YouTube video ID from a URL
func extractYoutubeID(videoURL string) string {
	parsedURL, err := url.Parse(videoURL)
	if err != nil {
		return ""
	}

	// Handle youtu.be short URLs
	if strings.Contains(parsedURL.Host, "youtu.be") {
		path := strings.TrimPrefix(parsedURL.Path, "/")
		return strings.Split(path, "/")[0]
	}

	// Handle youtube.com URLs
	query := parsedURL.Query()
	if v := query.Get("v"); v != "" {
		return v
	}

	// Handle shorts URLs
	if strings.Contains(parsedURL.Path, "/shorts/") {
		parts := strings.Split(parsedURL.Path, "/shorts/")
		if len(parts) > 1 {
			return strings.Split(parts[1], "/")[0]
		}
	}

	return ""
}

// GetIncompleteDownloads returns all downloads that are not completed (for restoring queue)
func (a *App) GetIncompleteDownloads() ([]db.Download, error) {
	if a.db == nil {
		return nil, nil
	}
	return a.db.GetIncompleteDownloads()
}

// DownloadResult is exposed to frontend for queue restoration
type DownloadResult struct {
	ID           string  `json:"id"`
	URL          string  `json:"url"`
	Status       string  `json:"status"`
	Progress     float64 `json:"progress"`
	Title        string  `json:"title"`
	Channel      string  `json:"channel"`
	ThumbnailURL string  `json:"thumbnail_url"`
	FormatID     string  `json:"format_id"`
	Quality      string  `json:"quality"`
	ErrorMessage string  `json:"error_message"`
	YoutubeID    string  `json:"youtube_id"`
}

// GetDownloadQueue returns incomplete downloads for frontend to restore
// This is called by the frontend after it's ready, instead of using events
func (a *App) GetDownloadQueue() ([]DownloadResult, error) {
	logger := applog.GetLogger()
	
	if a.db == nil {
		logger.Warn("Download", "GetDownloadQueue called but db is nil")
		return nil, nil
	}
	
	downloads, err := a.db.GetIncompleteDownloads()
	if err != nil {
		logger.Error("Download", "Failed to get download queue", err)
		return nil, err
	}
	
	if len(downloads) == 0 {
		logger.Debug("Download", "GetDownloadQueue: no incomplete downloads found")
		return nil, nil
	}
	
	logger.Info("Download", "Returning download queue to frontend", map[string]int{
		"count": len(downloads),
	})
	
	// Reset downloads that were 'downloading' to 'pending' so they can be retried
	for _, dl := range downloads {
		logger.Debug("Download", "Checking download status for reset", map[string]string{
			"id":     dl.ID,
			"status": dl.Status,
		})
		if dl.Status == "downloading" {
			logger.Info("Download", "Resetting stuck download to pending", map[string]string{
				"id": dl.ID,
			})
			if err := a.db.UpdateDownloadStatus(dl.ID, "pending"); err != nil {
				logger.Error("Download", "Failed to reset download status", err, map[string]string{
					"id": dl.ID,
				})
			} else {
				dl.Status = "pending"
				logger.Debug("Download", "Successfully reset download to pending", map[string]string{
					"id": dl.ID,
				})
			}
		}
	}
	
	// Convert to DownloadResult
	results := make([]DownloadResult, 0, len(downloads))
	for _, dl := range downloads {
		result := DownloadResult{
			ID:        dl.ID,
			URL:       dl.URL,
			Status:    dl.Status,
			Progress:  dl.Progress,
			YoutubeID: extractYoutubeID(dl.URL),
		}
		
		if dl.Title != nil {
			result.Title = *dl.Title
		}
		if dl.Channel != nil {
			result.Channel = *dl.Channel
		}
		if dl.ThumbnailURL != nil {
			result.ThumbnailURL = *dl.ThumbnailURL
		}
		if dl.FormatID != nil {
			result.FormatID = *dl.FormatID
		}
		if dl.Quality != nil {
			result.Quality = *dl.Quality
		}
		if dl.ErrorMessage != nil {
			result.ErrorMessage = *dl.ErrorMessage
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// StartProcessingDownloads tells the backend to start processing the queue
// This should be called by the frontend after restoring the queue
func (a *App) StartProcessingDownloads() {
	logger := applog.GetLogger()
	logger.Info("Download", "StartProcessingDownloads called by frontend")
	
	if a.db == nil {
		logger.Error("Download", "Cannot start processing - db is nil", fmt.Errorf("database not initialized"))
		return
	}
	
	// Count pending downloads
	pending, err := a.db.GetPendingDownloads(100)
	if err != nil {
		logger.Error("Download", "Failed to count pending downloads", err)
	} else {
		logger.Info("Download", "Found pending downloads", map[string]int{
			"count": len(pending),
		})
	}
	
	go a.processDownloads()
}

// RestoreDownloadQueue is deprecated - use GetDownloadQueue + StartProcessingDownloads
func (a *App) RestoreDownloadQueue() error {
	// This is now handled by GetDownloadQueue which is called by frontend
	return nil
}

// ClearDownloadCache removes all download records from the database
func (a *App) ClearDownloadCache() error {
	logger := applog.GetLogger()
	
	if a.db == nil {
		return nil
	}
	
	if err := a.db.ClearAllDownloads(); err != nil {
		logger.Error("Download", "Failed to clear download cache", err)
		return err
	}
	
	logger.Info("Download", "Download cache cleared")
	return nil
}

// ClearCompletedDownloadsCache removes only completed download records
func (a *App) ClearCompletedDownloadsCache() error {
	logger := applog.GetLogger()
	
	if a.db == nil {
		return nil
	}
	
	if err := a.db.ClearCompletedDownloads(); err != nil {
		logger.Error("Download", "Failed to clear completed downloads cache", err)
		return err
	}
	
	logger.Info("Download", "Completed downloads cache cleared")
	return nil
}
