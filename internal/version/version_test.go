package version

import "testing"

func TestVersion(t *testing.T) {
	t.Logf("Version: %s", GetVersion())
	t.Logf("Full: %+v", GetFullVersion())

	if GetVersion() == "" {
		t.Error("Version should not be empty")
	}
}
