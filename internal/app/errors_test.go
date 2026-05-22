package app

import (
	"errors"
	"testing"
)

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "download not found",
			err:      ErrDownloadNotFound,
			expected: true,
		},
		{
			name:     "video not found",
			err:      ErrVideoNotFound,
			expected: true,
		},
		{
			name:     "file not found",
			err:      ErrFileNotFound,
			expected: true,
		},
		{
			name:     "wrapped download not found",
			err:      errors.New("wrapped: " + ErrDownloadNotFound.Error()),
			expected: false, // errors.Is would be needed for wrapped errors
		},
		{
			name:     "database not initialized",
			err:      ErrDatabaseNotInitialized,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "random error",
			err:      errors.New("something went wrong"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFoundError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsInfrastructureError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "database not initialized",
			err:      ErrDatabaseNotInitialized,
			expected: true,
		},
		{
			name:     "ytdl not initialized",
			err:      ErrYtdlNotInitialized,
			expected: true,
		},
		{
			name:     "config not initialized",
			err:      ErrConfigNotInitialized,
			expected: true,
		},
		{
			name:     "logger not initialized",
			err:      ErrLoggerNotInitialized,
			expected: false, // not included in IsInfrastructureError
		},
		{
			name:     "download not found",
			err:      ErrDownloadNotFound,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "random error",
			err:      errors.New("something went wrong"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInfrastructureError(tt.err)
			if result != tt.expected {
				t.Errorf("IsInfrastructureError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Verify all expected error constants exist and are non-nil
	errors := []error{
		ErrDatabaseNotInitialized,
		ErrYtdlNotInitialized,
		ErrConfigNotInitialized,
		ErrLoggerNotInitialized,
		ErrDownloadNotFound,
		ErrDownloadAlreadyExists,
		ErrInvalidURL,
		ErrDownloadFailed,
		ErrDownloadPaused,
		ErrDownloadCancelled,
		ErrVideoNotFound,
		ErrFileNotFound,
		ErrFileNotManaged,
		ErrInvalidSetting,
		ErrPresetNotFound,
		ErrDiskFull,
		ErrPathNotWritable,
		ErrSystemDirectory,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("Expected error constant to be non-nil")
		}
		if err.Error() == "" {
			t.Error("Expected error message to be non-empty")
		}
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{ErrDatabaseNotInitialized, "database not initialized"},
		{ErrYtdlNotInitialized, "yt-dlp client not initialized"},
		{ErrConfigNotInitialized, "config not initialized"},
		{ErrDownloadNotFound, "download not found"},
		{ErrVideoNotFound, "video not found"},
		{ErrFileNotFound, "video file not found"},
		{ErrInvalidURL, "invalid YouTube URL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Error message = %q, want %q", tt.err.Error(), tt.expected)
			}
		})
	}
}
