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
	// MaxDownload is how many entries AddPlaylistDownload will queue at most
	MaxDownload int `json:"max_download"`
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
		ID:          info.ID,
		Title:       info.Title,
		Channel:     info.Channel,
		Count:       info.Count,
		Entries:     entries,
		MaxDownload: maxPlaylistItems,
	}, nil
}

// maxPlaylistItems caps how many playlist entries are queued at once.
// Auto-generated playlists (YouTube Mix/Radio, RD*) are effectively
// endless - without a cap a single click queues hundreds of downloads
const maxPlaylistItems = 50

// AddPlaylistDownload queues the videos of a playlist as individual
// downloads (at most maxPlaylistItems) and returns how many were added.
// Entries carry their titles, duplicates already in the queue are skipped,
// and the scheduler is started once for the whole batch
func (a *App) AddPlaylistDownload(videoURL string, formatID string, quality string) (int, error) {
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
	if len(entries) > maxPlaylistItems {
		logger.Info("Download", "Playlist truncated to first entries", map[string]interface{}{
			"total": len(entries),
			"cap":   maxPlaylistItems,
		})
		entries = entries[:maxPlaylistItems]
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
	for _, entry := range entries {
		watchURL := "https://www.youtube.com/watch?v=" + entry.ID

		existing, err := a.db.GetActiveDownloadByURL(watchURL)
		if err != nil {
			logger.Error("Download", "Failed to check for existing download", err)
			continue
		}
		if existing != nil {
			continue // already queued
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
		if err := a.db.CreateDownload(download); err != nil {
			logger.Error("Download", "Failed to create playlist download record", err)
			continue
		}
		added++
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "download:added", download)
		}
	}

	return added
}
