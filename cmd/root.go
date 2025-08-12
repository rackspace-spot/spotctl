// package cmd

// import (
// 	"fmt"
// 	"os"
// 	"strings"

// 	"github.com/rackspace-spot/spot-go-sdk/pkg/httpclient"
// 	"github.com/spf13/cobra"
// )

// var (
// 	outputFormat string
// 	verbose      bool // Global verbose flag
// )

// // rootCmd represents the base command when called without any subcommands
// var rootCmd = &cobra.Command{
// 	Use:   "spotctl",
// 	Short: "Rackspace Spot CLI - Manage your Spot resources",
// 	Long: `A command-line interface for managing Rackspace Spot resources.
// 	This CLI provides easy way to manage cloudspaces,node pools, and other Spot resources.`,
// 	Version: "0.1.0",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		if len(args) == 0 {
// 			_ = cmd.Help() // Display help message if no subcommand is provided
// 			os.Exit(0)     // Exit after showing help
// 		}
// 	},
// }

// // Execute adds all child commands to the root command and sets flags appropriately.
// func Execute() {
// 	// Set verbosity level for HTTP client
// 	httpclient.SetVerbose(verbose)

// 	if err := rootCmd.Execute(); err != nil {
// 		// Check if the error is a usage error
// 		if strings.Contains(err.Error(), "unknown command") || strings.Contains(err.Error(), "required flag") {
// 			_ = rootCmd.Help() // Display help message for usage errors
// 			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
// 			os.Exit(1)
// 		} else {
// 			// Print operational errors without showing help
// 			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
// 			os.Exit(1)
// 		}
// 	}
// }

// func init() {
// 	// Add global flags
// 	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, table, yaml)")
// 	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output") // Global verbose flag
// }

package cmd

import (
	"fmt"
	"os"

	"github.com/rackspace-spot/spot-go-sdk/pkg/httpclient"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	verbose      bool // Global verbose flag
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spotctl",
	Short: "Rackspace Spot CLI - Manage your Spot resources",
	Long: `A command-line interface for managing Rackspace Spot resources.
This CLI provides an easy way to manage cloudspaces, node pools, and other Spot resources.`,
	Version: "0.1.0",
	// This runs only if no subcommand is provided
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Set verbosity level for HTTP client
	httpclient.SetVerbose(verbose)

	// Silence usage globally; let Cobra show usage only on flag/arg parsing errors
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true // Stop Cobra from automatically showing usage on errors
	if err := rootCmd.Execute(); err != nil {
		// For all runtime errors, just print them cleanly
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, table, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output") // Global verbose flag
}
