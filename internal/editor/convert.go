package editor

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"yted/internal/db"
)

// executeConvert converts video to different format/codec
func (e *Editor) executeConvert(ctx context.Context, video *db.Video, settings db.EditSettings, progressFn ProgressCallback) (string, error) {
	if e.ffmpegPath == "" {
		return "", fmt.Errorf("ffmpeg not available")
	}

	// Determine output format
	outputFormat := "mp4"
	if settings.OutputFormat != nil && *settings.OutputFormat != "" {
		outputFormat = *settings.OutputFormat
	}

	// Determine output path
	base := strings.TrimSuffix(filepath.Base(video.FilePath), filepath.Ext(video.FilePath))
	outputPath := filepath.Join(filepath.Dir(video.FilePath), base+"_converted."+outputFormat)

	// Build FFmpeg arguments
	args := []string{"-y", "-i", video.FilePath}

	// Video codec selection
	videoCodec := e.getVideoCodec(settings)
	if videoCodec != "" {
		args = append(args, "-c:v", videoCodec)
	}

	// Quality settings
	crf := 23
	if settings.OutputQuality != nil && *settings.OutputQuality > 0 {
		crf = *settings.OutputQuality
	}

	if videoCodec == "libx264" || videoCodec == "libx265" {
		args = append(args, "-crf", fmt.Sprintf("%d", crf))
		args = append(args, "-preset", "fast")
	}

	// Resolution scaling
	if settings.OutputResolution != nil && *settings.OutputResolution != "original" {
		scale := e.getResolutionScale(*settings.OutputResolution)
		if scale != "" {
			args = append(args, "-vf", scale)
		}
	}

	// Audio codec
	audioCodec := e.getAudioCodec(settings, outputFormat)
	if audioCodec != "" {
		args = append(args, "-c:a", audioCodec)
		if audioCodec == "aac" {
			args = append(args, "-b:a", "128k")
		}
	}

	// Remove audio if requested
	if settings.RemoveAudio != nil && *settings.RemoveAudio {
		args = append(args, "-an")
	}

	// Format-specific options
	args = e.addFormatOptions(args, outputFormat)

	args = append(args, outputPath)

	// Execute command
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return "", fmt.Errorf("conversion cancelled")
		}
		return "", fmt.Errorf("conversion failed: %w", err)
	}

	if progressFn != nil {
		progressFn(1.0, "")
	}

	return outputPath, nil
}

// getVideoCodec returns the appropriate video codec based on settings
func (e *Editor) getVideoCodec(settings db.EditSettings) string {
	if settings.OutputCodec == nil {
		return "libx264" // Default to H.264
	}

	switch *settings.OutputCodec {
	case "h264":
		return "libx264"
	case "h265", "hevc":
		return "libx265"
	case "vp9":
		return "libvpx-vp9"
	case "av1":
		return "libaom-av1"
	default:
		return "libx264"
	}
}

// getAudioCodec returns the appropriate audio codec based on format
func (e *Editor) getAudioCodec(settings db.EditSettings, outputFormat string) string {
	if settings.RemoveAudio != nil && *settings.RemoveAudio {
		return "" // No audio
	}

	switch outputFormat {
	case "mp4", "mov":
		return "aac"
	case "webm":
		return "libopus"
	case "mkv":
		return "copy" // Keep original for MKV
	case "avi":
		return "mp3"
	case "gif":
		return "" // No audio for GIF
	default:
		return "aac"
	}
}

// getResolutionScale returns the scale filter for the target resolution
func (e *Editor) getResolutionScale(resolution string) string {
	switch resolution {
	case "2160p", "4K":
		return "scale=3840:2160:force_original_aspect_ratio=decrease,pad=3840:2160:(ow-iw)/2:(oh-ih)/2"
	case "1440p", "2K":
		return "scale=2560:1440:force_original_aspect_ratio=decrease,pad=2560:1440:(ow-iw)/2:(oh-ih)/2"
	case "1080p":
		return "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2"
	case "720p":
		return "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2"
	case "480p":
		return "scale=854:480:force_original_aspect_ratio=decrease,pad=854:480:(ow-iw)/2:(oh-ih)/2"
	case "360p":
		return "scale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2"
	default:
		return ""
	}
}

