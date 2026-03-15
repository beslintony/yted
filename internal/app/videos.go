package app

import (
	"yted/internal/db"
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

// ListVideos returns videos from the library
func (a *App) ListVideos(opts ListVideosOptions) ([]VideoResult, error) {
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

	videos, err := a.db.ListVideos(dbOpts)
	if err != nil {
		return nil, err
	}

	results := make([]VideoResult, len(videos))
	for i, v := range videos {
		results[i] = videoToResult(v)
	}

	return results, nil
}

// GetVideo returns a single video
func (a *App) GetVideo(id string) (*VideoResult, error) {
	if a.db == nil {
		return nil, nil
	}

	video, err := a.db.GetVideo(id)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, nil
	}

	result := videoToResult(*video)
	return &result, nil
}

// GetVideoByYoutubeID returns a video by YouTube ID
func (a *App) GetVideoByYoutubeID(youtubeID string) (*VideoResult, error) {
	if a.db == nil {
		return nil, nil
	}

	video, err := a.db.GetVideoByYoutubeID(youtubeID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, nil
	}

	result := videoToResult(*video)
	return &result, nil
}

// DeleteVideo removes a video from the library
func (a *App) DeleteVideo(id string) error {
	if a.db == nil {
		return nil
	}

	return a.db.DeleteVideo(id)
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
