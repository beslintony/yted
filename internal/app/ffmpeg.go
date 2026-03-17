package app

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"yted/internal/log"
)

// FFmpegManager handles ffmpeg binary detection and management
type FFmpegManager struct {
	binPath string
	logger  *log.Logger
}

// NewFFmpegManager creates a new ffmpeg manager
func NewFFmpegManager() *FFmpegManager {
	return &FFmpegManager{
		logger: log.GetLogger(),
	}
}

// Find searches for ffmpeg in PATH and common locations
func (f *FFmpegManager) Find() string {
	// Check if we already found it
	if f.binPath != "" {
		return f.binPath
	}
	
	// First try PATH
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		f.binPath = path
		f.logger.Info("FFmpeg", "Found ffmpeg in PATH", map[string]string{"path": path})
		return path
	}
	
	// Check common locations
	commonPaths := f.getCommonPaths()
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			f.binPath = path
			f.logger.Info("FFmpeg", "Found ffmpeg in common location", map[string]string{"path": path})
			return path
		}
	}
	
	f.logger.Warn("FFmpeg", "FFmpeg not found - video/audio may not merge properly", nil)
	return ""
}

// getCommonPaths returns common ffmpeg installation paths
func (f *FFmpegManager) getCommonPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/opt/homebrew/bin/ffmpeg",
			"/usr/local/bin/ffmpeg",
			"/usr/bin/ffmpeg",
		}
	case "linux":
		return []string{
			"/usr/bin/ffmpeg",
			"/usr/local/bin/ffmpeg",
			"/snap/bin/ffmpeg",
			"/app/bin/ffmpeg", // Flatpak
		}
	case "windows":
		return []string{
			`C:\ffmpeg\bin\ffmpeg.exe`,
			`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
			`C:\Program Files (x86)\ffmpeg\bin\ffmpeg.exe`,
		}
	default:
		return nil
	}
}

// IsAvailable returns true if ffmpeg is available
func (f *FFmpegManager) IsAvailable() bool {
	return f.Find() != ""
}

// GetVersion returns the ffmpeg version
func (f *FFmpegManager) GetVersion() string {
	path := f.Find()
	if path == "" {
		return ""
	}
	
	cmd := exec.Command(path, "-version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	// Parse version from first line
	lines := string(output)
	if len(lines) > 0 {
		parts := strings.Split(lines, "\n")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}

// GetPath returns the ffmpeg binary path
func (f *FFmpegManager) GetPath() string {
	return f.Find()
}

// MergeVideoAudio merges video and audio files using ffmpeg
func (f *FFmpegManager) MergeVideoAudio(videoPath, audioPath, outputPath string) error {
	path := f.Find()
	if path == "" {
		return fmt.Errorf("ffmpeg not available")
	}
	
	args := []string{
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "copy",
		"-shortest",
		"-y",
		outputPath,
	}
	
	cmd := exec.Command(path, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg merge failed: %w\nOutput: %s", err, string(output))
	}
	
	return nil
}

// InstallInstructions returns instructions for installing ffmpeg
func (f *FFmpegManager) InstallInstructions() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install ffmpeg using: brew install ffmpeg"
	case "linux":
		return "Install ffmpeg using: sudo apt install ffmpeg (Debian/Ubuntu) or sudo dnf install ffmpeg (Fedora)"
	case "windows":
		return "Download ffmpeg from https://ffmpeg.org/download.html and add to PATH"
	default:
		return "Please install ffmpeg from https://ffmpeg.org/download.html"
	}
}
