// Package version provides version information for spotctl.
package version

import (
	"bytes"
	"os/exec"
	"strings"
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
    // If ldflags provided a version, prefer it
    if Version != "" && Version != "dev" {
        return Version
    }
    // Otherwise, try to detect the latest git tag from the repository
    if v := detectLatestGitTag(); v != "" {
        return v
    }
    // Fallback to whatever is set (likely "dev")
    return Version
}

// detectLatestGitTag tries to get the latest tag using git commands.
// It first tries `git describe --tags --abbrev=0`, and if that fails,
// falls back to `git tag --sort=-creatordate` and picks the first tag.
func detectLatestGitTag() string {
    // Try: git describe --tags --abbrev=0
    if out, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output(); err == nil {
        tag := strings.TrimSpace(string(out))
        return tag
    }
    // Fallback: git tag --sort=-creatordate | head -n1
    cmd := exec.Command("git", "tag", "--sort=-creatordate")
    var stdout bytes.Buffer
    cmd.Stdout = &stdout
    if err := cmd.Run(); err == nil {
        // Take the first non-empty line
        lines := strings.Split(stdout.String(), "\n")
        for _, l := range lines {
            l = strings.TrimSpace(l)
            if l != "" {
                return l
            }
        }
    }
    return ""
}
