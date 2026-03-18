package app

import (
	"runtime"
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
	case "darwin":
		if !contains(t, instructions, "brew") {
			t.Error("macOS instructions should mention brew")
		}
	case "linux":
		if !contains(t, instructions, "apt") && !contains(t, instructions, "dnf") {
			t.Error("Linux instructions should mention apt or dnf")
		}
	case "windows":
		if !contains(t, instructions, "ffmpeg.org") {
			t.Error("Windows instructions should mention ffmpeg.org")
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

// Helper function
func contains(t *testing.T, s, substr string) bool {
	t.Helper()
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
