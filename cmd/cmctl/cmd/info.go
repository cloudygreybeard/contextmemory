package cmd

import (
	"fmt"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show storage information",
	Long: `Display information about the storage system including location, 
memory count, and total storage size.

Example:
  cmctl info`,
	RunE: runInfo,
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Get storage info
	info, err := fs.GetStorageInfo()
	if err != nil {
		return fmt.Errorf("failed to get storage info: %w", err)
	}

	fmt.Printf("Storage Directory:\t%s\n", info.StorageDir)
	fmt.Printf("Total Memories:\t\t%d\n", info.MemoriesCount)
	fmt.Printf("Storage Size:\t\t%.1f KB\n", float64(info.TotalSize)/1024)

	return nil
}
