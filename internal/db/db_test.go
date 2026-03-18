package db

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	
	// Create temporary directory for database
	tmpDir, err := os.MkdirTemp("", "yted-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Clean up after test
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	
	db, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	
	t.Cleanup(func() {
		db.Close()
	})
	
	return db
}

func TestNew(t *testing.T) {
	db := setupTestDB(t)
	
	if db == nil {
		t.Fatal("Expected non-nil DB")
	}
	
	if db.conn == nil {
		t.Fatal("Expected non-nil connection")
	}
}

func TestCreateAndGetDownload(t *testing.T) {
	db := setupTestDB(t)
	
	now := time.Now()
	download := &Download{
		ID:           "test-id-123",
		URL:          "https://youtube.com/watch?v=test",
		Status:       "pending",
		Progress:     0,
		Title:        strPtr("Test Video"),
		Channel:      strPtr("Test Channel"),
		ThumbnailURL: strPtr("https://example.com/thumb.jpg"),
		FormatID:     strPtr("best"),
		Quality:      strPtr("1080p"),
		Duration:     durationPtr(300),
		ErrorMessage: nil,
		CreatedAt:    now,
		StartedAt:    nil,
		CompletedAt:  nil,
	}
	
	// Create download
	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}
	
	// Retrieve download
	retrieved, err := db.GetDownload("test-id-123")
	if err != nil {
		t.Fatalf("Failed to get download: %v", err)
	}
	
	if retrieved == nil {
		t.Fatal("Expected non-nil retrieved download")
	}
	
	if retrieved.ID != download.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, download.ID)
	}
	
	if retrieved.URL != download.URL {
		t.Errorf("URL mismatch: got %s, want %s", retrieved.URL, download.URL)
	}
	
	if retrieved.Status != download.Status {
		t.Errorf("Status mismatch: got %s, want %s", retrieved.Status, download.Status)
	}
	
	if *retrieved.Title != *download.Title {
		t.Errorf("Title mismatch: got %s, want %s", *retrieved.Title, *download.Title)
	}
}

func TestGetDownloadNotFound(t *testing.T) {
	db := setupTestDB(t)
	
	retrieved, err := db.GetDownload("non-existent-id")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if retrieved != nil {
		t.Error("Expected nil for non-existent download")
	}
}

func TestUpdateDownloadStatus(t *testing.T) {
	db := setupTestDB(t)
	
	// Create initial download
	download := &Download{
		ID:        "status-test",
		URL:       "https://youtube.com/watch?v=test",
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
	}
	
	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}
	
	// Update status to downloading
	err = db.UpdateDownloadStatus("status-test", "downloading")
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}
	
	retrieved, _ := db.GetDownload("status-test")
	if retrieved.Status != "downloading" {
		t.Errorf("Status not updated: got %s, want downloading", retrieved.Status)
	}
	
	// Update status to completed
	err = db.CompleteDownload("status-test")
	if err != nil {
		t.Fatalf("Failed to complete download: %v", err)
	}
	
	retrieved, _ = db.GetDownload("status-test")
	if retrieved.Status != "completed" {
		t.Errorf("Status not updated: got %s, want completed", retrieved.Status)
	}
	
	if retrieved.Progress != 100 {
		t.Errorf("Progress not set to 100: got %f", retrieved.Progress)
	}
	
	if retrieved.CompletedAt == nil {
		t.Error("CompletedAt not set")
	}
}

func TestFailDownload(t *testing.T) {
	db := setupTestDB(t)
	
	// Create initial download
	download := &Download{
		ID:        "fail-test",
		URL:       "https://youtube.com/watch?v=test",
		Status:    "downloading",
		Progress:  50,
		CreatedAt: time.Now(),
	}
	
	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}
	
	// Fail the download
	err = db.FailDownload("fail-test", "HTTP 416 error: Range not satisfiable")
	if err != nil {
		t.Fatalf("Failed to fail download: %v", err)
	}
	
	retrieved, _ := db.GetDownload("fail-test")
	if retrieved.Status != "error" {
		t.Errorf("Status not updated: got %s, want error", retrieved.Status)
	}
	
	if retrieved.ErrorMessage == nil {
		t.Fatal("ErrorMessage not set")
	}
	
	if *retrieved.ErrorMessage != "HTTP 416 error: Range not satisfiable" {
		t.Errorf("ErrorMessage mismatch: got %s", *retrieved.ErrorMessage)
	}
}

