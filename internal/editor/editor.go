// Package editor provides video editing functionality using FFmpeg
package editor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"yted/internal/config"
	"yted/internal/db"
	"yted/internal/log"
)

// ProgressCallback is called during edit progress
type ProgressCallback func(progress float64, eta string)

// Editor handles video editing operations
type Editor struct {
	ffmpegPath   string
	db           *db.DB
	config       *config.Manager
	logger       *log.Logger
	queue        *EditQueue
	activeJobs   map[string]*EditJob
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// EditJob represents an active editing job
type EditJob struct {
	ID        string
	VideoID   string
	Operation string
	Settings  db.EditSettings
	Status    string
	Progress  float64
	Cancel    context.CancelFunc
}

// New creates a new Editor instance
func New(ffmpegPath string, database *db.DB, cfg *config.Manager) *Editor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Editor{
		ffmpegPath: ffmpegPath,
		db:         database,
		config:     cfg,
		logger:     log.GetLogger(),
		queue:      NewQueue(2), // 2 concurrent edit jobs
		activeJobs: make(map[string]*EditJob),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the editor worker queue
func (e *Editor) Start() {
	e.queue.Start()
}

// Stop stops the editor and cancels all active jobs
func (e *Editor) Stop() {
	e.cancel()
	e.queue.Stop()

	// Cancel all active jobs
	e.mu.Lock()
	for _, job := range e.activeJobs {
		if job.Cancel != nil {
			job.Cancel()
		}
	}
	e.mu.Unlock()
}

// SetFFmpegPath updates the FFmpeg binary path
func (e *Editor) SetFFmpegPath(path string) {
	e.ffmpegPath = path
}

// SubmitJob submits a new edit job to the queue
func (e *Editor) SubmitJob(videoID, operation string, settings db.EditSettings) (string, error) {
	jobID := uuid.New().String()

	// Get source video info
	video, err := e.db.GetVideoWithHash(videoID)
	if err != nil {
		return "", fmt.Errorf("failed to get video: %w", err)
	}
	if video == nil {
		return "", fmt.Errorf("video not found")
	}

	// Serialize settings
	settingsJSON, err := db.SettingsToJSON(settings)
	if err != nil {
		return "", fmt.Errorf("failed to serialize settings: %w", err)
	}

	// Create database record
	dbJob := &db.EditJob{
		ID:            jobID,
		SourceVideoID: videoID,
		Status:        "pending",
		Operation:     operation,
		Settings:      settingsJSON,
		Progress:      0,
		CreatedAt:     time.Now(),
	}

	if err := e.db.CreateEditJob(dbJob); err != nil {
		return "", fmt.Errorf("failed to create edit job: %w", err)
	}

	// Add to processing queue
	e.queue.Submit(&EditTask{
		JobID:     jobID,
		Video:     video,
		Operation: operation,
		Settings:  settings,
		Editor:    e,
	})

	e.logger.Info("Editor", "Edit job submitted", map[string]string{
		"job_id":    jobID,
		"video_id":  videoID,
		"operation": operation,
	})

	return jobID, nil
}

// GetJobStatus returns the status of an edit job
func (e *Editor) GetJobStatus(jobID string) (*db.EditJob, error) {
	return e.db.GetEditJob(jobID)
}

// ListJobs lists all edit jobs for a video
func (e *Editor) ListJobs(videoID string) ([]db.EditJob, error) {
	return e.db.ListEditJobs(videoID)
}

// CancelJob cancels an active edit job
func (e *Editor) CancelJob(jobID string) error {
	e.mu.Lock()
	job, exists := e.activeJobs[jobID]
	e.mu.Unlock()

	if exists && job.Cancel != nil {
		job.Cancel()
	}

	return nil
}

// GetVideoMetadata extracts metadata from a video file using ffprobe
func (e *Editor) GetVideoMetadata(filePath string) (*VideoMetadata, error) {
	ffprobePath := strings.Replace(e.ffmpegPath, "ffmpeg", "ffprobe", 1)
	if ffprobePath == e.ffmpegPath {
		// Try to find ffprobe in same directory
		ffprobePath = filepath.Join(filepath.Dir(e.ffmpegPath), "ffprobe")
		if runtime.GOOS == "windows" {
			ffprobePath += ".exe"
		}
	}

	// Check if ffprobe exists
	if _, err := os.Stat(ffprobePath); os.IsNotExist(err) {
		// Fall back to ffmpeg for basic info
		return e.getMetadataFromFFmpeg(filePath)
	}

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	cmd := exec.Command(ffprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	return parseFFProbeOutput(output)
}

// getMetadataFromFFmpeg gets basic metadata using ffmpeg
func (e *Editor) getMetadataFromFFmpeg(filePath string) (*VideoMetadata, error) {
	args := []string{
		"-i", filePath,
		"-f", "null",
		"-",
	}

	cmd := exec.Command(e.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err == nil || strings.Contains(string(output), "Input #0") {
		// Parse output for basic info
		return parseFFmpegOutput(string(output), filePath)
	}

	return nil, fmt.Errorf("ffmpeg failed: %w", err)
}

// VideoMetadata contains video file metadata
type VideoMetadata struct {
	Duration    float64 `json:"duration"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FPS         float64 `json:"fps"`
	Bitrate     int64   `json:"bitrate"`
	Codec       string  `json:"codec"`
	AudioCodec  string  `json:"audio_codec,omitempty"`
	AudioBitrate int64  `json:"audio_bitrate,omitempty"`
	HasAudio    bool    `json:"has_audio"`
}

func parseFFProbeOutput(output []byte) (*VideoMetadata, error) {
	// Simple parsing - in production, use proper JSON unmarshaling
	meta := &VideoMetadata{}
	outputStr := string(output)

	// Extract duration
	if idx := strings.Index(outputStr, "\"duration\":\""); idx != -1 {
		start := idx + len("\"duration\":\"")
		end := strings.Index(outputStr[start:], "\"")
		if end != -1 {
			duration, _ := strconv.ParseFloat(outputStr[start:start+end], 64)
			meta.Duration = duration
		}
	}

	// Extract width/height from video stream
	if idx := strings.Index(outputStr, "\"width\":"); idx != -1 {
		start := idx + len("\"width\":")
		end := strings.IndexAny(outputStr[start:], ",}")
		if end != -1 {
			meta.Width, _ = strconv.Atoi(strings.TrimSpace(outputStr[start : start+end]))
		}
	}

	if idx := strings.Index(outputStr, "\"height\":"); idx != -1 {
		start := idx + len("\"height\":")
		end := strings.IndexAny(outputStr[start:], ",}")
		if end != -1 {
			meta.Height, _ = strconv.Atoi(strings.TrimSpace(outputStr[start : start+end]))
		}
	}

	// Extract codec
	if idx := strings.Index(outputStr, "\"codec_name\":\""); idx != -1 {
		start := idx + len("\"codec_name\":\"")
		end := strings.Index(outputStr[start:], "\"")
		if end != -1 {
			meta.Codec = outputStr[start : start+end]
		}
	}

	return meta, nil
}

func parseFFmpegOutput(output, filePath string) (*VideoMetadata, error) {
	meta := &VideoMetadata{}

	// Parse duration
	if idx := strings.Index(output, "Duration: "); idx != -1 {
		start := idx + len("Duration: ")
		end := strings.Index(output[start:], ",")
		if end != -1 {
			durationStr := output[start : start+end]
			meta.Duration = parseDuration(durationStr)
		}
	}

	// Parse resolution
	if idx := strings.Index(output, "Video:"); idx != -1 {
		videoSection := output[idx:]
		if idx2 := strings.Index(videoSection, ","); idx2 != -1 {
			videoInfo := videoSection[:idx2]
			// Look for resolution pattern like 1920x1080
			parts := strings.Fields(videoInfo)
			for _, part := range parts {
				if strings.Contains(part, "x") {
					dims := strings.Split(part, "x")
					if len(dims) == 2 {
						meta.Width, _ = strconv.Atoi(dims[0])
						meta.Height, _ = strconv.Atoi(dims[1])
						break
					}
				}
			}
		}
	}

	return meta, nil
}

func parseDuration(durationStr string) float64 {
	// Format: HH:MM:SS.ms
	parts := strings.Split(durationStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return hours*3600 + minutes*60 + seconds
}

// GeneratePreview generates a preview frame for the given settings
func (e *Editor) GeneratePreview(filePath string, settings db.EditSettings, timestamp float64) (string, error) {
	if e.ffmpegPath == "" {
		return "", fmt.Errorf("ffmpeg not available")
	}

	// Create temp directory for previews
	previewDir := filepath.Join(os.TempDir(), "yted_previews")
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Generate unique filename
	previewFile := filepath.Join(previewDir, fmt.Sprintf("preview_%d_%d.jpg", time.Now().Unix(), randInt()))

	// Build filter chain
	filters := []string{}

	// Apply effects
	if settings.Brightness != nil || settings.Contrast != nil || settings.Saturation != nil {
		eq := "eq="
		params := []string{}
		if settings.Brightness != nil {
			params = append(params, fmt.Sprintf("brightness=%.2f", *settings.Brightness))
		}
		if settings.Contrast != nil {
			params = append(params, fmt.Sprintf("contrast=%.2f", *settings.Contrast))
		}
		if settings.Saturation != nil {
			params = append(params, fmt.Sprintf("saturation=%.2f", *settings.Saturation))
		}
		filters = append(filters, eq+strings.Join(params, ":"))
	}

	// Apply crop
	if settings.CropX != nil && settings.CropY != nil && settings.CropWidth != nil && settings.CropHeight != nil {
		filters = append(filters, fmt.Sprintf("crop=%d:%d:%d:%d", *settings.CropWidth, *settings.CropHeight, *settings.CropX, *settings.CropY))
	}

	// Apply rotation
	if settings.Rotation != nil && *settings.Rotation != 0 {
		transpose := "0"
		switch *settings.Rotation {
		case 90:
			transpose = "1"
		case 180:
			transpose = "2,transpose=2"
		case 270:
			transpose = "3"
		}
		if transpose != "0" {
			if strings.Contains(transpose, ",") {
				filters = append(filters, strings.Split(transpose, ",")...)
			} else {
				filters = append(filters, "transpose="+transpose)
			}
		}
	}

	// Build command
	args := []string{
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", filePath,
		"-vframes", "1",
		"-q:v", "2",
	}

	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	args = append(args, "-y", previewFile)

	cmd := exec.Command(e.ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("preview generation failed: %w", err)
	}

	return previewFile, nil
}

func randInt() int {
	return int(time.Now().UnixNano() % 10000)
}
