package app

import (
	"strings"
	"testing"

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
	a := &App{ytdl: ytdl.NewClient(&ytdl.ClientConfig{})}

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
