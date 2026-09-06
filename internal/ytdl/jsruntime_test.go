package ytdl

import (
	"errors"
	"reflect"
	"testing"
)

func TestSelectJSRuntimes(t *testing.T) {
	present := func(names ...string) func(string) (string, error) {
		have := map[string]bool{}
		for _, n := range names {
			have[n] = true
		}
		return func(name string) (string, error) {
			if have[name] {
				return "/usr/bin/" + name, nil
			}
			return "", errors.New("not found")
		}
	}

	tests := []struct {
		name       string
		candidates []string
		installed  []string
		want       []string
	}{
		{"all present keeps preference order", preferredJSRuntimes, []string{"deno", "node", "bun"}, []string{"deno", "node", "bun"}},
		{"node only", preferredJSRuntimes, []string{"node"}, []string{"node"}},
		{"none installed", preferredJSRuntimes, nil, nil},
		{"unknown candidates ignored", []string{"deno", "quux"}, []string{"deno"}, []string{"deno"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectJSRuntimes(tt.candidates, present(tt.installed...)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectJSRuntimes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectJSRuntimesRespectsLookPathError(t *testing.T) {
	failAll := func(string) (string, error) { return "", errors.New("nope") }
	if got := selectJSRuntimes(preferredJSRuntimes, failAll); len(got) != 0 {
		t.Errorf("selectJSRuntimes() = %v, want empty", got)
	}
}

func TestLastNonEmptyLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single path", "/tmp/a/video.mp4\n", "/tmp/a/video.mp4"},
		{"trailing blank lines", "/tmp/a/video.mp4\n\n", "/tmp/a/video.mp4"},
		{"extractor noise before path", "[youtube] Extracting URL: https://x\n/tmp/a/video.mp4\n", "/tmp/a/video.mp4"},
		{"empty", "", ""},
		{"only whitespace", "  \n\t\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lastNonEmptyLine(tt.input); got != tt.want {
				t.Errorf("lastNonEmptyLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveJSRuntimesCaches(t *testing.T) {
	c := NewClient(&ClientConfig{})
	first := c.resolveJSRuntimes()
	second := c.resolveJSRuntimes()
	if !reflect.DeepEqual(first, second) {
		t.Errorf("resolveJSRuntimes() not stable: %v vs %v", first, second)
	}
	// In this environment node exists (PATH), deno/bun do not.
	for _, rt := range first {
		if rt != "node" && rt != "deno" && rt != "bun" {
			t.Errorf("resolveJSRuntimes() returned unexpected runtime %q", rt)
		}
	}
}
