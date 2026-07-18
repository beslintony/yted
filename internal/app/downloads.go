package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"yted/internal/db"
	applog "yted/internal/log"
	"yted/internal/ytdl"
)

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

// GetVideoInfo extracts video information from URL with caching
// Cache TTL: 5 minutes to prevent duplicate fetches
func (a *App) GetVideoInfo(videoURL string) (*VideoInfoResult, error) {
	logger := applog.GetLogger()

	if a.ytdl == nil {
		err := fmt.Errorf("ytdl client not initialized")
		logger.Error("Download", "ytdl client not ready", err)
		return nil, err
	}

	if !ytdl.IsValidURL(videoURL) {
		err := fmt.Errorf("invalid URL")
		logger.Warn("Download", "Invalid URL provided", map[string]string{"url": videoURL})
		return nil, err
	}

	// Normalize URL for cache key
	cacheKey := normalizeURLForCache(videoURL)

	// Check cache first
	a.videoInfoCacheMu.RLock()
	cached, found := a.videoInfoCache[cacheKey]
	a.videoInfoCacheMu.RUnlock()

	if found && time.Now().Before(cached.expiresAt) {
		logger.Info("Download", "Video info cache hit", map[string]string{
			"url":   videoURL,
			"title": cached.info.Title,
		})
		return cached.info, nil
	}

	// Clean URL before fetching (strip playlist params, etc.)
	cleanURL := cleanYouTubeURL(videoURL)
	if cleanURL != videoURL {
		logger.Debug("Download", "URL cleaned for video info fetch", map[string]string{
			"original": videoURL,
			"cleaned":  cleanURL,
		})
	}

	logger.Info("Download", "Fetching video info", map[string]string{"url": cleanURL})

	ctx, cancel := context.WithTimeout(a.ctx, 2*time.Minute)
	defer cancel()

	info, err := a.ytdl.GetInfo(ctx, cleanURL)
	if err != nil {
		logger.Error("Download", "Failed to get video info", err, map[string]string{"url": videoURL})
		return nil, err
	}

	result := &VideoInfoResult{
		ID:          info.ID,
		Title:       info.Title,
		Channel:     info.Channel,
		ChannelID:   info.ChannelID,
		Duration:    info.Duration,
		Description: info.Description,
		Thumbnail:   info.Thumbnail,
		Formats:     info.Formats,
	}

	// Store in cache with 5-minute TTL
	a.videoInfoCacheMu.Lock()
	a.videoInfoCache[cacheKey] = videoInfoCacheEntry{
		info:      result,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	a.videoInfoCacheMu.Unlock()

	logger.Info("Download", "Video info fetched and cached", map[string]string{
		"id":    info.ID,
		"title": info.Title,
	})

	return result, nil
}

// AddDownload adds a new download to the queue
func (a *App) AddDownload(videoURL string, formatID string, quality string) (string, error) {
	logger := applog.GetLogger()

	logger.Info("Download", "Adding download", map[string]string{
		"url":      videoURL,
		"formatID": formatID,
		"quality":  quality,
	})

	if a.db == nil {
		logger.Error("Download", "Database not ready", ErrDatabaseNotInitialized)
		return "", ErrDatabaseNotInitialized
	}

	if a.ytdl == nil {
		logger.Error("Download", "ytdl client not ready", ErrYtdlNotInitialized)
		return "", ErrYtdlNotInitialized
	}

	// Check for existing active download to prevent duplicates
	// Use a timeout to prevent hanging on database lock
	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()

	existingDownload, err := a.db.GetActiveDownloadByURL(videoURL)
	if err != nil {
		logger.Error("Download", "Failed to check for existing download", err)
		return "", err
	}
	if existingDownload != nil {
		logger.Info("Download", "Download already in queue, returning existing ID", map[string]string{
			"url": videoURL,
			"id":  existingDownload.ID,
		})
		return existingDownload.ID, nil
	}

	// Verify context wasn't cancelled
	if ctx.Err() != nil {
		return "", fmt.Errorf("operation cancelled or timed out")
	}

	download := &db.Download{
		ID:       uuid.New().String(),
		URL:      videoURL,
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
		return nil, fmt.Errorf("database not initialized")
	}
	return a.db.ListDownloads(status)
}

// PauseDownload pauses a download by cancelling its context
func (a *App) PauseDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Cancel the active download if it's running
	a.activeDownloadsMu.RLock()
	cancel, exists := a.activeDownloads[id]
	a.activeDownloadsMu.RUnlock()

	if exists && cancel != nil {
		cancel()
		logger.Info("Download", "Download cancelled for pause", map[string]string{"id": id})
	}

	if err := a.db.UpdateDownloadStatus(id, "paused"); err != nil {
		logger.Error("Download", "Failed to pause download", err, map[string]string{"id": id})
		return err
	}

	logger.Info("Download", "Download paused", map[string]string{"id": id})
	runtime.EventsEmit(a.ctx, "download:paused", id)

	// Fill the slot freed by pausing
	go a.processDownloads()

	return nil
}

