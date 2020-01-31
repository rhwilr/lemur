package build

import (
	"fmt"
)

var (
	Major byte = 0
	Minor byte = 0
	Patch byte = 0

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

// FullVersion returns the full version and commit hash
func FullVersion() string {
	return fmt.Sprintf("%d.%d.%d@%s", Major, Minor, Patch, GitCommit)
}
