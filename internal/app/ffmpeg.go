package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goRuntime "runtime"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"yted/internal/log"
)

// FFmpegLocation represents a found FFmpeg binary location
type FFmpegLocation struct {
	Path      string `json:"path"`
	Version   string `json:"version"`
	IsValid   bool   `json:"isValid"`
	Source    string `json:"source"` // "custom", "bundled", "path", "common"
}

// FFmpegCheckResult provides detailed FFmpeg status for UI
type FFmpegCheckResult struct {
	Installed      bool              `json:"installed"`
	Version        string            `json:"version"`
	Path           string            `json:"path"`
	AllLocations   []FFmpegLocation  `json:"allLocations"`   // All found locations
	SelectedIndex  int               `json:"selectedIndex"`  // Which location is selected (-1 if none)
	CanAutoInstall bool              `json:"canAutoInstall"`
	InstallMethod  string            `json:"installMethod"`  // "package_manager", "download", "manual"
	InstallCommand string            `json:"installCommand"` // OS-specific command
	InstallGuide   string            `json:"installGuide"`   // Markdown guide text
	DownloadURL    string            `json:"downloadURL"`    // Direct download link
	RequiresAdmin  bool              `json:"requiresAdmin"`
}

// InstallGuide provides OS-specific installation instructions
type InstallGuide struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Steps              []string `json:"steps"`
	Command            string   `json:"command"`
	CommandDescription string   `json:"commandDescription"`
	AlternativeURL     string   `json:"alternativeURL"`
	Tips               []string `json:"tips"`
}

// FFmpegManager handles ffmpeg binary detection and management
type FFmpegManager struct {
	binPath    string
	customPath string
	logger     *log.Logger
	ctx        context.Context
}

// NewFFmpegManager creates a new ffmpeg manager
func NewFFmpegManager() *FFmpegManager {
	return &FFmpegManager{
		logger: log.GetLogger(),
	}
}

// SetCustomPath sets a custom ffmpeg path from user configuration
func (f *FFmpegManager) SetCustomPath(path string) {
	f.customPath = path
	f.binPath = "" // Reset cached path so Find() will check new path
	if path != "" {
		f.logger.Info("FFmpeg", "Custom ffmpeg path set", map[string]string{"path": path})
	}
}

// validatePath checks if a path is a valid ffmpeg binary and returns version info
func (f *FFmpegManager) validatePath(path string) (bool, string) {
	if path == "" {
		return false, ""
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false, ""
	}
	// Actually run ffmpeg -version to verify it works
	cmd := exec.Command(path, "-version")
	output, err := cmd.Output()
	if err != nil {
		return false, ""
	}
	// Verify output contains "ffmpeg version"
	outputStr := string(output)
	if !strings.Contains(outputStr, "ffmpeg version") {
		return false, ""
	}
	// Extract version from first line
	lines := strings.Split(outputStr, "\n")
	if len(lines) > 0 {
		return true, lines[0]
	}
	return true, ""
}

