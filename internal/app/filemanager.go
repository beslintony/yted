package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"yted/internal/config"
	applog "yted/internal/log"
)

const YTED_FOLDER_NAME = "YTed"

// FileWarning represents a warning about file operations
type FileWarning struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	IsCritical  bool   `json:"is_critical"`
}

// FileManager handles download folder management and safety checks
type FileManager struct {
	config *config.Manager
}

// NewFileManager creates a new file manager
func NewFileManager(cfg *config.Manager) *FileManager {
	return &FileManager{
		config: cfg,
	}
}

// SetDownloadPath sets the user's selected path and creates YTed subfolder
// Returns warnings if there are any issues with the path
func (fm *FileManager) SetDownloadPath(userPath string) (string, []FileWarning, error) {
	logger := applog.GetLogger()
	var warnings []FileWarning

	// Validate the path
	if userPath == "" {
		return "", nil, fmt.Errorf("path cannot be empty")
	}

	// Check if path exists
	info, err := os.Stat(userPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("path does not exist: %s", userPath)
		}
		return "", nil, fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return "", nil, fmt.Errorf("path is not a directory: %s", userPath)
	}

	// Check if path is writable
	testFile := filepath.Join(userPath, ".write_test")
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		os.Remove(testFile)
	} else {
		warnings = append(warnings, FileWarning{
			Type:       "permission",
			Title:      "Directory Not Writable",
			Message:    "The selected directory may not be writable. Downloads may fail.",
			IsCritical: true,
		})
	}

	// Check available disk space
	var stat os.FileInfo
	stat, err = os.Stat(userPath)
	_ = stat // We already have info from earlier

	// Check for common system directories (safety warning)
	systemDirs := []string{
		"/", "C:", "C:\\", "/root", "/bin", "/sbin", "/usr", "/etc",
		"/lib", "/lib64", "/boot", "/dev", "/proc", "/sys",
	}
	for _, sysDir := range systemDirs {
		if strings.EqualFold(userPath, sysDir) || strings.HasPrefix(userPath, sysDir+string(filepath.Separator)) {
			warnings = append(warnings, FileWarning{
				Type:       "system",
				Title:      "System Directory Warning",
				Message:    fmt.Sprintf("You have selected a system directory (%s). This is not recommended and may cause issues.", userPath),
				IsCritical: true,
			})
			break
		}
	}

	// Check for home directory (informational)
	homeDir, _ := os.UserHomeDir()
	if userPath == homeDir {
		warnings = append(warnings, FileWarning{
			Type:       "location",
			Title:      "Home Directory Selected",
			Message:    "Downloads will be saved to " + filepath.Join(userPath, YTED_FOLDER_NAME) + ". This will keep your home directory organized.",
			IsCritical: false,
		})
	}

	// Store the user's selected path
	fm.config.Update(func(cfg *config.Config) {
		cfg.UserSelectedPath = userPath
		// The actual download path is always the YTed subfolder
		cfg.DownloadPath = filepath.Join(userPath, YTED_FOLDER_NAME)
	})

	// Create the YTed subfolder
	ytedPath := filepath.Join(userPath, YTED_FOLDER_NAME)
	if err := os.MkdirAll(ytedPath, 0755); err != nil {
		return "", warnings, fmt.Errorf("failed to create YTed folder: %w", err)
	}

	logger.Info("FileManager", "Download path set", map[string]string{
		"user_path":    userPath,
		"yted_path":    ytedPath,
		"warning_count": fmt.Sprintf("%d", len(warnings)),
	})

	return ytedPath, warnings, nil
}

// GetDownloadPath returns the actual download path (YTed subfolder)
func (fm *FileManager) GetDownloadPath() string {
	cfg := fm.config.Get()
	if cfg.DownloadPath == "" && cfg.UserSelectedPath != "" {
		// Migrate old config
		return filepath.Join(cfg.UserSelectedPath, YTED_FOLDER_NAME)
	}
	return cfg.DownloadPath
}

