// Package version provides version information for spotctl.
package version

import (
	"runtime"
)

// These variables are set via ldflags during build time
var (
	// Version is the semantic version of the build
	Version = "dev"

	// Commit is the git commit hash of the build
	Commit = ""

	// BuildDate is the date when the binary was built
	BuildDate = ""

	// GoVersion is the version of Go used to build the binary
	GoVersion = runtime.Version()
)

// GetVersion returns the version string (just the version/tag)
func GetVersion() string {
	return Version
}