// ScanAllLocations scans all possible ffmpeg locations and returns found ones with version info
func (f *FFmpegManager) ScanAllLocations() []FFmpegLocation {
	var locations []FFmpegLocation
	seen := make(map[string]bool)

	f.logger.Info("FFmpeg", "Scanning for FFmpeg binaries...", nil)

	// 1. Check custom path
	if f.customPath != "" {
		f.logger.Info("FFmpeg", "Checking custom path", map[string]string{"path": f.customPath})
		if valid, version := f.validatePath(f.customPath); valid {
			extractedVersion := f.extractVersion(version)
			locations = append(locations, FFmpegLocation{
				Path:    f.customPath,
				Version: extractedVersion,
				IsValid: true,
				Source:  "custom",
			})
			seen[f.customPath] = true
			f.logger.Info("FFmpeg", "Found valid FFmpeg at custom path", map[string]string{
				"path":    f.customPath,
				"version": extractedVersion,
			})
		} else {
			f.logger.Warn("FFmpeg", "Custom path not valid", map[string]string{"path": f.customPath})
		}
	}

	// 2. Check bundled paths
	bundledPaths := f.getBundledPaths()
	f.logger.Info("FFmpeg", "Checking bundled paths", map[string]int{"count": len(bundledPaths)})
	for _, path := range bundledPaths {
		if seen[path] {
			continue
		}
		if valid, version := f.validatePath(path); valid {
			extractedVersion := f.extractVersion(version)
			locations = append(locations, FFmpegLocation{
				Path:    path,
				Version: extractedVersion,
				IsValid: true,
				Source:  "bundled",
			})
			seen[path] = true
			f.logger.Info("FFmpeg", "Found valid bundled FFmpeg", map[string]string{
				"path":    path,
				"version": extractedVersion,
			})
		}
	}

	// 3. Check PATH
	f.logger.Info("FFmpeg", "Checking system PATH", nil)
	if path, err := exec.LookPath("ffmpeg"); err == nil && !seen[path] {
		if valid, version := f.validatePath(path); valid {
			extractedVersion := f.extractVersion(version)
			locations = append(locations, FFmpegLocation{
				Path:    path,
				Version: extractedVersion,
				IsValid: true,
				Source:  "path",
			})
			seen[path] = true
			f.logger.Info("FFmpeg", "Found valid FFmpeg in PATH", map[string]string{
				"path":    path,
				"version": extractedVersion,
			})
		}
	} else if err != nil {
		f.logger.Info("FFmpeg", "FFmpeg not found in PATH", map[string]string{"error": err.Error()})
	}

	// 4. Check common locations
	commonPaths := f.getCommonPaths()
	f.logger.Info("FFmpeg", "Checking common installation paths", map[string]int{"count": len(commonPaths)})
	for _, path := range commonPaths {
		if seen[path] {
			continue
		}
		if valid, version := f.validatePath(path); valid {
			extractedVersion := f.extractVersion(version)
			locations = append(locations, FFmpegLocation{
				Path:    path,
				Version: extractedVersion,
				IsValid: true,
				Source:  "common",
			})
			seen[path] = true
			f.logger.Info("FFmpeg", "Found valid FFmpeg in common location", map[string]string{
				"path":    path,
				"version": extractedVersion,
			})
		}
	}

	if len(locations) == 0 {
		f.logger.Warn("FFmpeg", "No FFmpeg binaries found", map[string]interface{}{
			"customPath":   f.customPath,
			"bundledChecked": len(bundledPaths),
			"commonChecked": len(commonPaths),
		})
	} else {
		f.logger.Info("FFmpeg", "FFmpeg scan complete", map[string]interface{}{
			"found":        len(locations),
			"selectedPath": f.binPath,
		})
	}

	return locations
}

// extractVersion extracts just the version number from ffmpeg -version output
func (f *FFmpegManager) extractVersion(versionLine string) string {
	// Input: "ffmpeg version 6.1.1-3ubuntu5 Copyright (c) 2000-2023..."
	// Output: "6.1.1-3ubuntu5"
	if strings.HasPrefix(versionLine, "ffmpeg version ") {
		parts := strings.SplitN(versionLine, " ", 3)
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return versionLine
}

// Find searches for ffmpeg in custom path, PATH and common locations
// Validates the binary by running 'ffmpeg -version' to ensure it works
func (f *FFmpegManager) Find() string {
	// Check if we already found and validated it
	if f.binPath != "" {
		return f.binPath
	}

	// First try custom path if set
	if f.customPath != "" {
		if valid, _ := f.validatePath(f.customPath); valid {
			f.binPath = f.customPath
			f.logger.Info("FFmpeg", "Found and validated ffmpeg at custom path", map[string]string{"path": f.customPath})
			return f.customPath
		}
		f.logger.Warn("FFmpeg", "Custom ffmpeg path not valid, falling back to auto-detect", map[string]string{"path": f.customPath})
	}

	// Try bundled ffmpeg shipped with the app package
	for _, path := range f.getBundledPaths() {
		if valid, _ := f.validatePath(path); valid {
			f.binPath = path
			f.logger.Info("FFmpeg", "Found bundled ffmpeg", map[string]string{"path": path})
			return path
		}
	}

	// Try PATH
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		if valid, _ := f.validatePath(path); valid {
			f.binPath = path
			f.logger.Info("FFmpeg", "Found and validated ffmpeg in PATH", map[string]string{"path": path})
			return path
		}
	}

	// Check common locations
	commonPaths := f.getCommonPaths()
	for _, path := range commonPaths {
		if valid, _ := f.validatePath(path); valid {
			f.binPath = path
			f.logger.Info("FFmpeg", "Found and validated ffmpeg in common location", map[string]string{"path": path})
			return path
		}
	}

	f.logger.Warn("FFmpeg", "FFmpeg not found or not working - video/audio may not merge properly", nil)
	return ""
}

