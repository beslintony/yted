package editor

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"yted/internal/db"
)

// executeCrop performs video trimming and spatial cropping
func (e *Editor) executeCrop(ctx context.Context, video *db.Video, settings db.EditSettings, progressFn ProgressCallback) (string, error) {
	if e.ffmpegPath == "" {
		return "", fmt.Errorf("ffmpeg not available")
	}

	// Determine output path
	ext := filepath.Ext(video.FilePath)
	base := strings.TrimSuffix(filepath.Base(video.FilePath), ext)
	outputPath := filepath.Join(filepath.Dir(video.FilePath), base+"_cropped"+ext)

	// Build FFmpeg arguments
	args := []string{"-y", "-i", video.FilePath}

	// Time-based trimming (seek before input for faster processing)
	if settings.CropStart != nil && *settings.CropStart > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", *settings.CropStart))
	}

	// Build filter chain
	filters := []string{}

	// Spatial cropping
	if settings.CropX != nil && settings.CropY != nil && settings.CropWidth != nil && settings.CropHeight != nil {
		filters = append(filters, fmt.Sprintf("crop=%d:%d:%d:%d",
			*settings.CropWidth, *settings.CropHeight, *settings.CropX, *settings.CropY))
	}

	// Apply filters if any
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	// Duration limit (after input for accuracy)
	if settings.CropEnd != nil && settings.CropStart != nil && *settings.CropEnd > *settings.CropStart {
		duration := *settings.CropEnd - *settings.CropStart
		args = append(args, "-t", fmt.Sprintf("%.3f", duration))
	}

	// Video codec settings (re-encode for crop)
	args = append(args,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-c:a", "copy", // Copy audio without re-encoding
		"-movflags", "+faststart",
	)

	args = append(args, outputPath)

	// Execute command with progress monitoring
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)

	// Run command
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return "", fmt.Errorf("crop cancelled")
		}
		return "", fmt.Errorf("crop failed: %w", err)
	}

	// Report completion
	if progressFn != nil {
		progressFn(1.0, "")
	}

	return outputPath, nil
}

// GetCropPresets returns common crop aspect ratio presets
func GetCropPresets() map[string]struct {
	Name   string
	Width  int
	Height int
} {
	return map[string]struct {
		Name   string
		Width  int
		Height int
	}{
		"16:9": {Name: "Widescreen (16:9)", Width: 16, Height: 9},
		"4:3":  {Name: "Standard (4:3)", Width: 4, Height: 3},
		"1:1":  {Name: "Square (1:1)", Width: 1, Height: 1},
		"9:16": {Name: "Vertical (9:16)", Width: 9, Height: 16},
		"21:9": {Name: "Cinema (21:9)", Width: 21, Height: 9},
		"free": {Name: "Freeform", Width: 0, Height: 0},
	}
}

// CalculateCropRegion calculates crop dimensions for an aspect ratio
func CalculateCropRegion(videoWidth, videoHeight int, aspectRatio string) (x, y, width, height int) {
	presets := GetCropPresets()
	preset, exists := presets[aspectRatio]
	if !exists || preset.Width == 0 {
		// Freeform - return full video
		return 0, 0, videoWidth, videoHeight
	}

	// Calculate crop dimensions to fit the aspect ratio
	targetRatio := float64(preset.Width) / float64(preset.Height)
	currentRatio := float64(videoWidth) / float64(videoHeight)

	if currentRatio > targetRatio {
		// Video is wider - crop width
		height = videoHeight
		width = int(float64(height) * targetRatio)
		x = (videoWidth - width) / 2
		y = 0
	} else {
		// Video is taller - crop height
		width = videoWidth
		height = int(float64(width) / targetRatio)
		x = 0
		y = (videoHeight - height) / 2
	}

	return x, y, width, height
}

// FormatTime converts seconds to HH:MM:SS.ms format
func FormatTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, ms)
}

// ParseTime parses HH:MM:SS.ms format to seconds
func ParseTime(timeStr string) (float64, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format")
	}

	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}

	return hours*3600 + minutes*60 + seconds, nil
}
