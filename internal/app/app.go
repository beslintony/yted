package app

import (
	"context"
	"log"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"yted/internal/config"
	"yted/internal/db"
	"yted/internal/ytdl"
)

// App struct
type App struct {
	ctx    context.Context
	db     *db.DB
	config *config.Manager
	ytdl   *ytdl.Client
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Get app data directory
	appDataDir, err := config.GetAppDataDir()
	if err != nil {
		log.Printf("Error getting app data dir: %v", err)
		return
	}

	// Initialize config
	cfgManager, err := config.NewManager(appDataDir)
	if err != nil {
		log.Printf("Error initializing config: %v", err)
		return
	}
	
	if err := cfgManager.Load(); err != nil {
		log.Printf("Error loading config: %v", err)
	}
	a.config = cfgManager

	// Initialize database
	database, err := db.New(appDataDir)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		return
	}
	a.db = database

	// Initialize ytdl client
	ytdlConfig := &ytdl.ClientConfig{
		DownloadPath:     cfgManager.Get().DownloadPath,
		FilenameTemplate: cfgManager.Get().FilenameTemplate,
		ProxyURL:         cfgManager.Get().ProxyURL,
		SpeedLimitKbps:   cfgManager.Get().SpeedLimitKbps,
	}
	a.ytdl = ytdl.NewClient(ytdlConfig)

	// Ensure download directory exists
	if err := os.MkdirAll(cfgManager.Get().DownloadPath, 0755); err != nil {
		log.Printf("Error creating download directory: %v", err)
	}

	log.Println("YTed started successfully")
}

// Shutdown is called when the app shuts down
func (a *App) Shutdown(ctx context.Context) {
	if a.config != nil {
		if err := a.config.Save(); err != nil {
			log.Printf("Error saving config: %v", err)
		}
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

// DOMReady is called when the frontend is ready
func (a *App) DOMReady(ctx context.Context) {
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
