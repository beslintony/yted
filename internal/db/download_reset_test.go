package db

import (
	"testing"
	"time"
)

func TestResetStuckDownloading(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	seed := []*Download{
		{ID: "stuck-1", URL: "https://youtube.com/watch?v=1", Status: "downloading", CreatedAt: now},
		{ID: "stuck-2", URL: "https://youtube.com/watch?v=2", Status: "downloading", CreatedAt: now},
		{ID: "stuck-3", URL: "https://youtube.com/watch?v=3", Status: "downloading", CreatedAt: now},
		{ID: "pending-1", URL: "https://youtube.com/watch?v=4", Status: "pending", CreatedAt: now},
		{ID: "done-1", URL: "https://youtube.com/watch?v=5", Status: "completed", CreatedAt: now},
		{ID: "failed-1", URL: "https://youtube.com/watch?v=6", Status: "error", CreatedAt: now},
		{ID: "paused-1", URL: "https://youtube.com/watch?v=7", Status: "paused", CreatedAt: now},
	}
	for _, d := range seed {
		if err := db.CreateDownload(d); err != nil {
			t.Fatalf("failed to seed download %s: %v", d.ID, err)
		}
	}

	// Single call must fix all stuck rows
	n, err := db.ResetStuckDownloading()
	if err != nil {
		t.Fatalf("ResetStuckDownloading() error: %v", err)
	}
	if n != 3 {
		t.Errorf("ResetStuckDownloading() affected %d rows, want 3", n)
	}

	for _, id := range []string{"stuck-1", "stuck-2", "stuck-3"} {
		got, err := db.GetDownload(id)
		if err != nil {
			t.Fatalf("GetDownload(%s) error: %v", id, err)
		}
		if got.Status != "pending" {
			t.Errorf("download %s status = %q, want pending", id, got.Status)
		}
	}

	// Untouched statuses must be preserved
	want := map[string]string{
		"pending-1": "pending",
		"done-1":    "completed",
		"failed-1":  "error",
		"paused-1":  "paused",
	}
	for id, status := range want {
		got, err := db.GetDownload(id)
		if err != nil {
			t.Fatalf("GetDownload(%s) error: %v", id, err)
		}
		if got.Status != status {
			t.Errorf("download %s status = %q, want %q", id, got.Status, status)
		}
	}

	// Idempotent: second call affects nothing
	n, err = db.ResetStuckDownloading()
	if err != nil {
		t.Fatalf("second ResetStuckDownloading() error: %v", err)
	}
	if n != 0 {
		t.Errorf("second ResetStuckDownloading() affected %d rows, want 0", n)
	}
}

func TestListVideoFilePaths(t *testing.T) {
	db := setupTestDB(t)

	videos := []*Video{
		{ID: "v1", YoutubeID: "yt1", Title: "Video 1", FilePath: "/downloads/a.mp4", DownloadedAt: time.Now()},
		{ID: "v2", YoutubeID: "yt2", Title: "Video 2", FilePath: "/downloads/b.mp3", DownloadedAt: time.Now()},
		{ID: "v3", YoutubeID: "yt3", Title: "Video 3", FilePath: "", DownloadedAt: time.Now()},
	}
	for _, v := range videos {
		if err := db.CreateVideo(v); err != nil {
			t.Fatalf("failed to seed video %s: %v", v.ID, err)
		}
	}

	paths, err := db.ListVideoFilePaths()
	if err != nil {
		t.Fatalf("ListVideoFilePaths() error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("ListVideoFilePaths() returned %d paths, want 2 (empty excluded)", len(paths))
	}
	found := map[string]bool{}
	for _, p := range paths {
		found[p] = true
	}
	if !found["/downloads/a.mp4"] || !found["/downloads/b.mp3"] {
		t.Errorf("ListVideoFilePaths() = %v, want both seeded paths", paths)
	}
}

func TestUpdateDownloadMetadataPreservesStatus(t *testing.T) {
	db := setupTestDB(t)

	d := &Download{
		ID:        "meta-1",
		URL:       "https://youtube.com/watch?v=meta",
		Status:    "downloading",
		CreatedAt: time.Now(),
	}
	if err := db.CreateDownload(d); err != nil {
		t.Fatalf("failed to seed download: %v", err)
	}
	if err := db.StartDownload("meta-1"); err != nil {
		t.Fatalf("StartDownload() error: %v", err)
	}

	if err := db.UpdateDownloadMetadata("meta-1", "Title", "Channel", "https://example.com/t.jpg", 321); err != nil {
		t.Fatalf("UpdateDownloadMetadata() error: %v", err)
	}

	got, err := db.GetDownload("meta-1")
	if err != nil {
		t.Fatalf("GetDownload() error: %v", err)
	}
	if got.Status != "downloading" {
		t.Errorf("status = %q, want downloading (must not be clobbered)", got.Status)
	}
	if got.StartedAt == nil {
		t.Error("StartedAt was clobbered to nil")
	}
	if got.Title == nil || *got.Title != "Title" {
		t.Errorf("Title = %v, want Title", got.Title)
	}
	if got.Duration == nil || *got.Duration != 321 {
		t.Errorf("Duration = %v, want 321", got.Duration)
	}
}
