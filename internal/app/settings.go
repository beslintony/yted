package app

import (
	"yted/internal/config"
)

// GetSettings returns the current user settings
func (a *App) GetSettings() (*config.Config, error) {
	if a.config == nil {
		return nil, nil
	}
	return a.config.Get(), nil
}

// SaveSettings saves the user settings
func (a *App) SaveSettings(settings *config.Config) error {
	if a.config == nil {
		return nil
	}

	a.config.Update(func(cfg *config.Config) {
		*cfg = *settings
	})

	return a.config.Save()
}

// UpdateSetting updates a single setting
func (a *App) UpdateSetting(key string, value interface{}) error {
	if a.config == nil {
		return nil
	}

	a.config.Update(func(cfg *config.Config) {
		switch key {
		case "download_path":
			if v, ok := value.(string); ok {
				cfg.DownloadPath = v
			}
		case "max_concurrent_downloads":
			if v, ok := value.(float64); ok {
				cfg.MaxConcurrentDownloads = int(v)
			}
		case "default_quality":
			if v, ok := value.(string); ok {
				cfg.DefaultQuality = v
			}
		case "filename_template":
			if v, ok := value.(string); ok {
				cfg.FilenameTemplate = v
			}
		case "theme":
			if v, ok := value.(string); ok {
				cfg.Theme = v
			}
		case "accent_color":
			if v, ok := value.(string); ok {
				cfg.AccentColor = v
			}
		case "sidebar_collapsed":
			if v, ok := value.(bool); ok {
				cfg.SidebarCollapsed = v
			}
		case "default_volume":
			if v, ok := value.(float64); ok {
				cfg.DefaultVolume = int(v)
			}
		case "remember_position":
			if v, ok := value.(bool); ok {
				cfg.RememberPosition = v
			}
		case "speed_limit_kbps":
			if value == nil {
				cfg.SpeedLimitKbps = nil
			} else if v, ok := value.(float64); ok {
				limit := int(v)
				cfg.SpeedLimitKbps = &limit
			}
		case "proxy_url":
			if value == nil {
				cfg.ProxyURL = nil
			} else if v, ok := value.(string); ok {
				cfg.ProxyURL = &v
			}
		case "log_export_path":
			if v, ok := value.(string); ok {
				cfg.LogExportPath = v
			}
		case "log_path":
			if v, ok := value.(string); ok {
				cfg.LogPath = v
			}
		case "max_log_sessions":
			if v, ok := value.(float64); ok {
				cfg.MaxLogSessions = int(v)
				// Update logger's max sessions in real-time
				if a.logger != nil {
					a.logger.SetMaxSessions(int(v))
				}
			}
		}
	})

	return a.config.Save()
}

// GetDownloadPresets returns the download presets
func (a *App) GetDownloadPresets() ([]config.DownloadPreset, error) {
	if a.config == nil {
		return nil, nil
	}
	return a.config.Get().DownloadPresets, nil
}

// AddDownloadPreset adds a new download preset
func (a *App) AddDownloadPreset(preset config.DownloadPreset) error {
	if a.config == nil {
		return nil
	}

	a.config.Update(func(cfg *config.Config) {
		cfg.DownloadPresets = append(cfg.DownloadPresets, preset)
	})

	return a.config.Save()
}

// RemoveDownloadPreset removes a download preset
func (a *App) RemoveDownloadPreset(id string) error {
	if a.config == nil {
		return nil
	}

	a.config.Update(func(cfg *config.Config) {
		presets := make([]config.DownloadPreset, 0, len(cfg.DownloadPresets))
		for _, p := range cfg.DownloadPresets {
			if p.ID != id {
				presets = append(presets, p)
			}
		}
		cfg.DownloadPresets = presets
	})

	return a.config.Save()
}

// UpdateDownloadPreset updates a download preset
func (a *App) UpdateDownloadPreset(id string, preset config.DownloadPreset) error {
	if a.config == nil {
		return nil
	}

	a.config.Update(func(cfg *config.Config) {
		for i, p := range cfg.DownloadPresets {
			if p.ID == id {
				cfg.DownloadPresets[i] = preset
				break
			}
		}
	})

	return a.config.Save()
}
