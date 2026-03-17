package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CreateDownload adds a new download job
func (db *DB) CreateDownload(download *Download) error {
	query := `
		INSERT INTO downloads (id, url, status, progress, title, channel, 
			thumbnail_url, format_id, quality, duration, error_message, created_at, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		download.ID, download.URL, download.Status, download.Progress,
		download.Title, download.Channel, download.ThumbnailURL,
		download.FormatID, download.Quality, download.Duration, download.ErrorMessage,
		download.CreatedAt, download.StartedAt, download.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create download: %w", err)
	}
	return nil
}

// GetActiveDownloadByURL checks if there's an active download for the given URL
// Returns the download if found, nil if not found
func (db *DB) GetActiveDownloadByURL(url string) (*Download, error) {
	query := `
		SELECT id, url, status, progress, title, channel, thumbnail_url,
			format_id, quality, duration, error_message, created_at, started_at, completed_at
		FROM downloads
		WHERE url = ? AND status IN ('pending', 'downloading')
		LIMIT 1
	`
	row := db.conn.QueryRow(query, url)

	var d Download
	var title, channel, thumbnailURL, formatID, quality, errorMsg sql.NullString
	var duration sql.NullInt64
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&d.ID, &d.URL, &d.Status, &d.Progress,
		&title, &channel, &thumbnailURL, &formatID, &quality, &duration, &errorMsg,
		&d.CreatedAt, &startedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing download: %w", err)
	}

	d.Title = stringPtr(title)
	d.Channel = stringPtr(channel)
	d.ThumbnailURL = stringPtr(thumbnailURL)
	d.FormatID = stringPtr(formatID)
	d.Quality = stringPtr(quality)
	d.Duration = intPtr(duration)
	d.ErrorMessage = stringPtr(errorMsg)
	d.StartedAt = timePtr(startedAt)
	d.CompletedAt = timePtr(completedAt)

	return &d, nil
}

// GetDownload retrieves a download by ID
func (db *DB) GetDownload(id string) (*Download, error) {
	query := `
		SELECT id, url, status, progress, title, channel, thumbnail_url,
			format_id, quality, duration, error_message, created_at, started_at, completed_at
		FROM downloads WHERE id = ?
	`
	row := db.conn.QueryRow(query, id)

	var d Download
	var title, channel, thumbnailURL, formatID, quality, errorMsg sql.NullString
	var duration sql.NullInt64
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&d.ID, &d.URL, &d.Status, &d.Progress,
		&title, &channel, &thumbnailURL, &formatID, &quality, &duration, &errorMsg,
		&d.CreatedAt, &startedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get download: %w", err)
	}

	d.Title = stringPtr(title)
	d.Channel = stringPtr(channel)
	d.ThumbnailURL = stringPtr(thumbnailURL)
	d.FormatID = stringPtr(formatID)
	d.Quality = stringPtr(quality)
	d.Duration = intPtr(duration)
	d.ErrorMessage = stringPtr(errorMsg)
	d.StartedAt = timePtr(startedAt)
	d.CompletedAt = timePtr(completedAt)

	return &d, nil
}

// ListDownloads retrieves downloads filtered by status
func (db *DB) ListDownloads(status ...string) ([]Download, error) {
	var query string
	var args []interface{}

	if len(status) == 0 {
		query = `
			SELECT id, url, status, progress, title, channel, thumbnail_url,
				format_id, quality, duration, error_message, created_at, started_at, completed_at
			FROM downloads
			ORDER BY created_at DESC
		`
	} else {
		placeholders := make([]string, len(status))
		for i, s := range status {
			placeholders[i] = "?"
			args = append(args, s)
		}
		query = fmt.Sprintf(`
			SELECT id, url, status, progress, title, channel, thumbnail_url,
				format_id, quality, duration, error_message, created_at, started_at, completed_at
			FROM downloads
			WHERE status IN (%s)
			ORDER BY created_at DESC
		`, fmt.Sprintf("%s", strings.Join(placeholders, ",")))
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list downloads: %w", err)
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var d Download
		var title, channel, thumbnailURL, formatID, quality, errorMsg sql.NullString
		var duration sql.NullInt64
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&d.ID, &d.URL, &d.Status, &d.Progress,
			&title, &channel, &thumbnailURL, &formatID, &quality, &duration, &errorMsg,
			&d.CreatedAt, &startedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan download: %w", err)
		}

		d.Title = stringPtr(title)
		d.Channel = stringPtr(channel)
		d.ThumbnailURL = stringPtr(thumbnailURL)
		d.FormatID = stringPtr(formatID)
		d.Quality = stringPtr(quality)
		d.Duration = intPtr(duration)
		d.ErrorMessage = stringPtr(errorMsg)
		d.StartedAt = timePtr(startedAt)
		d.CompletedAt = timePtr(completedAt)

		downloads = append(downloads, d)
	}

	return downloads, nil
}

// UpdateDownload updates a download
func (db *DB) UpdateDownload(download *Download) error {
	query := `
		UPDATE downloads SET
			status = ?, progress = ?, title = ?, channel = ?, thumbnail_url = ?,
			format_id = ?, quality = ?, duration = ?, error_message = ?, started_at = ?, completed_at = ?
		WHERE id = ?
	`
	_, err := db.conn.Exec(query,
		download.Status, download.Progress, download.Title, download.Channel,
		download.ThumbnailURL, download.FormatID, download.Quality, download.Duration,
		download.ErrorMessage, download.StartedAt, download.CompletedAt,
		download.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update download: %w", err)
	}
	return nil
}

// UpdateDownloadProgress updates only the progress
func (db *DB) UpdateDownloadProgress(id string, progress float64) error {
	query := `UPDATE downloads SET progress = ? WHERE id = ?`
	_, err := db.conn.Exec(query, progress, id)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}
	return nil
}

// UpdateDownloadStatus updates only the status
func (db *DB) UpdateDownloadStatus(id string, status string) error {
	query := `UPDATE downloads SET status = ? WHERE id = ?`
	_, err := db.conn.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

// StartDownload marks a download as started
func (db *DB) StartDownload(id string) error {
	now := time.Now()
	query := `UPDATE downloads SET status = 'downloading', started_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	return nil
}

