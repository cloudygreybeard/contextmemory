package cmd

import (
	"fmt"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check storage health",
	Long: `Check if the storage system is accessible and healthy.

Example:
  cmctl health`,
	RunE: runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Check health
	if err := fs.Health(); err != nil {
		fmt.Printf("Storage health: Unhealthy\n")
		if !IsQuiet() {
			fmt.Printf("Error: %v\n", err)
		}
		return err
	}

	fmt.Printf("Storage health: OK\n")
	return nil
}
