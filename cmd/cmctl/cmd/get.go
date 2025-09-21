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

Example:
  cmctl get mem_abc123_def456`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	rootCmd.AddCommand(getCmd)
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

	// Display memory
	fmt.Printf("Name:\t\t%s\n", memory.Name)
	fmt.Printf("ID:\t\t%s\n", memory.ID)
	fmt.Printf("Labels:\t\t%s\n", formatLabels(memory.Labels))
	fmt.Printf("Created:\t%s\n", memory.CreatedAt.Format("2006-01-02T15:04:05Z"))
	fmt.Printf("Updated:\t%s\n", memory.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	fmt.Printf("\nContent:\n")
	fmt.Printf("--------\n")
	fmt.Println(memory.Content)

	return nil
}
