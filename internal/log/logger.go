package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp  time.Time       `json:"timestamp"`
	Level      LogLevel        `json:"level"`
	Component  string          `json:"component"`
	Message    string          `json:"message"`
	Data       json.RawMessage `json:"data,omitempty"`
	Error      string          `json:"error,omitempty"`
	StackTrace string          `json:"stack_trace,omitempty"`
}

type Logger struct {
	mu             sync.RWMutex
	entries        []LogEntry
	maxEntries     int
	logDir         string
	maxSessions    int
	currentSession string
	listeners      []func(LogEntry)
}

var (
	instance *Logger
	once     sync.Once
)

// GetLogger returns the singleton logger instance
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			entries:     make([]LogEntry, 0, 1000),
			maxEntries:  10000,
			maxSessions: 10,
			listeners:   make([]func(LogEntry), 0),
		}
	})
	return instance
}

// SetLogDir sets the directory for log file output and initializes a new session
func (l *Logger) SetLogDir(dir string) error {
	return l.SetLogDirWithSessions(dir, 10)
}

// SetLogDirWithSessions sets the log directory and configures session retention
func (l *Logger) SetLogDirWithSessions(dir string, maxSessions int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logDir = dir
	l.maxSessions = maxSessions

	// Create session timestamp
	l.currentSession = time.Now().Format("2006-01-02_15-04-05")

	// Ensure log directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create session subdirectory
	sessionDir := filepath.Join(dir, l.currentSession)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Cleanup old sessions
	go l.cleanupOldSessions()

	return nil
}

// SetMaxSessions updates the maximum number of sessions to keep
func (l *Logger) SetMaxSessions(count int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.maxSessions = count
	go l.cleanupOldSessions()
}

// cleanupOldSessions removes old session directories, keeping only maxSessions
func (l *Logger) cleanupOldSessions() {
	if l.logDir == "" || l.maxSessions <= 0 {
		return
	}

	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return
	}

	// Collect session directories
	var sessions []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) == 19 && entry.Name()[10] == '_' { // Format: 2006-01-02_15-04-05
			sessions = append(sessions, entry)
		}
	}

	// If we have more sessions than allowed, delete the oldest ones
	if len(sessions) > l.maxSessions {
		// Sort by name (chronological since format is YYYY-MM-DD_HH-MM-SS)
		for i := 0; i < len(sessions)-1; i++ {
			for j := i + 1; j < len(sessions); j++ {
				if sessions[i].Name() > sessions[j].Name() {
					sessions[i], sessions[j] = sessions[j], sessions[i]
				}
			}
		}

		// Delete oldest sessions
		toDelete := len(sessions) - l.maxSessions
		for i := 0; i < toDelete; i++ {
			sessionDir := filepath.Join(l.logDir, sessions[i].Name())
			os.RemoveAll(sessionDir)
			fmt.Printf("[Logger] Cleaned up old session: %s\n", sessions[i].Name())
		}
	}
}

// GetCurrentSessionDir returns the current session's log directory
func (l *Logger) GetCurrentSessionDir() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.logDir == "" || l.currentSession == "" {
		return ""
	}
	return filepath.Join(l.logDir, l.currentSession)
}

// AddListener adds a callback for new log entries (for frontend notifications)
func (l *Logger) AddListener(fn func(LogEntry)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.listeners = append(l.listeners, fn)
}

// RemoveListener removes a callback
func (l *Logger) RemoveListener(fn func(LogEntry)) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, listener := range l.listeners {
		if fmt.Sprintf("%p", listener) == fmt.Sprintf("%p", fn) {
			l.listeners = append(l.listeners[:i], l.listeners[i+1:]...)
			break
		}
	}
}

// log creates a log entry and notifies listeners
func (l *Logger) log(level LogLevel, component, message string, data interface{}, err error) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   message,
	}

	if data != nil {
		if jsonData, err := json.Marshal(data); err == nil {
			entry.Data = jsonData
		}
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.mu.Lock()

	// Add to in-memory buffer
	l.entries = append(l.entries, entry)

	// Trim if exceeding max
	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[len(l.entries)-l.maxEntries:]
	}

	// Notify listeners
	listeners := make([]func(LogEntry), len(l.listeners))
	copy(listeners, l.listeners)

	l.mu.Unlock()

	// Notify outside of lock
	for _, fn := range listeners {
		go fn(entry)
	}

	// Also write to file if configured
	if l.logDir != "" {
		go l.writeToFile(entry)
	}

	// Also print to stdout for development
	fmt.Printf("[%s] %s - %s: %s\n", entry.Timestamp.Format("15:04:05"), level, component, message)
}

func (l *Logger) writeToFile(entry LogEntry) {
	l.mu.RLock()
	logDir := l.logDir
	session := l.currentSession
	l.mu.RUnlock()

	if logDir == "" {
		return
	}

	// Use session subdirectory, fallback to main log dir if no session
	var filename string
	if session != "" {
		sessionDir := filepath.Join(logDir, session)
		filename = filepath.Join(sessionDir, "app.log")
	} else {
		filename = filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	jsonEntry, _ := json.Marshal(entry)
	file.Write(jsonEntry)
	file.Write([]byte("\n"))
}

// GetSessionDirs returns all session directories
func (l *Logger) GetSessionDirs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.logDir == "" {
		return nil
	}

	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return nil
	}

	var sessions []string
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) == 19 && entry.Name()[10] == '_' {
			sessions = append(sessions, filepath.Join(l.logDir, entry.Name()))
		}
	}

	// Sort newest first
	for i, j := 0, len(sessions)-1; i < j; i, j = i+1, j-1 {
		sessions[i], sessions[j] = sessions[j], sessions[i]
	}

	return sessions
}

// Debug logs a debug message
func (l *Logger) Debug(component, message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.log(DEBUG, component, message, d, nil)
}

// Info logs an info message
func (l *Logger) Info(component, message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.log(INFO, component, message, d, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(component, message string, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.log(WARN, component, message, d, nil)
}

// Error logs an error message
func (l *Logger) Error(component, message string, err error, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	l.log(ERROR, component, message, d, err)
}

// GetRecentEntries returns the most recent log entries
func (l *Logger) GetRecentEntries(count int, level ...LogLevel) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if count > len(l.entries) {
		count = len(l.entries)
	}

	entries := l.entries[len(l.entries)-count:]

	// Filter by level if specified
	if len(level) > 0 {
		filtered := make([]LogEntry, 0, len(entries))
		levelMap := make(map[LogLevel]bool)
		for _, l := range level {
			levelMap[l] = true
		}

		for _, entry := range entries {
			if levelMap[entry.Level] {
				filtered = append(filtered, entry)
			}
		}
		return filtered
	}

	return entries
}

// GetAllEntries returns all log entries
func (l *Logger) GetAllEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]LogEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Clear clears all log entries
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = make([]LogEntry, 0, 1000)
}

// Export exports logs to a file
func (l *Logger) Export(filepath string) error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, entry := range l.entries {
		jsonEntry, _ := json.Marshal(entry)
		file.Write(jsonEntry)
		file.Write([]byte("\n"))
	}

	return nil
}
