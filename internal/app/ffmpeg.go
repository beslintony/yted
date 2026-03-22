package app

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"yted/internal/log"
)

// FFmpegCheckResult provides detailed FFmpeg status for UI
type FFmpegCheckResult struct {
	Installed      bool   `json:"installed"`
	Version        string `json:"version"`
	Path           string `json:"path"`
	CanAutoInstall bool   `json:"canAutoInstall"`
	InstallMethod  string `json:"installMethod"`  // "package_manager", "download", "manual"
	InstallCommand string `json:"installCommand"` // OS-specific command
	InstallGuide   string `json:"installGuide"`   // Markdown guide text
	DownloadURL    string `json:"downloadURL"`    // Direct download link
	RequiresAdmin  bool   `json:"requiresAdmin"`
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
	ctx        interface{}} // Context for opening URLs}}

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

// Find searches for ffmpeg in custom path, PATH and common locations
// Validates the binary by running 'ffmpeg -version' to ensure it works
func (f *FFmpegManager) Find() string {
	// Check if we already found and validated it
	if f.binPath != "" {
		return f.binPath
	}

	// Helper to validate a path works by running ffmpeg -version
	validatePath := func(path string) bool {
		if path == "" {
			return false
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return false
		}
		// Actually run ffmpeg -version to verify it works
		cmd := exec.Command(path, "-version")
		output, err := cmd.Output()
		if err != nil {
			f.logger.Warn("FFmpeg", "Binary found but failed validation", map[string]string{
				"path":  path,
				"error": err.Error(),
			})
			return false
		}
		// Verify output contains "ffmpeg version"
		if !strings.Contains(string(output), "ffmpeg version") {
			f.logger.Warn("FFmpeg", "Binary output doesn't look like ffmpeg", map[string]string{
				"path": path,
			})
			return false
		}
		return true
	}

	// First try custom path if set
	if f.customPath != "" {
		if validatePath(f.customPath) {
			f.binPath = f.customPath
			f.logger.Info("FFmpeg", "Found and validated ffmpeg at custom path", map[string]string{"path": f.customPath})
			return f.customPath
		}
		f.logger.Warn("FFmpeg", "Custom ffmpeg path not valid, falling back to auto-detect", map[string]string{"path": f.customPath})
	}

	// Try PATH
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		if validatePath(path) {
			f.binPath = path
			f.logger.Info("FFmpeg", "Found and validated ffmpeg in PATH", map[string]string{"path": path})
			return path
		}
	}

	// Check common locations
	commonPaths := f.getCommonPaths()
	for _, path := range commonPaths {
		if validatePath(path) {
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

// SetContext sets the Wails context for opening URLs
func (f *FFmpegManager) SetContext(ctx interface{}) {
	f.ctx = ctx
}

// CheckFFmpegWithGuidance returns detailed FFmpeg status with install guidance
func (f *FFmpegManager) CheckFFmpegWithGuidance() FFmpegCheckResult {
	path := f.Find()
	version := f.GetVersion()

	result := FFmpegCheckResult{
		Installed: path != "",
		Version:   version,
		Path:      path,
	}

	if !result.Installed {
		guide := f.GetInstallGuide()
		result.InstallMethod = f.getInstallMethod()
		result.InstallCommand = guide.Command
		result.InstallGuide = f.formatInstallGuide(guide)
		result.DownloadURL = guide.AlternativeURL
		result.RequiresAdmin = runtime.GOOS != "darwin" || !f.isHomebrewAvailable()
		result.CanAutoInstall = runtime.GOOS == "darwin" && f.isHomebrewAvailable()
	}

	return result
}

// GetInstallGuide returns OS-specific installation instructions
func (f *FFmpegManager) GetInstallGuide() InstallGuide {
	switch runtime.GOOS {
	case "darwin":
		return f.getMacOSInstallGuide()
	case "linux":
		return f.getLinuxInstallGuide()
	case "windows":
		return f.getWindowsInstallGuide()
	default:
		return InstallGuide{
			Title:           "Install FFmpeg",
			Description:     "Please install FFmpeg for your operating system",
			Steps:           []string{"Visit the FFmpeg download page", "Download the appropriate version", "Follow the installation instructions"},
			Command:         "",
			AlternativeURL:  "https://ffmpeg.org/download.html",
			CommandDescription: "See website for instructions",
			Tips:            []string{"Make sure to add FFmpeg to your system PATH"},
		}
	}
}

func (f *FFmpegManager) getMacOSInstallGuide() InstallGuide {
	tips := []string{
		"Installation may take 5-10 minutes depending on your connection",
		"Homebrew will automatically add FFmpeg to your PATH",
	}

	if !f.isHomebrewAvailable() {
		tips = append([]string{"First install Homebrew from https://brew.sh"}, tips...)
	}

	return InstallGuide{
		Title:              "Install FFmpeg on macOS",
		Description:        "FFmpeg can be easily installed using Homebrew package manager",
		Steps:              []string{"Open Terminal", "Run the command below", "Wait for installation to complete", "Restart YTed"},
		Command:            "brew install ffmpeg",
		CommandDescription: "Copy and paste this command in Terminal",
		AlternativeURL:     "https://ffmpeg.org/download.html#build-mac",
		Tips:               tips,
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
		Title:              "Install FFmpeg on Windows",
		Description:        "Download and add FFmpeg to your system PATH",
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

func (f *FFmpegManager) isHomebrewAvailable() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func (f *FFmpegManager) getInstallMethod() string {
	switch runtime.GOOS {
	case "darwin":
		if f.isHomebrewAvailable() {
			return "package_manager"
		}
		return "manual"
	case "linux":
		return "package_manager"
	case "windows":
		return "download"
	default:
		return "manual"
	}
}

func (f *FFmpegManager) formatInstallGuide(guide InstallGuide) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", guide.Title))
	sb.WriteString(fmt.Sprintf("%s\n\n", guide.Description))
	sb.WriteString("## Steps:\n")
	for i, step := range guide.Steps {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
	}
	if guide.Command != "" {
		sb.WriteString(fmt.Sprintf("\n**Command:** `%s`\n", guide.Command))
	}
	if len(guide.Tips) > 0 {
		sb.WriteString("\n## Tips:\n")
		for _, tip := range guide.Tips {
			sb.WriteString(fmt.Sprintf("- %s\n", tip))
		}
	}
	return sb.String()
}

// OpenDownloadPage opens the FFmpeg download page in browser
func (f *FFmpegManager) OpenDownloadPage() {
	if f.ctx != nil {
		runtime.BrowserOpenURL(f.ctx.(interface{ Context() interface{} }).Context(), "https://ffmpeg.org/download.html")
	}
}
