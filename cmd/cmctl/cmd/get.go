package cmd

import (
	"fmt"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:   "get [memory-id]",
	Short: "Get memories or specific memory by ID",
	Long: `Retrieve and display memories. Without arguments, lists all memories.
With a memory ID, retrieves a specific memory.

Performance Options:
  --include-content=false   Fast metadata-only listing (names, labels, timestamps)
  --no-index               Force file-based loading (slower but more robust)

Examples:
  cmctl get                                     # List all memories
  cmctl get --include-content=false             # Fast metadata-only listing
  cmctl get --show-id                           # List all memories with IDs
  cmctl get --labels "type=test"                # List memories with specific labels
  cmctl get -o json                             # List all memories as JSON
  cmctl get mem_abc123_def456                   # Get specific memory
  cmctl get mem_abc123_def456 -o yaml          # Get specific memory as YAML
  cmctl get mem_abc123_def456 -o jsonpath='{.spec.content}'  # Extract content using JSONPath`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGet,
}

var (
	getOutputFlag     string
	getShowID         bool
	getLabels         string
	getIncludeContent bool
	getNoIndex        bool
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&getOutputFlag, "output", "o", "", "Output format: table|json|yaml|jsonpath=<template>|go-template=<template>")
	getCmd.Flags().BoolVar(&getShowID, "show-id", false, "Show memory IDs when listing memories")
	getCmd.Flags().StringVarP(&getLabels, "labels", "l", "", "Label selector for filtering (format: key1=value1,key2=value2)")
	getCmd.Flags().BoolVar(&getIncludeContent, "include-content", true, "Include full memory content (disable for faster metadata-only listing)")
	getCmd.Flags().BoolVar(&getNoIndex, "no-index", false, "Disable index-based optimizations (force file-based loading)")
}

func runGet(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Parse output format
	outputOpts, err := ParseOutputFormat(getOutputFlag)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// If no memory ID provided, or filtering flags are used, list memories
	if len(args) == 0 || getLabels != "" {
		return runGetList(fs, outputOpts)
	}

	// Otherwise, get specific memory
	memoryID := args[0]
	return runGetSingle(fs, memoryID, outputOpts)
}

func runGetList(fs *storage.FileStorage, outputOpts OutputOptions) error {
	var memories []storage.Memory
	var err error

	if getLabels != "" {
		// Use search with label filtering
		labelSelector := parseLabels(getLabels)
		if len(labelSelector) == 0 {
			return fmt.Errorf("invalid label selector format: %s", getLabels)
		}

		searchReq := storage.SearchRequest{
			LabelSelector:  labelSelector,
			Limit:          -1, // No limit for get command
			UseIndex:       !getNoIndex,
			IncludeContent: getIncludeContent,
		}
		searchRes, err := fs.Search(searchReq)
		if err != nil {
			return fmt.Errorf("failed to search memories: %w", err)
		}
		memories = searchRes.Memories
	} else {
		// List all memories with performance options
		listOpts := storage.ListOptions{
			IncludeContent: getIncludeContent,
			UseIndex:       !getNoIndex,
		}
		memories, err = fs.ListWithOptions(listOpts)
		if err != nil {
			return fmt.Errorf("failed to list memories: %w", err)
		}
	}

	// Format and print output using the list document format
	output, err := FormatMemoryList(memories, outputOpts, getShowID)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}

func runGetSingle(fs *storage.FileStorage, memoryID string, outputOpts OutputOptions) error {
	// Get memory
	memory, err := fs.Get(memoryID)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}

	if memory == nil {
		return fmt.Errorf("memory not found: %s", memoryID)
	}

	// Format and print output
	output, err := FormatSingleMemory(memory, outputOpts)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}
