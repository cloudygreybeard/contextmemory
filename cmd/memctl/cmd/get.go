package cmd

import (
	"fmt"

	"contextmemory/cmd/memctl/internal/client"
	"contextmemory/cmd/memctl/internal/output"

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
  --limit                   Limit number of results returned

Examples:
  memctl get                                     # List all memories
  memctl get --include-content=false             # Fast metadata-only listing
  memctl get --show-id                           # List all memories with IDs
  memctl get --labels "type=test"                # List memories with specific labels
  memctl get -o json                             # List all memories as JSON
  memctl get mem_abc123_def456                   # Get specific memory
  memctl get mem_abc123_def456 -o yaml          # Get specific memory as YAML
  memctl get mem_abc123_def456 -o jsonpath='{.spec.content}'  # Extract content using JSONPath`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGet,
}

var (
	getOutputFlag     string
	getShowID         bool
	getLabels         string
	getIncludeContent bool
	getLimit          int
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&getOutputFlag, "output", "o", "", "Output format: table|json|yaml|jsonpath=<template>|go-template=<template>")
	getCmd.Flags().BoolVar(&getShowID, "show-id", false, "Show memory IDs when listing memories")
	getCmd.Flags().StringVarP(&getLabels, "labels", "l", "", "Label selector for filtering (format: key1=value1,key2=value2)")
	getCmd.Flags().BoolVar(&getIncludeContent, "include-content", true, "Include full memory content (disable for faster metadata-only listing)")
	getCmd.Flags().IntVar(&getLimit, "limit", 0, "Limit number of results (0 means no limit)")
}

func runGet(cmd *cobra.Command, args []string) error {
	// Find MCP server
	serverPath, err := client.FindMCPServer(viper.GetString("mcp-server"))
	if err != nil {
		return fmt.Errorf("failed to find MCP server: %w", err)
	}

	// Create MCP client
	mcpClient := client.NewMCPClient(serverPath, viper.GetInt("verbosity"))

	// Parse output format
	outputOpts, err := output.ParseOutputFormat(getOutputFlag)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
	}

	// Get namespace
	namespace := viper.GetString("namespace")

	// If no memory ID provided, or filtering flags are used, list memories
	if len(args) == 0 || getLabels != "" {
		return runGetList(mcpClient, namespace, outputOpts)
	}

	// Otherwise, get specific memory
	memoryID := args[0]
	return runGetSingle(mcpClient, namespace, memoryID, outputOpts)
}

func runGetList(mcpClient *client.MCPClient, namespace string, outputOpts output.OutputOptions) error {
	// Parse label selector
	labelSelector := client.ParseLabels(getLabels)

	// Make the request
	response, err := mcpClient.ListMemories(namespace, labelSelector, getIncludeContent, getLimit)
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	// Format output
	switch outputOpts.Format {
	case output.OutputFormatTable:
		result, err := output.FormatMemoryList(response, outputOpts, getShowID)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(result)
	default:
		result, err := output.FormatMemoryList(response, outputOpts, getShowID)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(result)
	}

	return nil
}

func runGetSingle(mcpClient *client.MCPClient, namespace, memoryID string, outputOpts output.OutputOptions) error {
	// Make the request
	response, err := mcpClient.GetMemory(namespace, memoryID)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}

	// Format output
	switch outputOpts.Format {
	case output.OutputFormatTable:
		result, err := output.FormatSingleMemory(response, outputOpts)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(result)
	default:
		result, err := output.FormatSingleMemory(response, outputOpts)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Print(result)
	}

	return nil
}
