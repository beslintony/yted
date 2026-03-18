package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"strings"
	"sync"
	"time"

	"yted/internal/config"
	"yted/internal/db"
	applog "yted/internal/log"
	"yted/internal/version"
	"yted/internal/ytdl"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// videoInfoCacheEntry stores cached video info with TTL
type videoInfoCacheEntry struct {
	info      *VideoInfoResult
	expiresAt time.Time
}

// App struct
type App struct {
	ctx    context.Context
	db     *db.DB
	config *config.Manager
	ytdl   *ytdl.Client
	logger *applog.Logger
	fm     *FileManager
	ffmpeg *FFmpegManager

	// Mutex to prevent concurrent download processing
	downloadMu sync.Mutex

	// Active downloads map for cancellation (pause/resume)
	activeDownloads   map[string]context.CancelFunc
	activeDownloadsMu sync.RWMutex

	// Graceful shutdown support
	activeDownloadsWg sync.WaitGroup
	shutdownCh        chan struct{}

	// Video info cache to prevent duplicate fetches
	videoInfoCache   map[string]videoInfoCacheEntry
	videoInfoCacheMu sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		activeDownloads:  make(map[string]context.CancelFunc),
		shutdownCh:       make(chan struct{}),
		videoInfoCache:   make(map[string]videoInfoCacheEntry),
	}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.logger = applog.GetLogger()

	// Get app data directory
	appDataDir, err := config.GetAppDataDir()
	if err != nil {
		a.logger.Error("App", "Failed to get app data dir", err)
		return
	}

	// Initialize config first (needed for log settings)
	cfgManager, err := config.NewManager(appDataDir)
	if err != nil {
		a.logger.Error("Config", "Failed to initialize config", err)
		return
	}

	if err := cfgManager.Load(); err != nil {
		a.logger.Error("Config", "Failed to load config", err)
	}
	a.config = cfgManager

	// Initialize logger with configured log directory and session management
	logDir := cfgManager.Get().LogPath
	if logDir == "" {
		// Fallback to default log path
		logDir = filepath.Join(appDataDir, ".logs")
	}
	maxSessions := cfgManager.Get().MaxLogSessions
	if maxSessions < 1 {
		maxSessions = 10
	}
	if err := a.logger.SetLogDirWithSessions(logDir, maxSessions); err != nil {
		a.logger.Error("App", "Failed to set log directory", err)
		// Fallback to basic logging without sessions
		if err := a.logger.SetLogDir(logDir); err != nil {
			a.logger.Error("App", "Failed to set fallback log directory", err)
		}
	}

	a.logger.Info("App", "Starting YTed", map[string]string{
		"version":     version.GetVersion(),
		"appDir":      appDataDir,
		"logDir":      logDir,
		"maxSessions": fmt.Sprintf("%d", maxSessions),
	})
	a.logger.Info("Config", "Configuration loaded", map[string]interface{}{
		"downloadPath": cfgManager.Get().DownloadPath,
		"theme":        cfgManager.Get().Theme,
	})

	// Initialize database
	database, err := db.New(appDataDir)
	if err != nil {
		a.logger.Error("Database", "Failed to initialize database", err)
		return
	}
	a.db = database
	a.logger.Info("Database", "Database initialized")

	// Initialize file manager
	a.fm = NewFileManager(cfgManager)
	if err := a.fm.EnsureYTedFolder(); err != nil {
		a.logger.Error("App", "Failed to ensure YTed folder", err)
	} else {
		a.logger.Info("App", "File manager initialized", map[string]string{
			"path": a.fm.GetDownloadPath(),
		})
	}

	// Initialize ytdl client
	ytdlConfig := &ytdl.ClientConfig{
		DownloadPath:     a.fm.GetDownloadPath(),
		FilenameTemplate: cfgManager.Get().FilenameTemplate,
		ProxyURL:         cfgManager.Get().ProxyURL,
		SpeedLimitKbps:   cfgManager.Get().SpeedLimitKbps,
	}
	a.ytdl = ytdl.NewClient(ytdlConfig)
	a.logger.Info("YTDLP", "yt-dlp client initialized")

	// Install yt-dlp binary
	if err := a.ytdl.Install(ctx); err != nil {
		a.logger.Error("YTDLP", "Failed to install yt-dlp", err)
	} else {
		a.logger.Info("YTDLP", "yt-dlp installed successfully")
	}

	// Check for yt-dlp updates in the background (non-blocking)
	// yt-dlp requires frequent updates to handle YouTube changes
	go a.checkForYtdlpUpdate()

	// Find and configure ffmpeg for video/audio merging
	a.ffmpeg = NewFFmpegManager()
	
	// Set custom path from config if available
	if cfgManager.Get().FFmpegPath != "" {
		a.ffmpeg.SetCustomPath(cfgManager.Get().FFmpegPath)
	}
	
	if ffmpegPath := a.ffmpeg.Find(); ffmpegPath != "" {
		a.ytdl.SetFFmpegPath(ffmpegPath)
		a.logger.Info("FFmpeg", "FFmpeg configured for video/audio merging", map[string]string{
			"path": ffmpegPath,
		})
	} else {
		a.logger.Warn("FFmpeg", "FFmpeg not found - video and audio may be downloaded separately", map[string]string{
			"install_instructions": a.ffmpeg.InstallInstructions(),
		})
	}

	// Ensure download directory exists
	if err := os.MkdirAll(a.fm.GetDownloadPath(), 0755); err != nil {
		a.logger.Error("App", "Failed to create download directory", err)
	}

	// Verify and repair any downloads that have files but wrong status
	if err := a.VerifyAndRepairDownloads(); err != nil {
		a.logger.Error("App", "Failed to verify downloads", err)
	}

	// Clean up any existing duplicate library entries
	if err := a.CleanUpDuplicateVideos(); err != nil {
		a.logger.Error("App", "Failed to clean up duplicate videos", err)
	}

	// Note: Download queue restoration is now handled by the frontend
	// calling GetDownloadQueue() when it's ready

	a.logger.Info("App", "YTed started successfully")
}

