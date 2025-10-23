package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information (set by main)
var (
	clientVersion = "dev"
	clientCommit  = "unknown"
	clientDate    = "unknown"
)

// SetVersionInfo sets the version information from main
func SetVersionInfo(version, commit, date string) {
	clientVersion = version
	clientCommit = commit
	clientDate = date
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memctl",
	Short: "ContextMemory administration CLI",
	Long: `memctl provides administrative operations for ContextMemory.

Similar to kubectl, memctl offers a command-line interface for managing
conversation memories, with support for CRUD operations, search, and
label-based organization.

Examples:
  memctl get                           # List all memories
  memctl get my-memory                # Get specific memory
  memctl create --name "notes"        # Create memory from stdin
  memctl search --query "meeting"     # Search memories
  memctl delete my-memory             # Delete memory`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.contextmemory/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&storageDir, "storage-dir", "", "storage directory (default is $HOME/.contextmemory)")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "memory namespace")
	rootCmd.PersistentFlags().StringVar(&mcpServer, "mcp-server", "", "MCP server binary path")
	
	// Bind flags to viper
	viper.BindPFlag("storage-dir", rootCmd.PersistentFlags().Lookup("storage-dir"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("mcp-server", rootCmd.PersistentFlags().Lookup("mcp-server"))
}

var cfgFile string
var storageDir string
var namespace string
var mcpServer string

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".contextmemory" (without extension).
		viper.AddConfigPath(home + "/.contextmemory")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}