// ResumeDownload resumes a paused download
func (a *App) ResumeDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return fmt.Errorf("database not initialized")
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
// Note: Progress is NOT reset to allow yt-dlp's --continue to resume partial downloads
func (a *App) RetryDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Clear error message but preserve progress for resume
	if err := a.db.ClearDownloadError(id); err != nil {
		logger.Error("Download", "Failed to clear download error", err, map[string]string{"id": id})
		return err
	}

	if err := a.db.UpdateDownloadStatus(id, "pending"); err != nil {
		logger.Error("Download", "Failed to retry download", err, map[string]string{"id": id})
		return err
	}

	// Note: We intentionally do NOT reset progress to 0 here
	// yt-dlp's --continue flag will resume from the partial file
	// and the progress will update when download starts again
	logger.Info("Download", "Download retry initiated (preserving progress for resume)", map[string]string{
		"id": id,
	})
	runtime.EventsEmit(a.ctx, "download:retried", id)

	// Try to process downloads
	go a.processDownloads()

	return nil
}

// CancelDownload cancels and removes a download
// CancelDownload removes a download from the queue, stopping its worker
// if one is running
func (a *App) CancelDownload(id string) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Stop the worker first if it's running
	a.activeDownloadsMu.Lock()
	if cancel, ok := a.activeDownloads[id]; ok {
		cancel()
		delete(a.activeDownloads, id)
	}
	a.activeDownloadsMu.Unlock()

	if err := a.db.DeleteDownload(id); err != nil {
		logger.Error("Download", "Failed to cancel download", err, map[string]string{"id": id})
		return err
	}

	logger.Info("Download", "Download cancelled", map[string]string{"id": id})
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "download:cancelled", id)
	}
	return nil
}

// ClearCompletedDownloads removes all completed downloads
func (a *App) ClearCompletedDownloads() error {
	logger := applog.GetLogger()

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := a.db.DeleteCompletedDownloads(); err != nil {
		logger.Error("Download", "Failed to clear completed downloads", err)
		return err
	}

	logger.Info("Download", "Completed downloads cleared")
	runtime.EventsEmit(a.ctx, "downloads:cleared", nil)
	return nil
}

// ValidateURL checks if a URL is valid
func (a *App) ValidateURL(videoURL string) bool {
	return ytdl.IsValidURL(videoURL)
}

// GetIncompleteDownloads returns all downloads that are not completed (for restoring queue)
func (a *App) GetIncompleteDownloads() ([]db.Download, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
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
		return nil, fmt.Errorf("database not initialized")
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

	// Reset downloads that were 'downloading' to 'pending' ONLY if no active worker exists
	// This prevents spawning duplicate workers if the app was restarted
	for _, dl := range downloads {
		if dl.Status == "downloading" {
			a.activeDownloadsMu.RLock()
			_, hasActiveWorker := a.activeDownloads[dl.ID]
			a.activeDownloadsMu.RUnlock()

			if hasActiveWorker {
				logger.Info("Download", "Download has active worker, keeping as downloading", map[string]string{
					"id": dl.ID,
				})
			} else {
				logger.Info("Download", "Resetting stuck download to pending (no active worker)", map[string]string{
					"id": dl.ID,
				})
				if err := a.db.UpdateDownloadStatus(dl.ID, "pending"); err != nil {
					logger.Error("Download", "Failed to reset download status", err, map[string]string{
						"id": dl.ID,
					})
				} else {
					dl.Status = "pending"
				}
			}
		}
	}

	// Convert to DownloadResult
	results := make([]DownloadResult, 0, len(downloads))
	for _, dl := range downloads {
		result := convertDownloadToResult(dl)
		results = append(results, result)
	}

	return results, nil
}

// convertDownloadToResult converts a db.Download to DownloadResult
func convertDownloadToResult(dl db.Download) DownloadResult {
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

	return result
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

	go a.processDownloads()
}

// RestoreDownloadQueue is deprecated - use GetDownloadQueue + StartProcessingDownloads
func (a *App) RestoreDownloadQueue() error {
	// This is now handled by GetDownloadQueue which is called by frontend
	return nil
}

// ClearDownloadCache removes all download records from the database and
// stops any running workers
func (a *App) ClearDownloadCache() error {
	logger := applog.GetLogger()

	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Stop all running workers first
	a.activeDownloadsMu.Lock()
	for id, cancel := range a.activeDownloads {
		cancel()
		delete(a.activeDownloads, id)
	}
	a.activeDownloadsMu.Unlock()

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
		return fmt.Errorf("database not initialized")
	}

	if err := a.db.ClearCompletedDownloads(); err != nil {
		logger.Error("Download", "Failed to clear completed downloads cache", err)
		return err
	}

	logger.Info("Download", "Completed downloads cache cleared")
	return nil
}
