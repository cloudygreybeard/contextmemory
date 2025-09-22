package main

import (
	"os"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