// getCommonPaths returns common ffmpeg installation paths
func (f *FFmpegManager) getCommonPaths() []string {
	switch goRuntime.GOOS {
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

func (f *FFmpegManager) getBundledPaths() []string {
	executableName := "ffmpeg"
	if goRuntime.GOOS == "windows" {
		executableName = "ffmpeg.exe"
	}

	executablePath, err := os.Executable()
	if err != nil {
		return nil
	}

	executableDir := filepath.Dir(executablePath)
	candidates := []string{
		filepath.Join(executableDir, executableName),
		filepath.Join(executableDir, "ffmpeg", executableName),
		filepath.Join(executableDir, "bin", executableName),
		filepath.Join(executableDir, "resources", executableName),
		filepath.Join(executableDir, "resources", "ffmpeg", executableName),
	}

	return candidates
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
	switch goRuntime.GOOS {
	case "linux":
		return "Install ffmpeg using: sudo apt install ffmpeg (Debian/Ubuntu) or sudo dnf install ffmpeg (Fedora)"
	case "windows":
		return "Download ffmpeg from https://ffmpeg.org/download.html and add to PATH"
	default:
		return "Unsupported OS. YTed supports Linux and Windows only."
	}
}

// SetContext sets the Wails context for opening URLs
func (f *FFmpegManager) SetContext(ctx context.Context) {
	f.ctx = ctx
}

// CheckFFmpegWithGuidance returns detailed FFmpeg status with install guidance
// Scans all common locations and returns found binaries with version info
func (f *FFmpegManager) CheckFFmpegWithGuidance() FFmpegCheckResult {
	// Scan all locations first
	allLocations := f.ScanAllLocations()

	// Find the currently selected path
	selectedPath := f.Find()
	selectedIndex := -1
	var selectedVersion string

	for i, loc := range allLocations {
		if loc.Path == selectedPath {
			selectedIndex = i
			selectedVersion = loc.Version
			break
		}
	}

	result := FFmpegCheckResult{
		Installed:     selectedPath != "",
		Version:       selectedVersion,
		Path:          selectedPath,
		AllLocations:  allLocations,
		SelectedIndex: selectedIndex,
	}

	if !result.Installed {
		guide := f.GetInstallGuide()
		result.InstallMethod = f.getInstallMethod()
		result.InstallCommand = guide.Command
		result.InstallGuide = f.formatInstallGuide(guide)
		result.DownloadURL = guide.AlternativeURL
		result.RequiresAdmin = goRuntime.GOOS == "linux" || goRuntime.GOOS == "windows"
		result.CanAutoInstall = false
	}

	return result
}

// GetInstallGuide returns OS-specific installation instructions
func (f *FFmpegManager) GetInstallGuide() InstallGuide {
	switch goRuntime.GOOS {
	case "linux":
		return f.getLinuxInstallGuide()
	case "windows":
		return f.getWindowsInstallGuide()
	default:
		return InstallGuide{
			Title:              "Unsupported Operating System",
			Description:        "YTed supports Linux and Windows only.",
			Steps:              []string{"Use YTed on a supported Linux or Windows environment."},
			Command:            "",
			AlternativeURL:     "",
			CommandDescription: "No install command available",
			Tips:               []string{"FFmpeg guidance is only provided for Linux and Windows."},
		}
	}
}

func (f *FFmpegManager) getLinuxInstallGuide() InstallGuide {
	distro := f.detectLinuxDistro()

	switch distro {
	case "ubuntu", "debian":
		return InstallGuide{
			Title:              "Install FFmpeg on Ubuntu/Debian",
			Description:        "Install FFmpeg using apt package manager",
			Steps:              []string{"Open Terminal", "Run: sudo apt update", "Run the command below", "Restart YTed"},
			Command:            "sudo apt install ffmpeg",
			CommandDescription: "Copy and paste this command in Terminal",
			AlternativeURL:     "https://ffmpeg.org/download.html#build-linux",
			Tips: []string{
				"You may need to enter your password",
				"For older Ubuntu versions, you may need to add a PPA",
			},
		}
	case "fedora":
		return InstallGuide{
			Title:              "Install FFmpeg on Fedora",
			Description:        "Install FFmpeg using dnf package manager",
			Steps:              []string{"Open Terminal", "Run the command below", "Restart YTed"},
			Command:            "sudo dnf install ffmpeg",
			CommandDescription: "Copy and paste this command in Terminal",
			AlternativeURL:     "https://ffmpeg.org/download.html#build-linux",
			Tips: []string{
				"You may need to enter your password",
				"Enable RPM Fusion if ffmpeg is not found: sudo dnf install https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-$(rpm -E %fedora).noarch.rpm",
			},
		}
	case "arch":
		return InstallGuide{
			Title:              "Install FFmpeg on Arch Linux",
			Description:        "Install FFmpeg using pacman package manager",
			Steps:              []string{"Open Terminal", "Run the command below", "Restart YTed"},
			Command:            "sudo pacman -S ffmpeg",
			CommandDescription: "Copy and paste this command in Terminal",
			AlternativeURL:     "https://ffmpeg.org/download.html#build-linux",
			Tips: []string{
				"You may need to enter your password",
			},
		}
	default:
		return InstallGuide{
			Title:              "Install FFmpeg on Linux",
			Description:        "Install FFmpeg using your distribution's package manager",
			Steps:              []string{"Open Terminal", "Run the appropriate command for your distro", "Restart YTed"},
			Command:            "sudo apt install ffmpeg  # or: sudo dnf install ffmpeg  # or: sudo pacman -S ffmpeg",
			CommandDescription: "Use the command for your distribution",
			AlternativeURL:     "https://ffmpeg.org/download.html#build-linux",
			Tips: []string{
				"Ubuntu/Debian: sudo apt install ffmpeg",
				"Fedora: sudo dnf install ffmpeg",
				"Arch: sudo pacman -S ffmpeg",
			},
		}
	}
}

func (f *FFmpegManager) getWindowsInstallGuide() InstallGuide {
	return InstallGuide{
		Title:       "Install FFmpeg on Windows",
		Description: "Download and add FFmpeg to your system PATH",
		Steps: []string{
			"Download FFmpeg from the link below",
			"Extract the zip file to C:\\ffmpeg",
			"Add C:\\ffmpeg\\bin to your system PATH",
			"Restart YTed",
		},
		Command:            "",
		CommandDescription: "Download from the link and follow the included README",
		AlternativeURL:     "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip",
		Tips: []string{
			"Choose the 'essentials' build for smaller download",
			"You can also use winget: winget install Gyan.FFmpeg",
			"After adding to PATH, restart your command prompt",
			"To add to PATH: Settings > System > About > Advanced system settings > Environment Variables",
		},
	}
}

func (f *FFmpegManager) detectLinuxDistro() string {
	// Try to detect the Linux distribution
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}
	if _, err := os.Stat("/etc/ubuntu-release"); err == nil {
		return "ubuntu"
	}
	if _, err := os.Stat("/etc/fedora-release"); err == nil {
		return "fedora"
	}
	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return "arch"
	}
	return "unknown"
}

func (f *FFmpegManager) getInstallMethod() string {
	switch goRuntime.GOOS {
	case "linux":
		return "package_manager"
	case "windows":
		return "download"
	default:
		return "unsupported"
	}
}

func (f *FFmpegManager) formatInstallGuide(guide InstallGuide) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", guide.Title)
	fmt.Fprintf(&sb, "%s\n\n", guide.Description)
	sb.WriteString("## Steps:\n")
	for i, step := range guide.Steps {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, step)
	}
	if guide.Command != "" {
		fmt.Fprintf(&sb, "\n**Command:** `%s`\n", guide.Command)
	}
	if len(guide.Tips) > 0 {
		sb.WriteString("\n## Tips:\n")
		for _, tip := range guide.Tips {
			fmt.Fprintf(&sb, "- %s\n", tip)
		}
	}
	return sb.String()
}

// OpenDownloadPage opens the FFmpeg download page in browser
func (f *FFmpegManager) OpenDownloadPage() {
	if f.ctx != nil {
		runtime.BrowserOpenURL(f.ctx, "https://ffmpeg.org/download.html")
	}
}
