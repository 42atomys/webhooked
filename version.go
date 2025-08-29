package webhooked

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of webhooked
	// This can be overridden at build time with -ldflags "-X github.com/42atomys/webhooked.Version=x.y.z"
	Version = "dev"

	// GitCommit is the git commit hash, set at build time
	GitCommit = "unknown"

	// BuildDate is the build date, set at build time
	BuildDate = "unknown"
)

// BuildInfo returns formatted build information
func BuildInfo() string {
	return fmt.Sprintf("webhooked %s (commit: %s, built: %s, go: %s)",
		Version, GitCommit, BuildDate, runtime.Version())
}

// VersionInfo returns version information as a map
func VersionInfo() map[string]string {
	return map[string]string{
		"version":   Version,
		"commit":    GitCommit,
		"buildDate": BuildDate,
		"goVersion": runtime.Version(),
		"goOS":      runtime.GOOS,
		"goArch":    runtime.GOARCH,
	}
}
