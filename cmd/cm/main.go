package main

import (
	"os"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
