package cmd

import (
	"fmt"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all memories",
	Long: `List all memories with their basic information.

Examples:
  cmctl list                              # List memories without IDs
  cmctl list --show-id                    # List memories with IDs
  cmctl list -o json                      # Output as JSON
  cmctl list -o yaml                      # Output as YAML
  cmctl list -o jsonpath='{.items[*].metadata.name}'     # JSONPath output
  cmctl list -o go-template='{{range .items}}{{.spec.name}}{{"\n"}}{{end}}'  # Go template`,
	RunE: runList,
}

var (
	showID     bool
	outputFlag string
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&showID, "show-id", false, "Show memory IDs in the output")
	listCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output format: table|json|yaml|jsonpath=<template>|go-template=<template>")
}

func runList(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// List memories
	memories, err := fs.List()
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	// Parse output format
	outputOpts, err := ParseOutputFormat(outputFlag)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Format and print output
	output, err := FormatMemoryList(memories, outputOpts, showID)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}
