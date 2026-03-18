package app

import "errors"

// Common application errors for better error handling and frontend communication
var (
	// Infrastructure errors
	ErrDatabaseNotInitialized = errors.New("database not initialized")
	ErrYtdlNotInitialized     = errors.New("yt-dlp client not initialized")
	ErrConfigNotInitialized   = errors.New("config not initialized")
	ErrLoggerNotInitialized   = errors.New("logger not initialized")

	// Download errors
	ErrDownloadNotFound      = errors.New("download not found")
	ErrDownloadAlreadyExists = errors.New("download already in queue")
	ErrInvalidURL            = errors.New("invalid YouTube URL")
	ErrDownloadFailed        = errors.New("download failed")
	ErrDownloadPaused        = errors.New("download paused")
	ErrDownloadCancelled     = errors.New("download cancelled")

	// Video errors
	ErrVideoNotFound   = errors.New("video not found")
	ErrFileNotFound    = errors.New("video file not found")
	ErrFileNotManaged  = errors.New("file is not managed by YTed")

	// Settings errors
	ErrInvalidSetting = errors.New("invalid setting value")
	ErrPresetNotFound = errors.New("download preset not found")

	// Storage errors
	ErrDiskFull       = errors.New("insufficient disk space")
	ErrPathNotWritable = errors.New("download path not writable")
	ErrSystemDirectory = errors.New("cannot use system directory")
)

// IsNotFoundError checks if an error is a "not found" type error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrDownloadNotFound) ||
		errors.Is(err, ErrVideoNotFound) ||
		errors.Is(err, ErrFileNotFound)
}

// IsInfrastructureError checks if an error is an infrastructure/setup error
func IsInfrastructureError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrDatabaseNotInitialized) ||
		errors.Is(err, ErrYtdlNotInitialized) ||
		errors.Is(err, ErrConfigNotInitialized)
}
