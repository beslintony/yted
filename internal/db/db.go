package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the database connection
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(appDataDir string) (*DB, error) {
	dbPath := filepath.Join(appDataDir, "yted.db")
	conn, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate runs database migrations
func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS videos (
			id TEXT PRIMARY KEY,
			youtube_id TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			channel TEXT,
			channel_id TEXT,
			duration INTEGER,
			description TEXT,
			thumbnail_url TEXT,
			file_path TEXT NOT NULL,
			file_size INTEGER,
			file_hash TEXT,
			is_managed BOOLEAN DEFAULT 1,
			format TEXT,
			quality TEXT,
			downloaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			watch_position INTEGER DEFAULT 0,
			watch_count INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS downloads (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			progress REAL DEFAULT 0,
			title TEXT,
			channel TEXT,
			thumbnail_url TEXT,
			format_id TEXT,
			quality TEXT,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_videos_channel ON videos(channel)`,
		`CREATE INDEX IF NOT EXISTS idx_videos_downloaded_at ON videos(downloaded_at)`,
		`CREATE INDEX IF NOT EXISTS idx_downloads_status ON downloads(status)`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// Video represents a downloaded video
type Video struct {
	ID             string    `json:"id"`
	YoutubeID      string    `json:"youtube_id"`
	Title          string    `json:"title"`
	Channel        string    `json:"channel"`
	ChannelID      string    `json:"channel_id"`
	Duration       int       `json:"duration"`
	Description    string    `json:"description"`
	ThumbnailURL   string    `json:"thumbnail_url"`
	FilePath       string    `json:"file_path"`
	FileSize       int64     `json:"file_size"`
	FileHash       string    `json:"file_hash"`    // Unique content ID: YouTube ID + Format (allows multiple versions)
	IsManaged      bool      `json:"is_managed"`   // Whether file is in YTed managed folder
	Format         string    `json:"format"`
	Quality        string    `json:"quality"`
	DownloadedAt   time.Time `json:"downloaded_at"`
	WatchPosition  int       `json:"watch_position"`
	WatchCount     int       `json:"watch_count"`
}

// Download represents a download job
type Download struct {
	ID            string     `json:"id"`
	URL           string     `json:"url"`
	Status        string     `json:"status"`
	Progress      float64    `json:"progress"`
	Title         *string    `json:"title"`
	Channel       *string    `json:"channel"`
	ThumbnailURL  *string    `json:"thumbnail_url"`
	FormatID      *string    `json:"format_id"`
	Quality       *string    `json:"quality"`
	ErrorMessage  *string    `json:"error_message"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

// format helpers for nullable fields
func stringPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func timePtr(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}
