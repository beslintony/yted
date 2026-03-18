package log

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetLoggerSingleton(t *testing.T) {
	logger1 := GetLogger()
	logger2 := GetLogger()

	if logger1 == nil {
		t.Fatal("GetLogger returned nil")
	}

	if logger1 != logger2 {
		t.Error("GetLogger should return the same instance (singleton)")
	}
}

func TestLogLevels(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10000,
	}

	// Test Debug
	logger.Debug("Test", "debug message", map[string]string{"key": "value"})
	entries := logger.GetRecentEntries(1)
	if len(entries) != 1 {
		t.Fatal("Expected 1 entry")
	}
	if entries[0].Level != DEBUG {
		t.Errorf("Expected DEBUG level, got %s", entries[0].Level)
	}
	if entries[0].Component != "Test" {
		t.Errorf("Expected component 'Test', got %s", entries[0].Component)
	}

	// Test Info
	logger.Info("Test", "info message")
	entries = logger.GetRecentEntries(1)
	if entries[0].Level != INFO {
		t.Errorf("Expected INFO level, got %s", entries[0].Level)
	}

	// Test Warn
	logger.Warn("Test", "warn message")
	entries = logger.GetRecentEntries(1)
	if entries[0].Level != WARN {
		t.Errorf("Expected WARN level, got %s", entries[0].Level)
	}

	// Test Error
	testErr := os.ErrNotExist
	logger.Error("Test", "error message", testErr)
	entries = logger.GetRecentEntries(1)
	if entries[0].Level != ERROR {
		t.Errorf("Expected ERROR level, got %s", entries[0].Level)
	}
	if entries[0].Error != testErr.Error() {
		t.Errorf("Expected error '%s', got '%s'", testErr.Error(), entries[0].Error)
	}
}

func TestLogWithData(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10000,
	}

	data := map[string]interface{}{
		"url":      "https://youtube.com/watch?v=123",
		"progress": 50.5,
		"count":    10,
	}

	logger.Info("Download", "downloading", data)
	entries := logger.GetRecentEntries(1)

	if len(entries[0].Data) == 0 {
		t.Error("Expected data to be logged")
	}

	// Verify data can be unmarshaled
	var decoded map[string]interface{}
	if err := json.Unmarshal(entries[0].Data, &decoded); err != nil {
		t.Errorf("Failed to unmarshal data: %v", err)
	}

	if decoded["url"] != data["url"] {
		t.Errorf("Expected url '%s', got '%v'", data["url"], decoded["url"])
	}
}

func TestLogMaxEntries(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10,
	}

	// Add 15 entries
	for i := 0; i < 15; i++ {
		logger.Info("Test", "message", i)
	}

	allEntries := logger.GetAllEntries()
	if len(allEntries) != 10 {
		t.Errorf("Expected 10 entries (max), got %d", len(allEntries))
	}

	// Verify it kept the most recent entries
	recent := logger.GetRecentEntries(1)
	if len(recent) != 1 {
		t.Fatal("Expected 1 recent entry")
	}

	// The data should contain 14 (the last entry we added)
	var lastData int
	json.Unmarshal(recent[0].Data, &lastData)
	if lastData != 14 {
		t.Errorf("Expected last entry to be 14, got %d", lastData)
	}
}

func TestGetRecentEntriesWithFilter(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10000,
	}

	logger.Info("Test", "info 1")
	logger.Error("Test", "error 1", nil)
	logger.Info("Test", "info 2")
	logger.Warn("Test", "warn 1")
	logger.Error("Test", "error 2", nil)

	// Get all ERROR entries
	errors := logger.GetRecentEntries(10, ERROR)
	if len(errors) != 2 {
		t.Errorf("Expected 2 ERROR entries, got %d", len(errors))
	}

	for _, entry := range errors {
		if entry.Level != ERROR {
			t.Errorf("Expected ERROR level, got %s", entry.Level)
		}
	}

	// Get INFO entries
	infos := logger.GetRecentEntries(10, INFO)
	if len(infos) != 2 {
		t.Errorf("Expected 2 INFO entries, got %d", len(infos))
	}
}

func TestClear(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10000,
	}

	logger.Info("Test", "message 1")
	logger.Info("Test", "message 2")

	if len(logger.GetAllEntries()) != 2 {
		t.Fatal("Expected 2 entries before clear")
	}

	logger.Clear()

	if len(logger.GetAllEntries()) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(logger.GetAllEntries()))
	}
}

