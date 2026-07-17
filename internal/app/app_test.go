package app

import "testing"

func TestHasPathTraversal(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"plain file", "/home/user/Downloads/video.mp4", false},
		{"ellipsis in filename", "/home/user/Downloads/YTed/My Family Made A Chat To Erase Me... They Forgot I Was In It [abc123][399+140].mp4", false},
		{"dots in directory", "/home/user/my..dir/video.mp4", false},
		{"trailing dots", "/home/user/video...mp4", false},
		{"hidden file", "/home/user/.config/app.conf", false},
		{"traversal at start", "../../etc/passwd", true},
		{"traversal segment", "/home/user/../../../etc/passwd", true},
		{"windows separator", `..\..\windows\system32`, true},
		{"single dot segments", "/home/./user/./video.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasPathTraversal(tt.path); got != tt.want {
				t.Errorf("hasPathTraversal(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
