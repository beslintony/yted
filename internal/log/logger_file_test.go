package log

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// waitForLogLines polls the session log file until it holds at least n lines
// (writes are async via goroutine) or the timeout expires.
func waitForLogLines(t *testing.T, path string, n int) []byte {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		content, err := os.ReadFile(path)
		if err == nil {
			lines := 0
			for _, b := range content {
				if b == '\n' {
					lines++
				}
			}
			if lines >= n {
				return content
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %d log lines in %s", n, path)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestSingleHandleNWrites(t *testing.T) {
	dir := t.TempDir()
	logger := &Logger{
		entries:     make([]LogEntry, 0, 1000),
		maxEntries:  10000,
		maxSessions: 10,
	}
	if err := logger.SetLogDir(dir); err != nil {
		t.Fatalf("SetLogDir() error: %v", err)
	}
	t.Cleanup(func() { _ = logger.Close() })

	const n = 50
	for i := 0; i < n; i++ {
		logger.Info("Test", "message", map[string]int{"i": i})
	}

	sessionDir := logger.GetCurrentSessionDir()
	if sessionDir == "" {
		t.Fatal("session dir is empty")
	}
	logPath := filepath.Join(sessionDir, "app.log")
	content := waitForLogLines(t, logPath, n)

	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	if lines != n {
		t.Errorf("log file has %d lines, want %d", lines, n)
	}

	// Every line must be a valid JSON log entry
	lineStart := 0
	seen := make(map[int]bool)
	for i, b := range content {
		if b != '\n' {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal(content[lineStart:i], &entry); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", lines, err)
		}
		var data map[string]int
		if err := json.Unmarshal(entry.Data, &data); err != nil {
			t.Fatalf("line %d data is not valid JSON: %v", lines, err)
		}
		seen[data["i"]] = true
		lineStart = i + 1
	}
	if len(seen) != n {
		t.Errorf("saw %d distinct entries, want %d (all writes must land)", len(seen), n)
	}

	// A single handle must be open for the session (not one open per line)
	logger.fileMu.Lock()
	openPath := logger.logFilePath
	hasHandle := logger.logFile != nil
	logger.fileMu.Unlock()
	if !hasHandle {
		t.Error("expected an open log file handle after writes")
	}
	if openPath != logPath {
		t.Errorf("open handle path = %q, want %q", openPath, logPath)
	}
}

func TestLogFileReopensOnSessionChange(t *testing.T) {
	dir := t.TempDir()
	logger := &Logger{
		entries:     make([]LogEntry, 0, 1000),
		maxEntries:  10000,
		maxSessions: 10,
	}
	if err := logger.SetLogDir(dir); err != nil {
		t.Fatalf("SetLogDir() error: %v", err)
	}
	t.Cleanup(func() { _ = logger.Close() })

	logger.Info("Test", "first")
	firstDir := logger.GetCurrentSessionDir()
	waitForLogLines(t, filepath.Join(firstDir, "app.log"), 1)

	// Switching sessions must close the old handle and lazily open a new one.
	// Note: session names have 1s resolution, so two rapid SetLogDir calls
	// may land in the same session dir — expect 2 lines then, else 1.
	if err := logger.SetLogDir(dir); err != nil {
		t.Fatalf("second SetLogDir() error: %v", err)
	}
	logger.Info("Test", "second")
	secondDir := logger.GetCurrentSessionDir()
	wantLines := 1
	if secondDir == firstDir {
		wantLines = 2
	}
	waitForLogLines(t, filepath.Join(secondDir, "app.log"), wantLines)

	logger.fileMu.Lock()
	openPath := logger.logFilePath
	logger.fileMu.Unlock()
	if openPath != filepath.Join(secondDir, "app.log") {
		t.Errorf("open handle path = %q, want current session file", openPath)
	}
}
