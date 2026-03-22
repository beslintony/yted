package editor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"yted/internal/db"
)

// executeWatermark adds text or image watermarks to video
func (e *Editor) executeWatermark(ctx context.Context, video *db.Video, settings db.EditSettings, progressFn ProgressCallback) (string, error) {
	if e.ffmpegPath == "" {
		return "", fmt.Errorf("ffmpeg not available")
	}

	if settings.WatermarkType == nil {
		return "", fmt.Errorf("watermark type not specified")
	}

	ext := filepath.Ext(video.FilePath)
	base := strings.TrimSuffix(filepath.Base(video.FilePath), ext)
	outputPath := filepath.Join(filepath.Dir(video.FilePath), base+"_watermarked"+ext)

	var args []string

	switch *settings.WatermarkType {
	case "text":
		args = e.buildTextWatermarkArgs(video.FilePath, settings, outputPath)
	case "image":
		args = e.buildImageWatermarkArgs(video.FilePath, settings, outputPath)
	default:
		return "", fmt.Errorf("unknown watermark type: %s", *settings.WatermarkType)
	}

	if args == nil {
		return "", fmt.Errorf("failed to build watermark arguments")
	}

	// Execute command
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return "", fmt.Errorf("watermark cancelled")
		}
		return "", fmt.Errorf("watermark failed: %w", err)
	}

	if progressFn != nil {
		progressFn(1.0, "")
	}

	return outputPath, nil
}

// buildTextWatermarkArgs builds FFmpeg arguments for text watermark
func (e *Editor) buildTextWatermarkArgs(inputPath string, settings db.EditSettings, outputPath string) []string {
	text := "YTed"
	if settings.WatermarkText != nil && *settings.WatermarkText != "" {
		text = *settings.WatermarkText
	}

	position := "bottom-right"
	if settings.WatermarkPosition != nil {
		position = *settings.WatermarkPosition
	}

	opacity := 0.5
	if settings.WatermarkOpacity != nil {
		opacity = *settings.WatermarkOpacity
	}

	fontSize := 24
	if settings.WatermarkSize != nil && *settings.WatermarkSize > 0 {
		fontSize = *settings.WatermarkSize
	}

	// Calculate position
	x, y := e.calculateWatermarkPosition(position, fontSize)

	// Escape text for FFmpeg
	text = strings.ReplaceAll(text, ":", `\\:`)
	text = strings.ReplaceAll(text, "'", `\\'`)

	// Build drawtext filter
	drawtext := fmt.Sprintf("drawtext=text='%s':fontsize=%d:fontcolor=white@%.2f:x=%s:y=%s",
		text, fontSize, opacity, x, y)

	args := []string{
		"-y",
		"-i", inputPath,
		"-vf", drawtext,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-c:a", "copy",
		"-movflags", "+faststart",
		outputPath,
	}

	return args
}

// buildImageWatermarkArgs builds FFmpeg arguments for image watermark
func (e *Editor) buildImageWatermarkArgs(inputPath string, settings db.EditSettings, outputPath string) []string {
	if settings.WatermarkImage == nil || *settings.WatermarkImage == "" {
		return nil
	}

	imagePath := *settings.WatermarkImage
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil
	}

	position := "bottom-right"
	if settings.WatermarkPosition != nil {
		position = *settings.WatermarkPosition
	}

	opacity := 0.5
	if settings.WatermarkOpacity != nil {
		opacity = *settings.WatermarkOpacity
	}

	scale := 100
	if settings.WatermarkSize != nil && *settings.WatermarkSize > 0 {
		scale = *settings.WatermarkSize
	}

	// Calculate position
	x, y := e.calculateOverlayPosition(position)

	// Build overlay filter with opacity
	overlay := fmt.Sprintf("[1:v]format=rgba,colorchannelmixer=aa=%.2f[logo];[0:v][logo]overlay=%s:%s[OUT]",
		opacity, x, y)

	// If scaling is needed
	if scale != 100 {
		overlay = fmt.Sprintf("[1:v]format=rgba,colorchannelmixer=aa=%.2f,scale=%d:-1[logo];[0:v][logo]overlay=%s:%s[OUT]",
			opacity, scale, x, y)
	}

	args := []string{
		"-y",
		"-i", inputPath,
		"-i", imagePath,
		"-filter_complex", overlay,
		"-map", "[OUT]",
		"-map", "0:a?",
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-c:a", "copy",
		"-movflags", "+faststart",
		outputPath,
	}

	return args
}

// calculateWatermarkPosition calculates position for text watermark
func (e *Editor) calculateWatermarkPosition(position string, fontSize int) (x, y string) {
	padding := fontSize / 2

	switch position {
	case "top-left":
		return fmt.Sprintf("%d", padding), fmt.Sprintf("%d", padding)
	case "top-center":
		return "(w-text_w)/2", fmt.Sprintf("%d", padding)
	case "top-right":
		return fmt.Sprintf("w-text_w-%d", padding), fmt.Sprintf("%d", padding)
	case "center-left":
		return fmt.Sprintf("%d", padding), "(h-text_h)/2"
	case "center":
		return "(w-text_w)/2", "(h-text_h)/2"
	case "center-right":
		return fmt.Sprintf("w-text_w-%d", padding), "(h-text_h)/2"
	case "bottom-left":
		return fmt.Sprintf("%d", padding), fmt.Sprintf("h-text_h-%d", padding)
	case "bottom-center":
		return "(w-text_w)/2", fmt.Sprintf("h-text_h-%d", padding)
	case "bottom-right":
		return fmt.Sprintf("w-text_w-%d", padding), fmt.Sprintf("h-text_h-%d", padding)
	default:
		return fmt.Sprintf("w-text_w-%d", padding), fmt.Sprintf("h-text_h-%d", padding)
	}
}

// calculateOverlayPosition calculates position for image overlay
func (e *Editor) calculateOverlayPosition(position string) (x, y string) {
	padding := 10

	switch position {
	case "top-left":
		return fmt.Sprintf("%d", padding), fmt.Sprintf("%d", padding)
	case "top-center":
		return "(W-w)/2", fmt.Sprintf("%d", padding)
	case "top-right":
		return fmt.Sprintf("W-w-%d", padding), fmt.Sprintf("%d", padding)
	case "center-left":
		return fmt.Sprintf("%d", padding), "(H-h)/2"
	case "center":
		return "(W-w)/2", "(H-h)/2"
	case "center-right":
		return fmt.Sprintf("W-w-%d", padding), "(H-h)/2"
	case "bottom-left":
		return fmt.Sprintf("%d", padding), fmt.Sprintf("H-h-%d", padding)
	case "bottom-center":
		return "(W-w)/2", fmt.Sprintf("H-h-%d", padding)
	case "bottom-right":
		return fmt.Sprintf("W-w-%d", padding), fmt.Sprintf("H-h-%d", padding)
	default:
		return fmt.Sprintf("W-w-%d", padding), fmt.Sprintf("H-h-%d", padding)
	}
}

// GetWatermarkPositions returns available watermark positions
func GetWatermarkPositions() map[string]string {
	return map[string]string{
		"top-left":     "Top Left",
		"top-center":   "Top Center",
		"top-right":    "Top Right",
		"center-left":  "Center Left",
		"center":       "Center",
		"center-right": "Center Right",
		"bottom-left":  "Bottom Left",
		"bottom-center": "Bottom Center",
		"bottom-right": "Bottom Right",
	}
}
