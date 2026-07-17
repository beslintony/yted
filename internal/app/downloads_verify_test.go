package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"yted/internal/config"
	"yted/internal/db"
)

// setupVerifyTestApp creates an app with a temp database and a config whose
// download path is a temp directory
func setupVerifyTestApp(t *testing.T) (*App, string) {
	t.Helper()

	appDataDir := t.TempDir()
	database, err := db.New(appDataDir)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	cfgManager, err := config.NewManager(appDataDir)
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	downloadPath := t.TempDir()
	cfgManager.Update(func(c *config.Config) {
		c.DownloadPath = downloadPath
	})

	return &App{db: database, config: cfgManager}, downloadPath
}

func createVerifyTestDownload(t *testing.T, a *App, id, status, errorMsg string) {
	t.Helper()

	title := "Test Video"
	formatID := "18"
	quality := "best"
	dl := &db.Download{
		ID:        id,
		URL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		Status:    status,
		Title:     &title,
		FormatID:  &formatID,
		Quality:   &quality,
		CreatedAt: time.Now(),
	}
	if errorMsg != "" {
		dl.ErrorMessage = &errorMsg
	}
	if err := a.db.CreateDownload(dl); err != nil {
		t.Fatalf("failed to create download %s: %v", id, err)
	}
}

func TestVerifyAndRepairDownloads(t *testing.T) {
	a, downloadPath := setupVerifyTestApp(t)

	// File on disk matching the youtube ID + format of both downloads
	videoFile := filepath.Join(downloadPath, "Test Video [dQw4w9WgXcQ][18].mp4")
	if err := os.WriteFile(videoFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create video file: %v", err)
	}

	// An interrupted download with a file present gets repaired
	createVerifyTestDownload(t, a, "dl-pending", "pending", "")
	// A failed download must NOT be revived - the file may belong to a
	// different, successful download of the same video
	createVerifyTestDownload(t, a, "dl-failed", "error", "HTTP Error 403: Forbidden")

	if err := a.VerifyAndRepairDownloads(); err != nil {
		t.Fatalf("VerifyAndRepairDownloads() error: %v", err)
	}

	pending, err := a.db.GetDownload("dl-pending")
	if err != nil {
		t.Fatalf("GetDownload(dl-pending) error: %v", err)
	}
	if pending.Status != "completed" {
		t.Errorf("interrupted download status = %q, want completed", pending.Status)
	}

	failed, err := a.db.GetDownload("dl-failed")
	if err != nil {
		t.Fatalf("GetDownload(dl-failed) error: %v", err)
	}
	if failed.Status != "error" {
		t.Errorf("failed download status = %q, want error (must not be auto-repaired)", failed.Status)
	}
	if failed.ErrorMessage == nil || *failed.ErrorMessage != "HTTP Error 403: Forbidden" {
		t.Errorf("failed download error message = %v, want preserved", failed.ErrorMessage)
	}
}

func TestVerifyAndRepairDownloadsSkipsCompleted(t *testing.T) {
	a, _ := setupVerifyTestApp(t)

	// No file on disk; a completed download must remain untouched
	createVerifyTestDownload(t, a, "dl-done", "completed", "")

	if err := a.VerifyAndRepairDownloads(); err != nil {
		t.Fatalf("VerifyAndRepairDownloads() error: %v", err)
	}

	done, err := a.db.GetDownload("dl-done")
	if err != nil {
		t.Fatalf("GetDownload(dl-done) error: %v", err)
	}
	if done.Status != "completed" {
		t.Errorf("completed download status = %q, want completed", done.Status)
	}
}
