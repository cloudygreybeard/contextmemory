package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"contextmemory/cmd/memctl/internal/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type VersionInfo struct {
	Client VersionDetails `json:"client"`
	Server VersionDetails `json:"server"`
}

type VersionDetails struct {
	Version   string `json:"version"`
	Status    string `json:"status"`
	Path      string `json:"path,omitempty"`
	GoVersion string `json:"goVersion,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Show version information for both memctl client and MCP server components.

Similar to kubectl version, this shows:
- Client version (memctl)
- Server version (MCP server, if accessible)

Examples:
  memctl version                    # Show both versions
  memctl version -o json            # JSON output`,
	RunE: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().StringP("output", "o", "text", "Output format (text, json)")
}

func runVersion(cmd *cobra.Command, args []string) error {
	outputFormat, _ := cmd.Flags().GetString("output")

	versionInfo := VersionInfo{
		Client: VersionDetails{
			Version:   clientVersion,
			Status:    "ok",
			GoVersion: runtime.Version(),
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
	}

	// Try to get server version
	serverPath, err := client.FindMCPServer(viper.GetString("mcp-server"))
	if err != nil {
		versionInfo.Server = VersionDetails{
			Version: "unknown",
			Status:  fmt.Sprintf("error: %v", err),
		}
	} else {
		versionInfo.Server = VersionDetails{
			Version: clientVersion, // Server should match client version
			Status:  "ok",
			Path:    serverPath,
		}

		// Try to get more details about the server binary
		if abs, err := filepath.Abs(serverPath); err == nil {
			versionInfo.Server.Path = abs
		}

		// Try to get actual version from server if possible
		// For now, assume server matches client version
	}

	// Format output
	switch outputFormat {
	case "json":
		output, err := json.MarshalIndent(versionInfo, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal version info: %w", err)
		}
		fmt.Println(string(output))
	default:
		fmt.Printf("Client Version: %s\n", versionInfo.Client.Version)
		fmt.Printf("Client Platform: %s\n", versionInfo.Client.Platform)
		fmt.Printf("Client Go Version: %s\n", versionInfo.Client.GoVersion)
		fmt.Println()
		fmt.Printf("Server Version: %s\n", versionInfo.Server.Version)
		fmt.Printf("Server Status: %s\n", versionInfo.Server.Status)
		if versionInfo.Server.Path != "" {
			fmt.Printf("Server Path: %s\n", versionInfo.Server.Path)
		}
	}

	return nil
}