func TestGetPendingDownloads(t *testing.T) {
	db := setupTestDB(t)
	
	// Create downloads with different statuses
	downloads := []*Download{
		{ID: "pending-1", URL: "https://youtube.com/watch?v=1", Status: "pending", CreatedAt: time.Now()},
		{ID: "pending-2", URL: "https://youtube.com/watch?v=2", Status: "pending", CreatedAt: time.Now().Add(1 * time.Second)},
		{ID: "downloading", URL: "https://youtube.com/watch?v=3", Status: "downloading", CreatedAt: time.Now()},
		{ID: "completed", URL: "https://youtube.com/watch?v=4", Status: "completed", CreatedAt: time.Now()},
		{ID: "error", URL: "https://youtube.com/watch?v=5", Status: "error", CreatedAt: time.Now()},
	}
	
	for _, d := range downloads {
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}
	
	// Get pending downloads
	pending, err := db.GetPendingDownloads(10)
	if err != nil {
		t.Fatalf("Failed to get pending downloads: %v", err)
	}
	
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending downloads, got %d", len(pending))
	}
	
	// Verify order (oldest first)
	if pending[0].ID != "pending-1" {
		t.Errorf("Expected pending-1 first, got %s", pending[0].ID)
	}
}

func TestGetPendingDownloadsLimit(t *testing.T) {
	db := setupTestDB(t)
	
	// Create multiple pending downloads
	for i := 0; i < 5; i++ {
		d := &Download{
			ID:        fmt.Sprintf("pending-%d", i),
			URL:       fmt.Sprintf("https://youtube.com/watch?v=%d", i),
			Status:    "pending",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}
	
	// Get with limit
	pending, err := db.GetPendingDownloads(3)
	if err != nil {
		t.Fatalf("Failed to get pending downloads: %v", err)
	}
	
	if len(pending) != 3 {
		t.Errorf("Expected 3 pending downloads, got %d", len(pending))
	}
}

func TestStartDownload(t *testing.T) {
	db := setupTestDB(t)
	
	download := &Download{
		ID:        "start-test",
		URL:       "https://youtube.com/watch?v=test",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	
	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}
	
	err = db.StartDownload("start-test")
	if err != nil {
		t.Fatalf("Failed to start download: %v", err)
	}
	
	retrieved, _ := db.GetDownload("start-test")
	if retrieved.Status != "downloading" {
		t.Errorf("Status not updated: got %s, want downloading", retrieved.Status)
	}
	
	if retrieved.StartedAt == nil {
		t.Error("StartedAt not set")
	}
}

func TestUpdateDownloadProgress(t *testing.T) {
	db := setupTestDB(t)
	
	download := &Download{
		ID:        "progress-test",
		URL:       "https://youtube.com/watch?v=test",
		Status:    "downloading",
		Progress:  0,
		CreatedAt: time.Now(),
	}
	
	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}
	
	// Update progress
	err = db.UpdateDownloadProgress("progress-test", 75)
	if err != nil {
		t.Fatalf("Failed to update progress: %v", err)
	}
	
	retrieved, _ := db.GetDownload("progress-test")
	if retrieved.Progress != 75 {
		t.Errorf("Progress not updated: got %f, want 75", retrieved.Progress)
	}
}

func TestDeleteDownload(t *testing.T) {
	db := setupTestDB(t)
	
	download := &Download{
		ID:        "delete-test",
		URL:       "https://youtube.com/watch?v=test",
		Status:    "completed",
		CreatedAt: time.Now(),
	}
	
	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}
	
	// Verify it exists
	retrieved, _ := db.GetDownload("delete-test")
	if retrieved == nil {
		t.Fatal("Download should exist before deletion")
	}
	
	// Delete it
	err = db.DeleteDownload("delete-test")
	if err != nil {
		t.Fatalf("Failed to delete download: %v", err)
	}
	
	// Verify it's gone
	retrieved, _ = db.GetDownload("delete-test")
	if retrieved != nil {
		t.Error("Download should be deleted")
	}
}

