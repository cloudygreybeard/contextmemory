package cmd

import (
	"fmt"

	"contextmemory/cmd/memctl/internal/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <memory-id>",
	Short: "Delete a memory",
	Long: `Delete a memory by its ID.

Examples:
  memctl delete mem_abc123_def456
  memctl delete my-memory-name`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	// Make the request
	response, err := mcpClient.DeleteMemory(namespace, memoryID)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	// Show success message
	if viper.GetInt("verbosity") >= 1 {
		fmt.Printf("Memory %s deleted successfully\n", memoryID)

		// Show additional details in verbose mode
		if viper.GetInt("verbosity") >= 2 && response.Result != nil {
			fmt.Printf("Response: %+v\n", response.Result)
		}
	}

	return nil
}
