package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// EditJob represents a video editing job
type EditJob struct {
	ID            string     `json:"id"`
	SourceVideoID string     `json:"source_video_id"` // Original video
	OutputVideoID *string    `json:"output_video_id"` // Result video (null until complete)
	Status        string     `json:"status"`          // pending, processing, completed, error
	Operation     string     `json:"operation"`       // crop, watermark, convert, effects, combine
	Settings      string     `json:"settings"`        // JSON-encoded EditSettings
	Progress      float64    `json:"progress"`
	ErrorMessage  *string    `json:"error_message"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

// EditSettings represents edit operation settings
type EditSettings struct {
	// Crop settings
	CropStart  *float64 `json:"crop_start,omitempty"`  // Start time in seconds
	CropEnd    *float64 `json:"crop_end,omitempty"`    // End time in seconds
	CropX      *int     `json:"crop_x,omitempty"`      // Crop region X
	CropY      *int     `json:"crop_y,omitempty"`      // Crop region Y
	CropWidth  *int     `json:"crop_width,omitempty"`  // Crop region width
	CropHeight *int     `json:"crop_height,omitempty"` // Crop region height

	// Watermark settings
	WatermarkType     *string  `json:"watermark_type,omitempty"`     // "text" or "image"
	WatermarkText     *string  `json:"watermark_text,omitempty"`
	WatermarkImage    *string  `json:"watermark_image,omitempty"`    // Path to image
	WatermarkPosition *string  `json:"watermark_position,omitempty"` // "top-left", "top-right", etc.
	WatermarkOpacity  *float64 `json:"watermark_opacity,omitempty"`  // 0-1
	WatermarkSize     *int     `json:"watermark_size,omitempty"`     // Font size or image scale

	// Convert settings
	OutputFormat     *string `json:"output_format,omitempty"`      // mp4, avi, mkv, mov, webm, gif
	OutputCodec      *string `json:"output_codec,omitempty"`       // h264, h265, vp9, av1
	OutputQuality    *int    `json:"output_quality,omitempty"`     // CRF value (18-28)
	OutputResolution *string `json:"output_resolution,omitempty"`  // "original", "1080p", "720p", etc.

	// Effects
	Brightness  *float64 `json:"brightness,omitempty"`   // -1 to 1
	Contrast    *float64 `json:"contrast,omitempty"`     // -1 to 1
	Saturation  *float64 `json:"saturation,omitempty"`   // 0 to 2
	Rotation    *int     `json:"rotation,omitempty"`     // 0, 90, 180, 270
	Speed       *float64 `json:"speed,omitempty"`        // 0.5 to 2.0
	Volume      *float64 `json:"volume,omitempty"`       // 0 to 2
	RemoveAudio *bool    `json:"remove_audio,omitempty"`

	// Output options
	OutputFilename  *string `json:"output_filename,omitempty"`
	ReplaceOriginal *bool   `json:"replace_original,omitempty"`
}

// CreateEditTables creates the edit_jobs table
func (db *DB) CreateEditTables() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS edit_jobs (
			id TEXT PRIMARY KEY,
			source_video_id TEXT NOT NULL,
			output_video_id TEXT,
			status TEXT DEFAULT 'pending',
			operation TEXT NOT NULL,
			settings TEXT,
			progress REAL DEFAULT 0,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (source_video_id) REFERENCES videos(id),
			FOREIGN KEY (output_video_id) REFERENCES videos(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_edit_jobs_source_video ON edit_jobs(source_video_id)`,
		`CREATE INDEX IF NOT EXISTS idx_edit_jobs_status ON edit_jobs(status)`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("failed to create edit tables: %w", err)
		}
	}
	return nil
}

