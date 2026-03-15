package ytdl

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lrstanley/go-ytdlp"
)

// Client wraps the go-ytdlp client
type Client struct {
	dl     *ytdlp.Command
	config *ClientConfig
}

// ClientConfig contains configuration for the ytdl client
type ClientConfig struct {
	DownloadPath     string
	FilenameTemplate string
	ProxyURL         *string
	SpeedLimitKbps   *int
}

// NewClient creates a new ytdl client
func NewClient(config *ClientConfig) *Client {
	return &Client{
		dl:     ytdlp.New(),
		config: config,
	}
}

// Install ensures yt-dlp is installed
func (c *Client) Install(ctx context.Context) error {
	log.Println("Checking yt-dlp installation...")
	
	// Try to find yt-dlp in PATH first
	if _, err := execLookPath("yt-dlp"); err == nil {
		log.Println("yt-dlp found in PATH")
		return nil
	}

	// Auto-install yt-dlp
	log.Println("yt-dlp not found, installing...")
	if err := ytdlp.Install(ctx, nil); err != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", err)
	}

	log.Println("yt-dlp installed successfully")
	return nil
}

// execLookPath is a helper to check if a command exists
func execLookPath(file string) (string, error) {
	// Simple check for common paths
	paths := []string{
		"/usr/local/bin/yt-dlp",
		"/usr/bin/yt-dlp",
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "yt-dlp"),
		filepath.Join(os.Getenv("HOME"), "go", "bin", "yt-dlp"),
	}
	
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	
	return "", fmt.Errorf("not found")
}

// VideoInfo represents video metadata
type VideoInfo struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Channel     string        `json:"channel"`
	ChannelID   string        `json:"channel_id"`
	Duration    int           `json:"duration"`
	Description string        `json:"description"`
	Thumbnail   string        `json:"thumbnail"`
	Formats     []FormatInfo  `json:"formats"`
}

// FormatInfo represents a video format
type FormatInfo struct {
	FormatID   string  `json:"format_id"`
	Ext        string  `json:"ext"`
	Resolution string  `json:"resolution"`
	FPS        float64 `json:"fps"`
	VCodec     string  `json:"vcodec"`
	ACodec     string  `json:"acodec"`
	FileSize   int64   `json:"filesize"`
	Quality    string  `json:"quality"`
}

// ProgressCallback is called during download progress
type ProgressCallback func(progress DownloadProgress)

// DownloadProgress represents download progress
type DownloadProgress struct {
	Percent    float64 `json:"percent"`
	Speed      string  `json:"speed"`
	ETA        string  `json:"eta"`
	Size       string  `json:"size"`
	Status     string  `json:"status"`
}

// GetInfo extracts video information from URL
func (c *Client) GetInfo(ctx context.Context, url string) (*VideoInfo, error) {
	result, err := c.dl.
		NoWarnings().
		Quiet().
		JSON().
		ExtractAudio(false).
		Run(ctx, url)

	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// Parse JSON output
	info, err := parseVideoInfo(result.Stdout)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// parseVideoInfo parses yt-dlp JSON output
func parseVideoInfo(data []byte) (*VideoInfo, error) {
	// Use yt-dlp's built-in JSON parsing
	var raw map[string]interface{}
	
	// For now, return a minimal implementation
	// Full implementation would parse the yt-dlp JSON output
	_ = raw
	
	return &VideoInfo{
		ID:          "",
		Title:       "",
		Channel:     "",
		ChannelID:   "",
		Duration:    0,
		Description: "",
		Thumbnail:   "",
		Formats:     []FormatInfo{},
	}, nil
}

// DownloadOptions contains download configuration
type DownloadOptions struct {
	Format    string
	Quality   string
	OutputDir string
	ProxyURL  *string
}

// Download downloads a video
func (c *Client) Download(ctx context.Context, url string, opts DownloadOptions, callback ProgressCallback) error {
	outputTemplate := filepath.Join(opts.OutputDir, c.config.FilenameTemplate)

	dl := c.dl.
		Output(outputTemplate).
		NoWarnings()

	// Apply format selection
	if opts.Quality == "audio" {
		dl = dl.ExtractAudio(true).AudioFormat("mp3")
	} else if opts.Format != "" {
		dl = dl.Format(opts.Format)
	}

	// Apply proxy if configured
	if opts.ProxyURL != nil && *opts.ProxyURL != "" {
		dl = dl.Proxy(*opts.ProxyURL)
	}

	// Apply global proxy if set
	if c.config.ProxyURL != nil && *c.config.ProxyURL != "" {
		dl = dl.Proxy(*c.config.ProxyURL)
	}

	// Add progress parsing
	if callback != nil {
		dl = dl.ProgressFunc(func(line ytdlp.ProgressLine) {
			progress := DownloadProgress{
				Percent: line.Percentage(),
				Speed:   line.Speed(),
				ETA:     line.ETA(),
				Status:  string(line.Status),
			}
			callback(progress)
		})
	}

	// Add speed limit if configured
	if c.config.SpeedLimitKbps != nil && *c.config.SpeedLimitKbps > 0 {
		limitStr := fmt.Sprintf("%dk", *c.config.SpeedLimitKbps)
		dl = dl.LimitRate(limitStr)
	}

	// Run download
	if _, err := dl.Run(ctx, url); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

// GetFormats returns available formats for a video
func (c *Client) GetFormats(ctx context.Context, url string) ([]FormatInfo, error) {
	result, err := c.dl.
		NoWarnings().
		Quiet().
		ListFormats(true).
		Run(ctx, url)

	if err != nil {
		return nil, fmt.Errorf("failed to get formats: %w", err)
	}

	// Parse formats from output
	formats := parseFormats(result.Stdout)
	return formats, nil
}

// parseFormats parses format list output
func parseFormats(data []byte) []FormatInfo {
	// Implementation would parse yt-dlp format list output
	_ = data
	return []FormatInfo{}
}

// Search searches for videos
func (c *Client) Search(ctx context.Context, query string, maxResults int) ([]VideoInfo, error) {
	// Use ytsearch: prefix
	searchURL := fmt.Sprintf("ytsearch%d:%s", maxResults, query)
	
	result, err := c.dl.
		NoWarnings().
		Quiet().
		JSON().
		FlatPlaylist(true).
		Run(ctx, searchURL)

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Parse results
	videos := parseSearchResults(result.Stdout)
	return videos, nil
}

// parseSearchResults parses search output
func parseSearchResults(data []byte) []VideoInfo {
	// Implementation would parse yt-dlp JSON output
	_ = data
	return []VideoInfo{}
}

// IsValidURL checks if a URL is a valid YouTube/video URL
func IsValidURL(url string) bool {
	validPrefixes := []string{
		"https://youtube.com/watch",
		"https://www.youtube.com/watch",
		"https://youtu.be/",
		"https://youtube.com/shorts/",
		"https://www.youtube.com/shorts/",
		"https://youtube.com/playlist",
		"https://www.youtube.com/playlist",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}

	return false
}

// FormatDuration formats seconds to human-readable duration
func FormatDuration(seconds int) string {
	d := time.Duration(seconds) * time.Second
	if d.Hours() >= 1 {
		return fmt.Sprintf("%d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	}
	return fmt.Sprintf("%d:%02d", int(d.Minutes()), int(d.Seconds())%60)
}

// FormatFileSize formats bytes to human-readable size
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
