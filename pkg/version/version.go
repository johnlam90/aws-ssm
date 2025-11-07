package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version of the application
	Version = "0.1.0"

	// GitCommit is the git commit hash (set by build flags)
	GitCommit = "unknown"

	// BuildDate is the build date (set by build flags)
	BuildDate = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// GetVersion returns the full version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns the version with additional build information
func GetFullVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)",
		Version, GitCommit, BuildDate, GoVersion)
}

// GetBuildInfo returns structured build information
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     GitCommit,
		"build_date": BuildDate,
		"go_version": GoVersion,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}
}