// CreateEditJob creates a new edit job
func (db *DB) CreateEditJob(job *EditJob) error {
	query := `
		INSERT INTO edit_jobs (id, source_video_id, output_video_id, status, operation, settings, progress, error_message, created_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		job.ID, job.SourceVideoID, job.OutputVideoID, job.Status, job.Operation, job.Settings,
		job.Progress, job.ErrorMessage, job.CreatedAt, job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create edit job: %w", err)
	}
	return nil
}

// GetEditJob retrieves an edit job by ID
func (db *DB) GetEditJob(id string) (*EditJob, error) {
	query := `
		SELECT id, source_video_id, output_video_id, status, operation, settings, progress, error_message, created_at, completed_at
		FROM edit_jobs WHERE id = ?
	`
	row := db.conn.QueryRow(query, id)

	var job EditJob
	var outputVideoID, errorMessage sql.NullString
	var completedAt sql.NullTime

	err := row.Scan(
		&job.ID, &job.SourceVideoID, &outputVideoID, &job.Status, &job.Operation,
		&job.Settings, &job.Progress, &errorMessage, &job.CreatedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edit job: %w", err)
	}

	if outputVideoID.Valid {
		job.OutputVideoID = &outputVideoID.String
	}
	if errorMessage.Valid {
		job.ErrorMessage = &errorMessage.String
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// UpdateEditJob updates an edit job
func (db *DB) UpdateEditJob(job *EditJob) error {
	query := `
		UPDATE edit_jobs SET
			status = ?, settings = ?, progress = ?, error_message = ?, completed_at = ?, output_video_id = ?
		WHERE id = ?
	`
	_, err := db.conn.Exec(query,
		job.Status, job.Settings, job.Progress, job.ErrorMessage,
		job.CompletedAt, job.OutputVideoID, job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update edit job: %w", err)
	}
	return nil
}

// UpdateEditJobProgress updates only the progress of an edit job
func (db *DB) UpdateEditJobProgress(id string, progress float64) error {
	query := `UPDATE edit_jobs SET progress = ? WHERE id = ?`
	_, err := db.conn.Exec(query, progress, id)
	if err != nil {
		return fmt.Errorf("failed to update edit job progress: %w", err)
	}
	return nil
}

// CompleteEditJob marks an edit job as completed
func (db *DB) CompleteEditJob(id string, outputVideoID string) error {
	now := time.Now()
	query := `UPDATE edit_jobs SET status = 'completed', output_video_id = ?, completed_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, outputVideoID, now, id)
	if err != nil {
		return fmt.Errorf("failed to complete edit job: %w", err)
	}
	return nil
}

// FailEditJob marks an edit job as failed
func (db *DB) FailEditJob(id string, errorMessage string) error {
	now := time.Now()
	query := `UPDATE edit_jobs SET status = 'error', error_message = ?, completed_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, errorMessage, now, id)
	if err != nil {
		return fmt.Errorf("failed to fail edit job: %w", err)
	}
	return nil
}

// ListEditJobs retrieves all edit jobs for a video
func (db *DB) ListEditJobs(videoID string) ([]EditJob, error) {
	query := `
		SELECT id, source_video_id, output_video_id, status, operation, settings, progress, error_message, created_at, completed_at
		FROM edit_jobs WHERE source_video_id = ?
		ORDER BY created_at DESC
	`
	rows, err := db.conn.Query(query, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to list edit jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []EditJob
	for rows.Next() {
		var job EditJob
		var outputVideoID, errorMessage sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&job.ID, &job.SourceVideoID, &outputVideoID, &job.Status, &job.Operation,
			&job.Settings, &job.Progress, &errorMessage, &job.CreatedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edit job: %w", err)
		}

		if outputVideoID.Valid {
			job.OutputVideoID = &outputVideoID.String
		}
		if errorMessage.Valid {
			job.ErrorMessage = &errorMessage.String
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// DeleteEditJob deletes an edit job
func (db *DB) DeleteEditJob(id string) error {
	query := `DELETE FROM edit_jobs WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete edit job: %w", err)
	}
	return nil
}

// GetActiveEditJobs retrieves all active (pending or processing) edit jobs
func (db *DB) GetActiveEditJobs() ([]EditJob, error) {
	query := `
		SELECT id, source_video_id, output_video_id, status, operation, settings, progress, error_message, created_at, completed_at
		FROM edit_jobs WHERE status IN ('pending', 'processing')
		ORDER BY created_at ASC
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active edit jobs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var jobs []EditJob
	for rows.Next() {
		var job EditJob
		var outputVideoID, errorMessage sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&job.ID, &job.SourceVideoID, &outputVideoID, &job.Status, &job.Operation,
			&job.Settings, &job.Progress, &errorMessage, &job.CreatedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan edit job: %w", err)
		}

		if outputVideoID.Valid {
			job.OutputVideoID = &outputVideoID.String
		}
		if errorMessage.Valid {
			job.ErrorMessage = &errorMessage.String
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// SettingsToJSON converts EditSettings to JSON string
func SettingsToJSON(settings EditSettings) (string, error) {
	data, err := json.Marshal(settings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %w", err)
	}
	return string(data), nil
}

// SettingsFromJSON parses JSON string to EditSettings
func SettingsFromJSON(data string) (EditSettings, error) {
	var settings EditSettings
	if err := json.Unmarshal([]byte(data), &settings); err != nil {
		return settings, fmt.Errorf("failed to unmarshal settings: %w", err)
	}
	return settings, nil
}