// Shutdown is called when the app shuts down
func (a *App) Shutdown(ctx context.Context) {
	a.logger.Info("App", "Shutting down YTed")

	// Signal shutdown to all goroutines
	close(a.shutdownCh)

	// Cancel all active downloads gracefully
	a.activeDownloadsMu.Lock()
	activeCount := len(a.activeDownloads)
	for id, cancel := range a.activeDownloads {
		a.logger.Info("Download", "Cancelling active download", map[string]string{"id": id})
		cancel()
	}
	a.activeDownloadsMu.Unlock()

	// Wait for active downloads to finish (with timeout)
	if activeCount > 0 {
		a.logger.Info("App", "Waiting for downloads to finish", map[string]int{"count": activeCount})
		
		done := make(chan struct{})
		go func() {
			a.activeDownloadsWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			a.logger.Info("App", "All downloads finished gracefully")
		case <-time.After(5 * time.Second):
			a.logger.Warn("App", "Timeout waiting for downloads, forcing shutdown")
		}
	}

	if a.config != nil {
		if err := a.config.Save(); err != nil {
			a.logger.Error("Config", "Failed to save config", err)
		}
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Error("Database", "Failed to close database", err)
		}
	}

	a.logger.Info("App", "YTed shutdown complete")
}

// checkForYtdlpUpdate checks for and installs yt-dlp updates
// Runs in the background - yt-dlp requires frequent updates for YouTube compatibility
func (a *App) checkForYtdlpUpdate() {
	if a.ytdl == nil {
		return
	}

	// Check current version first
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	version, err := a.ytdl.GetVersion(ctx)
	cancel()
	if err != nil {
		a.logger.Warn("YTDLP", "Could not get current version", err)
	} else {
		a.logger.Info("YTDLP", "Current version", map[string]string{"version": version})
	}

	// Attempt update
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	updated, err := a.ytdl.Update(ctx)
	if err != nil {
		a.logger.Error("YTDLP", "Update check failed", err)
		return
	}

	if updated {
		a.logger.Info("YTDLP", "yt-dlp updated to latest version")
		runtime.EventsEmit(a.ctx, "ytdlp:updated", nil)
	} else {
		a.logger.Info("YTDLP", "yt-dlp is up to date")
	}
}

// DOMReady is called when the frontend is ready
func (a *App) DOMReady(ctx context.Context) {
	a.logger.Info("App", "Frontend DOM ready")
	// Frontend is ready, can emit events now
	runtime.EventsEmit(a.ctx, "app:ready", nil)
}

// GetVersion returns the app version
func (a *App) GetVersion() string {
	return version.GetVersion()
}

