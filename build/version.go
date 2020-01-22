package build

import (
	"fmt"
)

var (
	Major = 1
	Minor = 1
	Patch = 10

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

// FullVersion returns the full version and commit hash
func FullVersion() string {
	return fmt.Sprintf("%d.%d.%d@%s", Major, Minor, Patch, GitCommit)
}