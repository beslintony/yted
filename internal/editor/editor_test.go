package editor

import (
	"testing"

	"yted/internal/db"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		ffmpegPath string
		wantNil    bool
	}{
		{
			name:       "creates editor with empty ffmpeg path",
			ffmpegPath: "",
			wantNil:    false,
		},
		{
			name:       "creates editor with ffmpeg path",
			ffmpegPath: "/usr/bin/ffmpeg",
			wantNil:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We're testing with nil db and config for unit tests
			// In production, these would be actual instances
			editor := New(tt.ffmpegPath, nil, nil)
			if (editor == nil) != tt.wantNil {
				t.Errorf("New() = %v, want nil=%v", editor, tt.wantNil)
			}
			if editor != nil {
				if editor.ffmpegPath != tt.ffmpegPath {
					t.Errorf("ffmpegPath = %v, want %v", editor.ffmpegPath, tt.ffmpegPath)
				}
			}
		})
	}
}

func TestSetFFmpegPath(t *testing.T) {
	editor := New("/initial/path", nil, nil)
	
	newPath := "/new/ffmpeg/path"
	editor.SetFFmpegPath(newPath)
	
	if editor.ffmpegPath != newPath {
		t.Errorf("SetFFmpegPath() failed, got %v, want %v", editor.ffmpegPath, newPath)
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		wantErr  bool
	}{
		{"00:01:30.500", 90.5, false},
		{"01:00:00.000", 3600, false},
		{"00:00:45.250", 45.25, false},
		{"invalid", 0, true},
		{"00:30", 0, true}, // Missing seconds
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseTime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTime(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseTime(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseTime(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{90.5, "00:01:30.500"},
		{3600, "01:00:00.000"},
		{45.25, "00:00:45.250"},
		{0, "00:00:00.000"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatTime(tt.seconds)
			if result != tt.expected {
				t.Errorf("FormatTime(%v) = %v, want %v", tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestCalculateCropRegion(t *testing.T) {
	tests := []struct {
		name         string
		videoWidth   int
		videoHeight  int
		aspectRatio  string
		wantX        int
		wantY        int
		wantWidth    int
		wantHeight   int
	}{
		{
			name:        "16:9 crop on 4:3 video",
			videoWidth:  1440,
			videoHeight: 1080,
			aspectRatio: "16:9",
			wantX:       0,
			wantY:       135,
			wantWidth:   1440,
			wantHeight:  810,
		},
		{
			name:        "1:1 crop on 16:9 video",
			videoWidth:  1920,
			videoHeight: 1080,
			aspectRatio: "1:1",
			wantX:       420,
			wantY:       0,
			wantWidth:   1080,
			wantHeight:  1080,
		},
		{
			name:        "freeform returns full video",
			videoWidth:  1920,
			videoHeight: 1080,
			aspectRatio: "free",
			wantX:       0,
			wantY:       0,
			wantWidth:   1920,
			wantHeight:  1080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, width, height := CalculateCropRegion(tt.videoWidth, tt.videoHeight, tt.aspectRatio)
			if x != tt.wantX || y != tt.wantY || width != tt.wantWidth || height != tt.wantHeight {
				t.Errorf("CalculateCropRegion() = (%d, %d, %d, %d), want (%d, %d, %d, %d)",
					x, y, width, height, tt.wantX, tt.wantY, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestGetCropPresets(t *testing.T) {
	presets := GetCropPresets()
	
	expectedPresets := []string{"16:9", "4:3", "1:1", "9:16", "21:9", "free"}
	
	for _, preset := range expectedPresets {
		if _, exists := presets[preset]; !exists {
			t.Errorf("GetCropPresets() missing preset: %s", preset)
		}
	}
	
	// Verify specific preset values
	if preset, exists := presets["16:9"]; exists {
		if preset.Name != "Widescreen (16:9)" {
			t.Errorf("16:9 preset name = %v, want 'Widescreen (16:9)'", preset.Name)
		}
		if preset.Width != 16 || preset.Height != 9 {
			t.Errorf("16:9 preset dimensions = (%d, %d), want (16, 9)", preset.Width, preset.Height)
		}
	}
}

func TestGetWatermarkPositions(t *testing.T) {
	positions := GetWatermarkPositions()
	
	expectedPositions := []string{
		"top-left", "top-center", "top-right",
		"center-left", "center", "center-right",
		"bottom-left", "bottom-center", "bottom-right",
	}
	
	for _, pos := range expectedPositions {
		if _, exists := positions[pos]; !exists {
			t.Errorf("GetWatermarkPositions() missing position: %s", pos)
		}
	}
	
	// Verify specific position labels
	if positions["bottom-right"] != "Bottom Right" {
		t.Errorf("bottom-right label = %v, want 'Bottom Right'", positions["bottom-right"])
	}
}

func TestGetSupportedFormats(t *testing.T) {
	formats := GetSupportedFormats()
	
	expectedFormats := []string{"mp4", "mkv", "webm", "mov", "avi", "gif"}
	
	for _, format := range expectedFormats {
		if _, exists := formats[format]; !exists {
			t.Errorf("GetSupportedFormats() missing format: %s", format)
		}
	}
	
	// Verify MP4 format
	if mp4, exists := formats["mp4"]; exists {
		if mp4.Name != "MP4" {
			t.Errorf("mp4.Name = %v, want 'MP4'", mp4.Name)
		}
		if len(mp4.Codecs) == 0 {
			t.Error("mp4.Codecs should not be empty")
		}
	}
}

func TestGetSupportedCodecs(t *testing.T) {
	codecs := GetSupportedCodecs()
	
	expectedCodecs := []string{"h264", "h265", "vp9", "av1"}
	
	for _, codec := range expectedCodecs {
		if _, exists := codecs[codec]; !exists {
			t.Errorf("GetSupportedCodecs() missing codec: %s", codec)
		}
	}
	
	// Verify H.264 codec
	if h264, exists := codecs["h264"]; exists {
		if h264.Quality != "good" {
			t.Errorf("h264.Quality = %v, want 'good'", h264.Quality)
		}
		if h264.Speed != "fast" {
			t.Errorf("h264.Speed = %v, want 'fast'", h264.Speed)
		}
	}
}

func TestGetEffectRanges(t *testing.T) {
	ranges := GetEffectRanges()
	
	expectedEffects := []string{"brightness", "contrast", "saturation", "speed", "volume"}
	
	for _, effect := range expectedEffects {
		if _, exists := ranges[effect]; !exists {
			t.Errorf("GetEffectRanges() missing effect: %s", effect)
		}
	}
	
	// Verify brightness range
	if brightness, exists := ranges["brightness"]; exists {
		if brightness.Min != -1.0 || brightness.Max != 1.0 {
			t.Errorf("brightness range = (%v, %v), want (-1.0, 1.0)", brightness.Min, brightness.Max)
		}
		if brightness.Default != 0.0 {
			t.Errorf("brightness.Default = %v, want 0.0", brightness.Default)
		}
	}
}

func TestValidateEffectSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings db.EditSettings
		wantErrs int
	}{
		{
			name:     "valid settings",
			settings: db.EditSettings{},
			wantErrs: 0,
		},
		{
			name: "brightness out of range",
			settings: db.EditSettings{
				Brightness: ptrFloat64(2.0),
			},
			wantErrs: 1,
		},
		{
			name: "contrast too low",
			settings: db.EditSettings{
				Contrast: ptrFloat64(-2.0),
			},
			wantErrs: 1,
		},
		{
			name: "saturation out of range",
			settings: db.EditSettings{
				Saturation: ptrFloat64(3.0),
			},
			wantErrs: 1,
		},
		{
			name: "multiple invalid settings",
			settings: db.EditSettings{
				Brightness: ptrFloat64(2.0),
				Contrast:   ptrFloat64(-2.0),
				Speed:      ptrFloat64(3.0),
			},
			wantErrs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateEffectSettings(tt.settings)
			if len(errors) != tt.wantErrs {
				t.Errorf("ValidateEffectSettings() returned %d errors, want %d: %v", len(errors), tt.wantErrs, errors)
			}
		})
	}
}

func TestEstimateOutputSize(t *testing.T) {
	tests := []struct {
		name          string
		inputSize     int64
		inputDuration float64
		settings      db.EditSettings
		minExpected   int64 // Minimum expected size (sanity check)
	}{
		{
			name:          "same quality",
			inputSize:     100000000, // 100MB
			inputDuration: 600,       // 10 minutes
			settings:      db.EditSettings{},
			minExpected:   10000000, // At least 10MB
		},
		{
			name:          "H.265 compression",
			inputSize:     100000000,
			inputDuration: 600,
			settings: db.EditSettings{
				OutputCodec: ptrString("h265"),
			},
			minExpected: 5000000, // Should be smaller than original
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateOutputSize(tt.inputSize, tt.inputDuration, tt.settings)
			if result < tt.minExpected {
				t.Errorf("EstimateOutputSize() = %v, want at least %v", result, tt.minExpected)
			}
			// Sanity check: should not exceed input size significantly
			if result > tt.inputSize*10 {
				t.Errorf("EstimateOutputSize() = %v, seems too large (input was %v)", result, tt.inputSize)
			}
		})
	}
}

// Helper functions
func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrString(s string) *string {
	return &s
}
