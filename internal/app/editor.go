package app

import (
	"fmt"
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"yted/internal/db"
	"yted/internal/editor"
)

// VideoMetadataResult is exposed to frontend
type VideoMetadataResult struct {
	Duration     float64 `json:"duration"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	FPS          float64 `json:"fps"`
	Bitrate      int64   `json:"bitrate"`
	Codec        string  `json:"codec"`
	AudioCodec   string  `json:"audio_codec,omitempty"`
	HasAudio     bool    `json:"has_audio"`
}

// EditJobResult is exposed to frontend
type EditJobResult struct {
	ID            string   `json:"id"`
	SourceVideoID string   `json:"source_video_id"`
	OutputVideoID *string  `json:"output_video_id,omitempty"`
	Status        string   `json:"status"`
	Operation     string   `json:"operation"`
	Progress      float64  `json:"progress"`
	ErrorMessage  *string  `json:"error_message,omitempty"`
	CreatedAt     int64    `json:"created_at"`
	CompletedAt   *int64   `json:"completed_at,omitempty"`
}

// EditSettingsInput is received from frontend
type EditSettingsInput struct {
	// Crop
	CropStart  *float64 `json:"crop_start,omitempty"`
	CropEnd    *float64 `json:"crop_end,omitempty"`
	CropX      *int     `json:"crop_x,omitempty"`
	CropY      *int     `json:"crop_y,omitempty"`
	CropWidth  *int     `json:"crop_width,omitempty"`
	CropHeight *int     `json:"crop_height,omitempty"`

	// Watermark
	WatermarkType     *string  `json:"watermark_type,omitempty"`
	WatermarkText     *string  `json:"watermark_text,omitempty"`
	WatermarkImage    *string  `json:"watermark_image,omitempty"`
	WatermarkPosition *string  `json:"watermark_position,omitempty"`
	WatermarkOpacity  *float64 `json:"watermark_opacity,omitempty"`
	WatermarkSize     *int     `json:"watermark_size,omitempty"`

	// Convert
	OutputFormat     *string `json:"output_format,omitempty"`
	OutputCodec      *string `json:"output_codec,omitempty"`
	OutputQuality    *int    `json:"output_quality,omitempty"`
	OutputResolution *string `json:"output_resolution,omitempty"`

	// Effects
	Brightness   *float64 `json:"brightness,omitempty"`
	Contrast     *float64 `json:"contrast,omitempty"`
	Saturation   *float64 `json:"saturation,omitempty"`
	Rotation     *int     `json:"rotation,omitempty"`
	Speed        *float64 `json:"speed,omitempty"`
	Volume       *float64 `json:"volume,omitempty"`
	RemoveAudio  *bool    `json:"remove_audio,omitempty"`

	// Output
	OutputFilename  *string `json:"output_filename,omitempty"`
	ReplaceOriginal *bool   `json:"replace_original,omitempty"`
}

// EditPresetResult is exposed to frontend
type EditPresetResult struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Icon        string             `json:"icon"`
	Operation   string             `json:"operation"`
	Settings    EditSettingsInput  `json:"settings"`
}

// InitEditor initializes the video editor (called during app startup)
func (a *App) initEditor() {
	if a.ffmpeg == nil || a.db == nil {
		a.logger.Warn("Editor", "Cannot initialize editor - dependencies not ready")
		return
	}

	ffmpegPath := a.ffmpeg.GetPath()
	a.editor = editor.New(ffmpegPath, a.db, a.config)
	a.editor.Start()

	a.logger.Info("Editor", "Video editor initialized", map[string]string{
		"ffmpeg_path": ffmpegPath,
	})
}

// CheckFFmpegWithGuidance returns detailed FFmpeg status with install guidance
func (a *App) CheckFFmpegWithGuidance() FFmpegCheckResult {
	if a.ffmpeg == nil {
		return FFmpegCheckResult{
			Installed: false,
			InstallGuide: "FFmpeg manager not initialized",
		}
	}
	return a.ffmpeg.CheckFFmpegWithGuidance()
}

// GetFFmpegInstallGuide returns OS-specific installation instructions
func (a *App) GetFFmpegInstallGuide() InstallGuide {
	if a.ffmpeg == nil {
		return InstallGuide{}
	}
	return a.ffmpeg.GetInstallGuide()
}

// OpenFFmpegDownloadPage opens the FFmpeg download page in browser
func (a *App) OpenFFmpegDownloadPage() {
	if a.ctx != nil {
		runtime.BrowserOpenURL(a.ctx, "https://ffmpeg.org/download.html")
	}
}

// GetVideoMetadata retrieves metadata for a video file
func (a *App) GetVideoMetadata(videoID string) (*VideoMetadataResult, error) {
	logger := a.logger

	if a.editor == nil {
		return nil, fmt.Errorf("editor not initialized")
	}

	// Get video from database
	video, err := a.db.GetVideoWithHash(videoID)
	if err != nil {
		logger.Error("Editor", "Failed to get video", err)
		return nil, err
	}
	if video == nil {
		return nil, fmt.Errorf("video not found")
	}

	// Get metadata using ffprobe/ffmpeg
	metadata, err := a.editor.GetVideoMetadata(video.FilePath)
	if err != nil {
		logger.Error("Editor", "Failed to get video metadata", err)
		// Return basic metadata from database
		return &VideoMetadataResult{
			Duration: float64(video.Duration),
			Codec:    video.Format,
		}, nil
	}

	return &VideoMetadataResult{
		Duration:     metadata.Duration,
		Width:        metadata.Width,
		Height:       metadata.Height,
		FPS:          metadata.FPS,
		Bitrate:      metadata.Bitrate,
		Codec:        metadata.Codec,
		AudioCodec:   metadata.AudioCodec,
		HasAudio:     metadata.HasAudio,
	}, nil
}

// SubmitEditJob submits a new video editing job
func (a *App) SubmitEditJob(videoID string, operation string, settings EditSettingsInput) (string, error) {
	logger := a.logger

	if a.editor == nil {
		return "", fmt.Errorf("editor not initialized")
	}

	// Convert input to db settings
	dbSettings := convertSettingsInput(settings)

	// Submit job
	jobID, err := a.editor.SubmitJob(videoID, operation, dbSettings)
	if err != nil {
		logger.Error("Editor", "Failed to submit edit job", err)
		return "", err
	}

	logger.Info("Editor", "Edit job submitted", map[string]string{
		"job_id":    jobID,
		"video_id":  videoID,
		"operation": operation,
	})

	// Emit event
	runtime.EventsEmit(a.ctx, "editor:job_started", map[string]interface{}{
		"job_id":    jobID,
		"video_id":  videoID,
		"operation": operation,
	})

	return jobID, nil
}

// GetEditJobStatus returns the status of an edit job
func (a *App) GetEditJobStatus(jobID string) (*EditJobResult, error) {
	if a.editor == nil {
		return nil, fmt.Errorf("editor not initialized")
	}

	job, err := a.editor.GetJobStatus(jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}

	return convertEditJobToResult(*job), nil
}

// ListEditJobs lists all edit jobs for a video
func (a *App) ListEditJobs(videoID string) ([]EditJobResult, error) {
	if a.editor == nil {
		return nil, fmt.Errorf("editor not initialized")
	}

	jobs, err := a.editor.ListJobs(videoID)
	if err != nil {
		return nil, err
	}

	results := make([]EditJobResult, len(jobs))
	for i, job := range jobs {
		results[i] = *convertEditJobToResult(job)
	}

	return results, nil
}

// CancelEditJob cancels an active edit job
func (a *App) CancelEditJob(jobID string) error {
	logger := a.logger

	if a.editor == nil {
		return fmt.Errorf("editor not initialized")
	}

	if err := a.editor.CancelJob(jobID); err != nil {
		logger.Error("Editor", "Failed to cancel edit job", err)
		return err
	}

	logger.Info("Editor", "Edit job cancelled", map[string]string{"job_id": jobID})
	return nil
}

// DeleteEditJob deletes an edit job record
func (a *App) DeleteEditJob(jobID string) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	return a.db.DeleteEditJob(jobID)
}

// GetEditPresets returns available edit presets
func (a *App) GetEditPresets() ([]EditPresetResult, error) {
	presets := []EditPresetResult{
		{
			ID:          "trim",
			Name:        "Trim Video",
			Description: "Remove unwanted start/end sections",
			Icon:        "scissors",
			Operation:   "crop",
			Settings:    EditSettingsInput{},
		},
		{
			ID:          "crop-16-9",
			Name:        "Crop to Widescreen",
			Description: "Crop to 16:9 aspect ratio",
			Icon:        "crop",
			Operation:   "crop",
			Settings:    EditSettingsInput{},
		},
		{
			ID:          "watermark-text",
			Name:        "Add Text Watermark",
			Description: "Add a text overlay to your video",
			Icon:        "text",
			Operation:   "watermark",
			Settings: EditSettingsInput{
				WatermarkType:     strPtr("text"),
				WatermarkPosition: strPtr("bottom-right"),
				WatermarkOpacity:  floatPtr(0.5),
				WatermarkSize:     intPtr(24),
			},
		},
		{
			ID:          "compress",
			Name:        "Compress Video",
			Description: "Reduce file size with H.265 codec",
			Icon:        "file-compression",
			Operation:   "convert",
			Settings: EditSettingsInput{
				OutputCodec:   strPtr("h265"),
				OutputQuality: intPtr(28),
			},
		},
		{
			ID:          "gif",
			Name:        "Convert to GIF",
			Description: "Create an animated GIF",
			Icon:        "file-gif",
			Operation:   "convert",
			Settings: EditSettingsInput{
				OutputFormat: strPtr("gif"),
			},
		},
		{
			ID:          "rotate-90",
			Name:        "Rotate 90°",
			Description: "Rotate video 90 degrees clockwise",
			Icon:        "rotate-clockwise",
			Operation:   "effects",
			Settings: EditSettingsInput{
				Rotation: intPtr(90),
			},
		},
		{
			ID:          "speed-up",
			Name:        "Speed Up 2x",
			Description: "Double the playback speed",
			Icon:        "player-play",
			Operation:   "effects",
			Settings: EditSettingsInput{
				Speed: floatPtr(2.0),
			},
		},
		{
			ID:          "mute",
			Name:        "Remove Audio",
			Description: "Create a silent video",
			Icon:        "volume-off",
			Operation:   "effects",
			Settings: EditSettingsInput{
				RemoveAudio: boolPtr(true),
			},
		},
	}

	return presets, nil
}

// PreviewEdit generates a preview frame for the given settings
func (a *App) PreviewEdit(videoID string, operation string, settings EditSettingsInput) (string, error) {
	logger := a.logger

	if a.editor == nil {
		return "", fmt.Errorf("editor not initialized")
	}

	// Get video
	video, err := a.db.GetVideoWithHash(videoID)
	if err != nil {
		return "", err
	}
	if video == nil {
		return "", fmt.Errorf("video not found")
	}

	// Convert settings
	dbSettings := convertSettingsInput(settings)

	// Generate preview at 25% duration
	metadata, _ := a.editor.GetVideoMetadata(video.FilePath)
	timestamp := 1.0
	if metadata != nil && metadata.Duration > 0 {
		timestamp = metadata.Duration * 0.25
	}

	previewPath, err := a.editor.GeneratePreview(video.FilePath, dbSettings, timestamp)
	if err != nil {
		logger.Error("Editor", "Failed to generate preview", err)
		return "", err
	}

	return previewPath, nil
}

// GetCropPresets returns available crop aspect ratio presets
func (a *App) GetCropPresets() map[string]map[string]interface{} {
	presets := editor.GetCropPresets()
	result := make(map[string]map[string]interface{})
	for key, preset := range presets {
		result[key] = map[string]interface{}{
			"name":   preset.Name,
			"width":  preset.Width,
			"height": preset.Height,
		}
	}
	return result
}

// GetWatermarkPositions returns available watermark positions
func (a *App) GetWatermarkPositions() map[string]string {
	return editor.GetWatermarkPositions()
}

// GetSupportedFormats returns supported output formats
func (a *App) GetSupportedFormats() map[string]map[string]interface{} {
	formats := editor.GetSupportedFormats()
	result := make(map[string]map[string]interface{})
	for key, format := range formats {
		result[key] = map[string]interface{}{
			"name":        format.Name,
			"extension":   format.Extension,
			"description": format.Description,
			"codecs":      format.Codecs,
		}
	}
	return result
}

// GetSupportedCodecs returns supported video codecs
func (a *App) GetSupportedCodecs() map[string]map[string]interface{} {
	codecs := editor.GetSupportedCodecs()
	result := make(map[string]map[string]interface{})
	for key, codec := range codecs {
		result[key] = map[string]interface{}{
			"name":        codec.Name,
			"description": codec.Description,
			"quality":     codec.Quality,
			"speed":       codec.Speed,
		}
	}
	return result
}

// GetEffectRanges returns valid ranges for effect parameters
func (a *App) GetEffectRanges() map[string]map[string]interface{} {
	ranges := editor.GetEffectRanges()
	result := make(map[string]map[string]interface{})
	for key, r := range ranges {
		result[key] = map[string]interface{}{
			"min":         r.Min,
			"max":         r.Max,
			"default":     r.Default,
			"step":        r.Step,
			"description": r.Description,
		}
	}
	return result
}

// GetRotationOptions returns available rotation angles
func (a *App) GetRotationOptions() []map[string]interface{} {
	options := editor.GetRotationOptions()
	result := make([]map[string]interface{}, len(options))
	for i, opt := range options {
		result[i] = map[string]interface{}{
			"value":       opt.Value,
			"label":       opt.Label,
			"description": opt.Description,
		}
	}
	return result
}

// Helper functions

func convertSettingsInput(input EditSettingsInput) db.EditSettings {
	return db.EditSettings{
		CropStart:         input.CropStart,
		CropEnd:           input.CropEnd,
		CropX:             input.CropX,
		CropY:             input.CropY,
		CropWidth:         input.CropWidth,
		CropHeight:        input.CropHeight,
		WatermarkType:     input.WatermarkType,
		WatermarkText:     input.WatermarkText,
		WatermarkImage:    input.WatermarkImage,
		WatermarkPosition: input.WatermarkPosition,
		WatermarkOpacity:  input.WatermarkOpacity,
		WatermarkSize:     input.WatermarkSize,
		OutputFormat:      input.OutputFormat,
		OutputCodec:       input.OutputCodec,
		OutputQuality:     input.OutputQuality,
		OutputResolution:  input.OutputResolution,
		Brightness:        input.Brightness,
		Contrast:          input.Contrast,
		Saturation:        input.Saturation,
		Rotation:          input.Rotation,
		Speed:             input.Speed,
		Volume:            input.Volume,
		RemoveAudio:       input.RemoveAudio,
		OutputFilename:    input.OutputFilename,
		ReplaceOriginal:   input.ReplaceOriginal,
	}
}

func convertEditJobToResult(job db.EditJob) *EditJobResult {
	result := &EditJobResult{
		ID:            job.ID,
		SourceVideoID: job.SourceVideoID,
		Status:        job.Status,
		Operation:     job.Operation,
		Progress:      job.Progress,
		CreatedAt:     job.CreatedAt.Unix(),
	}

	if job.OutputVideoID != nil {
		result.OutputVideoID = job.OutputVideoID
	}
	if job.ErrorMessage != nil {
		result.ErrorMessage = job.ErrorMessage
	}
	if job.CompletedAt != nil {
		completedAt := job.CompletedAt.Unix()
		result.CompletedAt = &completedAt
	}

	return result
}

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
