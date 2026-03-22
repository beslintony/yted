package editor

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"yted/internal/db"
)

// executeEffects applies video effects (brightness, contrast, etc.)
func (e *Editor) executeEffects(ctx context.Context, video *db.Video, settings db.EditSettings, progressFn ProgressCallback) (string, error) {
	if e.ffmpegPath == "" {
		return "", fmt.Errorf("ffmpeg not available")
	}

	ext := filepath.Ext(video.FilePath)
	base := strings.TrimSuffix(filepath.Base(video.FilePath), ext)
	outputPath := filepath.Join(filepath.Dir(video.FilePath), base+"_effects"+ext)

	// Build video filter chain
	filters := []string{}

	// EQ filter (brightness, contrast, saturation)
	eqParams := []string{}
	if settings.Brightness != nil {
		eqParams = append(eqParams, fmt.Sprintf("brightness=%.2f", *settings.Brightness))
	}
	if settings.Contrast != nil {
		eqParams = append(eqParams, fmt.Sprintf("contrast=%.2f", *settings.Contrast))
	}
	if settings.Saturation != nil {
		eqParams = append(eqParams, fmt.Sprintf("saturation=%.2f", *settings.Saturation))
	}
	if len(eqParams) > 0 {
		filters = append(filters, "eq="+strings.Join(eqParams, ":"))
	}

	// Rotation
	if settings.Rotation != nil && *settings.Rotation != 0 {
		transpose := "0"
		switch *settings.Rotation {
		case 90:
			transpose = "1"
		case 180:
			// For 180 degrees, we need to transpose twice
			filters = append(filters, "transpose=2", "transpose=2")
		case 270:
			transpose = "3"
		}
		if transpose != "0" && *settings.Rotation != 180 {
			filters = append(filters, "transpose="+transpose)
		}
	}

	// Speed change
	atempo := ""
	if settings.Speed != nil && *settings.Speed != 1.0 {
		speed := *settings.Speed
		// Video speed filter
		if speed > 0.5 && speed <= 2.0 {
			filters = append(filters, fmt.Sprintf("setpts=PTS/%f", speed))
			// Audio tempo filter (limit to 0.5-2.0 range)
			if speed >= 0.5 && speed <= 2.0 {
				atempo = fmt.Sprintf("atempo=%f", speed)
			} else if speed < 0.5 {
				atempo = "atempo=0.5"
			} else {
				atempo = "atempo=2.0"
			}
		}
	}

	// Build command
	args := []string{"-y", "-i", video.FilePath}

	// Apply video filters
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	// Apply audio filters
	if atempo != "" {
		args = append(args, "-af", atempo)
	}

	// Remove audio if requested
	if settings.RemoveAudio != nil && *settings.RemoveAudio {
		args = append(args, "-an")
	} else if settings.Volume != nil && *settings.Volume != 1.0 {
		// Volume adjustment
		volume := *settings.Volume
		args = append(args, "-af", fmt.Sprintf("volume=%f", volume))
	}

	// Video codec settings
	args = append(args,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
	)

	// Audio codec (copy if no audio filters)
	if atempo == "" && (settings.Volume == nil || *settings.Volume == 1.0) {
		args = append(args, "-c:a", "copy")
	} else if !(*settings.RemoveAudio) {
		args = append(args, "-c:a", "aac", "-b:a", "128k")
	}

	args = append(args, "-movflags", "+faststart", outputPath)

	// Execute command
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return "", fmt.Errorf("effects application cancelled")
		}
		return "", fmt.Errorf("effects application failed: %w", err)
	}

	if progressFn != nil {
		progressFn(1.0, "")
	}

	return outputPath, nil
}

// executeCombine applies multiple operations in sequence
func (e *Editor) executeCombine(ctx context.Context, video *db.Video, settings db.EditSettings, progressFn ProgressCallback) (string, error) {
	// This is a placeholder for combining multiple operations
	// In a real implementation, you might chain operations or build a complex filter graph
	// For now, we'll just apply effects as the primary operation
	return e.executeEffects(ctx, video, settings, progressFn)
}

// GetEffectRanges returns the valid ranges for effect parameters
func GetEffectRanges() map[string]struct {
	Min         float64
	Max         float64
	Default     float64
	Step        float64
	Description string
}{
	return map[string]struct {
		Min         float64
		Max         float64
		Default     float64
		Step        float64
		Description string
	}{
		"brightness": {
			Min:         -1.0,
			Max:         1.0,
			Default:     0.0,
			Step:        0.1,
			Description: "Adjust brightness (-1.0 to 1.0)",
		},
		"contrast": {
			Min:         -1.0,
			Max:         1.0,
			Default:     0.0,
			Step:        0.1,
			Description: "Adjust contrast (-1.0 to 1.0)",
		},
		"saturation": {
			Min:         0.0,
			Max:         2.0,
			Default:     1.0,
			Step:        0.1,
			Description: "Adjust saturation (0.0 to 2.0)",
		},
		"speed": {
			Min:         0.5,
			Max:         2.0,
			Default:     1.0,
			Step:        0.1,
			Description: "Playback speed (0.5x to 2.0x)",
		},
		"volume": {
			Min:         0.0,
			Max:         2.0,
			Default:     1.0,
			Step:        0.1,
			Description: "Volume level (0.0 to 2.0)",
		},
	}
}

// GetRotationOptions returns available rotation angles
func GetRotationOptions() []struct {
	Value       int
	Label       string
	Description string
}{
	return []struct {
		Value       int
		Label       string
		Description string
	}{
		{0, "No Rotation", "Keep original orientation"},
		{90, "90° Clockwise", "Rotate 90 degrees right"},
		{180, "180°", "Flip upside down"},
		{270, "90° Counter-Clockwise", "Rotate 90 degrees left"},
	}
}

// ValidateEffectSettings validates effect settings are within valid ranges
func ValidateEffectSettings(settings db.EditSettings) []string {
	errors := []string{}
	ranges := GetEffectRanges()

	if settings.Brightness != nil {
		r := ranges["brightness"]
		if *settings.Brightness < r.Min || *settings.Brightness > r.Max {
			errors = append(errors, fmt.Sprintf("Brightness must be between %.1f and %.1f", r.Min, r.Max))
		}
	}

	if settings.Contrast != nil {
		r := ranges["contrast"]
		if *settings.Contrast < r.Min || *settings.Contrast > r.Max {
			errors = append(errors, fmt.Sprintf("Contrast must be between %.1f and %.1f", r.Min, r.Max))
		}
	}

	if settings.Saturation != nil {
		r := ranges["saturation"]
		if *settings.Saturation < r.Min || *settings.Saturation > r.Max {
			errors = append(errors, fmt.Sprintf("Saturation must be between %.1f and %.1f", r.Min, r.Max))
		}
	}

	if settings.Speed != nil {
		r := ranges["speed"]
		if *settings.Speed < r.Min || *settings.Speed > r.Max {
			errors = append(errors, fmt.Sprintf("Speed must be between %.1f and %.1f", r.Min, r.Max))
		}
	}

	if settings.Volume != nil {
		r := ranges["volume"]
		if *settings.Volume < r.Min || *settings.Volume > r.Max {
			errors = append(errors, fmt.Sprintf("Volume must be between %.1f and %.1f", r.Min, r.Max))
		}
	}

	return errors
}
