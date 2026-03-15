package app

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal.txt", "normal.txt"},
		{"file/with/slashes.txt", "file-with-slashes.txt"},
		{"file:with:colons.txt", "file-with-colons.txt"},
		{"file*with*stars.txt", "file-with-stars.txt"},
		{"file?with?questions.txt", "file-with-questions.txt"},
		{`file"with"quotes.txt`, "file'with'quotes.txt"},
		{"file<with>brackets.txt", "file-with-brackets.txt"},
		{"file|with|pipes.txt", "file-with-pipes.txt"},
		{"file\\with\\backslashes.txt", "file-with-backslashes.txt"},
	}

	for _, tt := range tests {
		result := SanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeFilenameMaxLength(t *testing.T) {
	// Test that long filenames are truncated
	longName := ""
	for i := 0; i < 200; i++ {
		longName += "a"
	}
	result := SanitizeFilename(longName)
	if len(result) > 100 {
		t.Errorf("SanitizeFilename length = %d, want <= 100", len(result))
	}
}
