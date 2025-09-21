package cmd

import (
	"fmt"
	"strings"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search memories",
	Long: `Search memories by text query and/or label selectors.

Examples:
  cmctl search --query "authentication"                        # Search by text
  cmctl search --labels "type=session"                         # Search by labels
  cmctl search --query "API" --labels "type=code" --limit 5    # Combined search
  cmctl search --query "auth" -o json                          # JSON output
  cmctl search -q "session" -o jsonpath='{.items[*].spec.name}' # Extract names`,
	RunE: runSearch,
}

var (
	searchQuery      string
	searchLabels     string
	searchLimit      int
	searchOutputFlag string
)

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Text search query")
	searchCmd.Flags().StringVarP(&searchLabels, "labels", "l", "", "Label selector (format: key1=value1,key2=value2)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Limit results")
	searchCmd.Flags().StringVarP(&searchOutputFlag, "output", "o", "", "Output format: table|json|yaml|jsonpath=<template>|go-template=<template>")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Parse label selector
	labelSelector := make(map[string]string)
	if searchLabels != "" {
		pairs := strings.Split(searchLabels, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				labelSelector[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Create search request
	req := storage.SearchRequest{
		Query:         searchQuery,
		LabelSelector: labelSelector,
		Limit:         searchLimit,
	}

	// Search memories
	result, err := fs.Search(req)
	if err != nil {
		return fmt.Errorf("failed to search memories: %w", err)
	}

	// Parse output format
	outputOpts, err := ParseOutputFormat(searchOutputFlag)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Format and print output
	output, err := FormatMemoryList(result.Memories, outputOpts, false)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Print(output)
	return nil
}