// CompleteDownload marks a download as completed
func (db *DB) CompleteDownload(id string) error {
	now := time.Now()
	query := `UPDATE downloads SET status = 'completed', progress = 100, completed_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("failed to complete download: %w", err)
	}
	return nil
}

// FailDownload marks a download as failed
func (db *DB) FailDownload(id string, errorMsg string) error {
	query := `UPDATE downloads SET status = 'error', error_message = ? WHERE id = ?`
	_, err := db.conn.Exec(query, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to fail download: %w", err)
	}
	return nil
}

// DeleteDownload removes a download
func (db *DB) DeleteDownload(id string) error {
	query := `DELETE FROM downloads WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete download: %w", err)
	}
	return nil
}

// DeleteCompletedDownloads removes all completed downloads
func (db *DB) DeleteCompletedDownloads() error {
	query := `DELETE FROM downloads WHERE status = 'completed'`
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete completed downloads: %w", err)
	}
	return nil
}

// GetPendingDownloads returns downloads that should be started
func (db *DB) GetPendingDownloads(limit int) ([]Download, error) {
	query := `
		SELECT id, url, status, progress, title, channel, thumbnail_url,
			format_id, quality, duration, error_message, created_at, started_at, completed_at
		FROM downloads
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT ?
	`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending downloads: %w", err)
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var d Download
		var title, channel, thumbnailURL, formatID, quality, errorMsg sql.NullString
		var duration sql.NullInt64
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&d.ID, &d.URL, &d.Status, &d.Progress,
			&title, &channel, &thumbnailURL, &formatID, &quality, &duration, &errorMsg,
			&d.CreatedAt, &startedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan download: %w", err)
		}

		d.Title = stringPtr(title)
		d.Channel = stringPtr(channel)
		d.ThumbnailURL = stringPtr(thumbnailURL)
		d.FormatID = stringPtr(formatID)
		d.Quality = stringPtr(quality)
		d.Duration = intPtr(duration)
		d.ErrorMessage = stringPtr(errorMsg)
		d.StartedAt = timePtr(startedAt)
		d.CompletedAt = timePtr(completedAt)

		downloads = append(downloads, d)
	}

	return downloads, nil
}

// CountActiveDownloads returns the number of active downloads
func (db *DB) CountActiveDownloads() (int, error) {
	query := `SELECT COUNT(*) FROM downloads WHERE status = 'downloading'`
	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active downloads: %w", err)
	}
	return count, nil
}

// GetIncompleteDownloads returns all downloads that are not completed
func (db *DB) GetIncompleteDownloads() ([]Download, error) {
	query := `
		SELECT id, url, status, progress, title, channel, thumbnail_url,
			format_id, quality, duration, error_message, created_at, started_at, completed_at
		FROM downloads
		WHERE status != 'completed'
		ORDER BY created_at DESC
	`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get incomplete downloads: %w", err)
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var d Download
		var title, channel, thumbnailURL, formatID, quality, errorMsg sql.NullString
		var duration sql.NullInt64
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&d.ID, &d.URL, &d.Status, &d.Progress,
			&title, &channel, &thumbnailURL, &formatID, &quality, &duration, &errorMsg,
			&d.CreatedAt, &startedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan download: %w", err)
		}

		d.Title = stringPtr(title)
		d.Channel = stringPtr(channel)
		d.ThumbnailURL = stringPtr(thumbnailURL)
		d.FormatID = stringPtr(formatID)
		d.Quality = stringPtr(quality)
		d.Duration = intPtr(duration)
		d.ErrorMessage = stringPtr(errorMsg)
		d.StartedAt = timePtr(startedAt)
		d.CompletedAt = timePtr(completedAt)

		downloads = append(downloads, d)
	}

	return downloads, nil
}

// ClearAllDownloads removes all download records
func (db *DB) ClearAllDownloads() error {
	query := `DELETE FROM downloads`
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clear downloads: %w", err)
	}
	return nil
}

// ClearCompletedDownloads removes all completed downloads
func (db *DB) ClearCompletedDownloads() error {
	query := `DELETE FROM downloads WHERE status = 'completed'`
	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clear completed downloads: %w", err)
	}
	return nil
}
