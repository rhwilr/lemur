package main

import (
	"fmt"
)

var (
	// Version release version
	Version = "0.0.1"

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

// FullVersion returns the full version and commit hash
func FullVersion() string {
	return fmt.Sprintf("%s@%s", Version, GitCommit)
}