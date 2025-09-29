package main

import (
	"os"

	"contextmemory/cmd/memctl/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
