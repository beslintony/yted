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