// addFormatOptions adds format-specific encoding options
func (e *Editor) addFormatOptions(args []string, format string) []string {
	switch format {
	case "mp4":
		args = append(args, "-movflags", "+faststart")
	case "webm":
		args = append(args, "-deadline", "good", "-cpu-used", "2")
	case "gif":
		args = append(args, "-vf", "fps=30,scale=480:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse=dither=bayer")
		args = append(args, "-loop", "0")
	}
	return args
}

// GetSupportedFormats returns list of supported output formats
func GetSupportedFormats() map[string]struct {
	Name        string
	Extension   string
	Description string
	Codecs      []string
}{
	return map[string]struct {
		Name        string
		Extension   string
		Description string
		Codecs      []string
	}{
		"mp4": {
			Name:        "MP4",
			Extension:   "mp4",
			Description: "Most compatible format for all devices",
			Codecs:      []string{"h264", "h265"},
		},
		"mkv": {
			Name:        "MKV",
			Extension:   "mkv",
			Description: "Matroska container, supports any codec",
			Codecs:      []string{"h264", "h265", "vp9", "av1"},
		},
		"webm": {
			Name:        "WebM",
			Extension:   "webm",
			Description: "Web-optimized format",
			Codecs:      []string{"vp9"},
		},
		"mov": {
			Name:        "QuickTime",
			Extension:   "mov",
			Description: "Apple QuickTime format",
			Codecs:      []string{"h264", "h265"},
		},
		"avi": {
			Name:        "AVI",
			Extension:   "avi",
			Description: "Legacy Windows format",
			Codecs:      []string{"h264"},
		},
		"gif": {
			Name:        "GIF",
			Extension:   "gif",
			Description: "Animated image (no audio)",
			Codecs:      []string{},
		},
	}
}

// GetSupportedCodecs returns list of supported video codecs
func GetSupportedCodecs() map[string]struct {
	Name        string
	Description string
	Quality     string // good, better, best
	Speed       string // fast, medium, slow
}{
	return map[string]struct {
		Name        string
		Description string
		Quality     string
		Speed       string
	}{
		"h264": {
			Name:        "H.264 (AVC)",
			Description: "Best compatibility, slightly larger files",
			Quality:     "good",
			Speed:       "fast",
		},
		"h265": {
			Name:        "H.265 (HEVC)",
			Description: "Better compression, may not work on older devices",
			Quality:     "better",
			Speed:       "medium",
		},
		"vp9": {
			Name:        "VP9",
			Description: "Open format, best for web streaming",
			Quality:     "better",
			Speed:       "slow",
		},
		"av1": {
			Name:        "AV1",
			Description: "Next-gen format, best compression but slow",
			Quality:     "best",
			Speed:       "slow",
		},
	}
}

// EstimateOutputSize estimates the output file size based on settings
func EstimateOutputSize(inputSize int64, inputDuration float64, settings db.EditSettings) int64 {
	if inputDuration <= 0 {
		return inputSize
	}

	// Base bitrate estimation (very rough)
	baseBitrate := float64(inputSize*8) / inputDuration // bits per second

	// Adjust for codec efficiency
	codecMultiplier := 1.0
	if settings.OutputCodec != nil {
		switch *settings.OutputCodec {
		case "h265":
			codecMultiplier = 0.6 // 40% smaller than H.264
		case "vp9":
			codecMultiplier = 0.5
		case "av1":
			codecMultiplier = 0.4
		}
	}

	// Adjust for quality
	qualityMultiplier := 1.0
	if settings.OutputQuality != nil {
		// Lower CRF = higher quality = larger file
		// Typical CRF range: 18-28
		qualityMultiplier = float64(*settings.OutputQuality) / 23.0
	}

	// Adjust for resolution
	resMultiplier := 1.0
	if settings.OutputResolution != nil && *settings.OutputResolution != "original" {
		switch *settings.OutputResolution {
		case "1080p":
			resMultiplier = 0.5 // Roughly half the pixels of 4K
		case "720p":
			resMultiplier = 0.25
		case "480p":
			resMultiplier = 0.15
		case "360p":
			resMultiplier = 0.08
		}
	}

	estimatedBitrate := baseBitrate * codecMultiplier * qualityMultiplier * resMultiplier
	estimatedSize := int64(estimatedBitrate * inputDuration / 8)

	// Sanity check: don't estimate less than 10% of original
	if estimatedSize < inputSize/10 {
		estimatedSize = inputSize / 10
	}

	return estimatedSize
}
