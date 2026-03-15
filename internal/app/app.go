package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"yted/internal/config"
	"yted/internal/db"
	applog "yted/internal/log"
	"yted/internal/ytdl"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	db     *db.DB
	config *config.Manager
	ytdl   *ytdl.Client
	logger *applog.Logger
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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
		"version":     "1.0.0",
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

	// Initialize ytdl client
	ytdlConfig := &ytdl.ClientConfig{
		DownloadPath:     cfgManager.Get().DownloadPath,
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

	// Ensure download directory exists
	if err := os.MkdirAll(cfgManager.Get().DownloadPath, 0755); err != nil {
		a.logger.Error("App", "Failed to create download directory", err)
	}

	a.logger.Info("App", "YTed started successfully")
}

// Shutdown is called when the app shuts down
func (a *App) Shutdown(ctx context.Context) {
	a.logger.Info("App", "Shutting down YTed")

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
}

// DOMReady is called when the frontend is ready
func (a *App) DOMReady(ctx context.Context) {
	a.logger.Info("App", "Frontend DOM ready")
	// Frontend is ready, can emit events now
	runtime.EventsEmit(a.ctx, "app:ready", nil)
}

// GetVersion returns the app version
func (a *App) GetVersion() string {
	return "1.0.0"
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

// ShowSaveDialog shows a save file dialog
func (a *App) ShowSaveDialog(defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
	})
}

// OpenFile opens a file with the default application
func (a *App) OpenFile(path string) {
	runtime.BrowserOpenURL(a.ctx, "file://"+path)
}

// OpenFolder opens a folder in the file manager
func (a *App) OpenFolder(path string) {
	runtime.BrowserOpenURL(a.ctx, "file://"+path)
}
