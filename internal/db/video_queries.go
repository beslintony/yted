package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// GetVideoWithHash retrieves a video by ID including hash and managed status
func (db *DB) GetVideoWithHash(id string) (*Video, error) {
	query := `
		SELECT id, youtube_id, title, channel, channel_id, duration, description,
			thumbnail_url, file_path, file_size, file_hash, is_managed, format, quality, downloaded_at,
			watch_position, watch_count
		FROM videos WHERE id = ?
	`
	row := db.conn.QueryRow(query, id)

	var v Video
	err := row.Scan(
		&v.ID, &v.YoutubeID, &v.Title, &v.Channel, &v.ChannelID,
		&v.Duration, &v.Description, &v.ThumbnailURL, &v.FilePath,
		&v.FileSize, &v.FileHash, &v.IsManaged, &v.Format, &v.Quality, &v.DownloadedAt,
		&v.WatchPosition, &v.WatchCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}
	return &v, nil
}

// GetVideoByFileHash retrieves a video by its file hash (YouTube ID)
func (db *DB) GetVideoByFileHash(hash string) (*Video, error) {
	query := `
		SELECT id, youtube_id, title, channel, channel_id, duration, description,
			thumbnail_url, file_path, file_size, file_hash, is_managed, format, quality, downloaded_at,
			watch_position, watch_count
		FROM videos WHERE file_hash = ?
	`
	row := db.conn.QueryRow(query, hash)

	var v Video
	err := row.Scan(
		&v.ID, &v.YoutubeID, &v.Title, &v.Channel, &v.ChannelID,
		&v.Duration, &v.Description, &v.ThumbnailURL, &v.FilePath,
		&v.FileSize, &v.FileHash, &v.IsManaged, &v.Format, &v.Quality, &v.DownloadedAt,
		&v.WatchPosition, &v.WatchCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get video by hash: %w", err)
	}
	return &v, nil
}

// ListVideosWithHash retrieves videos with full info including hash and managed status
func (db *DB) ListVideosWithHash(opts ListVideosOptions) ([]Video, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if opts.Search != "" {
		where = append(where, "(title LIKE ? OR channel LIKE ?)")
		searchTerm := "%" + opts.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	if opts.Channel != "" {
		where = append(where, "channel = ?")
		args = append(args, opts.Channel)
	}

	orderBy := "downloaded_at"
	switch opts.SortBy {
	case "title":
		orderBy = "title"
	case "channel":
		orderBy = "channel"
	case "duration":
		orderBy = "duration"
	}

	direction := "ASC"
	if opts.SortDesc {
		direction = "DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, youtube_id, title, channel, channel_id, duration, description,
			thumbnail_url, file_path, file_size, file_hash, is_managed, format, quality, downloaded_at,
			watch_position, watch_count
		FROM videos
		WHERE %s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, strings.Join(where, " AND "), orderBy, direction)

	args = append(args, opts.Limit, opts.Offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list videos: %w", err)
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		err := rows.Scan(
			&v.ID, &v.YoutubeID, &v.Title, &v.Channel, &v.ChannelID,
			&v.Duration, &v.Description, &v.ThumbnailURL, &v.FilePath,
			&v.FileSize, &v.FileHash, &v.IsManaged, &v.Format, &v.Quality, &v.DownloadedAt,
			&v.WatchPosition, &v.WatchCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, v)
	}

	return videos, nil
}

// DeleteVideoAndFile removes video from DB and returns file path for deletion
func (db *DB) DeleteVideoAndFile(id string) (string, bool, error) {
	// First get the video info
	video, err := db.GetVideoWithHash(id)
	if err != nil {
		return "", false, err
	}
	if video == nil {
		return "", false, nil
	}

	// Delete from database
	query := `DELETE FROM videos WHERE id = ?`
	_, err = db.conn.Exec(query, id)
	if err != nil {
		return "", false, fmt.Errorf("failed to delete video: %w", err)
	}

	return video.FilePath, video.IsManaged, nil
}

// VerifyFileExists checks if the file still exists at the stored path
func (db *DB) VerifyFileExists(id string) (bool, error) {
	query := `SELECT file_path FROM videos WHERE id = ?`
	var filePath string
	err := db.conn.QueryRow(query, id).Scan(&filePath)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Check if file exists
	_, err = os.Stat(filePath)
	return !os.IsNotExist(err), nil
}
