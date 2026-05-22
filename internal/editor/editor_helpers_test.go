package editor

import (
	"testing"

	"yted/internal/db"
)

func TestParseFFProbeOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *VideoMetadata
	}{
		{
			name: "valid ffprobe output",
			input: `{
				"streams": [{
					"codec_name":"h264",
					"width":1920,
					"height":1080
				}],
				"format": {
					"duration":"3600.500"
				}
			}`,
			expected: &VideoMetadata{
				Duration: 3600.500,
				Width:    1920,
				Height:   1080,
				Codec:    "h264",
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: &VideoMetadata{},
		},
		{
			name: "partial output",
			input: `{
				"streams": [{
					"width": 1280
				}]
			}`,
			expected: &VideoMetadata{
				Width: 1280,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFFProbeOutput([]byte(tt.input))
			if err != nil {
				t.Fatalf("parseFFProbeOutput() unexpected error: %v", err)
			}
			if result.Duration != tt.expected.Duration {
				t.Errorf("Duration = %v, want %v", result.Duration, tt.expected.Duration)
			}
			if result.Width != tt.expected.Width {
				t.Errorf("Width = %v, want %v", result.Width, tt.expected.Width)
			}
			if result.Height != tt.expected.Height {
				t.Errorf("Height = %v, want %v", result.Height, tt.expected.Height)
			}
			if result.Codec != tt.expected.Codec {
				t.Errorf("Codec = %v, want %v", result.Codec, tt.expected.Codec)
			}
		})
	}
}

func TestParseFFmpegOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *VideoMetadata
	}{
		{
			name: "valid ffmpeg output",
			input: `ffmpeg version 4.4.2
  Duration: 01:30:00.00, start: 0.000000, bitrate: 5000 kb/s
  Stream #0:0(eng): Video: h264 1920x1080, yuv420p`,
			expected: &VideoMetadata{
				Duration: 5400,
				Width:    1920,
				Height:   1080,
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: &VideoMetadata{},
		},
		{
			name: "duration only",
			input: `ffmpeg version 4.4.2
  Duration: 00:05:30.50, start: 0.000000`,
			expected: &VideoMetadata{
				Duration: 330.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFFmpegOutput(tt.input, "test.mp4")
			if err != nil {
				t.Fatalf("parseFFmpegOutput() unexpected error: %v", err)
			}
			if result.Duration != tt.expected.Duration {
				t.Errorf("Duration = %v, want %v", result.Duration, tt.expected.Duration)
			}
			if result.Width != tt.expected.Width {
				t.Errorf("Width = %v, want %v", result.Width, tt.expected.Width)
			}
			if result.Height != tt.expected.Height {
				t.Errorf("Height = %v, want %v", result.Height, tt.expected.Height)
			}
		})
	}
}

func TestParseDurationHelper(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"00:01:30.500", 90.5},
		{"01:00:00.000", 3600},
		{"00:00:45.250", 45.25},
		{"02:30:15.000", 9015},
		{"invalid", 0},
		{"00:30", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseDuration(tt.input)
			if result != tt.expected {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetRotationOptions(t *testing.T) {
	options := GetRotationOptions()

	if len(options) != 4 {
		t.Errorf("GetRotationOptions() returned %d options, want 4", len(options))
	}

	expected := []struct {
		value int
		label string
	}{
		{0, "No Rotation"},
		{90, "90° Clockwise"},
		{180, "180°"},
		{270, "90° Counter-Clockwise"},
	}

	for i, exp := range expected {
		if options[i].Value != exp.value {
			t.Errorf("Option %d: Value = %d, want %d", i, options[i].Value, exp.value)
		}
		if options[i].Label != exp.label {
			t.Errorf("Option %d: Label = %q, want %q", i, options[i].Label, exp.label)
		}
	}
}

func TestCalculateWatermarkPosition(t *testing.T) {
	e := New("", nil, nil)

	tests := []struct {
		position string
		fontSize int
		wantX    string
		wantY    string
	}{
		{"top-left", 24, "12", "12"},
		{"top-center", 24, "(w-text_w)/2", "12"},
		{"top-right", 24, "w-text_w-12", "12"},
		{"center-left", 24, "12", "(h-text_h)/2"},
		{"center", 24, "(w-text_w)/2", "(h-text_h)/2"},
		{"center-right", 24, "w-text_w-12", "(h-text_h)/2"},
		{"bottom-left", 24, "12", "h-text_h-12"},
		{"bottom-center", 24, "(w-text_w)/2", "h-text_h-12"},
		{"bottom-right", 24, "w-text_w-12", "h-text_h-12"},
		{"unknown", 24, "w-text_w-12", "h-text_h-12"},
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			x, y := e.calculateWatermarkPosition(tt.position, tt.fontSize)
			if x != tt.wantX {
				t.Errorf("calculateWatermarkPosition() x = %q, want %q", x, tt.wantX)
			}
			if y != tt.wantY {
				t.Errorf("calculateWatermarkPosition() y = %q, want %q", y, tt.wantY)
			}
		})
	}
}

