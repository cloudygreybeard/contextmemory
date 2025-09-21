package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// VerbosityLevel represents different verbosity levels
type VerbosityLevel int

const (
	Quiet   VerbosityLevel = 0 // Only errors and essential output
	Normal  VerbosityLevel = 1 // Default level with standard messages
	Verbose VerbosityLevel = 2 // Debug info, config details, etc.
)

// GetVerbosity returns the current verbosity level
func GetVerbosity() VerbosityLevel {
	return VerbosityLevel(viper.GetInt("verbosity"))
}

// VPrintf prints formatted output only if verbosity level is met
func VPrintf(level VerbosityLevel, format string, args ...interface{}) {
	if GetVerbosity() >= level {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// VPrintln prints a line only if verbosity level is met
func VPrintln(level VerbosityLevel, args ...interface{}) {
	if GetVerbosity() >= level {
		fmt.Fprintln(os.Stderr, args...)
	}
}

// DebugPrintf prints debug information (verbosity >= 2)
func DebugPrintf(format string, args ...interface{}) {
	VPrintf(Verbose, "[DEBUG] "+format, args...)
}

// DebugPrintln prints debug information (verbosity >= 2)
func DebugPrintln(args ...interface{}) {
	if GetVerbosity() >= Verbose {
		debugArgs := append([]interface{}{"[DEBUG]"}, args...)
		VPrintln(Verbose, debugArgs...)
	}
}

// IsQuiet returns true if verbosity is set to quiet mode
func IsQuiet() bool {
	return GetVerbosity() == Quiet
}

// IsVerbose returns true if verbosity is >= 2
func IsVerbose() bool {
	return GetVerbosity() >= Verbose
}
