package app

import (
	"strings"
	"testing"
	"time"

	"yted/internal/db"
	"yted/internal/editor"
)

// setupEditorTestApp creates an app with a real (temp dir) database and an
// editor with no ffmpeg, which is enough for job-record level tests
func setupEditorTestApp(t *testing.T) *App {
	t.Helper()

	database, err := db.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	video := &db.Video{
		ID:           "vid-1",
		YoutubeID:    "yt-1",
		Title:        "Test Video",
		FilePath:     "/tmp/test.mp4",
		DownloadedAt: time.Now(),
	}
	if err := database.CreateVideo(video); err != nil {
		t.Fatalf("failed to create test video: %v", err)
	}

	return &App{
		db:     database,
		editor: editor.New("", database, nil),
	}
}

func TestGetEditOptions(t *testing.T) {
	a := &App{}
	opts := a.GetEditOptions()

	if len(opts.Formats) == 0 {
		t.Error("GetEditOptions() returned no formats")
	}
	if len(opts.Codecs) == 0 {
		t.Error("GetEditOptions() returned no codecs")
	}
	if len(opts.CropPresets) == 0 {
		t.Error("GetEditOptions() returned no crop presets")
	}
	if len(opts.EffectRanges) == 0 {
		t.Error("GetEditOptions() returned no effect ranges")
	}
	if len(opts.Rotations) == 0 {
		t.Error("GetEditOptions() returned no rotations")
	}
	if len(opts.WatermarkPositions) == 0 {
		t.Error("GetEditOptions() returned no watermark positions")
	}

	// Option lists are sorted by ID for a stable UI
	for i := 1; i < len(opts.Formats); i++ {
		if opts.Formats[i-1].ID > opts.Formats[i].ID {
			t.Errorf("formats not sorted: %q before %q", opts.Formats[i-1].ID, opts.Formats[i].ID)
		}
	}
	for i := 1; i < len(opts.EffectRanges); i++ {
		if opts.EffectRanges[i-1].ID > opts.EffectRanges[i].ID {
			t.Errorf("effect ranges not sorted: %q before %q", opts.EffectRanges[i-1].ID, opts.EffectRanges[i].ID)
		}
	}
}

func TestSubmitEditJob_InvalidOperation(t *testing.T) {
	a := setupEditorTestApp(t)

	_, err := a.SubmitEditJob("vid-1", "explode", db.EditSettings{})
	if err == nil {
		t.Fatal("SubmitEditJob() with unknown operation should fail")
	}
	if !strings.Contains(err.Error(), "unknown edit operation") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubmitEditJob_InvalidEffectSettings(t *testing.T) {
	a := setupEditorTestApp(t)

	outOfRange := 5.0
	_, err := a.SubmitEditJob("vid-1", "effects", db.EditSettings{Brightness: &outOfRange})
	if err == nil {
		t.Fatal("SubmitEditJob() with out-of-range brightness should fail")
	}
	if !strings.Contains(err.Error(), "invalid effect settings") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubmitEditJob_NoFFmpeg(t *testing.T) {
	a := setupEditorTestApp(t)
	// No ffmpeg path configured and no FFmpegManager -> editorReady must fail
	a.ffmpeg = NewFFmpegManager()
	a.ffmpeg.SetCustomPath("/nonexistent/ffmpeg")

	_, err := a.SubmitEditJob("vid-1", "crop", db.EditSettings{})
	if err == nil {
		t.Fatal("SubmitEditJob() without ffmpeg should fail")
	}
	if !strings.Contains(err.Error(), "ffmpeg") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEditJobBindings_NilEditor(t *testing.T) {
	a := &App{}

	if _, err := a.GetEditJobStatus("job-1"); err == nil {
		t.Error("GetEditJobStatus() with nil editor should fail")
	}
	if _, err := a.ListEditJobs("vid-1"); err == nil {
		t.Error("ListEditJobs() with nil editor should fail")
	}
	if err := a.CancelEditJob("job-1"); err == nil {
		t.Error("CancelEditJob() with nil editor should fail")
	}
}

func TestEditJobStatusAndList(t *testing.T) {
	a := setupEditorTestApp(t)

	job := &db.EditJob{
		ID:            "job-1",
		SourceVideoID: "vid-1",
		Status:        "pending",
		Operation:     "crop",
		Settings:      "{}",
		CreatedAt:     time.Now(),
	}
	if err := a.db.CreateEditJob(job); err != nil {
		t.Fatalf("failed to create edit job: %v", err)
	}

	got, err := a.GetEditJobStatus("job-1")
	if err != nil {
		t.Fatalf("GetEditJobStatus() error: %v", err)
	}
	if got.ID != "job-1" || got.Operation != "crop" || got.Status != "pending" {
		t.Errorf("GetEditJobStatus() = %+v", got)
	}

	jobs, err := a.ListEditJobs("vid-1")
	if err != nil {
		t.Fatalf("ListEditJobs() error: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != "job-1" {
		t.Errorf("ListEditJobs() = %+v", jobs)
	}

	// Cancelling an inactive job is a no-op, not an error
	if err := a.CancelEditJob("job-1"); err != nil {
		t.Errorf("CancelEditJob() error: %v", err)
	}
}
