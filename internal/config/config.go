package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config holds all user-configurable settings
type Config struct {
	// Downloads
	UserSelectedPath       string `json:"user_selected_path"` // Path user selected (parent)
	DownloadPath           string `json:"download_path"`      // Actual path we use (YTed subfolder)
	MaxConcurrentDownloads int    `json:"max_concurrent_downloads"`
	DefaultQuality         string `json:"default_quality"`
	FilenameTemplate       string `json:"filename_template"`

	// UI
	Theme            string `json:"theme"`
	AccentColor      string `json:"accent_color"`
	SidebarCollapsed bool   `json:"sidebar_collapsed"`

	// Player
	DefaultVolume    int  `json:"default_volume"`
	RememberPosition bool `json:"remember_position"`

	// Network
	SpeedLimitKbps *int    `json:"speed_limit_kbps"`
	ProxyURL       *string `json:"proxy_url"`

	// Logging
	LogPath        string `json:"log_path"`         // Internal log storage path
	LogExportPath  string `json:"log_export_path"`  // Export destination
	MaxLogSessions int    `json:"max_log_sessions"` // Number of sessions to keep (default: 10)

	// Presets
	DownloadPresets []DownloadPreset `json:"download_presets"`

	// FFmpeg
	FFmpegPath string `json:"ffmpeg_path,omitempty"` // Custom ffmpeg path (empty = auto-detect)
}

// DownloadPreset represents a saved download configuration
type DownloadPreset struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Format    string `json:"format"`
	Quality   string `json:"quality"`
	Extension string `json:"extension"`
}

// DefaultConfig returns the default configuration
func DefaultConfig(appDataDir string) *Config {
	homeDir, _ := os.UserHomeDir()
	defaultDownloadPath := filepath.Join(homeDir, "Downloads", "YTed")
	defaultLogExportPath := filepath.Join(homeDir, "Downloads")

	// Default log path is inside the app data directory as .logs
	defaultLogPath := filepath.Join(appDataDir, ".logs")

	return &Config{
		DownloadPath:           defaultDownloadPath,
		MaxConcurrentDownloads: 3,
		DefaultQuality:         "best",
		FilenameTemplate:       "%(title).60s [%(id)s][%(format_id)s].%(ext)s",
		Theme:                  "dark",
		AccentColor:            "#ff0000",
		SidebarCollapsed:       false,
		DefaultVolume:          80,
		RememberPosition:       true,
		SpeedLimitKbps:         nil,
		ProxyURL:               nil,
		LogPath:                defaultLogPath,
		LogExportPath:          defaultLogExportPath,
		MaxLogSessions:         10,
		DownloadPresets: []DownloadPreset{
			// High quality presets with explicit format selection
			{ID: "1", Name: "4K (2160p)", Format: "bestvideo[height<=2160][vcodec^=avc1]+bestaudio/bestvideo[height<=2160]+bestaudio", Quality: "2160p", Extension: "mp4"},
			{ID: "2", Name: "1440p (2K)", Format: "bestvideo[height<=1440][vcodec^=avc1]+bestaudio/bestvideo[height<=1440]+bestaudio", Quality: "1440p", Extension: "mp4"},
			{ID: "3", Name: "1080p HD", Format: "bestvideo[height<=1080][vcodec^=avc1]+bestaudio/bestvideo[height<=1080]+bestaudio", Quality: "1080p", Extension: "mp4"},
			{ID: "4", Name: "720p HD", Format: "bestvideo[height<=720][vcodec^=avc1]+bestaudio/bestvideo[height<=720]+bestaudio", Quality: "720p", Extension: "mp4"},
			{ID: "5", Name: "480p", Format: "bestvideo[height<=480]+bestaudio/best[height<=480]", Quality: "480p", Extension: "mp4"},
			{ID: "6", Name: "Best Available", Format: "bestvideo+bestaudio/best", Quality: "best", Extension: "mp4"},
			{ID: "7", Name: "Audio Only (MP3)", Format: "bestaudio", Quality: "audio", Extension: "mp3"},
			{ID: "8", Name: "Audio Only (M4A)", Format: "bestaudio[ext=m4a]", Quality: "audio", Extension: "m4a"},
		},
	}
}

// Manager handles configuration persistence
type Manager struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
}

// NewManager creates a new config manager
func NewManager(appDataDir string) (*Manager, error) {
	configDir := filepath.Join(appDataDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &Manager{
		configPath: filepath.Join(configDir, "settings.json"),
		config:     DefaultConfig(appDataDir),
	}, nil
}

// Load reads configuration from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			return nil
		}
		return fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Ensure defaults for new fields that might be empty in existing configs
	if cfg.LogPath == "" {
		cfg.LogPath = filepath.Join(filepath.Dir(m.configPath), "..", ".logs")
	}
	if cfg.MaxLogSessions == 0 {
		cfg.MaxLogSessions = 10
	}
	if cfg.LogExportPath == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.LogExportPath = filepath.Join(homeDir, "Downloads")
	}
	// Set default filename template if not configured
	// Format ID is included to allow multiple versions (e.g., 720p vs 1080p) of same video
	// Title is truncated to 60 chars to prevent "file name too long" errors
	if cfg.FilenameTemplate == "" {
		cfg.FilenameTemplate = "%(title).60s [%(id)s][%(format_id)s].%(ext)s"
	}

	m.config = &cfg
	return nil
}

// Save writes configuration to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Update updates the configuration
func (m *Manager) Update(fn func(*Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(m.config)
}

// GetAppDataDir returns the application data directory
func GetAppDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appDataDir := filepath.Join(homeDir, ".yted")
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return "", err
	}

	return appDataDir, nil
}
