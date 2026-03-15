package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	applog "yted/internal/log"
	"yted/internal/db"
	"yted/internal/ytdl"
)

// VideoInfoResult is exposed to frontend
type VideoInfoResult struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Channel     string           `json:"channel"`
	ChannelID   string           `json:"channel_id"`
	Duration    int              `json:"duration"`
	Description string           `json:"description"`
	Thumbnail   string           `json:"thumbnail"`
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
			"active":   activeCount,
			"max":      maxConcurrent,
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
		"count":         len(pending),
		"slotsAvailable": slotsAvailable,
	})

	// Start each download
	for _, dl := range pending {
		go a.startDownload(dl)
	}
}

// startDownload starts a single download
func (a *App) startDownload(dl db.Download) {
	logger := applog.GetLogger()
	ctx := context.Background()

	logger.Info("Download", "Starting download", map[string]string{
		"id":  dl.ID,
		"url": dl.URL,
	})

	// Mark as started
	if err := a.db.StartDownload(dl.ID); err != nil {
		logger.Error("Download", "Failed to mark download as started", err, map[string]string{"id": dl.ID})
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
	err := a.ytdl.Download(ctx, dl.URL, opts, progressCallback)
	if err != nil {
		logger.Error("Download", "Download failed", err, map[string]string{"id": dl.ID})
		a.db.FailDownload(dl.ID, err.Error())
		runtime.EventsEmit(a.ctx, "download:error", map[string]interface{}{
			"id":      dl.ID,
			"error":   err.Error(),
		})
		return
	}

	// Mark as completed
	if err := a.db.CompleteDownload(dl.ID); err != nil {
		logger.Error("Download", "Failed to mark download as completed", err, map[string]string{"id": dl.ID})
	}

	logger.Info("Download", "Download completed successfully", map[string]string{"id": dl.ID})
	runtime.EventsEmit(a.ctx, "download:completed", dl.ID)

	// Process more downloads
	a.processDownloads()
}

// ValidateURL checks if a URL is valid
func (a *App) ValidateURL(url string) bool {
	return ytdl.IsValidURL(url)
}
