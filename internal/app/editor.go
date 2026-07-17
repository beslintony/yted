package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"sort"

	"yted/internal/db"
	"yted/internal/editor"
)

// FormatOption describes a supported output container format
type FormatOption struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Extension   string   `json:"extension"`
	Description string   `json:"description"`
	Codecs      []string `json:"codecs"`
}

// CodecOption describes a supported video codec
type CodecOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quality     string `json:"quality"`
	Speed       string `json:"speed"`
}

// CropPresetOption describes an aspect-ratio crop preset
type CropPresetOption struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// EffectRangeOption describes an adjustable effect and its valid range
type EffectRangeOption struct {
	ID          string  `json:"id"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Default     float64 `json:"default"`
	Step        float64 `json:"step"`
	Description string  `json:"description"`
}

// RotationOption describes a rotation choice
type RotationOption struct {
	Value       int    `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// EditOptions aggregates all editor option providers for the frontend
type EditOptions struct {
	Formats            []FormatOption      `json:"formats"`
	Codecs             []CodecOption       `json:"codecs"`
	CropPresets        []CropPresetOption  `json:"crop_presets"`
	EffectRanges       []EffectRangeOption `json:"effect_ranges"`
	Rotations          []RotationOption    `json:"rotations"`
	WatermarkPositions map[string]string   `json:"watermark_positions"`
}

// editOperations lists the operations accepted by SubmitEditJob
var editOperations = map[string]bool{
	"crop":      true,
	"watermark": true,
	"convert":   true,
	"effects":   true,
	"combine":   true,
}

// editorReady checks that the editor and ffmpeg are available
func (a *App) editorReady() error {
	if a.editor == nil {
		return fmt.Errorf("video editor not initialized")
	}
	if a.ffmpeg == nil || !a.ffmpeg.IsAvailable() {
		return fmt.Errorf("ffmpeg is not available - install it to use the video editor")
	}
	return nil
}

// SubmitEditJob queues an edit job for a library video and returns the job ID
func (a *App) SubmitEditJob(videoID, operation string, settings db.EditSettings) (string, error) {
	if err := a.editorReady(); err != nil {
		return "", err
	}
	if !editOperations[operation] {
		return "", fmt.Errorf("unknown edit operation: %s", operation)
	}
	if operation == "effects" {
		if problems := editor.ValidateEffectSettings(settings); len(problems) > 0 {
			return "", fmt.Errorf("invalid effect settings: %s", problems[0])
		}
	}

	jobID, err := a.editor.SubmitJob(videoID, operation, settings)
	if err != nil {
		a.logger.Error("Editor", "Failed to submit edit job", err)
		return "", err
	}
	return jobID, nil
}

// GetEditJobStatus returns the current state of an edit job
func (a *App) GetEditJobStatus(jobID string) (*db.EditJob, error) {
	if a.editor == nil {
		return nil, fmt.Errorf("video editor not initialized")
	}
	return a.editor.GetJobStatus(jobID)
}

// ListEditJobs lists all edit jobs for a library video
func (a *App) ListEditJobs(videoID string) ([]db.EditJob, error) {
	if a.editor == nil {
		return nil, fmt.Errorf("video editor not initialized")
	}
	return a.editor.ListJobs(videoID)
}

// CancelEditJob cancels an active edit job
func (a *App) CancelEditJob(jobID string) error {
	if a.editor == nil {
		return fmt.Errorf("video editor not initialized")
	}
	return a.editor.CancelJob(jobID)
}

// GetEditVideoMetadata returns ffprobe metadata for a library video
func (a *App) GetEditVideoMetadata(videoID string) (*editor.VideoMetadata, error) {
	if err := a.editorReady(); err != nil {
		return nil, err
	}
	video, err := a.db.GetVideoWithHash(videoID)
	if err != nil || video == nil {
		return nil, fmt.Errorf("video not found")
	}
	return a.editor.GetVideoMetadata(video.FilePath)
}

// GenerateEditPreview renders a preview frame with the given settings applied
// and returns it as a data URL for direct use in an <img> element
func (a *App) GenerateEditPreview(videoID string, settings db.EditSettings, timestamp float64) (string, error) {
	if err := a.editorReady(); err != nil {
		return "", err
	}
	video, err := a.db.GetVideoWithHash(videoID)
	if err != nil || video == nil {
		return "", fmt.Errorf("video not found")
	}

	previewPath, err := a.editor.GeneratePreview(video.FilePath, settings, timestamp)
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(previewPath) }()

	data, err := os.ReadFile(previewPath)
	if err != nil {
		return "", fmt.Errorf("failed to read preview: %w", err)
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data), nil
}

// GetEditOptions returns the available formats, codecs, presets and ranges
// that drive the editor UI
func (a *App) GetEditOptions() EditOptions {
	formats := editor.GetSupportedFormats()
	formatOpts := make([]FormatOption, 0, len(formats))
	for id, f := range formats {
		formatOpts = append(formatOpts, FormatOption{
			ID:          id,
			Name:        f.Name,
			Extension:   f.Extension,
			Description: f.Description,
			Codecs:      f.Codecs,
		})
	}
	sort.Slice(formatOpts, func(i, j int) bool { return formatOpts[i].ID < formatOpts[j].ID })

	codecs := editor.GetSupportedCodecs()
	codecOpts := make([]CodecOption, 0, len(codecs))
	for id, c := range codecs {
		codecOpts = append(codecOpts, CodecOption{
			ID:          id,
			Name:        c.Name,
			Description: c.Description,
			Quality:     c.Quality,
			Speed:       c.Speed,
		})
	}
	sort.Slice(codecOpts, func(i, j int) bool { return codecOpts[i].ID < codecOpts[j].ID })

	presets := editor.GetCropPresets()
	presetOpts := make([]CropPresetOption, 0, len(presets))
	for id, p := range presets {
		presetOpts = append(presetOpts, CropPresetOption{
			ID:     id,
			Name:   p.Name,
			Width:  p.Width,
			Height: p.Height,
		})
	}
	sort.Slice(presetOpts, func(i, j int) bool { return presetOpts[i].ID < presetOpts[j].ID })

	ranges := editor.GetEffectRanges()
	rangeOpts := make([]EffectRangeOption, 0, len(ranges))
	for id, r := range ranges {
		rangeOpts = append(rangeOpts, EffectRangeOption{
			ID:          id,
			Min:         r.Min,
			Max:         r.Max,
			Default:     r.Default,
			Step:        r.Step,
			Description: r.Description,
		})
	}
	sort.Slice(rangeOpts, func(i, j int) bool { return rangeOpts[i].ID < rangeOpts[j].ID })

	rotations := editor.GetRotationOptions()
	rotationOpts := make([]RotationOption, 0, len(rotations))
	for _, r := range rotations {
		rotationOpts = append(rotationOpts, RotationOption{
			Value:       r.Value,
			Label:       r.Label,
			Description: r.Description,
		})
	}

	return EditOptions{
		Formats:            formatOpts,
		Codecs:             codecOpts,
		CropPresets:        presetOpts,
		EffectRanges:       rangeOpts,
		Rotations:          rotationOpts,
		WatermarkPositions: editor.GetWatermarkPositions(),
	}
}
