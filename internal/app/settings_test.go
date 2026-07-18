package app

import (
	"testing"

	"yted/internal/config"
	"yted/internal/ytdl"
)

func TestUpdateSettingCookies(t *testing.T) {
	cfgManager, err := config.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}

	a := &App{
		config: cfgManager,
		ytdl:   ytdl.NewClient(&ytdl.ClientConfig{}),
	}

	if err := a.UpdateSetting("cookies_browser", "firefox"); err != nil {
		t.Fatalf("UpdateSetting(cookies_browser) error: %v", err)
	}
	if got := a.config.Get().CookiesBrowser; got != "firefox" {
		t.Errorf("CookiesBrowser = %q, want firefox", got)
	}

	if err := a.UpdateSetting("cookies_file", "/tmp/cookies.txt"); err != nil {
		t.Fatalf("UpdateSetting(cookies_file) error: %v", err)
	}
	if got := a.config.Get().CookiesFile; got != "/tmp/cookies.txt" {
		t.Errorf("CookiesFile = %q, want /tmp/cookies.txt", got)
	}

	// SaveSettings also pushes cookies into the ytdl client (no error path
	// to assert beyond not panicking with a live client)
	cfg := a.config.Get()
	if err := a.SaveSettings(cfg); err != nil {
		t.Fatalf("SaveSettings() error: %v", err)
	}
}
