package app

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	if a.ytdl == nil {
		return nil, fmt.Errorf("ytdl client not initialized")
	}

	if !ytdl.IsValidURL(url) {
		return nil, fmt.Errorf("invalid URL")
	}

	ctx := context.Background()
	info, err := a.ytdl.GetInfo(ctx, url)
	if err != nil {
		return nil, err
	}

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
	if a.db == nil {
		return "", fmt.Errorf("database not initialized")
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
		return "", err
	}

	// Emit event to notify frontend
	runtime.EventsEmit(a.ctx, "download:added", download)

	// Try to start the download if under limit
	go a.processDownloads()

	return download.ID, nil
}

// GetDownloads returns all downloads
func (a *App) GetDownloads() ([]db.Download, error) {
	if a.db == nil {
		return nil, nil
	}
	return a.db.ListDownloads()
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
	if a.db == nil {
		return nil
	}
	
	if err := a.db.UpdateDownloadStatus(id, "paused"); err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "download:paused", id)
	return nil
}

// ResumeDownload resumes a paused download
func (a *App) ResumeDownload(id string) error {
	if a.db == nil {
		return nil
	}
	
	if err := a.db.UpdateDownloadStatus(id, "pending"); err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "download:resumed", id)

	// Try to process downloads
	go a.processDownloads()

	return nil
}

// RetryDownload retries a failed download
func (a *App) RetryDownload(id string) error {
	if a.db == nil {
		return nil
	}
	
	if err := a.db.UpdateDownloadStatus(id, "pending"); err != nil {
		return err
	}

	if err := a.db.UpdateDownloadProgress(id, 0); err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "download:retried", id)

	// Try to process downloads
	go a.processDownloads()

	return nil
}

// CancelDownload cancels and removes a download
func (a *App) CancelDownload(id string) error {
	if a.db == nil {
		return nil
	}
	
	if err := a.db.DeleteDownload(id); err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "download:cancelled", id)
	return nil
}

// ClearCompletedDownloads removes all completed downloads
func (a *App) ClearCompletedDownloads() error {
	if a.db == nil {
		return nil
	}
	
	if err := a.db.DeleteCompletedDownloads(); err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "downloads:cleared", nil)
	return nil
}

// processDownloads starts pending downloads up to the concurrent limit
func (a *App) processDownloads() {
	if a.db == nil || a.config == nil || a.ytdl == nil {
		return
	}

	maxConcurrent := a.config.Get().MaxConcurrentDownloads
	if maxConcurrent < 1 {
		maxConcurrent = 3
	}

	// Count active downloads
	activeCount, err := a.db.CountActiveDownloads()
	if err != nil {
		log.Printf("Error counting active downloads: %v", err)
		return
	}

	// Calculate how many we can start
	slotsAvailable := maxConcurrent - activeCount
	if slotsAvailable <= 0 {
		return
	}

	// Get pending downloads
	pending, err := a.db.GetPendingDownloads(slotsAvailable)
	if err != nil {
		log.Printf("Error getting pending downloads: %v", err)
		return
	}

	// Start each download
	for _, dl := range pending {
		go a.startDownload(dl)
	}
}

// startDownload starts a single download
func (a *App) startDownload(dl db.Download) {
	ctx := context.Background()

	// Mark as started
	if err := a.db.StartDownload(dl.ID); err != nil {
		log.Printf("Error starting download: %v", err)
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
	}

	// Perform download
	err := a.ytdl.Download(ctx, dl.URL, opts, progressCallback)
	if err != nil {
		log.Printf("Download failed: %v", err)
		a.db.FailDownload(dl.ID, err.Error())
		runtime.EventsEmit(a.ctx, "download:error", map[string]interface{}{
			"id":      dl.ID,
			"error":   err.Error(),
		})
		return
	}

	// Mark as completed
	if err := a.db.CompleteDownload(dl.ID); err != nil {
		log.Printf("Error completing download: %v", err)
	}

	runtime.EventsEmit(a.ctx, "download:completed", dl.ID)

	// Process more downloads
	a.processDownloads()
}

// ValidateURL checks if a URL is valid
func (a *App) ValidateURL(url string) bool {
	return ytdl.IsValidURL(url)
}
