package db

import (
	"fmt"
	"testing"
	"time"
)

func TestRequiredIndexesExist(t *testing.T) {
	database := setupTestDB(t)

	required := map[string]bool{
		// Pre-existing indexes (must be kept)
		"idx_videos_channel":         false,
		"idx_videos_downloaded_at":   false,
		"idx_downloads_status":       false,
		"idx_edit_jobs_source_video": false,
		"idx_edit_jobs_status":       false,
		// New perf indexes
		"idx_downloads_url_status": false,
		"idx_videos_youtube_id":    false,
		"idx_videos_file_hash":     false,
	}

	rows, err := database.conn.Query("SELECT name FROM sqlite_master WHERE type='index'")
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan index name: %v", err)
		}
		if _, ok := required[name]; ok {
			required[name] = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("failed to read indexes: %v", err)
	}

	for name, found := range required {
		if !found {
			t.Errorf("required index %s missing", name)
		}
	}
}

func TestConnectionPoolLimits(t *testing.T) {
	database := setupTestDB(t)

	stats := database.conn.Stats()
	if stats.MaxOpenConnections != 2 {
		t.Errorf("MaxOpenConnections = %d, want 2 (single-writer)", stats.MaxOpenConnections)
	}
}

func TestGetActiveDownloadURLs(t *testing.T) {
	database := setupTestDB(t)

	seed := []*Download{
		{ID: "u-pending", URL: "https://youtube.com/watch?v=1", Status: "pending", CreatedAt: time.Now()},
		{ID: "u-downloading", URL: "https://youtube.com/watch?v=2", Status: "downloading", CreatedAt: time.Now()},
		{ID: "u-completed", URL: "https://youtube.com/watch?v=3", Status: "completed", CreatedAt: time.Now()},
	}
	for _, d := range seed {
		if err := database.CreateDownload(d); err != nil {
			t.Fatalf("failed to seed download: %v", err)
		}
	}

	existing, err := database.GetActiveDownloadURLs([]string{
		"https://youtube.com/watch?v=1",
		"https://youtube.com/watch?v=2",
		"https://youtube.com/watch?v=3",
		"https://youtube.com/watch?v=missing",
	})
	if err != nil {
		t.Fatalf("GetActiveDownloadURLs() error: %v", err)
	}

	for _, url := range []string{"https://youtube.com/watch?v=1", "https://youtube.com/watch?v=2"} {
		if !existing[url] {
			t.Errorf("GetActiveDownloadURLs() missing active URL %s", url)
		}
	}
	for _, url := range []string{"https://youtube.com/watch?v=3", "https://youtube.com/watch?v=missing"} {
		if existing[url] {
			t.Errorf("GetActiveDownloadURLs() wrongly reports %s as active", url)
		}
	}

	empty, err := database.GetActiveDownloadURLs(nil)
	if err != nil {
		t.Fatalf("GetActiveDownloadURLs(nil) error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("GetActiveDownloadURLs(nil) = %v, want empty", empty)
	}
}

func TestCreateDownloadsBatch(t *testing.T) {
	database := setupTestDB(t)

	batch := make([]*Download, 0, 10)
	for i := 0; i < 10; i++ {
		batch = append(batch, &Download{
			ID:        fmt.Sprintf("batch-%d", i),
			URL:       fmt.Sprintf("https://youtube.com/watch?v=batch%d", i),
			Status:    "pending",
			CreatedAt: time.Now(),
		})
	}

	if err := database.CreateDownloads(batch); err != nil {
		t.Fatalf("CreateDownloads() error: %v", err)
	}

	for _, want := range batch {
		got, err := database.GetDownload(want.ID)
		if err != nil {
			t.Fatalf("GetDownload(%s) error: %v", want.ID, err)
		}
		if got == nil {
			t.Errorf("batch download %s missing", want.ID)
			continue
		}
		if got.URL != want.URL {
			t.Errorf("batch download %s URL = %s, want %s", want.ID, got.URL, want.URL)
		}
	}

	if err := database.CreateDownloads(nil); err != nil {
		t.Errorf("CreateDownloads(nil) error: %v", err)
	}
}
