package main

import (
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := newRootCmd(buildInfo{version: version, commit: commit, date: date})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "org2md: %v\n", err)
		os.Exit(1)
	}
}
