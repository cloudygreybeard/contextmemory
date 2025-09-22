package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	verbosity int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cmctl",
	Short: "ContextMemory - File-based memory management for LLM development workflows",
	Long: `ContextMemory is a clean, simple, file-based memory management system designed
for LLM development workflows. It provides CRUD operations for session contexts,
code snippets, and development notes with AI-assisted smart defaults.

Features:
- File-based storage (no servers or databases)
- AI-assisted name and label generation
- Flexible labeling system for organization
- Fast search and filtering
- Multiple output formats (JSON, YAML, JSONPath, Go templates)
- Cross-platform single binary

Verbosity levels:
- -v=0 (quiet): Only essential output
- -v=1 (normal): Standard messages (default)
- -v=2 (verbose): Debug info and config details`,
	Version: "0.6.3",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.contextmemory/config.yaml)")
	rootCmd.PersistentFlags().String("storage-dir", "", "storage directory (default is $HOME/.contextmemory)")
	rootCmd.PersistentFlags().String("provider", "file", "storage provider (file, s3, gcs, remote)")
	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", 1, "verbosity level (0=quiet, 1=normal, 2=verbose)")

	// Bind flags to viper
	if err := viper.BindPFlag("storage-dir", rootCmd.PersistentFlags().Lookup("storage-dir")); err != nil {
		// This should never happen for flags we define ourselves
		panic(fmt.Sprintf("failed to bind storage-dir flag: %v", err))
	}
	if err := viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider")); err != nil {
		panic(fmt.Sprintf("failed to bind provider flag: %v", err))
	}
	if err := viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity")); err != nil {
		panic(fmt.Sprintf("failed to bind verbosity flag: %v", err))
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".contextmemory".
		configDir := home + "/.contextmemory"
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// Only show config file info in verbose mode
		if viper.GetInt("verbosity") >= 2 {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
