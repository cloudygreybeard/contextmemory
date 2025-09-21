package cmd

import (
	"fmt"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get memory by ID",
	Long: `Retrieve and display a memory by its ID.

Examples:
  cmctl get mem_abc123_def456                    # Display memory in table format
  cmctl get mem_abc123_def456 -o json           # Output as JSON
  cmctl get mem_abc123_def456 -o yaml           # Output as YAML
  cmctl get mem_abc123_def456 -o jsonpath='{.spec.content}'  # Extract content using JSONPath`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

var (
	getOutputFlag string
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&getOutputFlag, "output", "o", "", "Output format: table|json|yaml|jsonpath=<template>|go-template=<template>")
}

func runGet(cmd *cobra.Command, args []string) error {
	memoryID := args[0]

	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Get memory
	memory, err := fs.Get(memoryID)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}

	if memory == nil {
		return fmt.Errorf("memory not found: %s", memoryID)
	}

	// Parse output format
	outputOpts, err := ParseOutputFormat(getOutputFlag)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Format and print output
	output, err := FormatSingleMemory(memory, outputOpts)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}
