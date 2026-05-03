package app

import (
	"runtime"
	"strings"
	"testing"
)

func TestFFmpegManagerFind(t *testing.T) {
	fm := NewFFmpegManager()

	// Just test that it doesn't panic
	path := fm.Find()

	// On CI systems, ffmpeg might not be installed
	// So we just verify the function works
	t.Logf("FFmpeg path: %s", path)
}

func TestFFmpegManagerIsAvailable(t *testing.T) {
	fm := NewFFmpegManager()

	// Should return boolean without error
	available := fm.IsAvailable()
	t.Logf("FFmpeg available: %v", available)
}

func TestFFmpegManagerGetPath(t *testing.T) {
	fm := NewFFmpegManager()

	// GetPath should call Find
	path := fm.GetPath()
	t.Logf("FFmpeg path from GetPath: %s", path)
}

func TestFFmpegManagerInstallInstructions(t *testing.T) {
	fm := NewFFmpegManager()

	instructions := fm.InstallInstructions()

	if instructions == "" {
		t.Error("Install instructions should not be empty")
	}

	// Check platform-specific instructions
	switch runtime.GOOS {
	case "linux":
		if !strings.Contains(instructions, "apt") && !strings.Contains(instructions, "dnf") {
			t.Error("Linux instructions should mention apt or dnf")
		}
	case "windows":
		if !strings.Contains(instructions, "ffmpeg.org") {
			t.Error("Windows instructions should mention ffmpeg.org")
		}
	default:
		if !strings.Contains(instructions, "Linux and Windows only") {
			t.Error("unsupported platform instructions should mention Linux and Windows only")
		}
	}

	t.Logf("Install instructions: %s", instructions)
}

func TestFFmpegManagerGetVersion(t *testing.T) {
	fm := NewFFmpegManager()

	// If ffmpeg is not available, should return empty string
	version := fm.GetVersion()
	t.Logf("FFmpeg version: %s", version)
}

func TestFFmpegManagerSetCustomPath(t *testing.T) {
	fm := NewFFmpegManager()

	// Test setting empty path
	fm.SetCustomPath("")
	if fm.customPath != "" {
		t.Error("Custom path should be empty")
	}

	// Test setting custom path (non-existent path should be handled gracefully)
	fm.SetCustomPath("/nonexistent/ffmpeg")
	if fm.customPath != "/nonexistent/ffmpeg" {
		t.Error("Custom path should be set")
	}

	// Verify binPath is reset when setting new custom path
	fm.binPath = "/some/path"
	fm.SetCustomPath("/another/path")
	if fm.binPath != "" {
		t.Error("binPath should be reset when setting new custom path")
	}
}

func TestFFmpegManagerValidatePath(t *testing.T) {
	fm := NewFFmpegManager()

	// Test empty path
	valid, version := fm.validatePath("")
	if valid {
		t.Error("Empty path should not be valid")
	}

	// Test non-existent path
	valid, version = fm.validatePath("/nonexistent/ffmpeg")
	if valid {
		t.Error("Non-existent path should not be valid")
	}

	// Test directory instead of file
	valid, version = fm.validatePath("/tmp")
	if valid {
		t.Error("Directory should not be valid")
	}

	t.Logf("Version from validatePath: %s", version)
}

func TestFFmpegManagerScanAllLocations(t *testing.T) {
	fm := NewFFmpegManager()

	// Scan all locations
	locations := fm.ScanAllLocations()

	// Should return a slice (may be empty if no ffmpeg found)
	if locations == nil {
		t.Error("ScanAllLocations should not return nil")
	}

	// Verify each location has required fields
	for _, loc := range locations {
		if loc.Path == "" {
			t.Error("Location should have a path")
		}
		if !loc.IsValid {
			t.Error("Returned location should be valid")
		}
		if loc.Source == "" {
			t.Error("Location should have a source")
		}
		// Source should be one of the expected values
		switch loc.Source {
		case "custom", "path", "common":
			// valid
		default:
			t.Errorf("Unexpected source: %s", loc.Source)
		}
	}

	t.Logf("Found %d FFmpeg locations", len(locations))
}

