package ytdl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/lrstanley/go-ytdlp"
)

// PlaylistEntry represents a single video in a playlist
type PlaylistEntry struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
}

// PlaylistInfo contains playlist metadata and its entries
type PlaylistInfo struct {
	ID      string          `json:"id"`
	Title   string          `json:"title"`
	Channel string          `json:"channel"`
	Count   int             `json:"count"`
	Entries []PlaylistEntry `json:"entries"`
}

// IsPlaylistURL reports whether the URL points at a YouTube playlist,
// either a /playlist page or any URL carrying a list= parameter
func IsPlaylistURL(urlStr string) bool {
	parsed, err := url.Parse(strings.TrimSpace(urlStr))
	if err != nil {
		return false
	}
	if parsed.Query().Get("list") != "" {
		return true
	}
	return strings.Contains(parsed.Path, "/playlist")
}

// rawPlaylistInfo mirrors yt-dlp's flat-playlist JSON output
type rawPlaylistInfo struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Channel       string `json:"channel"`
	Uploader      string `json:"uploader"`
	PlaylistCount int    `json:"playlist_count"`
	Entries       []struct {
		ID       string  `json:"id"`
		Title    string  `json:"title"`
		Duration float64 `json:"duration"`
	} `json:"entries"`
}

// GetPlaylistInfo fetches playlist metadata and entries (flat, no per-video
// requests) without downloading anything
func (c *Client) GetPlaylistInfo(ctx context.Context, url string) (*PlaylistInfo, error) {
	// Fresh command (like Download) so FlatPlaylist doesn't leak into the
	// shared c.dl builder used by GetInfo
	result, err := ytdlp.New().
		NoWarnings().
		Quiet().
		FlatPlaylist().
		DumpSingleJSON().
		Run(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist info: %w", err)
	}

	return parsePlaylistInfo(result.Stdout)
}

// parsePlaylistInfo parses yt-dlp flat-playlist JSON output
func parsePlaylistInfo(data string) (*PlaylistInfo, error) {
	var raw rawPlaylistInfo
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse playlist info: %w", err)
	}

	channel := raw.Channel
	if channel == "" {
		channel = raw.Uploader
	}

	entries := make([]PlaylistEntry, 0, len(raw.Entries))
	for _, e := range raw.Entries {
		if e.ID == "" {
			continue
		}
		entries = append(entries, PlaylistEntry{
			ID:       e.ID,
			Title:    e.Title,
			Duration: int(e.Duration),
		})
	}

	count := raw.PlaylistCount
	if count == 0 {
		count = len(entries)
	}

	return &PlaylistInfo{
		ID:      raw.ID,
		Title:   raw.Title,
		Channel: channel,
		Count:   count,
		Entries: entries,
	}, nil
}
