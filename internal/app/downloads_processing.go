package app

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"yted/internal/db"
	applog "yted/internal/log"
	"yted/internal/ytdl"
)

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
		// Skip if this download already has an active worker
		a.activeDownloadsMu.RLock()
		_, hasActiveWorker := a.activeDownloads[dl.ID]
		a.activeDownloadsMu.RUnlock()
		if hasActiveWorker {
			logger.Info("Download", "Download already has active worker, skipping", map[string]string{
				"id": dl.ID,
			})
			continue
		}

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

	// Register with WaitGroup for graceful shutdown
	a.activeDownloadsWg.Add(1)
	defer a.activeDownloadsWg.Done()

	// Check if shutdown is in progress
	select {
	case <-a.shutdownCh:
		logger.Info("Download", "Download cancelled due to shutdown", map[string]string{"id": dl.ID})
		return
	default:
	}

	// Create a context with timeout for the download
	// Store cancel function for pause/resume support
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Minute)
	a.activeDownloadsMu.Lock()
	a.activeDownloads[dl.ID] = cancel
	a.activeDownloadsMu.Unlock()
	defer func() {
		cancel()
		a.activeDownloadsMu.Lock()
		delete(a.activeDownloads, dl.ID)
		a.activeDownloadsMu.Unlock()
	}()

	logger.Info("Download", "Download worker starting", map[string]string{
		"id":  dl.ID,
		"url": dl.URL,
	})

	// Check if app context is still valid
	if a.ctx == nil {
		logger.Error("Download", "App context is nil", fmt.Errorf("app not initialized"))
		return
	}

	// Verify download is still in 'downloading' state (could have been cancelled/paused)
	currentDownload, err := a.db.GetDownload(dl.ID)
	if err != nil {
		logger.Error("Download", "Failed to verify download status", err, map[string]string{"id": dl.ID})
		return
	}
	if currentDownload == nil {
		logger.Info("Download", "Download no longer exists, aborting", map[string]string{"id": dl.ID})
		return
	}
	if currentDownload.Status != "downloading" {
		logger.Info("Download", "Download no longer active, aborting", map[string]string{
			"id":     dl.ID,
			"status": currentDownload.Status,
		})
		return
	}

	// Get video info if not already have title
	title, channel, thumbnail := extractDownloadInfo(&dl)

	if title == "" {
		info, err := a.ytdl.GetInfo(ctx, cleanYouTubeURL(dl.URL))
		if err == nil {
			title = info.Title
			channel = info.Channel
			thumbnail = info.Thumbnail

			// Update download with info
			dl.Title = &title
			dl.Channel = &channel
			dl.ThumbnailURL = &thumbnail
			dl.Duration = &info.Duration
			if updateErr := a.db.UpdateDownload(&dl); updateErr != nil {
				logger.Warn("Download", "Failed to update download info", map[string]string{"error": updateErr.Error()})
			}

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

	// FFmpeg is required for every output path (merging, remuxing, MP3
	// extraction). Without it yt-dlp leaves unmerged fragment files which
	// would get registered in the library as if they were the video.
	if a.ffmpeg == nil || !a.ffmpeg.IsAvailable() {
		err := fmt.Errorf("ffmpeg is not available - install ffmpeg to download videos")
		logger.Error("Download", "Download failed - ffmpeg unavailable", err, map[string]string{"id": dl.ID})
		if failErr := a.db.FailDownload(dl.ID, err.Error()); failErr != nil {
			logger.Error("Download", "Failed to mark download as failed", failErr, map[string]string{"id": dl.ID})
		}
		runtime.EventsEmit(a.ctx, "download:error", map[string]interface{}{
			"id":    dl.ID,
			"error": err.Error(),
		})
		return
	}

	// Download options
	format := ""
	if dl.FormatID != nil {
		format = *dl.FormatID
	}

	quality := ""
	if dl.Quality != nil {
		quality = *dl.Quality
	}

	opts := ytdl.DownloadOptions{
		Format:    format,
		Quality:   quality,
		OutputDir: a.config.Get().DownloadPath,
		ProxyURL:  a.config.Get().ProxyURL,
	}

	// Progress callback with debouncing to prevent frontend flooding
	// Throttle UI updates to max 2 per second (every 500ms)
	var lastEmitTime time.Time
	var lastProgress float64
	var lastLogTime time.Time
	progressCallback := func(progress ytdl.DownloadProgress) {
		// Always update database
		if updateErr := a.db.UpdateDownloadProgress(dl.ID, progress.Percent); updateErr != nil {
			logger.Debug("Download", "Failed to update progress", map[string]string{"error": updateErr.Error()})
		}

		// Throttle UI events: emit if:
		// 1. First event (lastEmitTime.IsZero)
		// 2. 500ms has passed since last emit
		// 3. Progress changed by more than 5%
		// 4. Download completed (100%)
		shouldEmit := lastEmitTime.IsZero() ||
			time.Since(lastEmitTime) > 500*time.Millisecond ||
			progress.Percent-lastProgress > 5 ||
			progress.Percent >= 100

		if shouldEmit {
			lastEmitTime = time.Now()
			lastProgress = progress.Percent
			runtime.EventsEmit(a.ctx, "download:progress", map[string]interface{}{
				"id":       dl.ID,
				"progress": progress.Percent,
				"speed":    progress.Speed,
				"eta":      progress.ETA,
			})
		}

		// Throttle debug logging to max 1 per second to reduce log spam
		if time.Since(lastLogTime) > time.Second {
			lastLogTime = time.Now()
			logger.Debug("Download", "Progress update", map[string]interface{}{
				"id":       dl.ID,
				"progress": progress.Percent,
			})
		}
	}

	// Perform download
	err = a.ytdl.Download(ctx, dl.URL, opts, progressCallback)
	if err != nil {
		// Check if this was a pause (context cancelled) vs a real error
		if ctx.Err() == context.Canceled {
			// Check current status - if paused, this was intentional
			currentDl, _ := a.db.GetDownload(dl.ID)
			if currentDl != nil && currentDl.Status == "paused" {
				logger.Info("Download", "Download stopped for pause", map[string]string{"id": dl.ID})
				return
			}
		}

		logger.Error("Download", "Download failed", err, map[string]string{"id": dl.ID})
		if failErr := a.db.FailDownload(dl.ID, err.Error()); failErr != nil {
			logger.Error("Download", "Failed to mark download as failed", failErr, map[string]string{"id": dl.ID})
		}
		runtime.EventsEmit(a.ctx, "download:error", map[string]interface{}{
			"id":    dl.ID,
			"error": err.Error(),
		})

		// Free the slot for the next download - without this a wave of
		// failures stalls the queue with everything stuck in pending
		go a.processDownloads()
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
	go a.processDownloads()
}

// extractDownloadInfo extracts info from a download record
func extractDownloadInfo(dl *db.Download) (title, channel, thumbnail string) {
	if dl.Title != nil {
		title = *dl.Title
	}
	if dl.Channel != nil {
		channel = *dl.Channel
	}
	if dl.ThumbnailURL != nil {
		thumbnail = *dl.ThumbnailURL
	}
	return
}