func TestCalculateOverlayPosition(t *testing.T) {
	e := New("", nil, nil)

	tests := []struct {
		position string
		wantX    string
		wantY    string
	}{
		{"top-left", "10", "10"},
		{"top-center", "(W-w)/2", "10"},
		{"top-right", "W-w-10", "10"},
		{"center-left", "10", "(H-h)/2"},
		{"center", "(W-w)/2", "(H-h)/2"},
		{"center-right", "W-w-10", "(H-h)/2"},
		{"bottom-left", "10", "H-h-10"},
		{"bottom-center", "(W-w)/2", "H-h-10"},
		{"bottom-right", "W-w-10", "H-h-10"},
		{"unknown", "W-w-10", "H-h-10"},
	}

	for _, tt := range tests {
		t.Run(tt.position, func(t *testing.T) {
			x, y := e.calculateOverlayPosition(tt.position)
			if x != tt.wantX {
				t.Errorf("calculateOverlayPosition() x = %q, want %q", x, tt.wantX)
			}
			if y != tt.wantY {
				t.Errorf("calculateOverlayPosition() y = %q, want %q", y, tt.wantY)
			}
		})
	}
}

func TestGetVideoCodec(t *testing.T) {
	e := New("", nil, nil)

	tests := []struct {
		codec    *string
		expected string
	}{
		{nil, "libx264"},
		{strPtr("h264"), "libx264"},
		{strPtr("h265"), "libx265"},
		{strPtr("hevc"), "libx265"},
		{strPtr("vp9"), "libvpx-vp9"},
		{strPtr("av1"), "libaom-av1"},
		{strPtr("unknown"), "libx264"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			settings := db.EditSettings{OutputCodec: tt.codec}
			result := e.getVideoCodec(settings)
			if result != tt.expected {
				t.Errorf("getVideoCodec() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetAudioCodec(t *testing.T) {
	e := New("", nil, nil)

	tests := []struct {
		name         string
		removeAudio  *bool
		outputFormat string
		expected     string
	}{
		{"mp4 with audio", boolPtr(false), "mp4", "aac"},
		{"webm with audio", boolPtr(false), "webm", "libopus"},
		{"mkv with audio", boolPtr(false), "mkv", "copy"},
		{"avi with audio", boolPtr(false), "avi", "mp3"},
		{"gif no audio", boolPtr(false), "gif", ""},
		{"remove audio", boolPtr(true), "mp4", ""},
		{"nil remove audio mp4", nil, "mp4", "aac"},
		{"unknown format", boolPtr(false), "xyz", "aac"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := db.EditSettings{
				RemoveAudio: tt.removeAudio,
			}
			result := e.getAudioCodec(settings, tt.outputFormat)
			if result != tt.expected {
				t.Errorf("getAudioCodec() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetResolutionScale(t *testing.T) {
	e := New("", nil, nil)

	tests := []struct {
		resolution string
		contains   string
	}{
		{"2160p", "3840:2160"},
		{"4K", "3840:2160"},
		{"1440p", "2560:1440"},
		{"2K", "2560:1440"},
		{"1080p", "1920:1080"},
		{"720p", "1280:720"},
		{"480p", "854:480"},
		{"360p", "640:360"},
		{"original", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.resolution, func(t *testing.T) {
			result := e.getResolutionScale(tt.resolution)
			if tt.contains != "" && result == "" {
				t.Errorf("getResolutionScale(%q) = empty, want to contain %q", tt.resolution, tt.contains)
			}
			if tt.contains != "" && result != "" && !contains(result, tt.contains) {
				t.Errorf("getResolutionScale(%q) = %q, want to contain %q", tt.resolution, result, tt.contains)
			}
			if tt.contains == "" && result != "" {
				t.Errorf("getResolutionScale(%q) = %q, want empty", tt.resolution, result)
			}
		})
	}
}

func TestAddFormatOptions(t *testing.T) {
	e := New("", nil, nil)

	tests := []struct {
		format   string
		contains []string
	}{
		{"mp4", []string{"-movflags", "+faststart"}},
		{"webm", []string{"-deadline", "good"}},
		{"gif", []string{"-loop", "0"}},
		{"mkv", nil},
		{"avi", nil},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			args := e.addFormatOptions([]string{}, tt.format)
			for _, expected := range tt.contains {
				found := false
				for _, arg := range args {
					if arg == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("addFormatOptions() missing %q in %v", expected, args)
				}
			}
		})
	}
}

func TestBuildTextWatermarkArgs(t *testing.T) {
	e := New("/usr/bin/ffmpeg", nil, nil)

	tests := []struct {
		name     string
		settings db.EditSettings
		contains []string
	}{
		{
			name: "default text watermark",
			settings: db.EditSettings{
				WatermarkType: ptrString("text"),
			},
			contains: []string{"-i", "-vf", "drawtext=text='YTed'", "-c:v", "libx264"},
		},
		{
			name: "custom text and position",
			settings: db.EditSettings{
				WatermarkType:     ptrString("text"),
				WatermarkText:     ptrString("My Channel"),
				WatermarkPosition: ptrString("top-left"),
				WatermarkOpacity:  ptrFloat64Helper(0.8),
				WatermarkSize:     ptrInt(36),
			},
			contains: []string{"drawtext=text='My Channel'", "fontsize=36", "fontcolor=white@0.80"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := e.buildTextWatermarkArgs("/path/to/video.mp4", tt.settings, "/path/to/output.mp4")
			for _, expected := range tt.contains {
				found := false
				for _, arg := range args {
					if contains(arg, expected) || arg == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildTextWatermarkArgs() missing %q in %v", expected, args)
				}
			}
		})
	}
}

func TestBuildImageWatermarkArgs(t *testing.T) {
	e := New("/usr/bin/ffmpeg", nil, nil)

	// Test with nil image path
	t.Run("nil image path", func(t *testing.T) {
		settings := db.EditSettings{
			WatermarkType: ptrString("image"),
		}
		args := e.buildImageWatermarkArgs("/path/to/video.mp4", settings, "/path/to/output.mp4")
		if args != nil {
			t.Error("buildImageWatermarkArgs() should return nil for nil image path")
		}
	})

	// Test with non-existent image
	t.Run("non-existent image", func(t *testing.T) {
		settings := db.EditSettings{
			WatermarkType: ptrString("image"),
			WatermarkImage: ptrString("/nonexistent/image.png"),
		}
		args := e.buildImageWatermarkArgs("/path/to/video.mp4", settings, "/path/to/output.mp4")
		if args != nil {
			t.Error("buildImageWatermarkArgs() should return nil for non-existent image")
		}
	})
}

func TestNewQueue(t *testing.T) {
	q := NewQueue(2)

	if q == nil {
		t.Fatal("NewQueue() returned nil")
	}
	if q.workers != 2 {
		t.Errorf("workers = %d, want 2", q.workers)
	}
	if q.tasks == nil {
		t.Error("tasks channel is nil")
	}
	if q.ctx == nil {
		t.Error("ctx is nil")
	}
}

func TestEditQueueSubmit(t *testing.T) {
	q := NewQueue(1)

	// Test submitting to closed queue
	q.cancel()

	err := q.Submit(&EditTask{
		JobID:     "test",
		Operation: "crop",
	})
	if err == nil {
		t.Error("Submit() to cancelled queue should return error")
	}
}

func TestRandInt(t *testing.T) {
	// Test that randInt returns different values
	val1 := randInt()
	val2 := randInt()

	// Very unlikely to be the same due to time-based
	if val1 == val2 {
		t.Logf("randInt returned same value twice: %d (unlikely but possible)", val1)
	}

	if val1 < 0 || val1 >= 10000 {
		t.Errorf("randInt() = %d, want value in range [0, 10000)", val1)
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func ptrInt(i int) *int {
	return &i
}

func ptrFloat64Helper(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
