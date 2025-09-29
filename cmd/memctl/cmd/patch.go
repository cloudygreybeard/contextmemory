package cmd

import (
	"fmt"

	"contextmemory/cmd/memctl/internal/client"
	"contextmemory/cmd/memctl/internal/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var patchCmd = &cobra.Command{
	Use:   "patch <memory-id>",
	Short: "Patch (partially update) a memory",
	Long: `Patch (partially update) a memory by its ID. This command allows you to 
modify specific fields without replacing the entire memory.

Examples:
  memctl patch mem_123 --name "New Name"                        # Update name only
  memctl patch mem_123 --labels "type=updated,status=ready"     # Update labels only  
  memctl patch mem_123 --content "New content"                  # Update content only
  memctl patch mem_123 --name "Updated" --labels "type=final"   # Update multiple fields`,
	Args: cobra.ExactArgs(1),
	RunE: runPatch,
}

var (
	patchName    string
	patchContent string
	patchLabels  string
	patchOutput  string
)

func init() {
	rootCmd.AddCommand(patchCmd)

	patchCmd.Flags().StringVar(&patchName, "name", "", "Update memory name")
	patchCmd.Flags().StringVarP(&patchContent, "content", "c", "", "Update memory content")
	patchCmd.Flags().StringVarP(&patchLabels, "labels", "l", "", "Update labels (format: key1=value1,key2=value2)")
	patchCmd.Flags().StringVarP(&patchOutput, "output", "o", "", "Output format: table|json|yaml")
}

func runPatch(cmd *cobra.Command, args []string) error {
	// Find MCP server
	serverPath, err := client.FindMCPServer(viper.GetString("mcp-server"))
	if err != nil {
		return fmt.Errorf("failed to find MCP server: %w", err)
	}

	// Create MCP client
	mcpClient := client.NewMCPClient(serverPath, viper.GetInt("verbosity"))

	// Get namespace and memory ID
	namespace := viper.GetString("namespace")
	memoryID := args[0]

	// Build patch object
	patch := make(map[string]interface{})

	if patchName != "" {
		patch["name"] = patchName
	}

	if patchContent != "" {
		patch["content"] = patchContent
	}

	if patchLabels != "" {
		labels := client.ParseLabels(patchLabels)
		if labels != nil {
			patch["labels"] = labels
		}
	}

	// Validate that at least one field is being patched
	if len(patch) == 0 {
		return fmt.Errorf("at least one field must be specified for patching (--name, --content, or --labels)")
	}

	// Make the request
	response, err := mcpClient.PatchMemory(namespace, memoryID, patch)
	if err != nil {
		return fmt.Errorf("failed to patch memory: %w", err)
	}

	// Parse output format
	outputOpts, err := output.ParseOutputFormat(patchOutput)
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
