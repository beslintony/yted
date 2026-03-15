package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// CreateVideo adds a new video to the library
func (db *DB) CreateVideo(video *Video) error {
	query := `
		INSERT INTO videos (id, youtube_id, title, channel, channel_id, duration, description, 
			thumbnail_url, file_path, file_size, file_hash, is_managed, format, quality, downloaded_at, watch_position, watch_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := db.conn.Exec(query,
		video.ID, video.YoutubeID, video.Title, video.Channel, video.ChannelID,
		video.Duration, video.Description, video.ThumbnailURL, video.FilePath,
		video.FileSize, video.FileHash, video.IsManaged, video.Format, video.Quality, video.DownloadedAt,
		video.WatchPosition, video.WatchCount,
	)
	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}
	return nil
}

// GetVideo retrieves a video by ID
func (db *DB) GetVideo(id string) (*Video, error) {
	query := `
		SELECT id, youtube_id, title, channel, channel_id, duration, description,
			thumbnail_url, file_path, file_size, format, quality, downloaded_at,
			watch_position, watch_count
		FROM videos WHERE id = ?
	`
	row := db.conn.QueryRow(query, id)

	var v Video
	err := row.Scan(
		&v.ID, &v.YoutubeID, &v.Title, &v.Channel, &v.ChannelID,
		&v.Duration, &v.Description, &v.ThumbnailURL, &v.FilePath,
		&v.FileSize, &v.Format, &v.Quality, &v.DownloadedAt,
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

// GetVideoByYoutubeID retrieves a video by YouTube ID
func (db *DB) GetVideoByYoutubeID(youtubeID string) (*Video, error) {
	query := `
		SELECT id, youtube_id, title, channel, channel_id, duration, description,
			thumbnail_url, file_path, file_size, format, quality, downloaded_at,
			watch_position, watch_count
		FROM videos WHERE youtube_id = ?
	`
	row := db.conn.QueryRow(query, youtubeID)

	var v Video
	err := row.Scan(
		&v.ID, &v.YoutubeID, &v.Title, &v.Channel, &v.ChannelID,
		&v.Duration, &v.Description, &v.ThumbnailURL, &v.FilePath,
		&v.FileSize, &v.Format, &v.Quality, &v.DownloadedAt,
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

// ListVideosOptions contains options for listing videos
type ListVideosOptions struct {
	Search   string
	Channel  string
	SortBy   string // date, title, channel, duration
	SortDesc bool
	Limit    int
	Offset   int
}

// ListVideos retrieves a list of videos
func (db *DB) ListVideos(opts ListVideosOptions) ([]Video, error) {
	// Build query
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

	// Sorting
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
			thumbnail_url, file_path, file_size, format, quality, downloaded_at,
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
			&v.FileSize, &v.Format, &v.Quality, &v.DownloadedAt,
			&v.WatchPosition, &v.WatchCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, v)
	}

	return videos, nil
}

// UpdateVideo updates a video
func (db *DB) UpdateVideo(video *Video) error {
	query := `
		UPDATE videos SET
			title = ?, channel = ?, channel_id = ?, duration = ?,
			description = ?, thumbnail_url = ?, file_path = ?, file_size = ?,
			format = ?, quality = ?, watch_position = ?, watch_count = ?
		WHERE id = ?
	`
	_, err := db.conn.Exec(query,
		video.Title, video.Channel, video.ChannelID, video.Duration,
		video.Description, video.ThumbnailURL, video.FilePath, video.FileSize,
		video.Format, video.Quality, video.WatchPosition, video.WatchCount,
		video.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}
	return nil
}

// UpdateWatchPosition updates the watch position for a video
func (db *DB) UpdateWatchPosition(id string, position int) error {
	query := `UPDATE videos SET watch_position = ? WHERE id = ?`
	_, err := db.conn.Exec(query, position, id)
	if err != nil {
		return fmt.Errorf("failed to update watch position: %w", err)
	}
	return nil
}

// IncrementWatchCount increments the watch count for a video
func (db *DB) IncrementWatchCount(id string) error {
	query := `UPDATE videos SET watch_count = watch_count + 1 WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to increment watch count: %w", err)
	}
	return nil
}

// DeleteVideo removes a video from the library
func (db *DB) DeleteVideo(id string) error {
	query := `DELETE FROM videos WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}
	return nil
}

// GetChannels returns a list of all channels in the library
func (db *DB) GetChannels() ([]string, error) {
	query := `SELECT DISTINCT channel FROM videos WHERE channel IS NOT NULL AND channel != '' ORDER BY channel`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// GetStats returns library statistics
func (db *DB) GetStats() (totalVideos int, totalSize int64, err error) {
	query := `SELECT COUNT(*), COALESCE(SUM(file_size), 0) FROM videos`
	row := db.conn.QueryRow(query)
	err = row.Scan(&totalVideos, &totalSize)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get stats: %w", err)
	}
	return totalVideos, totalSize, nil
}
