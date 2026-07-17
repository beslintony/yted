package app

import (
	"context"
	"fmt"
	"time"

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

// AddPlaylistDownload queues every video of a playlist as an individual
// download and returns how many were added. Duplicates already in the
// queue are skipped by AddDownload's existing check
func (a *App) AddPlaylistDownload(videoURL string, formatID string, quality string) (int, error) {
	logger := applog.GetLogger()

	if a.ytdl == nil {
		return 0, fmt.Errorf("ytdl client not initialized")
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

	logger.Info("Download", "Adding playlist download", map[string]interface{}{
		"url":      videoURL,
		"title":    info.Title,
		"entries":  len(info.Entries),
		"formatID": formatID,
		"quality":  quality,
	})

	added := 0
	for _, entry := range info.Entries {
		watchURL := "https://www.youtube.com/watch?v=" + entry.ID
		if _, err := a.AddDownload(watchURL, formatID, quality); err != nil {
			logger.Warn("Download", "Failed to add playlist entry", map[string]string{
				"entry": entry.ID,
				"error": err.Error(),
			})
			continue
		}
		added++
	}

	if added == 0 {
		return 0, fmt.Errorf("no videos could be added from this playlist")
	}

	logger.Info("Download", "Playlist download added", map[string]interface{}{
		"added":   added,
		"skipped": len(info.Entries) - added,
	})

	return added, nil
}
