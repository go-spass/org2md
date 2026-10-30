package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

// Stamped by the Makefile with -ldflags -X; a build that skips the Makefile
// recovers what it can from the embedded build info instead.
var (
	version = defaultVersion
	commit  = defaultCommit
	date    = defaultDate
)

func main() {
	info, _ := debug.ReadBuildInfo()
	cmd := newRootCmd(buildInfo{version: version, commit: commit, date: date}.fill(info))
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "org2md: %v\n", err)
		os.Exit(1)
	}
}
