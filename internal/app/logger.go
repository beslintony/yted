package app

import (
	"fmt"
	"path/filepath"
	"time"

	"yted/internal/config"
	"yted/internal/log"
)

// LogEntry represents a log entry for the frontend
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
}

// GetLogs returns recent log entries
func (a *App) GetLogs(count int) []LogEntry {
	logger := log.GetLogger()
	entries := logger.GetRecentEntries(count)

	result := make([]LogEntry, len(entries))
	for i, entry := range entries {
		result[i] = LogEntry{
			Timestamp: entry.Timestamp.Format("2006-01-02 15:04:05"),
			Level:     string(entry.Level),
			Component: entry.Component,
			Message:   entry.Message,
			Error:     entry.Error,
		}
	}

	return result
}

// GetAllLogs returns all log entries
func (a *App) GetAllLogs() []LogEntry {
	logger := log.GetLogger()
	entries := logger.GetAllEntries()

	result := make([]LogEntry, len(entries))
	for i, entry := range entries {
		result[i] = LogEntry{
			Timestamp: entry.Timestamp.Format("2006-01-02 15:04:05"),
			Level:     string(entry.Level),
			Component: entry.Component,
			Message:   entry.Message,
			Error:     entry.Error,
		}
	}

	return result
}

// ClearLogs clears all log entries
func (a *App) ClearLogs() {
	logger := log.GetLogger()
	logger.Clear()
}

// ExportLogs exports logs to a file
func (a *App) ExportLogs(customPath string) error {
	logger := log.GetLogger()

	// Use custom path if provided, otherwise use configured log export path
	exportPath := customPath
	if exportPath == "" && a.config != nil {
		exportPath = a.config.Get().LogExportPath
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := filepath.Join(exportPath, fmt.Sprintf("yted-logs-%s.json", timestamp))

	return logger.Export(filename)
}

// GetLogExportPath returns the configured log export path
func (a *App) GetLogExportPath() string {
	if a.config == nil {
		return ""
	}
	return a.config.Get().LogExportPath
}

// SetLogExportPath sets the log export path
func (a *App) SetLogExportPath(path string) error {
	if a.config == nil {
		return nil
	}

	a.config.Update(func(cfg *config.Config) {
		cfg.LogExportPath = path
	})

	return a.config.Save()
}
