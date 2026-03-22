// Package editor provides FFmpeg best practices and utilities
package editor

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// FFmpegBestPractices contains recommended settings for high-quality output
// Based on FFmpeg community best practices and VLC/MPV player compatibility
var FFmpegBestPractices = struct {
	// Video codecs - prioritized by compatibility
	VideoCodecs struct {
		H264     string // libx264 - best compatibility
		H265     string // libx265 - better compression
		VP9      string // libvpx-vp9 - web optimized
		AV1      string // libaom-av1 - future format
		Copy     string // copy - no re-encoding
	}

	// Audio codecs - prioritized by compatibility
	AudioCodecs struct {
		AAC      string // aac - best compatibility
		Opus     string // libopus - best quality at low bitrate
		MP3      string // libmp3lame - universal compatibility
		FLAC     string // flac - lossless
		Copy     string // copy - no re-encoding
	}

	// Pixel formats
	PixelFormats struct {
		YUV420P  string // yuv420p - most compatible (required for MP4/Web)
		YUV444P  string // yuv444p - higher quality, less compatible
		Auto     string // auto - let FFmpeg decide
	}

	// Presets (speed vs compression)
	Presets struct {
		UltraFast string // fastest, largest file
		SuperFast string
		VeryFast  string
		Faster    string
		Fast      string
		Medium    string // default, good balance
		Slow      string
		Slower    string
		VerySlow  string // slowest, smallest file
	}

	// Tune options for specific content
	Tune struct {
		Film       string // film - for high quality movie content
		Animation  string // animation - for cartoons
		Grain      string // grain - preserves film grain
		StillImage string // stillimage - for slideshow-like content
		FastDecode string // fastdecode - faster decoding
		ZeroLatency string // zerolatency - for streaming
	}

	// CRF values (quality vs file size)
	CRF struct {
		VisuallyLossless int // 18 - visually lossless
		HighQuality      int // 20 - high quality
		GoodQuality      int // 23 - default, good quality
		MediumQuality    int // 28 - smaller file, acceptable quality
		LowQuality       int // 35 - smallest file, lower quality
	}

	// Audio bitrates (kbps)
	AudioBitrates struct {
		Low    int // 96 - voice/podcast
		Medium int // 128 - standard music
		High   int // 192 - high quality
		Lossless int // 320 or use FLAC
	}
}{
	VideoCodecs: struct {
		H264     string
		H265     string
		VP9      string
		AV1      string
		Copy     string
	}{
		H264:  "libx264",
		H265:  "libx265",
		VP9:   "libvpx-vp9",
		AV1:   "libaom-av1",
		Copy:  "copy",
	},
	AudioCodecs: struct {
		AAC      string
		Opus     string
		MP3      string
		FLAC     string
		Copy     string
	}{
		AAC:   "aac",
		Opus:  "libopus",
		MP3:   "libmp3lame",
		FLAC:  "flac",
		Copy:  "copy",
	},
	PixelFormats: struct {
		YUV420P  string
		YUV444P  string
		Auto     string
	}{
		YUV420P: "yuv420p",
		YUV444P: "yuv444p",
		Auto:    "",
	},
	Presets: struct {
		UltraFast string
		SuperFast string
		VeryFast  string
		Faster    string
		Fast      string
		Medium    string
		Slow      string
		Slower    string
		VerySlow  string
	}{
		UltraFast:  "ultrafast",
		SuperFast:  "superfast",
		VeryFast:   "veryfast",
		Faster:     "faster",
		Fast:       "fast",
		Medium:     "medium",
		Slow:       "slow",
		Slower:     "slower",
		VerySlow:   "veryslow",
	},
	Tune: struct {
		Film       string
		Animation  string
		Grain      string
		StillImage string
		FastDecode string
		ZeroLatency string
	}{
		Film:        "film",
		Animation:   "animation",
		Grain:       "grain",
		StillImage:  "stillimage",
		FastDecode:  "fastdecode",
		ZeroLatency: "zerolatency",
	},
	CRF: struct {
		VisuallyLossless int
		HighQuality      int
		GoodQuality      int
		MediumQuality    int
		LowQuality       int
	}{
		VisuallyLossless: 18,
		HighQuality:      20,
		GoodQuality:      23,
		MediumQuality:    28,
		LowQuality:       35,
	},
	AudioBitrates: struct {
		Low      int
		Medium   int
		High     int
		Lossless int
	}{
		Low:      96,
		Medium:   128,
		High:     192,
		Lossless: 320,
	},
}

