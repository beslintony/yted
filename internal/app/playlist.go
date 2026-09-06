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

// PlaylistEntryResult is a single playlist video exposed to the frontend
type PlaylistEntryResult struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
}

// PlaylistInfoResult is playlist metadata exposed to the frontend
type PlaylistInfoResult struct {
	ID      string                `json:"id"`
	Title   string                `json:"title"`
	Channel string                `json:"channel"`
	Count   int                   `json:"count"`
	Entries []PlaylistEntryResult `json:"entries"`
}

// IsPlaylistURL reports whether the URL points at a YouTube playlist
func (a *App) IsPlaylistURL(videoURL string) bool {
	return ytdl.IsPlaylistURL(videoURL)
}

// GetPlaylistInfo fetches playlist metadata and entries for a playlist URL
func (a *App) GetPlaylistInfo(videoURL string) (*PlaylistInfoResult, error) {
	logger := applog.GetLogger()

	if a.ytdl == nil {
		err := fmt.Errorf("ytdl client not initialized")
		logger.Error("Download", "ytdl client not ready", err)
		return nil, err
	}

	if !ytdl.IsValidURL(videoURL) {
		return nil, fmt.Errorf("invalid URL")
	}
	if !ytdl.IsPlaylistURL(videoURL) {
		return nil, fmt.Errorf("not a playlist URL")
	}

	logger.Info("Download", "Fetching playlist info", map[string]string{"url": videoURL})

	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Minute)
	defer cancel()

	info, err := a.ytdl.GetPlaylistInfo(ctx, videoURL)
	if err != nil {
		logger.Error("Download", "Failed to fetch playlist info", err)
		return nil, err
	}

	entries := make([]PlaylistEntryResult, 0, len(info.Entries))
	for _, e := range info.Entries {
		entries = append(entries, PlaylistEntryResult{
			ID:       e.ID,
			Title:    e.Title,
			Duration: e.Duration,
		})
	}

	logger.Info("Download", "Playlist info fetched", map[string]interface{}{
		"title":   info.Title,
		"entries": len(entries),
	})

	return &PlaylistInfoResult{
		ID:      info.ID,
		Title:   info.Title,
		Channel: info.Channel,
		Count:   info.Count,
		Entries: entries,
	}, nil
}

// AddPlaylistDownload queues the videos of a playlist as individual
// downloads and returns how many were added. maxItems limits how many
// entries are queued (from the start of the playlist); maxItems <= 0
// queues the whole playlist. Entries carry their titles, duplicates
// already in the queue are skipped, and the scheduler is started once
// for the whole batch
func (a *App) AddPlaylistDownload(videoURL string, formatID string, quality string, maxItems int) (int, error) {
	logger := applog.GetLogger()

	if a.ytdl == nil {
		return 0, fmt.Errorf("ytdl client not initialized")
	}
	if a.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	if !ytdl.IsPlaylistURL(videoURL) {
		return 0, fmt.Errorf("not a playlist URL")
	}

	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Minute)
	defer cancel()

	info, err := a.ytdl.GetPlaylistInfo(ctx, videoURL)
	if err != nil {
		logger.Error("Download", "Failed to fetch playlist info for download", err)
		return 0, err
	}
	if len(info.Entries) == 0 {
		return 0, fmt.Errorf("playlist has no videos")
	}

	entries := info.Entries
	if maxItems > 0 && len(entries) > maxItems {
		logger.Info("Download", "Playlist truncated to first entries", map[string]interface{}{
			"total": len(entries),
			"cap":   maxItems,
		})
		entries = entries[:maxItems]
	}

	logger.Info("Download", "Adding playlist download", map[string]interface{}{
		"url":      videoURL,
		"title":    info.Title,
		"entries":  len(entries),
		"formatID": formatID,
		"quality":  quality,
	})

	added := a.queuePlaylistEntries(entries, formatID, quality)
	if added == 0 {
		return 0, fmt.Errorf("no videos could be added from this playlist")
	}

	// Start the scheduler once for the whole batch
	go a.processDownloads()

	logger.Info("Download", "Playlist download added", map[string]int{"added": added})
	return added, nil
}

// queuePlaylistEntries creates pending download records for playlist entries
// (with titles), skipping entries already in the queue, and emits a
// download:added event per new record
func (a *App) queuePlaylistEntries(entries []ytdl.PlaylistEntry, formatID, quality string) int {
	logger := applog.GetLogger()

	added := 0

	// Single pre-fetch of already-queued URLs to avoid an N+1 query pattern
	// (one GetActiveDownloadByURL per entry). Falls back to per-item checks
	// if the batch lookup fails so duplicates are still skipped.
	watchURLs := make([]string, 0, len(entries))
	for _, entry := range entries {
		watchURLs = append(watchURLs, "https://www.youtube.com/watch?v="+entry.ID)
	}
	existing, err := a.db.GetActiveDownloadURLs(watchURLs)
	if err != nil {
		logger.Error("Download", "Failed to pre-fetch existing downloads", err)
		existing = nil
	}
	// seen tracks URLs already queued (or added in this batch) so duplicates
	// within the same playlist batch are skipped like before
	seen := make(map[string]bool, len(entries))
	for url := range existing {
		seen[url] = true
	}

	pending := make([]*db.Download, 0, len(entries))
	for i, entry := range entries {
		watchURL := watchURLs[i]

		if seen[watchURL] {
			continue // already queued
		}
		if existing == nil {
			// Pre-fetch failed: fall back to the per-item check
			existingItem, err := a.db.GetActiveDownloadByURL(watchURL)
			if err != nil {
				logger.Error("Download", "Failed to check for existing download", err)
				continue
			}
			if existingItem != nil {
				seen[watchURL] = true
				continue // already queued
			}
		}

		title := entry.Title
		download := &db.Download{
			ID:       uuid.New().String(),
			URL:      watchURL,
			Status:   "pending",
			Progress: 0,
			Title:    &title,
			FormatID: &formatID,
			Quality:  &quality,
		}
		pending = append(pending, download)
		seen[watchURL] = true
	}

	// Insert the batch in a single transaction; on failure fall back to
	// per-item inserts to preserve the previous best-effort behavior
	created := pending
	if err := a.db.CreateDownloads(pending); err != nil {
		logger.Error("Download", "Failed to batch-create playlist download records", err)
		created = created[:0]
		for _, download := range pending {
			if err := a.db.CreateDownload(download); err != nil {
				logger.Error("Download", "Failed to create playlist download record", err)
				continue
			}
			created = append(created, download)
		}
	}
	for _, download := range created {
		added++
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "download:added", download)
		}
	}

	return added
}
