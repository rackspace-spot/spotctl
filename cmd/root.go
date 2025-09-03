package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/rackspace-spot/spotctl/internal/version"
	config "github.com/rackspace-spot/spotctl/pkg"

	"github.com/spf13/cobra"
	"k8s.io/klog"
)

var (
	outputFormat string
	verbosity    int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "spotctl",
	Short:   "Rackspace Spot CLI - Manage your Spot resources",
	Long:    `A command-line interface for managing Rackspace Spot resources. This CLI provides an easy way to manage cloudspaces, node pools, and other Spot resources.`,
	Version: version.GetVersion(),
	// Root PersistentPreRun runs before ANY subcommand
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLoggingFlags(verbosity)
		klog.V(1).Infof("Verbosity set to %d", verbosity)
		cfg, err := config.LoadConfig()
		if err != nil {
			klog.Errorf("Failed to load config: %v", err)
			os.Exit(1)
		}
		klog.V(1).Infof("Config loaded: %+v", cfg)
		// Inject default org from config if flag not set
		orgFlag := "org"
		orgVal, err := cmd.Flags().GetString(orgFlag)
		if err == nil && orgVal == "" && cfg.Org != "" {
			// Set flag value programmatically
			err := cmd.Flags().Set(orgFlag, cfg.Org)
			if err != nil {
				klog.Errorf("failed to set default flag %s: %v", orgFlag, err)
			}
		}

		// Inject default region from config if flag not set
		regionFlag := "region"
		regionVal, err := cmd.Flags().GetString(regionFlag)
		if err == nil && regionVal == "" && cfg.Region != "" {
			err := cmd.Flags().Set(regionFlag, cfg.Region)
			if err != nil {
				klog.Errorf("failed to set default flag %s: %v", regionFlag, err)
			}
		}
	},
	// This runs only if no subcommand is provided
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Set verbosity level for HTTP client
	//httpclient.SetVerbose(verbose)

	// Silence usage globally; let Cobra show usage only on flag/arg parsing errors
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true // Stop Cobra from automatically showing usage on errors
	if err := rootCmd.Execute(); err != nil {
		// For all runtime errors, just print them cleanly
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		defer klog.Flush() // ensure logs are written before exit
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().IntVarP(&verbosity, "v", "v", 0, "Log verbosity level (0=Errors only)")
	// Customize the version output format
	rootCmd.SetVersionTemplate("{{.Name}} version : {{.Version}}\n")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Initialize klog flags into global flagset
		klog.InitFlags(nil)

		// Apply verbosity from CLI flag to klog
		flag.Set("v", fmt.Sprintf("%d", verbosity))

		// Optional: always log to stderr (otherwise klog can default to files)
		flag.Set("logtostderr", "true")
	}

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, table, yaml)")
}

func initLoggingFlags(verbosity int) {
	// Reset the default global FlagSet to avoid "flag redefined" panic
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Initialize klog flags into the new flag set
	klog.InitFlags(nil)

	// Apply verbosity from CLI flag to klog
	_ = flag.Set("v", fmt.Sprintf("%d", verbosity))

	// Always log to stderr (otherwise klog can log to files by default)
	_ = flag.Set("logtostderr", "true")
}
