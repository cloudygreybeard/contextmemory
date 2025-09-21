package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudygreybeard/contextmemory-v2/cmd/cm/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new memory",
	Long: `Create a new memory with optional name, labels, and content.
Content can be provided via --content flag or piped from stdin.

Examples:
  cmctl create --name "API Notes" --content "REST endpoints..." --labels "type=notes,project=api"
  echo "Session context..." | cmctl create --name "Debug Session"
  cmctl create --content "$(cat notes.txt)" --labels "type=docs"`,
	RunE: runCreate,
}

var (
	createName    string
	createContent string
	createLabels  string
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Memory name")
	createCmd.Flags().StringVarP(&createContent, "content", "c", "", "Memory content (or pipe from stdin)")
	createCmd.Flags().StringVarP(&createLabels, "labels", "l", "", "Labels (format: key1=value1,key2=value2)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storageDir := viper.GetString("storage-dir")
	fs, err := storage.NewFileStorage(storageDir)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Get content from stdin if not provided via flag
	content := createContent
	if content == "" {
		stdinContent, err := readStdin()
		if err == nil && stdinContent != "" {
			content = stdinContent
		}
	}

	if content == "" {
		return fmt.Errorf("content is required (use --content or pipe from stdin)")
	}

	// Parse labels
	labels := make(map[string]string)
	if createLabels != "" {
		pairs := strings.Split(createLabels, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Create memory
	req := storage.CreateMemoryRequest{
		Name:    createName,
		Content: content,
		Labels:  labels,
	}

	memory, err := fs.Create(req)
	if err != nil {
		return fmt.Errorf("failed to create memory: %w", err)
	}

	// Output success message
	fmt.Printf("memory/%s created\n", memory.ID)
	if GetVerbosity() >= Normal {
		fmt.Printf("NAME\t%s\n", memory.Name)
		fmt.Printf("LABELS\t%s\n", formatLabels(memory.Labels))
		fmt.Printf("CREATED\t%s\n", memory.CreatedAt.Format("2006-01-02T15:04:05Z"))
	}

	return nil
}

func readStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// No piped input
		return "", nil
	}

	var content strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(content.String()), nil
}
