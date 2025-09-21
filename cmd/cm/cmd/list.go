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

Example:
  cmctl list`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
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

	if len(memories) == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	// Print header
	fmt.Printf("%-40s %-30s %-20s\n", "NAME", "LABELS", "AGE")
	
	// Print memories
	for _, memory := range memories {
		labels := formatLabelsCompact(memory.Labels)
		age := formatAge(memory.UpdatedAt)
		fmt.Printf("%-40s %-30s %-20s\n", 
			truncateString(memory.Name, 38), 
			truncateString(labels, 28), 
			age)
	}

	return nil
}
