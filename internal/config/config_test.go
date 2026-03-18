package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig(tempDir)

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Test default values
	if cfg.MaxConcurrentDownloads != 3 {
		t.Errorf("Expected MaxConcurrentDownloads=3, got %d", cfg.MaxConcurrentDownloads)
	}

	if cfg.DefaultQuality != "best" {
		t.Errorf("Expected DefaultQuality='best', got %s", cfg.DefaultQuality)
	}

	if cfg.Theme != "dark" {
		t.Errorf("Expected Theme='dark', got %s", cfg.Theme)
	}

	if cfg.DefaultVolume != 80 {
		t.Errorf("Expected DefaultVolume=80, got %d", cfg.DefaultVolume)
	}

	if cfg.MaxLogSessions != 10 {
		t.Errorf("Expected MaxLogSessions=10, got %d", cfg.MaxLogSessions)
	}

	// Test that download presets are populated
	if len(cfg.DownloadPresets) == 0 {
		t.Error("Expected DownloadPresets to be populated")
	}

	// Check for expected preset names
	expectedPresets := map[string]bool{
		"4K (2160p)":       false,
		"1080p HD":         false,
		"Audio Only (MP3)": false,
	}

	for _, preset := range cfg.DownloadPresets {
		if _, exists := expectedPresets[preset.Name]; exists {
			expectedPresets[preset.Name] = true
		}
	}

	for name, found := range expectedPresets {
		if !found {
			t.Errorf("Expected preset '%s' not found", name)
		}
	}
}

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)

	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.config == nil {
		t.Error("Manager config is nil")
	}

	if manager.configPath == "" {
		t.Error("Manager configPath is empty")
	}

	// Verify config directory was created
	configDir := filepath.Join(tempDir, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}
}

func TestManagerLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Modify config
	manager.config.DownloadPath = "/test/download/path"
	manager.config.MaxConcurrentDownloads = 5
	manager.config.Theme = "light"

	// Save config
	if err := manager.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(manager.configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Create new manager and load
	manager2, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager (2nd) failed: %v", err)
	}

	if err := manager2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded values
	cfg := manager2.Get()
	if cfg.DownloadPath != "/test/download/path" {
		t.Errorf("Expected DownloadPath='/test/download/path', got %s", cfg.DownloadPath)
	}

	if cfg.MaxConcurrentDownloads != 5 {
		t.Errorf("Expected MaxConcurrentDownloads=5, got %d", cfg.MaxConcurrentDownloads)
	}

	if cfg.Theme != "light" {
		t.Errorf("Expected Theme='light', got %s", cfg.Theme)
	}
}

func TestManagerLoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Load should not fail when file doesn't exist
	if err := manager.Load(); err != nil {
		t.Errorf("Load failed for non-existent file: %v", err)
	}

	// Should have default config
	cfg := manager.Get()
	if cfg.MaxConcurrentDownloads != 3 {
		t.Error("Expected default config after loading non-existent file")
	}
}

func TestManagerGetAndUpdate(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Get initial config
	cfg := manager.Get()
	initialValue := cfg.MaxConcurrentDownloads

	// Update config
	manager.Update(func(c *Config) {
		c.MaxConcurrentDownloads = 10
	})

	// Get updated config
	updatedCfg := manager.Get()
	if updatedCfg.MaxConcurrentDownloads != 10 {
		t.Errorf("Expected MaxConcurrentDownloads=10 after update, got %d", updatedCfg.MaxConcurrentDownloads)
	}

	// Original pointer should also reflect change (same underlying object)
	if cfg.MaxConcurrentDownloads == initialValue {
		t.Error("Original config pointer was not updated")
	}
}

func TestManagerConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Run concurrent reads and writes
	done := make(chan bool, 3)

	// Reader
	go func() {
		for i := 0; i < 100; i++ {
			_ = manager.Get()
		}
		done <- true
	}()

	// Writer
	go func() {
		for i := 0; i < 100; i++ {
			manager.Update(func(c *Config) {
				c.MaxConcurrentDownloads = i
			})
		}
		done <- true
	}()

	// Another reader
	go func() {
		for i := 0; i < 100; i++ {
			_ = manager.Get()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestGetAppDataDir(t *testing.T) {
	// This test might fail in CI environments without HOME
	dir, err := GetAppDataDir()
	if err != nil {
		t.Skipf("GetAppDataDir failed (might be CI environment): %v", err)
	}

	if dir == "" {
		t.Error("GetAppDataDir returned empty string")
	}

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("App data directory was not created")
	}

	// Verify it ends with .yted
	if filepath.Base(dir) != ".yted" {
		t.Errorf("Expected directory name '.yted', got %s", filepath.Base(dir))
	}
}

func TestConfigPresetsHaveRequiredFields(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig(tempDir)

	for i, preset := range cfg.DownloadPresets {
		if preset.ID == "" {
			t.Errorf("Preset %d: ID is empty", i)
		}
		if preset.Name == "" {
			t.Errorf("Preset %d: Name is empty", i)
		}
		if preset.Format == "" {
			t.Errorf("Preset %d: Format is empty", i)
		}
		if preset.Quality == "" {
			t.Errorf("Preset %d: Quality is empty", i)
		}
		if preset.Extension == "" {
			t.Errorf("Preset %d: Extension is empty", i)
		}
	}
}

func TestConfigDefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig(tempDir)

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"FilenameTemplate", cfg.FilenameTemplate, "%(title).60s [%(id)s][%(format_id)s].%(ext)s"},
		{"AccentColor", cfg.AccentColor, "#ff0000"},
		{"RememberPosition", cfg.RememberPosition, true},
		{"SidebarCollapsed", cfg.SidebarCollapsed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, tt.got)
			}
		})
	}
}
