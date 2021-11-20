package build

import (
	"fmt"
)

var (
	BinaryVersion byte = 1

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

// FullVersion returns the full version and commit hash
func FullVersion() string {
	return fmt.Sprintf("%s (bin:%02X)", GitCommit, BinaryVersion)
}
