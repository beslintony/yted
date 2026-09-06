package app

import (
	"context"
	"testing"
	"time"

	"yted/internal/db"
)

func createQueueTestDownload(t *testing.T, a *App, id, status string) {
	t.Helper()
	dl := &db.Download{
		ID:        id,
		URL:       "https://www.youtube.com/watch?v=" + id,
		Status:    status,
		CreatedAt: time.Now(),
	}
	if err := a.db.CreateDownload(dl); err != nil {
		t.Fatalf("failed to create download %s: %v", id, err)
	}
}

func TestGetDownloadQueueResetsStuck(t *testing.T) {
	a, _ := setupVerifyTestApp(t)
	a.activeDownloads = make(map[string]context.CancelFunc)

	createQueueTestDownload(t, a, "q-stuck-1", "downloading")
	createQueueTestDownload(t, a, "q-stuck-2", "downloading")
	createQueueTestDownload(t, a, "q-pending-1", "pending")

	queue, err := a.GetDownloadQueue()
	if err != nil {
		t.Fatalf("GetDownloadQueue() error: %v", err)
	}
	if len(queue) != 3 {
		t.Fatalf("GetDownloadQueue() returned %d items, want 3", len(queue))
	}

	// Returned results must reflect the reset (regression: old code mutated
	// a loop copy, so callers still saw 'downloading')
	for _, r := range queue {
		if r.ID == "q-stuck-1" || r.ID == "q-stuck-2" {
			if r.Status != "pending" {
				t.Errorf("queue result %s status = %q, want pending", r.ID, r.Status)
			}
		}
	}

	for _, id := range []string{"q-stuck-1", "q-stuck-2"} {
		got, err := a.db.GetDownload(id)
		if err != nil {
			t.Fatalf("GetDownload(%s) error: %v", id, err)
		}
		if got.Status != "pending" {
			t.Errorf("download %s status = %q, want pending", id, got.Status)
		}
	}
}

func TestGetDownloadQueueKeepsActiveWorker(t *testing.T) {
	a, _ := setupVerifyTestApp(t)
	a.activeDownloads = make(map[string]context.CancelFunc)

	createQueueTestDownload(t, a, "q-active", "downloading")
	createQueueTestDownload(t, a, "q-stuck", "downloading")

	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.activeDownloads["q-active"] = cancel

	queue, err := a.GetDownloadQueue()
	if err != nil {
		t.Fatalf("GetDownloadQueue() error: %v", err)
	}

	byID := map[string]string{}
	for _, r := range queue {
		byID[r.ID] = r.Status
	}
	if byID["q-active"] != "downloading" {
		t.Errorf("active download status = %q, want downloading (kept)", byID["q-active"])
	}
	if byID["q-stuck"] != "pending" {
		t.Errorf("stuck download status = %q, want pending", byID["q-stuck"])
	}

	got, _ := a.db.GetDownload("q-active")
	if got.Status != "downloading" {
		t.Errorf("db active download status = %q, want downloading", got.Status)
	}
	got, _ = a.db.GetDownload("q-stuck")
	if got.Status != "pending" {
		t.Errorf("db stuck download status = %q, want pending", got.Status)
	}
}