func TestFFmpegManagerExtractVersion(t *testing.T) {
	fm := NewFFmpegManager()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "ffmpeg version 6.1.1-3ubuntu5 Copyright (c) 2000-2023",
			expected: "6.1.1-3ubuntu5 Copyright (c) 2000-2023",
		},
		{
			input:    "ffmpeg version 5.1.2 Copyright (c) 2000-2022",
			expected: "5.1.2 Copyright (c) 2000-2022",
		},
		{
			input:    "ffmpeg version 4.4.2-0ubuntu0.22.04.1 Copyright (c)",
			expected: "4.4.2-0ubuntu0.22.04.1 Copyright (c)",
		},
		{
			input:    "invalid version string",
			expected: "invalid version string",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		result := fm.extractVersion(tt.input)
		if result != tt.expected {
			t.Errorf("extractVersion(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFFmpegManagerCheckFFmpegWithGuidance(t *testing.T) {
	fm := NewFFmpegManager()

	result := fm.CheckFFmpegWithGuidance()

	// Verify result has required fields
	if result.Installed && result.Path == "" {
		t.Error("If installed is true, path should not be empty")
	}

	if result.AllLocations == nil {
		t.Error("AllLocations should not be nil")
	}

	// If installed, selectedIndex should point to a valid location
	if result.Installed {
		if result.SelectedIndex < 0 || result.SelectedIndex >= len(result.AllLocations) {
			t.Errorf("SelectedIndex %d out of range for %d locations", result.SelectedIndex, len(result.AllLocations))
		}
	}

	// If not installed, should have install guidance
	if !result.Installed {
		if result.InstallGuide == "" {
			t.Error("Should have install guidance when not installed")
		}
		if result.InstallMethod == "" {
			t.Error("Should have install method when not installed")
		}
	}

	t.Logf("FFmpeg installed: %v, Path: %s, Locations found: %d", result.Installed, result.Path, len(result.AllLocations))
}

func TestFFmpegManagerGetInstallGuide(t *testing.T) {
	fm := NewFFmpegManager()

	guide := fm.GetInstallGuide()

	if guide.Title == "" {
		t.Error("Install guide should have a title")
	}
	if guide.Description == "" {
		t.Error("Install guide should have a description")
	}
	if len(guide.Steps) == 0 {
		t.Error("Install guide should have steps")
	}

	// Verify platform-specific content
	switch runtime.GOOS {
	case "linux":
		if !strings.Contains(guide.Title, "Linux") && !strings.Contains(guide.Title, "Ubuntu") && !strings.Contains(guide.Title, "Debian") {
			t.Error("Linux install guide title should mention Linux, Ubuntu, or Debian")
		}
	case "windows":
		if !strings.Contains(guide.Title, "Windows") {
			t.Error("Windows install guide title should mention Windows")
		}
	}

	t.Logf("Install guide: %s", guide.Title)
}

func TestFFmpegManagerGetCommonPaths(t *testing.T) {
	fm := NewFFmpegManager()

	paths := fm.getCommonPaths()

	// Should return paths for the current OS
	if paths == nil && (runtime.GOOS == "linux" || runtime.GOOS == "windows") {
		t.Error("Should return common paths for Linux and Windows")
	}

	// Verify paths are not empty on supported platforms
	if runtime.GOOS == "linux" {
		found := false
		for _, path := range paths {
			if strings.Contains(path, "ffmpeg") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Linux common paths should include ffmpeg")
		}
	}

	t.Logf("Common paths: %v", paths)
}

func TestFFmpegManagerDetectLinuxDistro(t *testing.T) {
	fm := NewFFmpegManager()

	// This test will only work meaningfully on Linux
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux distro detection test on non-Linux platform")
	}

	distro := fm.detectLinuxDistro()

	// Should return a known distro or "unknown"
	validDistros := []string{"debian", "ubuntu", "fedora", "arch", "unknown"}
	valid := false
	for _, d := range validDistros {
		if distro == d {
			valid = true
			break
		}
	}
	if !valid {
		t.Errorf("Unexpected distro: %s", distro)
	}

	t.Logf("Detected distro: %s", distro)
}

func TestFFmpegManagerGetInstallMethod(t *testing.T) {
	fm := NewFFmpegManager()

	method := fm.getInstallMethod()

	// Should return appropriate method for platform
	switch runtime.GOOS {
	case "linux":
		if method != "package_manager" {
			t.Errorf("Linux should use package_manager, got %s", method)
		}
	case "windows":
		if method != "download" {
			t.Errorf("Windows should use download, got %s", method)
		}
	default:
		if method != "unsupported" {
			t.Errorf("Other platforms should be unsupported, got %s", method)
		}
	}
}

func TestFFmpegManagerFormatInstallGuide(t *testing.T) {
	fm := NewFFmpegManager()

	guide := InstallGuide{
		Title:              "Test Guide",
		Description:        "Test description",
		Steps:              []string{"Step 1", "Step 2"},
		Command:            "sudo test command",
		CommandDescription: "Run this command",
		Tips:               []string{"Tip 1", "Tip 2"},
	}

	formatted := fm.formatInstallGuide(guide)

	// Should include all parts
	if !strings.Contains(formatted, guide.Title) {
		t.Error("Formatted guide should include title")
	}
	if !strings.Contains(formatted, guide.Description) {
		t.Error("Formatted guide should include description")
	}
	if !strings.Contains(formatted, guide.Command) {
		t.Error("Formatted guide should include command")
	}

	t.Logf("Formatted guide length: %d chars", len(formatted))
}
