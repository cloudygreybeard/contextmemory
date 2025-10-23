package main

import (
	"os"

	"contextmemory/cmd/memctl/cmd"
)

// Version information (set by build)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Set version information in cmd package
	cmd.SetVersionInfo(version, commit, date)
	
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