// BuildFFmpegArgs builds FFmpeg arguments following best practices
func BuildFFmpegArgs(options FFmpegOptions) []string {
	args := []string{"-y"}

	// Input options
	if options.SeekTime > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", options.SeekTime))
	}

	args = append(args, "-i", options.InputPath)

	// Video filter complex
	if len(options.VideoFilters) > 0 {
		args = append(args, "-vf", strings.Join(options.VideoFilters, ","))
	}

	// Audio filter
	if len(options.AudioFilters) > 0 {
		args = append(args, "-af", strings.Join(options.AudioFilters, ","))
	}

	// Video codec
	if options.VideoCodec != "" {
		args = append(args, "-c:v", options.VideoCodec)

		// Codec-specific options
		if options.VideoCodec == FFmpegBestPractices.VideoCodecs.H264 ||
			options.VideoCodec == FFmpegBestPractices.VideoCodecs.H265 {
			// Preset
			if options.Preset != "" {
				args = append(args, "-preset", options.Preset)
			} else {
				args = append(args, "-preset", FFmpegBestPractices.Presets.Fast)
			}

			// Tune
			if options.Tune != "" {
				args = append(args, "-tune", options.Tune)
			}

			// CRF
			if options.CRF > 0 {
				args = append(args, "-crf", fmt.Sprintf("%d", options.CRF))
			} else {
				args = append(args, "-crf", fmt.Sprintf("%d", FFmpegBestPractices.CRF.GoodQuality))
			}

			// Pixel format - ALWAYS use yuv420p for compatibility
			args = append(args, "-pix_fmt", FFmpegBestPractices.PixelFormats.YUV420P)
		}
	}

	// Audio codec
	if options.RemoveAudio {
		args = append(args, "-an")
	} else if options.AudioCodec != "" {
		args = append(args, "-c:a", options.AudioCodec)

		// Audio bitrate
		if options.AudioBitrate > 0 {
			args = append(args, "-b:a", fmt.Sprintf("%dk", options.AudioBitrate))
		}

		// Audio codec specific options
		if options.AudioCodec == FFmpegBestPractices.AudioCodecs.AAC {
			// Use high efficiency AAC
			args = append(args, "-aac_coder", "twoloop")
		}
	}

	// Duration limit
	if options.Duration > 0 {
		args = append(args, "-t", fmt.Sprintf("%.3f", options.Duration))
	}

	// Fast start for web optimization
	if options.FastStart {
		args = append(args, "-movflags", "+faststart")
	}

	// Metadata
	if len(options.Metadata) > 0 {
		for key, value := range options.Metadata {
			args = append(args, "-metadata", fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Output path
	args = append(args, options.OutputPath)

	return args
}

// FFmpegOptions contains all options for building FFmpeg command
type FFmpegOptions struct {
	InputPath      string
	OutputPath     string
	SeekTime       float64
	Duration       float64
	VideoCodec     string
	AudioCodec     string
	VideoFilters   []string
	AudioFilters   []string
	CRF            int
	Preset         string
	Tune           string
	AudioBitrate   int
	RemoveAudio    bool
	FastStart      bool
	Metadata       map[string]string
}

// GetRecommendedSettings returns recommended settings based on use case
func GetRecommendedSettings(useCase string) FFmpegOptions {
	switch useCase {
	case "high_quality":
		return FFmpegOptions{
			VideoCodec:   FFmpegBestPractices.VideoCodecs.H264,
			AudioCodec:   FFmpegBestPractices.AudioCodecs.AAC,
			CRF:          FFmpegBestPractices.CRF.VisuallyLossless,
			Preset:       FFmpegBestPractices.Presets.Slow,
			AudioBitrate: FFmpegBestPractices.AudioBitrates.High,
			FastStart:    true,
		}
	case "web_optimized":
		return FFmpegOptions{
			VideoCodec:   FFmpegBestPractices.VideoCodecs.H264,
			AudioCodec:   FFmpegBestPractices.AudioCodecs.AAC,
			CRF:          FFmpegBestPractices.CRF.GoodQuality,
			Preset:       FFmpegBestPractices.Presets.Fast,
			AudioBitrate: FFmpegBestPractices.AudioBitrates.Medium,
			FastStart:    true,
		}
	case "small_file":
		return FFmpegOptions{
			VideoCodec:   FFmpegBestPractices.VideoCodecs.H265,
			AudioCodec:   FFmpegBestPractices.AudioCodecs.AAC,
			CRF:          FFmpegBestPractices.CRF.MediumQuality,
			Preset:       FFmpegBestPractices.Presets.Medium,
			AudioBitrate: FFmpegBestPractices.AudioBitrates.Medium,
			FastStart:    true,
		}
	case "fast_processing":
		return FFmpegOptions{
			VideoCodec:   FFmpegBestPractices.VideoCodecs.H264,
			AudioCodec:   FFmpegBestPractices.AudioCodecs.Copy,
			CRF:          FFmpegBestPractices.CRF.GoodQuality,
			Preset:       FFmpegBestPractices.Presets.UltraFast,
			FastStart:    false,
		}
	default:
		return FFmpegOptions{
			VideoCodec:   FFmpegBestPractices.VideoCodecs.H264,
			AudioCodec:   FFmpegBestPractices.AudioCodecs.AAC,
			CRF:          FFmpegBestPractices.CRF.GoodQuality,
			Preset:       FFmpegBestPractices.Presets.Fast,
			AudioBitrate: FFmpegBestPractices.AudioBitrates.Medium,
			FastStart:    true,
		}
	}
}

// CheckFFmpegVersion checks if FFmpeg version meets minimum requirements
func CheckFFmpegVersion(ffmpegPath string) (string, error) {
	if ffmpegPath == "" {
		return "", fmt.Errorf("ffmpeg path not provided")
	}

	cmd := exec.Command(ffmpegPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get ffmpeg version: %w", err)
	}

	// Parse version from first line
	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("no output from ffmpeg")
	}

	// First line format: "ffmpeg version x.x.x Copyright..."
	parts := strings.Fields(lines[0])
	for i, part := range parts {
		if part == "version" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}

	return "", fmt.Errorf("could not parse version")
}

// GetOptimalThreadCount returns optimal thread count for FFmpeg
func GetOptimalThreadCount() int {
	// Use number of CPU cores, but cap at a reasonable number
	numCPU := runtime.NumCPU()
	if numCPU > 8 {
		return 8
	}
	if numCPU < 2 {
		return 2
	}
	return numCPU
}

// ValidateFilterGraph validates a filter graph string
func ValidateFilterGraph(filterGraph string) error {
	if filterGraph == "" {
		return nil
	}

	// Basic validation - check for common syntax errors
	openBrackets := strings.Count(filterGraph, "[")
	closeBrackets := strings.Count(filterGraph, "]")
	if openBrackets != closeBrackets {
		return fmt.Errorf("unbalanced brackets in filter graph")
	}

	return nil
}