func TestCountActiveDownloads(t *testing.T) {
	db := setupTestDB(t)
	
	// Create downloads with different statuses
	downloads := []*Download{
		{ID: "active-1", URL: "https://youtube.com/watch?v=1", Status: "downloading", CreatedAt: time.Now()},
		{ID: "active-2", URL: "https://youtube.com/watch?v=2", Status: "downloading", CreatedAt: time.Now()},
		{ID: "pending", URL: "https://youtube.com/watch?v=3", Status: "pending", CreatedAt: time.Now()},
		{ID: "completed", URL: "https://youtube.com/watch?v=4", Status: "completed", CreatedAt: time.Now()},
	}
	
	for _, d := range downloads {
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}
	
	count, err := db.CountActiveDownloads()
	if err != nil {
		t.Fatalf("Failed to count active downloads: %v", err)
	}
	
	if count != 2 {
		t.Errorf("Expected 2 active downloads, got %d", count)
	}
}

func TestGetActiveDownloadByURL(t *testing.T) {
	db := setupTestDB(t)

	// Create downloads with different statuses for same URL
	url := "https://youtube.com/watch?v=duplicate-test"
	
	downloads := []*Download{
		{ID: "pending-dl", URL: url, Status: "pending", CreatedAt: time.Now()},
		{ID: "downloading-dl", URL: url, Status: "downloading", CreatedAt: time.Now()},
		{ID: "completed-dl", URL: url, Status: "completed", CreatedAt: time.Now()},
		{ID: "error-dl", URL: url, Status: "error", CreatedAt: time.Now()},
	}

	for _, d := range downloads {
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}

	// Should find an active download (either pending or downloading)
	found, err := db.GetActiveDownloadByURL(url)
	if err != nil {
		t.Fatalf("Failed to get active download: %v", err)
	}
	if found == nil {
		t.Fatal("Expected to find an active download")
	}
	// Status should be either pending or downloading
	if found.Status != "pending" && found.Status != "downloading" {
		t.Errorf("Expected pending or downloading status, got %s", found.Status)
	}

	// Test with URL that has no active downloads
	otherURL := "https://youtube.com/watch?v=other"
	err = db.CreateDownload(&Download{
		ID:        "other-completed",
		URL:       otherURL,
		Status:    "completed",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	found, err = db.GetActiveDownloadByURL(otherURL)
	if err != nil {
		t.Fatalf("Failed to get active download: %v", err)
	}
	if found != nil {
		t.Error("Should not find active download for completed URL")
	}

	// Test with non-existent URL
	found, err = db.GetActiveDownloadByURL("https://youtube.com/watch?v=nonexistent")
	if err != nil {
		t.Fatalf("Failed to get active download: %v", err)
	}
	if found != nil {
		t.Error("Should not find download for non-existent URL")
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func durationPtr(i int) *int {
	return &i
}

// ============ NEW TESTS ============

func TestListDownloads(t *testing.T) {
	db := setupTestDB(t)

	// Create downloads with different statuses
	downloads := []*Download{
		{ID: "list-1", URL: "https://youtube.com/watch?v=1", Status: "pending", CreatedAt: time.Now()},
		{ID: "list-2", URL: "https://youtube.com/watch?v=2", Status: "downloading", CreatedAt: time.Now()},
		{ID: "list-3", URL: "https://youtube.com/watch?v=3", Status: "completed", CreatedAt: time.Now()},
	}

	for _, d := range downloads {
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}

	// List all downloads
	list, err := db.ListDownloads()
	if err != nil {
		t.Fatalf("Failed to list downloads: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 downloads, got %d", len(list))
	}

	// List with status filter
	pendingList, err := db.ListDownloads("pending")
	if err != nil {
		t.Fatalf("Failed to list pending downloads: %v", err)
	}

	if len(pendingList) != 1 {
		t.Errorf("Expected 1 pending download, got %d", len(pendingList))
	}
}

func TestUpdateDownload(t *testing.T) {
	db := setupTestDB(t)

	download := &Download{
		ID:        "update-test",
		URL:       "https://youtube.com/watch?v=test",
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
	}

	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	// Update download
	download.Status = "completed"
	download.Progress = 100
	download.Title = strPtr("Updated Title")

	err = db.UpdateDownload(download)
	if err != nil {
		t.Fatalf("Failed to update download: %v", err)
	}

	retrieved, _ := db.GetDownload("update-test")
	if retrieved.Status != "completed" {
		t.Errorf("Status not updated: got %s, want completed", retrieved.Status)
	}

	if *retrieved.Title != "Updated Title" {
		t.Errorf("Title not updated: got %s, want 'Updated Title'", *retrieved.Title)
	}
}

func TestGetIncompleteDownloads(t *testing.T) {
	db := setupTestDB(t)

	// Create downloads with different statuses
	downloads := []*Download{
		{ID: "inc-1", URL: "https://youtube.com/watch?v=1", Status: "pending", CreatedAt: time.Now()},
		{ID: "inc-2", URL: "https://youtube.com/watch?v=2", Status: "downloading", CreatedAt: time.Now()},
		{ID: "inc-3", URL: "https://youtube.com/watch?v=3", Status: "error", CreatedAt: time.Now()},
		{ID: "inc-4", URL: "https://youtube.com/watch?v=4", Status: "completed", CreatedAt: time.Now()},
	}

	for _, d := range downloads {
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}

	// Get incomplete downloads
	incomplete, err := db.GetIncompleteDownloads()
	if err != nil {
		t.Fatalf("Failed to get incomplete downloads: %v", err)
	}

	if len(incomplete) != 3 {
		t.Errorf("Expected 3 incomplete downloads, got %d", len(incomplete))
	}

	for _, d := range incomplete {
		if d.Status == "completed" {
			t.Errorf("Should not include completed downloads: %s", d.ID)
		}
	}
}

func TestClearDownloadError(t *testing.T) {
	db := setupTestDB(t)

	download := &Download{
		ID:           "clear-error-test",
		URL:          "https://youtube.com/watch?v=test",
		Status:       "error",
		Progress:     50,
		ErrorMessage: strPtr("Some error"),
		CreatedAt:    time.Now(),
	}

	err := db.CreateDownload(download)
	if err != nil {
		t.Fatalf("Failed to create download: %v", err)
	}

	// Clear error
	err = db.ClearDownloadError("clear-error-test")
	if err != nil {
		t.Fatalf("Failed to clear error: %v", err)
	}

	retrieved, _ := db.GetDownload("clear-error-test")
	if retrieved.ErrorMessage != nil && *retrieved.ErrorMessage != "" {
		t.Errorf("Error message should be cleared, got: %v", retrieved.ErrorMessage)
	}
}

func TestDeleteCompletedDownloads(t *testing.T) {
	db := setupTestDB(t)

	downloads := []*Download{
		{ID: "del-comp-1", URL: "https://youtube.com/watch?v=1", Status: "completed", CreatedAt: time.Now()},
		{ID: "del-comp-2", URL: "https://youtube.com/watch?v=2", Status: "completed", CreatedAt: time.Now()},
		{ID: "del-pending", URL: "https://youtube.com/watch?v=3", Status: "pending", CreatedAt: time.Now()},
	}

	for _, d := range downloads {
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}

	// Delete completed downloads
	err := db.DeleteCompletedDownloads()
	if err != nil {
		t.Fatalf("Failed to delete completed downloads: %v", err)
	}

	// Verify completed are deleted
	for _, id := range []string{"del-comp-1", "del-comp-2"} {
		retrieved, _ := db.GetDownload(id)
		if retrieved != nil {
			t.Errorf("Completed download %s should be deleted", id)
		}
	}

	// Verify pending still exists
	retrieved, _ := db.GetDownload("del-pending")
	if retrieved == nil {
		t.Error("Pending download should still exist")
	}
}

func TestClearAllDownloads(t *testing.T) {
	db := setupTestDB(t)

	downloads := []*Download{
		{ID: "clear-1", URL: "https://youtube.com/watch?v=1", Status: "completed", CreatedAt: time.Now()},
		{ID: "clear-2", URL: "https://youtube.com/watch?v=2", Status: "pending", CreatedAt: time.Now()},
	}

	for _, d := range downloads {
		err := db.CreateDownload(d)
		if err != nil {
			t.Fatalf("Failed to create download: %v", err)
		}
	}

	// Clear all downloads
	err := db.ClearAllDownloads()
	if err != nil {
		t.Fatalf("Failed to clear downloads: %v", err)
	}

	// Verify all deleted
	list, _ := db.ListDownloads()
	if len(list) != 0 {
		t.Errorf("Expected 0 downloads after clear, got %d", len(list))
	}
}

func TestVideoOperations(t *testing.T) {
	db := setupTestDB(t)

	// Create video
	video := &Video{
		ID:           "video-1",
		YoutubeID:    "youtube123",
		Title:        "Test Video",
		Channel:      "Test Channel",
		ChannelID:    "channel123",
		Duration:     300,
		Description:  "Test description",
		ThumbnailURL: "https://example.com/thumb.jpg",
		FilePath:     "/path/to/video.mp4",
		FileSize:     1024 * 1024 * 100, // 100MB
		FileHash:     "youtube123_best",
		IsManaged:    true,
		Format:       "best",
		Quality:      "1080p",
		DownloadedAt: time.Now(),
		WatchPosition: 0,
		WatchCount:    0,
	}

	err := db.CreateVideo(video)
	if err != nil {
		t.Fatalf("Failed to create video: %v", err)
	}

	// Get video
	retrieved, err := db.GetVideo("video-1")
	if err != nil {
		t.Fatalf("Failed to get video: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected non-nil video")
	}

	if retrieved.Title != video.Title {
		t.Errorf("Title mismatch: got %s, want %s", retrieved.Title, video.Title)
	}

	// Get video by YouTube ID
	byYoutubeID, err := db.GetVideosByYoutubeID("youtube123")
	if err != nil {
		t.Fatalf("Failed to get video by YouTube ID: %v", err)
	}

	if len(byYoutubeID) != 1 {
		t.Errorf("Expected 1 video, got %d", len(byYoutubeID))
	}

	// Get video by file hash
	byHash, err := db.GetVideoByFileHash("youtube123_best")
	if err != nil {
		t.Fatalf("Failed to get video by hash: %v", err)
	}

	if byHash == nil {
		t.Error("Expected to find video by hash")
	}

	// Update watch position
	err = db.UpdateWatchPosition("video-1", 150)
	if err != nil {
		t.Fatalf("Failed to update watch position: %v", err)
	}

	retrieved, _ = db.GetVideo("video-1")
	if retrieved.WatchPosition != 150 {
		t.Errorf("Watch position not updated: got %d, want 150", retrieved.WatchPosition)
	}

	// Update video
	video.Title = "Updated Title"
	err = db.UpdateVideo(video)
	if err != nil {
		t.Fatalf("Failed to update video: %v", err)
	}

	retrieved, _ = db.GetVideo("video-1")
	if retrieved.Title != "Updated Title" {
		t.Errorf("Title not updated: got %s", retrieved.Title)
	}

	// Delete video
	err = db.DeleteVideo("video-1")
	if err != nil {
		t.Fatalf("Failed to delete video: %v", err)
	}

	retrieved, _ = db.GetVideo("video-1")
	if retrieved != nil {
		t.Error("Video should be deleted")
	}
}

func TestGetStats(t *testing.T) {
	db := setupTestDB(t)

	// Create videos
	videos := []*Video{
		{
			ID:           "stats-1",
			YoutubeID:    "yt1",
			Title:        "Video 1",
			FilePath:     "/path/1.mp4",
			FileSize:     1000,
			DownloadedAt: time.Now(),
		},
		{
			ID:           "stats-2",
			YoutubeID:    "yt2",
			Title:        "Video 2",
			FilePath:     "/path/2.mp4",
			FileSize:     2000,
			DownloadedAt: time.Now(),
		},
	}

	for _, v := range videos {
		err := db.CreateVideo(v)
		if err != nil {
			t.Fatalf("Failed to create video: %v", err)
		}
	}

	// Get stats
	totalVideos, totalSize, err := db.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if totalVideos != 2 {
		t.Errorf("Expected 2 videos, got %d", totalVideos)
	}

	if totalSize != 3000 {
		t.Errorf("Expected total size 3000, got %d", totalSize)
	}
}

func TestListVideos(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	videos := []*Video{
		{
			ID:           "list-v-1",
			YoutubeID:    "yt1",
			Title:        "Alpha Video",
			Channel:      "Channel A",
			FilePath:     "/path/1.mp4",
			DownloadedAt: now,
		},
		{
			ID:           "list-v-2",
			YoutubeID:    "yt2",
			Title:        "Beta Video",
			Channel:      "Channel B",
			FilePath:     "/path/2.mp4",
			DownloadedAt: now.Add(time.Hour),
		},
		{
			ID:           "list-v-3",
			YoutubeID:    "yt3",
			Title:        "Gamma Video",
			Channel:      "Channel A",
			FilePath:     "/path/3.mp4",
			DownloadedAt: now.Add(2 * time.Hour),
		},
	}

	for _, v := range videos {
		err := db.CreateVideo(v)
		if err != nil {
			t.Fatalf("Failed to create video: %v", err)
		}
	}

	// List all videos
	list, err := db.ListVideos(ListVideosOptions{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list videos: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 videos, got %d", len(list))
	}

	// List with limit
	limited, err := db.ListVideos(ListVideosOptions{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("Failed to list videos: %v", err)
	}

	if len(limited) != 2 {
		t.Errorf("Expected 2 videos with limit, got %d", len(limited))
	}

	// List with channel filter
	channelA, err := db.ListVideos(ListVideosOptions{Limit: 10, Offset: 0, Channel: "Channel A"})
	if err != nil {
		t.Fatalf("Failed to list videos: %v", err)
	}

	if len(channelA) != 2 {
		t.Errorf("Expected 2 videos from Channel A, got %d", len(channelA))
	}

	// List with search
	search, err := db.ListVideos(ListVideosOptions{Limit: 10, Offset: 0, Search: "Alpha"})
	if err != nil {
		t.Fatalf("Failed to search videos: %v", err)
	}

	if len(search) != 1 {
		t.Errorf("Expected 1 video matching 'Alpha', got %d", len(search))
	}
}

func TestGetChannels(t *testing.T) {
	db := setupTestDB(t)

	videos := []*Video{
		{ID: "ch-1", YoutubeID: "yt1", Title: "Video 1", Channel: "Channel A", FilePath: "/1.mp4", DownloadedAt: time.Now()},
		{ID: "ch-2", YoutubeID: "yt2", Title: "Video 2", Channel: "Channel B", FilePath: "/2.mp4", DownloadedAt: time.Now()},
		{ID: "ch-3", YoutubeID: "yt3", Title: "Video 3", Channel: "Channel A", FilePath: "/3.mp4", DownloadedAt: time.Now()},
	}

	for _, v := range videos {
		err := db.CreateVideo(v)
		if err != nil {
			t.Fatalf("Failed to create video: %v", err)
		}
	}

	channels, err := db.GetChannels()
	if err != nil {
		t.Fatalf("Failed to get channels: %v", err)
	}

	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}

	// Verify both channels exist
	channelMap := make(map[string]bool)
	for _, ch := range channels {
		channelMap[ch] = true
	}

	if !channelMap["Channel A"] {
		t.Error("Expected Channel A in list")
	}
	if !channelMap["Channel B"] {
		t.Error("Expected Channel B in list")
	}
}
