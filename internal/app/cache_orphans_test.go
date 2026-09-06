package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"yted/internal/db"
)

func setupOrphanTestApp(t *testing.T) (*App, string) {
	t.Helper()

	a, downloadPath := setupVerifyTestApp(t)
	a.fm = NewFileManager(a.config)
	return a, downloadPath
}

func writeOrphanFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file %s: %v", path, err)
	}
}

func createOrphanTestVideo(t *testing.T, a *App, id, filePath string) {
	t.Helper()
	v := &db.Video{
		ID:           id,
		YoutubeID:    "yt-" + id,
		Title:        "Video " + id,
		FilePath:     filePath,
		DownloadedAt: time.Now(),
	}
	if err := a.db.CreateVideo(v); err != nil {
		t.Fatalf("failed to create video %s: %v", id, err)
	}
}

func TestOrphanScanSinglePass(t *testing.T) {
	a, downloadPath := setupOrphanTestApp(t)

	trackedPath := filepath.Join(downloadPath, "tracked.mp4")
	orphanPath := filepath.Join(downloadPath, "orphan.mp4")
	notesPath := filepath.Join(downloadPath, "notes.txt")

	writeOrphanFile(t, trackedPath, "tracked-content")
	writeOrphanFile(t, orphanPath, "orphan-content")
	writeOrphanFile(t, notesPath, "not-media")
	// Subdirectories are not scanned (non-recursive listing)
	subDir := filepath.Join(downloadPath, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	writeOrphanFile(t, filepath.Join(subDir, "nested.mp4"), "nested")

	createOrphanTestVideo(t, a, "v-tracked", trackedPath)

	count, size, err := a.findOrphanedFiles()
	if err != nil {
		t.Fatalf("findOrphanedFiles() error: %v", err)
	}
	if count != 1 {
		t.Errorf("findOrphanedFiles() count = %d, want 1 (only orphan.mp4)", count)
	}
	if size <= 0 {
		t.Errorf("findOrphanedFiles() size = %d, want > 0", size)
	}

	// Dry run: reports but deletes nothing
	result, err := a.CleanupOrphanedFiles(false)
	if err != nil {
		t.Fatalf("CleanupOrphanedFiles(false) error: %v", err)
	}
	if result["found"] != 1 {
		t.Errorf("dry-run found = %v, want 1", result["found"])
	}
	if result["deleted"] != 0 {
		t.Errorf("dry-run deleted = %v, want 0", result["deleted"])
	}
	if _, err := os.Stat(orphanPath); err != nil {
		t.Errorf("dry-run deleted orphan file, want it kept: %v", err)
	}

	// Delete mode: removes only the orphan
	result, err = a.CleanupOrphanedFiles(true)
	if err != nil {
		t.Fatalf("CleanupOrphanedFiles(true) error: %v", err)
	}
	if result["found"] != 1 {
		t.Errorf("delete found = %v, want 1", result["found"])
	}
	if result["deleted"] != 1 {
		t.Errorf("delete deleted = %v, want 1", result["deleted"])
	}
	if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
		t.Error("orphan file still exists after delete mode")
	}
	if _, err := os.Stat(trackedPath); err != nil {
		t.Errorf("tracked file was deleted, want it kept: %v", err)
	}
	if _, err := os.Stat(notesPath); err != nil {
		t.Errorf("non-media file was deleted, want it kept: %v", err)
	}

	// Second run finds nothing
	count, _, err = a.findOrphanedFiles()
	if err != nil {
		t.Fatalf("second findOrphanedFiles() error: %v", err)
	}
	if count != 0 {
		t.Errorf("second findOrphanedFiles() count = %d, want 0", count)
	}
}

func TestScanOrphanedEntriesPure(t *testing.T) {
	dir := t.TempDir()
	// Build real dir entries via temp files (no network, no db)
	for name, content := range map[string]string{
		"a.mp4": "12345",
		"b.txt": "nope",
		"c.MP3": "123",
	} {
		writeOrphanFile(t, filepath.Join(dir, name), content)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}

	// Tracked set uses lower-cased full paths (same convention as the scan)
	tracked := map[string]bool{
		strings.ToLower(filepath.Join(dir, "a.mp4")): true,
	}

	orphans, size := scanOrphanedEntries(entries, dir, tracked)
	if len(orphans) != 1 {
		t.Fatalf("scanOrphanedEntries() found %d orphans %v, want 1 (c.MP3, uppercase ext lower-cased)", len(orphans), orphans)
	}
	if size != 3 {
		t.Errorf("scanOrphanedEntries() size = %d, want 3", size)
	}
}
