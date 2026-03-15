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
	// First, handle schema changes for existing databases
	if err := db.migrateSchema(); err != nil {
		return fmt.Errorf("schema migration failed: %w", err)
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS videos (
			id TEXT PRIMARY KEY,
			youtube_id TEXT NOT NULL,
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
		// Schema migrations for existing databases
		`ALTER TABLE videos ADD COLUMN file_hash TEXT`,
		`ALTER TABLE videos ADD COLUMN is_managed BOOLEAN DEFAULT 1`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			// Ignore errors for optional migrations (e.g., column already exists)
			// Log but continue - this allows idempotent migrations
			continue
		}
	}

	return nil
}

// migrateSchema handles complex schema migrations
func (db *DB) migrateSchema() error {
	// Check if we need to remove the UNIQUE constraint from youtube_id
	// This is needed for supporting multiple versions of the same video

	// First, check if the table exists and has the constraint
	var count int
	err := db.conn.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='videos'",
	).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Table doesn't exist yet, will be created with correct schema
		return nil
	}

	// Check if file_hash column exists
	var hasFileHash int
	err = db.conn.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('videos') WHERE name='file_hash'",
	).Scan(&hasFileHash)
	if err != nil {
		return err
	}

	// Check if is_managed column exists
	var hasIsManaged int
	err = db.conn.QueryRow(
		"SELECT COUNT(*) FROM pragma_table_info('videos') WHERE name='is_managed'",
	).Scan(&hasIsManaged)
	if err != nil {
		return err
	}

	// If both columns exist, nothing to do
	if hasFileHash > 0 && hasIsManaged > 0 {
		return nil
	}

	// Need to recreate table to remove UNIQUE constraint and add columns
	// This is a complex migration - we'll do it in a transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create new table with correct schema
	_, err = tx.Exec(`
		CREATE TABLE videos_new (
			id TEXT PRIMARY KEY,
			youtube_id TEXT NOT NULL,
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
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// Copy data from old table
	_, err = tx.Exec(`
		INSERT INTO videos_new (
			id, youtube_id, title, channel, channel_id, duration, description,
			thumbnail_url, file_path, file_size, format, quality, downloaded_at,
			watch_position, watch_count
		)
		SELECT 
			id, youtube_id, title, channel, channel_id, duration, description,
			thumbnail_url, file_path, file_size, format, quality, downloaded_at,
			watch_position, watch_count
		FROM videos
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE videos")
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Rename new table
	_, err = tx.Exec("ALTER TABLE videos_new RENAME TO videos")
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	return tx.Commit()
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
