package app

import (
	"testing"
	"time"
)

func TestShouldWriteDBProgress(t *testing.T) {
	base := time.Now()
	tests := []struct {
		name     string
		lastTime time.Time
		lastProg float64
		now      time.Time
		prog     float64
		expected bool
	}{
		{
			name:     "first tick always writes",
			lastTime: time.Time{},
			lastProg: 0,
			now:      base,
			prog:     0,
			expected: true,
		},
		{
			name:     "throttled within 1s with small delta",
			lastTime: base,
			lastProg: 10,
			now:      base.Add(200 * time.Millisecond),
			prog:     10.5,
			expected: false,
		},
		{
			name:     "writes after 1s elapsed",
			lastTime: base,
			lastProg: 10,
			now:      base.Add(1100 * time.Millisecond),
			prog:     10.1,
			expected: true,
		},
		{
			name:     "writes on >=1% delta",
			lastTime: base,
			lastProg: 10,
			now:      base.Add(100 * time.Millisecond),
			prog:     11,
			expected: true,
		},
		{
			name:     "no write just under 1% delta",
			lastTime: base,
			lastProg: 10,
			now:      base.Add(100 * time.Millisecond),
			prog:     10.99,
			expected: false,
		},
		{
			name:     "always writes at 100%",
			lastTime: base,
			lastProg: 99.9,
			now:      base.Add(100 * time.Millisecond),
			prog:     100,
			expected: true,
		},
		{
			name:     "progress regression does not write",
			lastTime: base,
			lastProg: 50,
			now:      base.Add(100 * time.Millisecond),
			prog:     49,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldWriteDBProgress(tt.lastTime, tt.lastProg, tt.now, tt.prog)
			if got != tt.expected {
				t.Errorf("shouldWriteDBProgress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPruneVideoInfoCacheDropsExpired(t *testing.T) {
	now := time.Now()
	cache := map[string]videoInfoCacheEntry{
		"fresh":   {info: &VideoInfoResult{Title: "fresh"}, expiresAt: now.Add(time.Minute)},
		"expired": {info: &VideoInfoResult{Title: "old"}, expiresAt: now.Add(-time.Minute)},
	}

	pruneVideoInfoCache(cache, now)

	if len(cache) != 1 {
		t.Fatalf("pruneVideoInfoCache() kept %d entries, want 1", len(cache))
	}
	if _, ok := cache["fresh"]; !ok {
		t.Error("pruneVideoInfoCache() dropped the fresh entry")
	}
}

func TestPruneVideoInfoCacheCapsSize(t *testing.T) {
	now := time.Now()
	cache := make(map[string]videoInfoCacheEntry, videoInfoCacheMaxEntries+50)
	for i := 0; i < videoInfoCacheMaxEntries+50; i++ {
		key := string(rune('a'+i%26)) + string(rune('A'+i/26)) + string(rune('0'+i/676))
		cache[key] = videoInfoCacheEntry{
			info:      &VideoInfoResult{Title: key},
			expiresAt: now.Add(time.Duration(i) * time.Second),
		}
	}
	newestExpiry := now.Add(time.Duration(videoInfoCacheMaxEntries+49) * time.Second)

	pruneVideoInfoCache(cache, now)

	if len(cache) != videoInfoCacheMaxEntries {
		t.Fatalf("pruneVideoInfoCache() size = %d, want %d", len(cache), videoInfoCacheMaxEntries)
	}
	// The newest entry must survive eviction of oldest-first
	survived := false
	for _, entry := range cache {
		if entry.expiresAt.Equal(newestExpiry) {
			survived = true
			break
		}
	}
	if !survived {
		t.Error("pruneVideoInfoCache() evicted the newest entry instead of the oldest")
	}
}

func TestPruneVideoInfoCacheKeepsTTL(t *testing.T) {
	now := time.Now()
	cache := map[string]videoInfoCacheEntry{
		"a": {info: &VideoInfoResult{Title: "a"}, expiresAt: now.Add(videoInfoCacheTTL)},
	}

	pruneVideoInfoCache(cache, now)

	entry, ok := cache["a"]
	if !ok {
		t.Fatal("pruneVideoInfoCache() dropped a fresh entry")
	}
	if entry.expiresAt.Sub(now) != videoInfoCacheTTL {
		t.Errorf("pruneVideoInfoCache() changed TTL: got %v, want %v", entry.expiresAt.Sub(now), videoInfoCacheTTL)
	}
}
