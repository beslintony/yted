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
	DownloadPath           string           `json:"download_path"`
	MaxConcurrentDownloads int              `json:"max_concurrent_downloads"`
	DefaultQuality         string           `json:"default_quality"`
	FilenameTemplate       string           `json:"filename_template"`

	// UI
	Theme            string `json:"theme"`
	AccentColor      string `json:"accent_color"`
	SidebarCollapsed bool   `json:"sidebar_collapsed"`

	// Player
	DefaultVolume      int  `json:"default_volume"`
	RememberPosition   bool `json:"remember_position"`

	// Network
	SpeedLimitKbps *int    `json:"speed_limit_kbps"`
	ProxyURL       *string `json:"proxy_url"`

	// Presets
	DownloadPresets []DownloadPreset `json:"download_presets"`
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
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	defaultDownloadPath := filepath.Join(homeDir, "Downloads", "YTed")

	return &Config{
		DownloadPath:           defaultDownloadPath,
		MaxConcurrentDownloads: 3,
		DefaultQuality:         "best",
		FilenameTemplate:       "%(title)s.%(ext)s",
		Theme:                  "dark",
		AccentColor:            "#ff0000",
		SidebarCollapsed:       false,
		DefaultVolume:          80,
		RememberPosition:       true,
		SpeedLimitKbps:         nil,
		ProxyURL:               nil,
		DownloadPresets: []DownloadPreset{
			{ID: "1", Name: "Best Quality", Format: "best", Quality: "best", Extension: "mp4"},
			{ID: "2", Name: "1080p", Format: "bestvideo[height<=1080]+bestaudio", Quality: "1080p", Extension: "mp4"},
			{ID: "3", Name: "720p", Format: "bestvideo[height<=720]+bestaudio", Quality: "720p", Extension: "mp4"},
			{ID: "4", Name: "Audio Only", Format: "bestaudio", Quality: "audio", Extension: "mp3"},
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
		config:     DefaultConfig(),
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
