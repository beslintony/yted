// Package version provides version information for the application.
// Version is set at build time using ldflags.
package version

// Variables set at build time via ldflags
var (
	// Version is the application version (e.g., "1.0.0")
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"
)

// GetVersion returns the application version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns detailed version info
func GetFullVersion() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     Commit,
		"build_date": BuildDate,
	}
}