// GetYtdlpVersion returns the current yt-dlp version
func (a *App) GetYtdlpVersion() string {
	if a.ytdl == nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	version, err := a.ytdl.GetVersion(ctx)
	if err != nil {
		a.logger.Error("YTDLP", "Failed to get version", err)
		return ""
	}
	return version
}

// UpdateYtdlp manually triggers a yt-dlp update check
// Returns true if an update was performed
func (a *App) UpdateYtdlp() (bool, error) {
	if a.ytdl == nil {
		return false, fmt.Errorf("yt-dlp client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return a.ytdl.Update(ctx)
}

// GetAppName returns the app name
func (a *App) GetAppName() string {
	return "YTed"
}

// ShowError shows an error dialog
func (a *App) ShowError(message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   "Error",
		Message: message,
	})
}

// ShowOpenDirectoryDialog shows a directory picker
func (a *App) ShowOpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Download Directory",
	})
}

// ShowFFmpegDialog shows a file picker for ffmpeg binary
func (a *App) ShowFFmpegDialog() (string, error) {
	filters := []runtime.FileFilter{}
	
	// Add platform-specific filters
	if goRuntime.GOOS == "windows" {
		filters = append(filters, runtime.FileFilter{
			DisplayName: "Executable files (*.exe)",
			Pattern:     "*.exe",
		})
	}
	filters = append(filters, runtime.FileFilter{
		DisplayName: "All files",
		Pattern:     "*",
	})
	
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Select FFmpeg Binary",
		Filters: filters,
	})
}

// ShowSaveDialog shows a save file dialog
func (a *App) ShowSaveDialog(defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
	})
}

// OpenFile opens a file with the default application
func (a *App) OpenFile(path string) error {
	logger := applog.GetLogger()

	// Verify file exists
	if path == "" {
		return fmt.Errorf("file path is empty")
	}

	// Clean and validate the path
	path = filepath.Clean(path)

	// Security: Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid path: path traversal detected")
	}

	// Convert to absolute path for validation
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Error("App", "File not found", err, map[string]string{"path": absPath})
			return fmt.Errorf("file not found: %s", absPath)
		}
		logger.Error("App", "Cannot access file", err, map[string]string{"path": absPath})
		return fmt.Errorf("cannot access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", absPath)
	}

	logger.Info("App", "Opening file", map[string]string{"path": absPath})

	// Try native OS command first (using separate args to prevent injection)
	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", absPath)
	default: // Linux and others
		cmd = exec.Command("xdg-open", absPath)
	}

	if err := cmd.Start(); err != nil {
		// Fallback to BrowserOpenURL (especially for Linux sandboxed apps)
		logger.Warn("App", "Failed to open file with native command, trying BrowserOpenURL", map[string]string{
			"path":  absPath,
			"error": err.Error(),
		})

		fileURL := "file://" + absPath
		runtime.BrowserOpenURL(a.ctx, fileURL)
	}

	return nil
}

// OpenFolder opens the folder containing the file
func (a *App) OpenFolder(filePath string) error {
	logger := applog.GetLogger()

	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	// Clean and validate the path
	filePath = filepath.Clean(filePath)

	// Security: Check for path traversal attempts
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("invalid path: path traversal detected")
	}

	// Get the directory containing the file
	dir := filepath.Dir(filePath)

	// Convert to absolute path for validation
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify directory exists
	dirInfo, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Error("App", "Folder not found", err, map[string]string{"path": absDir})
			return fmt.Errorf("folder not found: %s", absDir)
		}
		logger.Error("App", "Cannot access folder", err, map[string]string{"path": absDir})
		return fmt.Errorf("cannot access folder: %w", err)
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absDir)
	}

	logger.Info("App", "Opening folder", map[string]string{"path": absDir})

	// Try native OS command first (using separate args to prevent injection)
	var cmd *exec.Cmd
	switch goRuntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absDir)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", absDir)
	default: // Linux and others
		cmd = exec.Command("xdg-open", absDir)
	}

	if err := cmd.Start(); err != nil {
		// Fallback to BrowserOpenURL (especially for Linux sandboxed apps)
		logger.Warn("App", "Failed to open folder with native command, trying BrowserOpenURL", map[string]string{
			"path":  absDir,
			"error": err.Error(),
		})

		folderURL := "file://" + absDir
		runtime.BrowserOpenURL(a.ctx, folderURL)
	}

	return nil
}
