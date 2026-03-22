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

// executeWatermarkEnhanced adds watermarks with professional quality settings
func (e *Editor) executeWatermarkEnhanced(ctx context.Context, video *db.Video, settings db.EditSettings, progressFn ProgressCallback) (string, error) {
	if e.ffmpegPath == "" {
		return "", fmt.Errorf("ffmpeg not available")
	}

	if settings.WatermarkType == nil {
		return "", fmt.Errorf("watermark type not specified")
	}

	// Determine output path
	ext := filepath.Ext(video.FilePath)
	base := strings.TrimSuffix(filepath.Base(video.FilePath), ext)
	outputPath := filepath.Join(filepath.Dir(video.FilePath), base+"_watermarked.mp4")

	// Get base options with best practices
	opts := GetRecommendedSettings("web_optimized")
	opts.InputPath = video.FilePath
	opts.OutputPath = outputPath

	// Build filter complex based on watermark type
	var filterComplex string
	switch *settings.WatermarkType {
	case "text":
		filterComplex = e.buildTextWatermarkFilter(settings)
	case "image":
		filterComplex = e.buildImageWatermarkFilter(settings)
	default:
		return "", fmt.Errorf("unknown watermark type: %s", *settings.WatermarkType)
	}

	if filterComplex == "" {
		return "", fmt.Errorf("failed to build watermark filter")
	}

	opts.VideoFilters = []string{filterComplex}

	// Build final command
	args := BuildFFmpegArgs(opts)

	// Execute with progress monitoring
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

// buildTextWatermarkFilter creates professional text watermark with proper formatting
func (e *Editor) buildTextWatermarkFilter(settings db.EditSettings) string {
	text := "YTed"
	if settings.WatermarkText != nil && *settings.WatermarkText != "" {
		text = *settings.WatermarkText
	}

	position := "bottom-right"
	if settings.WatermarkPosition != nil {
		position = *settings.WatermarkPosition
	}

	opacity := 0.7
	if settings.WatermarkOpacity != nil {
		opacity = *settings.WatermarkOpacity
	}

	fontSize := 24
	if settings.WatermarkSize != nil && *settings.WatermarkSize > 0 {
		fontSize = *settings.WatermarkSize
	}

	// Calculate position using FFmpeg expressions
	x, y := calculateTextPosition(position, fontSize)

	// Escape special characters in text
	text = escapeTextForFFmpeg(text)

	// Build drawtext filter with professional settings
	// Use box for better visibility
	boxOpacity := opacity * 0.3 // Box is more transparent than text
	
	drawtext := fmt.Sprintf(
		"drawtext=text='%s':fontsize=%d:fontcolor=white@%.2f:x=%s:y=%s:"+
			"box=1:boxcolor=black@%.2f:boxborderw=5:"+
			"fontfile=%s",
		text, fontSize, opacity, x, y, boxOpacity, getDefaultFont(),
	)

	return drawtext
}

// buildImageWatermarkFilter creates professional image watermark with proper alpha blending
func (e *Editor) buildImageWatermarkFilter(settings db.EditSettings) string {
	if settings.WatermarkImage == nil || *settings.WatermarkImage == "" {
		return ""
	}

	imagePath := *settings.WatermarkImage
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return ""
	}

	position := "bottom-right"
	if settings.WatermarkPosition != nil {
		position = *settings.WatermarkPosition
	}

	opacity := 0.7
	if settings.WatermarkOpacity != nil {
		opacity = *settings.WatermarkOpacity
	}

	scale := 100
	if settings.WatermarkSize != nil && *settings.WatermarkSize > 0 {
		scale = *settings.WatermarkSize
	}

	// Calculate position
	x, y := calculateOverlayPosition(position)

	// Build filter complex
	// [0:v] is main video, [1:v] is watermark image
	// Process watermark: apply opacity, scale if needed
	// Then overlay on main video
	
	var filters []string
	
	// Load and process watermark
	watermarkFilter := "[1:v]format=rgba"
	
	// Apply opacity
	if opacity < 1.0 {
		watermarkFilter += fmt.Sprintf(",colorchannelmixer=aa=%.2f", opacity)
	}
	
	// Apply scaling if needed
	if scale != 100 {
		watermarkFilter += fmt.Sprintf(",scale=%d:-1", scale)
	}
	
	watermarkFilter += "[logo]"
	filters = append(filters, watermarkFilter)
	
	// Overlay on main video
	overlayFilter := fmt.Sprintf("[0:v][logo]overlay=%s:%s:format=auto", x, y)
	
	// Final format conversion for compatibility
	overlayFilter += ",format=yuv420p"
	
	filters = append(filters, overlayFilter)

	return strings.Join(filters, ";")
}

// escapeTextForFFmpeg escapes special characters in text for FFmpeg drawtext
func escapeTextForFFmpeg(text string) string {
	// Escape FFmpeg special characters
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ":", "\\:")
	text = strings.ReplaceAll(text, "'", "\\'")
	text = strings.ReplaceAll(text, "%", "\\%")
	return text
}

// getDefaultFont returns the default font path for the current OS
func getDefaultFont() string {
	switch runtime.GOOS {
	case "darwin":
		return "/System/Library/Fonts/Helvetica.ttc"
	case "linux":
		// Try common Linux font paths
		candidates := []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/truetype/freefont/FreeSans.ttf",
		}
		for _, font := range candidates {
			if _, err := os.Stat(font); err == nil {
				return font
			}
		}
		return "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	case "windows":
		return "C:\\Windows\\Fonts\\arial.ttf"
	default:
		return ""
	}
}

// BuildWatermarkCommand builds a complete FFmpeg command for watermarking with multiple inputs
func (e *Editor) BuildWatermarkCommand(inputPath, outputPath string, settings db.EditSettings) ([]string, error) {
	if e.ffmpegPath == "" {
		return nil, fmt.Errorf("ffmpeg not available")
	}

	baseArgs := []string{"-y"}

	// Add main input
	baseArgs = append(baseArgs, "-i", inputPath)

	// Add watermark image if needed
	if settings.WatermarkType != nil && *settings.WatermarkType == "image" {
		if settings.WatermarkImage != nil && *settings.WatermarkImage != "" {
			baseArgs = append(baseArgs, "-i", *settings.WatermarkImage)
		}
	}

	// Build filter
	var filter string
	switch *settings.WatermarkType {
	case "text":
		filter = e.buildTextWatermarkFilter(settings)
	case "image":
		filter = e.buildImageWatermarkFilter(settings)
	}

	// Get recommended settings
	opts := GetRecommendedSettings("web_optimized")
	opts.InputPath = inputPath
	opts.OutputPath = outputPath
	opts.VideoFilters = []string{filter}

	// Build complete args
	args := BuildFFmpegArgs(opts)

	return args, nil
}
