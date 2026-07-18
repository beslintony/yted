package app

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"yted/internal/db"
	"yted/internal/ytdl"
)

func TestPlaylistBindingsNilClient(t *testing.T) {
	a := &App{}

	if _, err := a.GetPlaylistInfo("https://www.youtube.com/playlist?list=PLabc"); err == nil {
		t.Error("GetPlaylistInfo() with nil ytdl client should fail")
	}
	if _, err := a.AddPlaylistDownload("https://www.youtube.com/playlist?list=PLabc", "best", "best"); err == nil {
		t.Error("AddPlaylistDownload() with nil ytdl client should fail")
	}
}

func TestPlaylistBindingsURLValidation(t *testing.T) {
	// Client never executes yt-dlp here - validation happens before any run
	database, err := db.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	a := &App{db: database, ytdl: ytdl.NewClient(&ytdl.ClientConfig{})}

	if _, err := a.GetPlaylistInfo("https://www.youtube.com/watch?v=abc123xyz00"); err == nil ||
		!strings.Contains(err.Error(), "not a playlist URL") {
		t.Errorf("GetPlaylistInfo() with non-playlist URL: err = %v", err)
	}
	if _, err := a.AddPlaylistDownload("https://www.youtube.com/watch?v=abc123xyz00", "best", "best"); err == nil ||
		!strings.Contains(err.Error(), "not a playlist URL") {
		t.Errorf("AddPlaylistDownload() with non-playlist URL: err = %v", err)
	}
}

func TestIsPlaylistURLBinding(t *testing.T) {
	a := &App{}

	if !a.IsPlaylistURL("https://www.youtube.com/playlist?list=PLabc") {
		t.Error("IsPlaylistURL() = false for playlist page, want true")
	}
	if a.IsPlaylistURL("https://www.youtube.com/watch?v=abc123xyz00") {
		t.Error("IsPlaylistURL() = true for plain watch URL, want false")
	}
}

func TestQueuePlaylistEntries(t *testing.T) {
	database, err := db.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	a := &App{db: database}

	// 60 entries - one of them already queued
	preExistingURL := "https://www.youtube.com/watch?v=entry000000"
	preTitle := "Existing"
	preExisting := &db.Download{ID: "pre-1", URL: preExistingURL, Status: "pending", Title: &preTitle}
	if err := database.CreateDownload(preExisting); err != nil {
		t.Fatalf("failed to seed download: %v", err)
	}

	entries := make([]ytdl.PlaylistEntry, 0, 60)
	entries = append(entries, ytdl.PlaylistEntry{ID: "entry000000", Title: "Duplicate"})
	for i := 1; i < 60; i++ {
		entries = append(entries, ytdl.PlaylistEntry{
			ID:    fmt.Sprintf("entry%06d", i),
			Title: fmt.Sprintf("Video %d", i),
		})
	}

	added := a.queuePlaylistEntries(entries, "best", "best")
	if added != 59 {
		t.Errorf("queuePlaylistEntries() added %d, want 59 (1 duplicate skipped)", added)
	}

	// Titles must be stored with the records (no more "Loading...")
	downloads, err := database.GetIncompleteDownloads()
	if err != nil {
		t.Fatalf("GetIncompleteDownloads() error: %v", err)
	}
	for _, d := range downloads {
		if d.Title == nil || *d.Title == "" {
			t.Errorf("download %s has no title", d.ID)
		}
	}
}

func TestCancelDownloadStopsWorker(t *testing.T) {
	database, err := db.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	a := NewApp()
	a.db = database

	dl := &db.Download{ID: "dl-1", URL: "https://www.youtube.com/watch?v=abc123xyz00", Status: "downloading"}
	if err := database.CreateDownload(dl); err != nil {
		t.Fatalf("failed to seed download: %v", err)
	}

	var cancelled atomic.Bool
	a.activeDownloads["dl-1"] = func() { cancelled.Store(true) }

	if err := a.CancelDownload("dl-1"); err != nil {
		t.Fatalf("CancelDownload() error: %v", err)
	}
	if !cancelled.Load() {
		t.Error("CancelDownload() did not stop the active worker")
	}
	if _, ok := a.activeDownloads["dl-1"]; ok {
		t.Error("worker still registered after CancelDownload")
	}
	got, err := database.GetDownload("dl-1")
	if err != nil {
		t.Fatalf("GetDownload() error: %v", err)
	}
	if got != nil {
		t.Errorf("download row still exists after CancelDownload: %+v", got)
	}
}

func TestClearDownloadCacheStopsWorkers(t *testing.T) {
	database, err := db.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	a := NewApp()
	a.db = database

	var cancelled atomic.Int32
	for _, id := range []string{"w1", "w2"} {
		a.activeDownloads[id] = func() { cancelled.Add(1) }
	}

	if err := a.ClearDownloadCache(); err != nil {
		t.Fatalf("ClearDownloadCache() error: %v", err)
	}
	if cancelled.Load() != 2 {
		t.Errorf("ClearDownloadCache() stopped %d workers, want 2", cancelled.Load())
	}
	if len(a.activeDownloads) != 0 {
		t.Errorf("activeDownloads still has %d entries", len(a.activeDownloads))
	}
}
