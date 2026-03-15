package ytdl

import (
	"context"
	"encoding/json"
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

// Install ensures yt-dlp is installed via go-ytdlp auto-install
func (c *Client) Install(ctx context.Context) error {
	log.Println("Ensuring yt-dlp is available...")
	
	// go-ytdlp will auto-download yt-dlp if not found in PATH
	_, err := ytdlp.Install(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to install yt-dlp: %w", err)
	}

	log.Println("yt-dlp is ready")
	return nil
}

// VideoInfo represents video metadata
type VideoInfo struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Channel     string       `json:"channel"`
	ChannelID   string       `json:"channel_id"`
	Duration    int          `json:"duration"`
	Description string       `json:"description"`
	Thumbnail   string       `json:"thumbnail"`
	Formats     []FormatInfo `json:"formats"`
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
	Percent float64 `json:"percent"`
	Speed   string  `json:"speed"`
	ETA     string  `json:"eta"`
	Size    string  `json:"size"`
	Status  string  `json:"status"`
}

// rawVideoInfo is the yt-dlp JSON output structure
type rawVideoInfo struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Channel     string          `json:"channel"`
	ChannelID   string          `json:"channel_id"`
	Uploader    string          `json:"uploader"`
	UploaderID  string          `json:"uploader_id"`
	Duration    float64         `json:"duration"`
	Description string          `json:"description"`
	Thumbnail   string          `json:"thumbnail"`
	Formats     []rawFormatInfo `json:"formats"`
	Thumbnails  []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"thumbnails"`
}

type rawFormatInfo struct {
	FormatID   string  `json:"format_id"`
	Ext        string  `json:"ext"`
	Resolution string  `json:"resolution"`
	FPS        float64 `json:"fps"`
	VCodec     string  `json:"vcodec"`
	ACodec     string  `json:"acodec"`
	FileSize   int64   `json:"filesize"`
	FileSizeApprox int64 `json:"filesize_approx"`
	Quality    float64 `json:"quality"`
}

// GetInfo extracts video information from URL
func (c *Client) GetInfo(ctx context.Context, url string) (*VideoInfo, error) {
	result, err := c.dl.
		NoWarnings().
		Quiet().
		DumpSingleJSON().
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
func parseVideoInfo(data string) (*VideoInfo, error) {
	var raw rawVideoInfo
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse video info: %w", err)
	}

	// Get best thumbnail
	thumbnail := raw.Thumbnail
	if len(raw.Thumbnails) > 0 {
		// Find highest resolution thumbnail
		bestThumb := raw.Thumbnails[0]
		bestArea := bestThumb.Width * bestThumb.Height
		for _, t := range raw.Thumbnails[1:] {
			area := t.Width * t.Height
			if area > bestArea {
				bestThumb = t
				bestArea = area
			}
		}
		thumbnail = bestThumb.URL
	}

	// Parse formats
	formats := make([]FormatInfo, 0, len(raw.Formats))
	for _, f := range raw.Formats {
		size := f.FileSize
		if size == 0 {
			size = f.FileSizeApprox
		}
		formats = append(formats, FormatInfo{
			FormatID:   f.FormatID,
			Ext:        f.Ext,
			Resolution: f.Resolution,
			FPS:        f.FPS,
			VCodec:     f.VCodec,
			ACodec:     f.ACodec,
			FileSize:   size,
			Quality:    fmt.Sprintf("%.0f", f.Quality),
		})
	}

	channel := raw.Channel
	if channel == "" {
		channel = raw.Uploader
	}
	channelID := raw.ChannelID
	if channelID == "" {
		channelID = raw.UploaderID
	}

	return &VideoInfo{
		ID:          raw.ID,
		Title:       raw.Title,
		Channel:     channel,
		ChannelID:   channelID,
		Duration:    int(raw.Duration),
		Description: raw.Description,
		Thumbnail:   thumbnail,
		Formats:     formats,
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
	log.Printf("[YTDLP] Starting download for URL: %s", url)
	log.Printf("[YTDLP] Output directory: %s", opts.OutputDir)
	log.Printf("[YTDLP] Format: %s, Quality: %s", opts.Format, opts.Quality)

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputTemplate := filepath.Join(opts.OutputDir, c.config.FilenameTemplate)
	log.Printf("[YTDLP] Output template: %s", outputTemplate)
	log.Printf("[YTDLP] Config template: %s", c.config.FilenameTemplate)
	log.Printf("[YTDLP] Output dir: %s", opts.OutputDir)

	// Create a FRESH command for each download (don't reuse c.dl)
	dl := ytdlp.New().
		Output(outputTemplate).
		NoWarnings().
		NoOverwrites().
		YesPlaylist().
		Continue().
		TrimFileNames(100)

	// Apply format selection - ALWAYS set a format
	if opts.Quality == "audio" {
		log.Println("[YTDLP] Using audio-only format (mp3)")
		dl = dl.ExtractAudio().AudioFormat("mp3")
	} else if opts.Format != "" && opts.Format != "best" {
		// Use specific format if provided
		log.Printf("[YTDLP] Using specific format: %s", opts.Format)
		dl = dl.Format(opts.Format)
	} else {
		// Default to best quality video+audio
		log.Println("[YTDLP] Using default format: best")
		dl = dl.Format("best")
	}

	// Apply proxy if configured
	if opts.ProxyURL != nil && *opts.ProxyURL != "" {
		log.Printf("[YTDLP] Using proxy: %s", *opts.ProxyURL)
		dl = dl.Proxy(*opts.ProxyURL)
	}

	// Apply global proxy if set
	if c.config.ProxyURL != nil && *c.config.ProxyURL != "" {
		log.Printf("[YTDLP] Using global proxy: %s", *c.config.ProxyURL)
		dl = dl.Proxy(*c.config.ProxyURL)
	}

	// Add progress parsing
	if callback != nil {
		dl = dl.ProgressFunc(100*time.Millisecond, func(update ytdlp.ProgressUpdate) {
			var percent float64
			if update.TotalBytes > 0 {
				percent = float64(update.DownloadedBytes) / float64(update.TotalBytes) * 100
			}
			
			// Calculate speed and ETA
			var speed, eta string
			if !update.Started.IsZero() {
				elapsed := time.Since(update.Started).Seconds()
				if elapsed > 0 && update.DownloadedBytes > 0 {
					bytesPerSec := float64(update.DownloadedBytes) / elapsed
					speed = FormatFileSize(int64(bytesPerSec)) + "/s"
					
					if update.TotalBytes > 0 && bytesPerSec > 0 {
						remainingBytes := update.TotalBytes - update.DownloadedBytes
						remainingSecs := float64(remainingBytes) / bytesPerSec
						eta = time.Duration(remainingSecs * float64(time.Second)).String()
					}
				}
			}
			
			progress := DownloadProgress{
				Percent: percent,
				Status:  string(update.Status),
				Speed:   speed,
				ETA:     eta,
				Size:    FormatFileSize(int64(update.TotalBytes)),
			}
			log.Printf("[YTDLP] Progress: %.1f%% (status: %s, speed: %s, eta: %s)", percent, update.Status, speed, eta)
			callback(progress)
		})
	}

	// Add speed limit if configured
	if c.config.SpeedLimitKbps != nil && *c.config.SpeedLimitKbps > 0 {
		limitStr := fmt.Sprintf("%dk", *c.config.SpeedLimitKbps)
		log.Printf("[YTDLP] Using speed limit: %s", limitStr)
		dl = dl.LimitRate(limitStr)
	}

	// Run download
	log.Println("[YTDLP] Executing yt-dlp...")
	result, err := dl.Run(ctx, url)
	if err != nil {
		log.Printf("[YTDLP] Download failed: %v", err)
		return fmt.Errorf("download failed: %w", err)
	}

	log.Printf("[YTDLP] Download completed successfully")
	log.Printf("[YTDLP] Output: %s", result.Stdout)
	if result.Stderr != "" {
		log.Printf("[YTDLP] Stderr: %s", result.Stderr)
	}

	return nil
}

// GetFormats returns available formats for a video
func (c *Client) GetFormats(ctx context.Context, url string) ([]FormatInfo, error) {
	result, err := c.dl.
		NoWarnings().
		Quiet().
		ListFormats().
		Run(ctx, url)

	if err != nil {
		return nil, fmt.Errorf("failed to get formats: %w", err)
	}

	// Parse formats from output
	formats := parseFormatsList(result.Stdout)
	return formats, nil
}

// parseFormatsList parses format list output
func parseFormatsList(data string) []FormatInfo {
	// Format list is printed to stderr in list format
	// For now, return empty - full implementation would parse the table
	_ = data
	return []FormatInfo{}
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
