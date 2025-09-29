package cmd

import (
	"fmt"

	"contextmemory/cmd/memctl/internal/client"
	"contextmemory/cmd/memctl/internal/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search memories",
	Long: `Search memories by text query and/or label selectors.

Performance Options:
  --no-content   Fast metadata-only search (exclude memory content)

Examples:
  memctl search --query "authentication"                        # Search by text
  memctl search --labels "type=session"                         # Search by labels
  memctl search --labels "type=session" --no-content            # Metadata-only search
  memctl search --query "API" --labels "type=code" --limit 5    # Combined search
  memctl search --query "auth" -o json                          # JSON output
  memctl search -q "session" -o jsonpath='{.items[*].metadata.name}' # Extract names`,
	RunE: runSearch,
}

var (
	searchQuery      string
	searchLabels     string
	searchLimit      int
	searchOutputFlag string
	searchNoContent  bool
)

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Text search query")
	searchCmd.Flags().StringVarP(&searchLabels, "labels", "l", "", "Label selector (format: key1=value1,key2=value2)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Limit results")
	searchCmd.Flags().StringVarP(&searchOutputFlag, "output", "o", "", "Output format: table|json|yaml|jsonpath=<template>|go-template=<template>")
	searchCmd.Flags().BoolVar(&searchNoContent, "no-content", false, "Exclude memory content from results (faster for metadata-only searches)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Find MCP server
	serverPath, err := client.FindMCPServer(viper.GetString("mcp-server"))
	if err != nil {
		return fmt.Errorf("failed to find MCP server: %w", err)
	}

	// Create MCP client
	mcpClient := client.NewMCPClient(serverPath, viper.GetInt("verbosity"))

	// Parse output format
	outputOpts, err := output.ParseOutputFormat(searchOutputFlag)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get namespace
	namespace := viper.GetString("namespace")

	// Parse label selector
	labelSelector := client.ParseLabels(searchLabels)

	// Determine content inclusion (inverse of no-content flag)
	includeContent := !searchNoContent

	// Make the request
	response, err := mcpClient.SearchMemories(searchQuery, namespace, labelSelector, includeContent, searchLimit)
	if err != nil {
		return fmt.Errorf("failed to search memories: %w", err)
	}

	// Format output
	switch outputOpts.Format {
	case output.OutputFormatTable:
		result, err := output.FormatMemoryList(response, outputOpts, false) // Don't show ID by default in search
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(result)
	default:
		result, err := output.FormatMemoryList(response, outputOpts, false)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(result)
	}

	return nil
}
