package cmd

import (
	"fmt"
	"strings"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search memories",
	Long: `Search memories by text query and/or label selectors.

Examples:
  cmctl search --query "authentication"
  cmctl search --labels "type=session"
  cmctl search --query "API" --labels "type=code" --limit 5`,
	RunE: runSearch,
}

var (
	searchQuery  string
	searchLabels string
	searchLimit  int
)

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Text search query")
	searchCmd.Flags().StringVarP(&searchLabels, "labels", "l", "", "Label selector (format: key1=value1,key2=value2)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 10, "Limit results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Parse label selector
	labelSelector := make(map[string]string)
	if searchLabels != "" {
		pairs := strings.Split(searchLabels, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				labelSelector[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Create search request
	req := storage.SearchRequest{
		Query:         searchQuery,
		LabelSelector: labelSelector,
		Limit:         searchLimit,
	}

	// Search memories
	result, err := fs.Search(req)
	if err != nil {
		return fmt.Errorf("failed to search memories: %w", err)
	}

	if len(result.Memories) == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	// Print header
	fmt.Printf("%-40s %-25s %-15s %s\n", "NAME", "LABELS", "AGE", "PREVIEW")
	
	// Print matching memories
	for _, memory := range result.Memories {
		labels := formatLabelsCompact(memory.Labels)
		age := formatAge(memory.UpdatedAt)
		preview := memory.Content
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		
		fmt.Printf("%-40s %-25s %-15s %s\n", 
			truncateString(memory.Name, 38), 
			truncateString(labels, 23),
			age,
			truncateString(preview, 50))
	}

	return nil
}
