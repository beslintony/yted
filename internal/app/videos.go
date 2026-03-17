package app

import (
	"fmt"
	"os"

	"yted/internal/db"
	applog "yted/internal/log"
)

// VideoResult is exposed to frontend
type VideoResult struct {
	ID            string `json:"id"`
	YoutubeID     string `json:"youtube_id"`
	Title         string `json:"title"`
	Channel       string `json:"channel"`
	ChannelID     string `json:"channel_id"`
	Duration      int    `json:"duration"`
	Description   string `json:"description"`
	ThumbnailURL  string `json:"thumbnail_url"`
	FilePath      string `json:"file_path"`
	FileSize      int64  `json:"file_size"`
	Format        string `json:"format"`
	Quality       string `json:"quality"`
	DownloadedAt  int64  `json:"downloaded_at"`
	WatchPosition int    `json:"watch_position"`
	WatchCount    int    `json:"watch_count"`
}

// ListVideosOptions is exposed to frontend
type ListVideosOptions struct {
	Search   string `json:"search"`
	Channel  string `json:"channel"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// ListVideos returns videos from the library with file existence check
func (a *App) ListVideos(opts ListVideosOptions) ([]VideoResult, error) {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil, nil
	}

	dbOpts := db.ListVideosOptions{
		Search:   opts.Search,
		Channel:  opts.Channel,
		SortBy:   opts.SortBy,
		SortDesc: opts.SortDesc,
		Limit:    opts.Limit,
		Offset:   opts.Offset,
	}

	videos, err := a.db.ListVideosWithHash(dbOpts)
	if err != nil {
		return nil, err
	}

	results := make([]VideoResult, 0, len(videos))
	for _, v := range videos {
		// Check if file still exists for managed files
		if v.IsManaged && v.FilePath != "" {
			if _, err := os.Stat(v.FilePath); os.IsNotExist(err) {
				logger.Warn("Library", "Managed file missing, skipping", map[string]string{
					"video_id": v.ID,
					"path":     v.FilePath,
				})
				continue
			}
		}
		results = append(results, videoToResult(v))
	}

	return results, nil
}

// GetVideo returns a single video with file existence check
func (a *App) GetVideo(id string) (*VideoResult, error) {
	if a.db == nil {
		return nil, nil
	}

	video, err := a.db.GetVideoWithHash(id)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, nil
	}

	// Check if managed file exists
	if video.IsManaged && video.FilePath != "" {
		if _, err := os.Stat(video.FilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("video file not found: %s", video.FilePath)
		}
	}

	result := videoToResult(*video)
	return &result, nil
}

// GetVideoByYoutubeID returns a video by YouTube ID
func (a *App) GetVideoByYoutubeID(youtubeID string) (*VideoResult, error) {
	if a.db == nil {
		return nil, nil
	}

	video, err := a.db.GetVideoByFileHash(youtubeID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, nil
	}

	result := videoToResult(*video)
	return &result, nil
}

// DeleteVideo removes a video from the library and optionally deletes the file
func (a *App) DeleteVideo(id string, deleteFile bool) error {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil
	}

	// Get video info before deleting
	filePath, isManaged, err := a.db.DeleteVideoAndFile(id)
	if err != nil {
		return err
	}

	logger.Info("Library", "Video deleted from database", map[string]string{
		"id":   id,
		"file": filePath,
	})

	// Delete file only if it's managed and user confirmed
	if deleteFile && isManaged && filePath != "" {
		if err := os.Remove(filePath); err != nil {
			logger.Error("Library", "Failed to delete video file", err, map[string]string{
				"path": filePath,
			})
			return fmt.Errorf("failed to delete file: %w", err)
		}
		logger.Info("Library", "Video file deleted", map[string]string{
			"path": filePath,
		})
	}

	return nil
}

// DeleteVideoWithConfirmation removes a video and returns info for frontend confirmation
func (a *App) DeleteVideoWithConfirmation(id string) (map[string]interface{}, error) {
	logger := applog.GetLogger()

	if a.db == nil {
		return nil, nil
	}

	video, err := a.db.GetVideoWithHash(id)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return map[string]interface{}{
			"found": false,
		}, nil
	}

	// Check if file exists
	fileExists := false
	if video.FilePath != "" {
		_, err := os.Stat(video.FilePath)
		fileExists = !os.IsNotExist(err)
	}

	result := map[string]interface{}{
		"found":                true,
		"title":                video.Title,
		"isManaged":            video.IsManaged,
		"filePath":             video.FilePath,
		"fileExists":           fileExists,
		"fileSize":             video.FileSize,
		"requiresConfirmation": video.IsManaged && fileExists,
	}

	logger.Info("Library", "Video deletion info", result)

	return result, nil
}

// UpdateWatchPosition updates the watch position for a video
func (a *App) UpdateWatchPosition(id string, position int) error {
	if a.db == nil {
		return nil
	}

	return a.db.UpdateWatchPosition(id, position)
}

// GetChannels returns all unique channels
func (a *App) GetChannels() ([]string, error) {
	if a.db == nil {
		return nil, nil
	}

	return a.db.GetChannels()
}

// GetLibraryStats returns library statistics
func (a *App) GetLibraryStats() (map[string]interface{}, error) {
	if a.db == nil {
		return map[string]interface{}{
			"total_videos": 0,
			"total_size":   0,
		}, nil
	}

	totalVideos, totalSize, err := a.db.GetStats()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_videos": totalVideos,
		"total_size":   totalSize,
	}, nil
}

// videoToResult converts a db.Video to VideoResult
func videoToResult(v db.Video) VideoResult {
	return VideoResult{
		ID:            v.ID,
		YoutubeID:     v.YoutubeID,
		Title:         v.Title,
		Channel:       v.Channel,
		ChannelID:     v.ChannelID,
		Duration:      v.Duration,
		Description:   v.Description,
		ThumbnailURL:  v.ThumbnailURL,
		FilePath:      v.FilePath,
		FileSize:      v.FileSize,
		Format:        v.Format,
		Quality:       v.Quality,
		DownloadedAt:  v.DownloadedAt.Unix(),
		WatchPosition: v.WatchPosition,
		WatchCount:    v.WatchCount,
	}
}
