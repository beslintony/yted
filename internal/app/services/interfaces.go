package services

import (
	"context"

	"yted/internal/config"
	"yted/internal/ytdl"
)

// VideoResult represents a video in the library
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

// ListVideosOptions for filtering and sorting videos
type ListVideosOptions struct {
	Search   string `json:"search"`
	Channel  string `json:"channel"`
	SortBy   string `json:"sort_by"`
	SortDesc bool   `json:"sort_desc"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// VideoInfoResult contains information about a video
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

// DownloadResult represents a download in the queue
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

// CacheInfo provides information about cache status
type CacheInfo struct {
	DownloadCount      int   `json:"download_count"`
	CompletedCount     int   `json:"completed_count"`
	PendingCount       int   `json:"pending_count"`
	VideoCount         int   `json:"video_count"`
	TotalLibrarySize   int64 `json:"total_library_size"`
	OrphanedFilesCount int   `json:"orphaned_files_count"`
	OrphanedFilesSize  int64 `json:"orphaned_files_size"`
}

// LogEntry represents a log entry for the frontend
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
}

// DownloadService handles download operations
type DownloadService interface {
	AddDownload(ctx context.Context, videoURL, formatID, quality string) (string, error)
	PauseDownload(ctx context.Context, id string) error
	ResumeDownload(ctx context.Context, id string) error
	RetryDownload(ctx context.Context, id string) error
	CancelDownload(ctx context.Context, id string) error
	GetDownloads(ctx context.Context) ([]DownloadResult, error)
	GetDownloadsByStatus(ctx context.Context, status string) ([]DownloadResult, error)
	GetDownloadQueue(ctx context.Context) ([]DownloadResult, error)
	StartProcessingDownloads(ctx context.Context)
	GetVideoInfo(ctx context.Context, videoURL string) (*VideoInfoResult, error)
	ValidateURL(videoURL string) bool
	ClearCompletedDownloads(ctx context.Context) error
	ClearDownloadCache(ctx context.Context) error
	ClearCompletedDownloadsCache(ctx context.Context) error
	VerifyAndRepairDownloads(ctx context.Context) error
}

// VideoService handles video library operations
type VideoService interface {
	ListVideos(ctx context.Context, opts ListVideosOptions) ([]VideoResult, error)
	GetVideo(ctx context.Context, id string) (*VideoResult, error)
	GetVideoByYoutubeID(ctx context.Context, youtubeID string) (*VideoResult, error)
	DeleteVideo(ctx context.Context, id string, deleteFile bool) error
	DeleteVideoWithConfirmation(ctx context.Context, id string) (map[string]interface{}, error)
	UpdateWatchPosition(ctx context.Context, id string, position int) error
	GetChannels(ctx context.Context) ([]string, error)
	GetLibraryStats(ctx context.Context) (map[string]interface{}, error)
	CleanUpDuplicateVideos(ctx context.Context) error
	SyncDownloadWithFile(ctx context.Context, id string) error
}

// SettingsService handles configuration
type SettingsService interface {
	GetSettings(ctx context.Context) (*config.Config, error)
	SaveSettings(ctx context.Context, settings *config.Config) error
	UpdateSetting(ctx context.Context, key string, value interface{}) error
	GetDownloadPresets(ctx context.Context) ([]config.DownloadPreset, error)
	AddDownloadPreset(ctx context.Context, preset config.DownloadPreset) error
	RemoveDownloadPreset(ctx context.Context, id string) error
	UpdateDownloadPreset(ctx context.Context, id string, preset config.DownloadPreset) error
}

// CacheService handles cache operations
type CacheService interface {
	GetCacheInfo(ctx context.Context) (*CacheInfo, error)
	CleanupOrphanedFiles(ctx context.Context, deleteFiles bool) (map[string]interface{}, error)
	ClearTempFiles(ctx context.Context) (map[string]interface{}, error)
	RepairLibrary(ctx context.Context) (map[string]interface{}, error)
}

// LoggerService handles logging operations
type LoggerService interface {
	GetLogs(count int) []LogEntry
	GetAllLogs() []LogEntry
	ClearLogs()
	ExportLogs(ctx context.Context, customPath string) error
	GetLogExportPath(ctx context.Context) string
	SetLogExportPath(ctx context.Context, path string) error
}

// YtdlpService handles yt-dlp operations
type YtdlpService interface {
	GetYtdlpVersion(ctx context.Context) string
	UpdateYtdlp(ctx context.Context) (bool, error)
	CheckForUpdate(ctx context.Context)
}

// EventEmitter interface for emitting events to frontend
type EventEmitter interface {
	Emit(eventName string, data ...interface{})
}
