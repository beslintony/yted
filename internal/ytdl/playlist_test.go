package ytdl

import "testing"

func TestIsPlaylistURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"playlist page", "https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", true},
		{"watch with list", "https://www.youtube.com/watch?v=abc123xyz00&list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", true},
		{"list first param", "https://www.youtube.com/watch?list=PLabc&v=abc123xyz00", true},
		{"plain watch", "https://www.youtube.com/watch?v=abc123xyz00", false},
		{"short url", "https://youtu.be/abc123xyz00", false},
		{"shorts", "https://www.youtube.com/shorts/abc123xyz00", false},
		{"empty", "", false},
		{"not a url", "not a url at all", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPlaylistURL(tt.url); got != tt.want {
				t.Errorf("IsPlaylistURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestParsePlaylistInfo(t *testing.T) {
	data := `{
		"id": "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		"title": "Test Playlist",
		"channel": "Test Channel",
		"playlist_count": 2,
		"entries": [
			{"id": "video001", "title": "First Video", "duration": 123.0},
			{"id": "video002", "title": "Second Video", "duration": 456.7},
			{"id": "", "title": "Broken Entry"}
		]
	}`

	info, err := parsePlaylistInfo(data)
	if err != nil {
		t.Fatalf("parsePlaylistInfo() error: %v", err)
	}

	if info.ID != "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf" {
		t.Errorf("ID = %q", info.ID)
	}
	if info.Title != "Test Playlist" {
		t.Errorf("Title = %q", info.Title)
	}
	if info.Channel != "Test Channel" {
		t.Errorf("Channel = %q", info.Channel)
	}
	if info.Count != 2 {
		t.Errorf("Count = %d, want 2", info.Count)
	}
	// Entries with empty IDs are dropped
	if len(info.Entries) != 2 {
		t.Fatalf("len(Entries) = %d, want 2", len(info.Entries))
	}
	if info.Entries[0].ID != "video001" || info.Entries[0].Duration != 123 {
		t.Errorf("Entries[0] = %+v", info.Entries[0])
	}
	if info.Entries[1].Duration != 456 {
		t.Errorf("Entries[1].Duration = %d, want 456", info.Entries[1].Duration)
	}
}

func TestParsePlaylistInfoFallbacks(t *testing.T) {
	// No channel, no playlist_count -> uploader and entry count are used
	data := `{"id": "PLabc", "title": "Fallbacks", "uploader": "Uploader", "entries": [{"id": "v1", "title": "V1"}]}`

	info, err := parsePlaylistInfo(data)
	if err != nil {
		t.Fatalf("parsePlaylistInfo() error: %v", err)
	}
	if info.Channel != "Uploader" {
		t.Errorf("Channel = %q, want uploader fallback", info.Channel)
	}
	if info.Count != 1 {
		t.Errorf("Count = %d, want entry-count fallback of 1", info.Count)
	}
}
