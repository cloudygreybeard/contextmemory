package cmd

import (
	"fmt"
	"strings"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [memory-id]",
	Short: "Delete memories by ID or criteria",
	Long: `Delete one or more memories by ID or using label selectors.

Examples:
  cmctl delete memory/mem_12345678_90abcd    # Delete specific memory
  cmctl delete --labels "type=test"         # Delete all memories with type=test
  cmctl delete --all                        # Delete all memories (use with caution)`,
	RunE: runDelete,
}

var (
	deleteLabels string
	deleteAll    bool
	deleteForce  bool
)

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteLabels, "labels", "l", "", "Delete memories matching label selector (format: key1=value1,key2=value2)")
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "Delete all memories (dangerous)")
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompts")
}

func runDelete(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	verbosity := viper.GetInt("verbosity")

	// Handle different delete modes
	if len(args) == 1 {
		// Delete specific memory by ID
		memoryID := args[0]
		return deleteMemoryByID(fs, memoryID, verbosity)
	} else if deleteAll {
		// Delete all memories
		return deleteAllMemories(fs, verbosity)
	} else if deleteLabels != "" {
		// Delete by label selector
		return deleteMemoriesByLabels(fs, deleteLabels, verbosity)
	} else {
		return fmt.Errorf("must specify memory ID, --labels, or --all")
	}
}

func deleteMemoryByID(fs *storage.FileStorage, memoryID string, verbosity int) error {
	// Check if memory exists
	memory, err := fs.Get(memoryID)
	if err != nil {
		return fmt.Errorf("failed to get memory: %w", err)
	}
	if memory == nil {
		return fmt.Errorf("memory not found: %s", memoryID)
	}

	// Confirmation prompt (unless forced)
	if !deleteForce {
		if verbosity >= 1 {
			fmt.Printf("Are you sure you want to delete memory '%s'? (y/N): ", memory.Name)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Delete cancelled")
				return nil
			}
		}
	}

	// Delete the memory
	if err := fs.Delete(memoryID); err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	if verbosity >= 1 {
		fmt.Printf("Memory '%s' deleted successfully\n", memory.Name)
	}
	return nil
}

func deleteAllMemories(fs *storage.FileStorage, verbosity int) error {
	// Get all memories
	memories, err := fs.List()
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	if len(memories) == 0 {
		if verbosity >= 1 {
			fmt.Println("No memories to delete")
		}
		return nil
	}

	// Confirmation prompt (unless forced)
	if !deleteForce {
		if verbosity >= 1 {
			fmt.Printf("Are you sure you want to delete ALL %d memories? This cannot be undone! (y/N): ", len(memories))
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Delete cancelled")
				return nil
			}
		}
	}

	// Delete all memories
	deletedCount := 0
	for _, memory := range memories {
		if err := fs.Delete(memory.ID); err != nil {
			if verbosity >= 1 {
				fmt.Printf("Failed to delete memory '%s': %v\n", memory.Name, err)
			}
		} else {
			deletedCount++
			if verbosity >= 2 {
				fmt.Printf("Deleted: %s\n", memory.Name)
			}
		}
	}

	if verbosity >= 1 {
		fmt.Printf("Successfully deleted %d/%d memories\n", deletedCount, len(memories))
	}
	return nil
}

func deleteMemoriesByLabels(fs *storage.FileStorage, labelSelector string, verbosity int) error {
	// Parse label selector
	labels := parseLabels(labelSelector)
	if len(labels) == 0 {
		return fmt.Errorf("invalid label selector format: %s", labelSelector)
	}

	// Search for matching memories
	searchReq := storage.SearchRequest{
		LabelSelector: labels,
		Limit:         1000, // Large limit to get all matches
	}

	searchResp, err := fs.Search(searchReq)
	if err != nil {
		return fmt.Errorf("failed to search memories: %w", err)
	}

	if len(searchResp.Memories) == 0 {
		if verbosity >= 1 {
			fmt.Println("No memories found matching the label selector")
		}
		return nil
	}

	// Confirmation prompt (unless forced)
	if !deleteForce {
		if verbosity >= 1 {
			fmt.Printf("Found %d memories matching labels '%s'\n", len(searchResp.Memories), labelSelector)
			for _, memory := range searchResp.Memories {
				fmt.Printf("  - %s\n", memory.Name)
			}
			fmt.Print("Are you sure you want to delete these memories? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Delete cancelled")
				return nil
			}
		}
	}

	// Delete matching memories
	deletedCount := 0
	for _, memory := range searchResp.Memories {
		if err := fs.Delete(memory.ID); err != nil {
			if verbosity >= 1 {
				fmt.Printf("Failed to delete memory '%s': %v\n", memory.Name, err)
			}
		} else {
			deletedCount++
			if verbosity >= 2 {
				fmt.Printf("Deleted: %s\n", memory.Name)
			}
		}
	}

	if verbosity >= 1 {
		fmt.Printf("Successfully deleted %d/%d memories\n", deletedCount, len(searchResp.Memories))
	}
	return nil
}
