package cmd

import (
	"fmt"
	"os"

	"github.com/rackerlabs/spot-sdk/rxtspot/httpclient" // Updated import path
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	verbose      bool // Global verbose flag
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spot-cli",
	Short: "Rackspace Spot CLI - Manage your Spot resources",
	Long: `A command-line interface for managing Rackspace Spot resources.
		
This CLI provides easy access to create, list, and manage cloudspaces, 
node pools, and other Spot resources.`,
	Version: "0.1.0",
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no subcommand is provided
		_ = cmd.Help() // Display help message
		os.Exit(0)     // Exit after showing help
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Set verbosity level for HTTP client
	httpclient.SetVerbose(verbose)

	if err := rootCmd.Execute(); err != nil {
		// Print the error only once
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, table, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output") // Global verbose flag
}
