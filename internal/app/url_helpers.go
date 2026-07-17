package app

import (
	"net/url"
	"strings"
)

// normalizeURLForCache normalizes URL for cache key comparison
func normalizeURLForCache(urlStr string) string {
	// Clean the URL first (strip playlist params, etc.)
	return cleanYouTubeURL(urlStr)
}

// cleanYouTubeURL strips playlist and tracking parameters from YouTube URLs
// to ensure we always fetch single video info, not playlist info
func cleanYouTubeURL(urlStr string) string {
	// Trim whitespace
	urlStr = strings.TrimSpace(urlStr)

	// Parse the URL
	parsed, err := url.Parse(urlStr)
	if err != nil {
		// If parsing fails, try manual extraction for common formats
		return extractVideoIDManual(urlStr)
	}

	// For youtu.be short URLs, just return the base with path
	if parsed.Host == "youtu.be" || strings.HasSuffix(parsed.Host, ".youtu.be") {
		// Extract video ID from path
		parts := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			return "https://youtu.be/" + parts[0]
		}
		return urlStr
	}

	// For youtube.com URLs, keep only the 'v' parameter
	if strings.Contains(parsed.Host, "youtube.com") || strings.Contains(parsed.Host, "youtube") {
		query := parsed.Query()
		videoID := query.Get("v")
		if videoID != "" {
			// Rebuild URL with only video ID
			parsed.RawQuery = "v=" + videoID
			return parsed.String()
		}
	}

	return urlStr
}

// extractVideoIDManual manually extracts video ID from URL when parsing fails
func extractVideoIDManual(urlStr string) string {
	// Try to find v= parameter
	if idx := strings.Index(urlStr, "v="); idx != -1 {
		start := idx + 2
		end := strings.IndexAny(urlStr[start:], "&?# ")
		if end == -1 {
			return urlStr[:start] + urlStr[start:]
		}
		return urlStr[:start] + urlStr[start:start+end]
	}

	// Try youtu.be format
	if idx := strings.Index(urlStr, "youtu.be/"); idx != -1 {
		start := idx + 9
		end := strings.IndexAny(urlStr[start:], "?# ")
		if end == -1 {
			return "https://youtu.be/" + urlStr[start:]
		}
		return "https://youtu.be/" + urlStr[start:start+end]
	}

	return urlStr
}

// extractYoutubeID extracts the YouTube video ID from a URL
func extractYoutubeID(videoURL string) string {
	parsedURL, err := url.Parse(videoURL)
	if err != nil {
		return ""
	}

	// Handle youtu.be short URLs
	if strings.Contains(parsedURL.Host, "youtu.be") {
		path := strings.TrimPrefix(parsedURL.Path, "/")
		parts := strings.Split(path, "/")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	// Handle youtube.com URLs
	query := parsedURL.Query()
	if v := query.Get("v"); v != "" {
		return v
	}

	// Handle shorts URLs
	if strings.Contains(parsedURL.Path, "/shorts/") {
		parts := strings.Split(parsedURL.Path, "/shorts/")
		if len(parts) > 1 {
			return strings.Split(parts[1], "/")[0]
		}
	}

	return ""
}