// GetUserSelectedPath returns the original path selected by the user
func (fm *FileManager) GetUserSelectedPath() string {
	cfg := fm.config.Get()
	return cfg.UserSelectedPath
}

// IsManagedFile checks if a file is within our managed YTed folder
func (fm *FileManager) IsManagedFile(filePath string) bool {
	downloadPath := fm.GetDownloadPath()
	if downloadPath == "" {
		return false
	}

	// Normalize paths for comparison
	filePath = filepath.Clean(filePath)
	downloadPath = filepath.Clean(downloadPath)

	// Check if file is within the YTed folder
	return strings.HasPrefix(filePath, downloadPath+string(filepath.Separator)) ||
		filePath == downloadPath
}

// EnsureYTedFolder creates the YTed subfolder if it doesn't exist
func (fm *FileManager) EnsureYTedFolder() error {
	downloadPath := fm.GetDownloadPath()
	if downloadPath == "" {
		return fmt.Errorf("download path not configured")
	}

	return os.MkdirAll(downloadPath, 0755)
}

// DeleteManagedFile deletes a file only if it's within the managed YTed folder
// Returns detailed error information for user-friendly messages
func (fm *FileManager) DeleteManagedFile(filePath string) error {
	logger := applog.GetLogger()

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Security check - only delete files within our managed folder
	if !fm.IsManagedFile(filePath) {
		logger.Warn("FileManager", "Attempted to delete file outside managed folder", map[string]string{
			"path":         filePath,
			"managed_path": fm.GetDownloadPath(),
		})
		return fmt.Errorf("SECURITY: Cannot delete file outside the YTed managed folder. For your safety, please delete this file manually: %s", filePath)
	}

	// Attempt to delete
	if err := os.Remove(filePath); err != nil {
		logger.Error("FileManager", "Failed to delete file", err, map[string]string{
			"path": filePath,
		})
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info("FileManager", "File deleted successfully", map[string]string{
		"path": filePath,
	})

	return nil
}

// SafeDeleteWithWarning attempts to delete a file with comprehensive checks and warnings
func (fm *FileManager) SafeDeleteWithWarning(filePath string) (*FileWarning, error) {
	logger := applog.GetLogger()

	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	// Check if managed
	if !fm.IsManagedFile(filePath) {
		warning := &FileWarning{
			Type:       "security",
			Title:      "External File",
			Message:    fmt.Sprintf("This file is outside the YTed managed folder (%s). For your safety, it must be deleted manually.", fm.GetDownloadPath()),
			IsCritical: true,
		}
		logger.Warn("FileManager", "Blocked deletion of external file", map[string]string{
			"path": filePath,
		})
		return warning, fmt.Errorf("file is not managed by YTed")
	}

	// Check file size for large files
	if info.Size() > 1024*1024*1024 { // > 1GB
		warning := &FileWarning{
			Type:       "size",
			Title:      "Large File",
			Message:    fmt.Sprintf("This file is %s. It will be permanently deleted.", formatFileSize(info.Size())),
			IsCritical: false,
		}
		// Don't return error, just warning - actual deletion happens elsewhere
		return warning, nil
	}

	return nil, nil
}

// FileExists checks if a file exists
func (fm *FileManager) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileInfo returns file info for a managed file
func (fm *FileManager) GetFileInfo(filePath string) (os.FileInfo, error) {
	if !fm.IsManagedFile(filePath) {
		return nil, fmt.Errorf("file is not within the managed YTed folder")
	}

	return os.Stat(filePath)
}

// GetFolderSize calculates the total size of the YTed folder
func (fm *FileManager) GetFolderSize() (int64, error) {
	downloadPath := fm.GetDownloadPath()
	if downloadPath == "" {
		return 0, fmt.Errorf("download path not configured")
	}

	var totalSize int64
	err := filepath.Walk(downloadPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// SanitizeFilename removes characters that could be problematic in filenames
func SanitizeFilename(name string) string {
	// Replace problematic characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "'",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(name)
}

// formatFileSize formats bytes to human readable string
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
