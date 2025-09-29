package cmd

import (
	"fmt"

	"contextmemory/cmd/memctl/internal/client"
	"contextmemory/cmd/memctl/internal/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new memory",
	Long: `Create a new memory with optional name, labels, and content.
Content can be provided via --content flag or piped from stdin.

Examples:
  memctl create --name "API Notes" --content "REST endpoints..." --labels "type=notes,project=api"
  echo "Session context..." | memctl create --name "Debug Session"
  memctl create --content "$(cat notes.txt)" --labels "type=docs"
  memctl create --name "Meeting Notes" < notes.txt`,
	RunE: runCreate,
}

var (
	createName    string
	createContent string
	createLabels  string
	createOutput  string
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&createName, "name", "", "", "Memory name")
	createCmd.Flags().StringVarP(&createContent, "content", "c", "", "Memory content (or pipe from stdin)")
	createCmd.Flags().StringVarP(&createLabels, "labels", "l", "", "Labels (format: key1=value1,key2=value2)")
	createCmd.Flags().StringVarP(&createOutput, "output", "o", "", "Output format: table|json|yaml")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Find MCP server
	serverPath, err := client.FindMCPServer(viper.GetString("mcp-server"))
	if err != nil {
		return fmt.Errorf("failed to find MCP server: %w", err)
	}

	// Create MCP client
	mcpClient := client.NewMCPClient(serverPath, viper.GetInt("verbosity"))

	// Get namespace
	namespace := viper.GetString("namespace")

	// Get content from stdin if not provided via flag
	if createContent == "" {
		stdinContent, err := client.ReadStdin()
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		if stdinContent != "" {
			createContent = stdinContent
		}
	}

	// Validate required fields
	if createContent == "" {
		return fmt.Errorf("content is required. Provide via --content flag or pipe from stdin")
	}

	// Parse labels
	labels := client.ParseLabels(createLabels)

	// Make the request
	response, err := mcpClient.CreateMemory(namespace, createName, createContent, labels)
	if err != nil {
		return fmt.Errorf("failed to create memory: %w", err)
	}

	// Parse output format
	outputOpts, err := output.ParseOutputFormat(createOutput)
	if err != nil {
		return fmt.Errorf("invalid output format: %w", err)
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