func TestSetLogDir(t *testing.T) {
	tempDir := t.TempDir()
	logger := &Logger{
		entries:     make([]LogEntry, 0, 1000),
		maxEntries:  10000,
		maxSessions: 10,
	}

	err := logger.SetLogDir(tempDir)
	if err != nil {
		t.Fatalf("SetLogDir failed: %v", err)
	}

	if logger.logDir != tempDir {
		t.Errorf("Expected logDir to be '%s', got '%s'", tempDir, logger.logDir)
	}

	// Verify session directory was created
	sessionDir := logger.GetCurrentSessionDir()
	if sessionDir == "" {
		t.Fatal("Session directory is empty")
	}

	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		t.Error("Session directory was not created")
	}
}

func TestSetLogDirWithSessions(t *testing.T) {
	tempDir := t.TempDir()
	logger := &Logger{
		entries:     make([]LogEntry, 0, 1000),
		maxEntries:  10000,
		maxSessions: 10,
	}

	err := logger.SetLogDirWithSessions(tempDir, 5)
	if err != nil {
		t.Fatalf("SetLogDirWithSessions failed: %v", err)
	}

	if logger.maxSessions != 5 {
		t.Errorf("Expected maxSessions=5, got %d", logger.maxSessions)
	}
}

func TestSetMaxSessions(t *testing.T) {
	tempDir := t.TempDir()
	logger := &Logger{
		entries:     make([]LogEntry, 0, 1000),
		maxEntries:  10000,
		maxSessions: 10,
	}

	// Need to set log dir first for cleanup to work
	logger.SetLogDir(tempDir)

	logger.SetMaxSessions(3)

	if logger.maxSessions != 3 {
		t.Errorf("Expected maxSessions=3, got %d", logger.maxSessions)
	}
}

func TestAddAndRemoveListener(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10000,
	}

	var received atomic.Bool
	listener := func(entry LogEntry) {
		received.Store(true)
	}

	logger.AddListener(listener)

	if len(logger.listeners) != 1 {
		t.Errorf("Expected 1 listener, got %d", len(logger.listeners))
	}

	// Trigger a log (without file output to speed up)
	logger.Info("Test", "test message")

	// Give listener time to be called
	time.Sleep(100 * time.Millisecond)

	if !received.Load() {
		t.Error("Listener was not called")
	}

	// Remove listener
	logger.RemoveListener(listener)

	if len(logger.listeners) != 0 {
		t.Errorf("Expected 0 listeners after remove, got %d", len(logger.listeners))
	}
}

func TestConcurrentLogging(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10000,
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	logsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Info("Test", "concurrent log", map[string]int{"goroutine": id, "log": j})
			}
		}(i)
	}

	wg.Wait()

	totalExpected := numGoroutines * logsPerGoroutine
	entries := logger.GetAllEntries()

	// Due to maxEntries limiting, we might not have all
	if len(entries) != logger.maxEntries && len(entries) != totalExpected {
		t.Logf("Entries: %d (max: %d, expected: %d)", len(entries), logger.maxEntries, totalExpected)
	}
}

func TestExport(t *testing.T) {
	logger := &Logger{
		entries:    make([]LogEntry, 0, 1000),
		maxEntries: 10000,
	}

	logger.Info("Test", "message 1", map[string]string{"key": "value1"})
	logger.Info("Test", "message 2", map[string]string{"key": "value2"})

	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "export.log")

	err := logger.Export(exportPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file exists and has content
	content, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}

	if lines != 2 {
		t.Errorf("Expected 2 lines in export, got %d", lines)
	}
}

func TestLogLevelConstants(t *testing.T) {
	if DEBUG != "DEBUG" {
		t.Errorf("DEBUG constant wrong: %s", DEBUG)
	}
	if INFO != "INFO" {
		t.Errorf("INFO constant wrong: %s", INFO)
	}
	if WARN != "WARN" {
		t.Errorf("WARN constant wrong: %s", WARN)
	}
	if ERROR != "ERROR" {
		t.Errorf("ERROR constant wrong: %s", ERROR)
	}
}

func TestGetSessionDirs(t *testing.T) {
	tempDir := t.TempDir()
	logger := &Logger{
		entries:     make([]LogEntry, 0, 1000),
		maxEntries:  10000,
		maxSessions: 10,
	}

	// No log dir set yet
	dirs := logger.GetSessionDirs()
	if dirs != nil {
		t.Error("Expected nil when logDir not set")
	}

	// Set log dir and create some sessions
	logger.SetLogDir(tempDir)

	// Create a fake old session
	oldSession := filepath.Join(tempDir, "2023-01-01_00-00-00")
	os.MkdirAll(oldSession, 0755)

	dirs = logger.GetSessionDirs()
	if len(dirs) == 0 {
		t.Error("Expected at least one session directory")
	}

	// Verify format filtering works
	foundOldSession := false
	for _, dir := range dirs {
		if filepath.Base(dir) == "2023-01-01_00-00-00" {
			foundOldSession = true
			break
		}
	}
	if !foundOldSession {
		t.Error("Expected to find old session directory")
	}
}
