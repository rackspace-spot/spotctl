package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/google/uuid"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal"
	"github.com/rackspace-spot/spotctl/internal/ui"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

type interactiveModel struct {
	client      *internal.Client
	cfg         *config.SpotConfig
	params      createCloudspaceParams
	currentStep int
	steps       []func() error
	err         error
	cancelled   bool
}

// createCloudspaceParams holds all parameters needed for cloudspace creation
type createCloudspaceParams struct {
	Name                 string                     `json:"name" yaml:"name"`
	Org                  string                     `json:"org" yaml:"org"`
	Region               string                     `json:"region" yaml:"region"`
	KubernetesVersion    string                     `json:"kubernetesVersion" yaml:"kubernetesVersion"`
	PreemptionWebhookURL string                     `json:"preemptionWebhookURL" yaml:"preemptionWebhookURL"`
	CNI                  string                     `json:"cni" yaml:"cni"`
	ConfigPath           string                     `json:"-" yaml:"-"`
	SpotNodePools        []rxtspot.SpotNodePool     `json:"spotNodePools,omitempty" yaml:"spotNodePools,omitempty"`
	OnDemandNodePools    []rxtspot.OnDemandNodePool `json:"onDemandNodePools,omitempty" yaml:"onDemandNodePools,omitempty"`
}

const (
	HKG_HKG_1        = "hkg-hkg-1"
	US_CENTRAL_ORD_1 = "us-central-ord-1"
	AUS_SYD_1        = "aus-syd-1"
	UK_LON_1         = "uk-lon-1"
	US_EAST_IAD_1    = "us-east-iad-1"
	US_CENTRAL_DFW_1 = "us-central-dfw-1"
	US_CENTRAL_DFW_2 = "us-central-dfw-2"
	US_WEST_SJC_1    = "us-west-sjc-1"
)

// cloudspacesCmd represents the cloudspaces command
var cloudspacesCmd = &cobra.Command{
	Use:     "cloudspaces",
	Short:   "Manage cloudspaces",
	Long:    `Manage Rackspace Spot cloudspaces (Kubernetes clusters).`,
	Aliases: []string{"cs", "cloudspace"},
}

func init() {
	rootCmd.AddCommand(cloudspacesCmd)
	cloudspacesCmd.AddCommand(cloudspacesListCmd)
	cloudspacesCmd.AddCommand(cloudspacesCreateCmd)
	cloudspacesCmd.AddCommand(cloudspacesGetCmd)
	cloudspacesCmd.AddCommand(cloudspacesDeleteCmd)
	cloudspacesCmd.AddCommand(cloudspacesGetConfigCmd)

	// Add flags for cloudspaces list
	cloudspacesListCmd.Flags().String("org", "", "Organization ID")
	cloudspacesListCmd.Flags().StringP("output", "o", "json", "Output format (json, table, yaml)")

	// Add flags for cloudspaces create
	cloudspacesCreateCmd.Flags().String("name", "", "Cloudspace name")
	cloudspacesCreateCmd.Flags().String("org", "", "Organization ID")
	cloudspacesCreateCmd.Flags().String("region", "", "Region ")
	cloudspacesCreateCmd.Flags().StringP("kubernetes-version", "", "1.31.1", "Kubernetes version (default: 1.31.1)")
	cloudspacesCreateCmd.Flags().String("preemption-webhook-url", "", "Preemption webhook URL")

	cloudspacesCreateCmd.Flags().StringArray("spot-nodepool", []string{}, "Spot nodepool details in key=value format (e.g., desired=1,serverclass=gp.vs1.medium-ord,bidprice=0.08)")
	cloudspacesCreateCmd.Flags().StringArray("ondemand-nodepool", []string{}, "Ondemand nodepool details in key=value format (e.g., desired=1,serverclass=gp.vs1.medium-ord)")
	cloudspacesCreateCmd.Flags().String("config", "", "Path to config file (YAML or JSON)")
	cloudspacesCreateCmd.Flags().StringP("cni", "", "calico", "CNI (default: calico)")

	// Add flags for cloudspaces get
	cloudspacesGetCmd.Flags().String("name", "", "Cloudspace name (required)")
	cloudspacesGetCmd.Flags().String("org", "", "Organization ID")
	cloudspacesGetCmd.MarkFlagRequired("name")

	// Add flags for cloudspaces get-config
	cloudspacesGetConfigCmd.Flags().String("name", "", "Cloudspace name (required)")
	cloudspacesGetConfigCmd.Flags().String("org", "", "Organization ID")
	cloudspacesGetConfigCmd.Flags().String("file", "", "Output file name (default: <cloudspace_name>.yaml)")
	cloudspacesGetConfigCmd.MarkFlagRequired("name")

	// Add flags for cloudspaces delete
	cloudspacesDeleteCmd.Flags().String("name", "", "Cloudspace name (required)")
	cloudspacesDeleteCmd.Flags().String("org", "", "Organization ID")
	cloudspacesDeleteCmd.MarkFlagRequired("name")
	cloudspacesDeleteCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts; assume \"yes\" as answer")
}

// cloudspacesListCmd represents the cloudspaces list command
var cloudspacesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloudspaces",
	Long:  `List all cloudspaces in an organization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		org, _ := cmd.Flags().GetString("org")
		if org == "" {
			if err == nil && cfg.Org != "" {
				org = cfg.Org
			}
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotcli configure')")
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		cloudspaces, err := client.GetAPI().ListCloudspaces(context.Background(), org)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(cloudspaces, outputFormat)
	},
}

// cloudspacesDeleteCmd represents the cloudspaces delete command
var cloudspacesDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cloudspace",
	Long:  `Delete a cloudspace and all its resources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("name is required")
		}
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		org, _ := cmd.Flags().GetString("org")
		if org == "" {
			if err == nil && cfg.Org != "" {
				org = cfg.Org
			}
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotcli configure')")
		}
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			// Interactive prompt
			prompt := color.New(color.FgYellow).PrintfFunc()
			prompt("Are you sure you want to delete cloudspace '%s'? (y/N): ", name)

			var response string
			_, err := fmt.Scanln(&response)
			if err != nil || (response != "y" && response != "Y") {
				fmt.Println("Aborted.")
				return nil
			}
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		err = client.GetAPI().DeleteCloudspace(context.Background(), org, name)
		if err != nil {
			if rxtspot.IsNotFound(err) {
				return fmt.Errorf("cloudspace '%s' not found", name)
			}
			if rxtspot.IsForbidden(err) {
				return fmt.Errorf("forbidden: %w", err)
			}
			if rxtspot.IsConflict(err) {
				return fmt.Errorf("conflict: %w", err)
			}
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("Cloudspace '%s' deleted successfully\n", name)
		return nil
	},
}

// cloudspacesCreateCmd represents the cloudspaces create command
var cloudspacesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cloudspace",
	Long:  `Create a new Rackspace Spot cloudspace (Kubernetes cluster) with optional spot and on-demand node pools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create a cancellable context
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Handle interrupt signals
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\n\nOperation cancelled by user")
			cancel()
		}()
		// Get CLI configuration
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return fmt.Errorf("failed to get CLI configuration: %w", err)
		}

		// Initialize client
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		// Check if we're in interactive mode
		interactive := isInteractiveMode(cmd)

		// Load parameters based on mode
		var params *createCloudspaceParams
		if interactive {
			// Interactive mode - collect input from user
			params, err = collectInteractiveInput(client, cfg)
			if err != nil {
				return fmt.Errorf("failed to collect interactive input: %w", err)
			}
		} else {
			// Non-interactive mode - load from flags
			params, err = loadParamsFromFlags(cmd)
			if err != nil {
				return fmt.Errorf("failed to load parameters from flags: %w", err)
			}
		}

		// Set default values
		if params.Org == "" && cfg.Org != "" {
			params.Org = cfg.Org
		}
		if params.Region == "" && cfg.Region != "" {
			params.Region = cfg.Region
		}
		// Validate parameters
		if err := validateCreateParams(params, interactive); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// Check if context was cancelled before starting creation
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled")
		default:
			// Continue with creation
		}

		// Create cloudspace with all required fields
		cloudspace := rxtspot.CloudSpace{
			Name:                 params.Name,
			Org:                  params.Org,
			Region:               params.Region,
			KubernetesVersion:    params.KubernetesVersion,
			CNI:                  params.CNI,
			PreemptionWebhookURL: params.PreemptionWebhookURL,
		}

		// Temporary debug to verify values before API call
		fmt.Printf("Creating cloudspace: Name=%q Org=%q Region=%q K8s=%q CNI=%q\n",
			cloudspace.Name, cloudspace.Org, cloudspace.Region, cloudspace.KubernetesVersion, cloudspace.CNI)

		if err := client.GetAPI().CreateCloudspace(ctx, cloudspace); err != nil {
			return fmt.Errorf("failed to create cloudspace: %w", err)
		}
		// Create spot node pools if any
		for _, pool := range params.SpotNodePools {
			// Check if context was cancelled before each pool creation
			select {
			case <-ctx.Done():
				// Clean up the cloudspace if we're cancelled mid-creation
				if err := client.GetAPI().DeleteCloudspace(ctx, params.Org, params.Name); err != nil {
					klog.Warningf("Failed to clean up cloudspace after cancellation: %v", err)
				}
				return fmt.Errorf("operation cancelled during spot pool creation")
			default:
				// Continue with pool creation
			}

			// Ensure bid price is properly formatted
			bidPrice, err := validateBidPrice(pool.BidPrice)
			if err != nil {
				return fmt.Errorf("invalid bid price for pool %s: %w", pool.Name, err)
			}

			// Validate the bid price
			bidPrice, err = getBidPrice(bidPrice)
			if err != nil {
				return fmt.Errorf("invalid bid price for pool %s: %w", pool.Name, err)
			}
			if pool.Name == "" {
				pool.Name = uuid.NewString()
			}

			spotPool := rxtspot.SpotNodePool{
				Name:        pool.Name,
				Org:         params.Org,
				Cloudspace:  params.Name,
				ServerClass: pool.ServerClass,
				BidPrice:    bidPrice,
				Desired:     pool.Desired,
			}

			// Create the spot node pool with context
			createErr := client.GetAPI().CreateSpotNodePool(ctx, params.Org, spotPool)
			if createErr != nil {
				err = client.GetAPI().DeleteCloudspace(context.Background(), params.Org, params.Name)
				if err != nil {
					return fmt.Errorf("failed to delete cloudspace %s: %w", params.Name, err)
				}
				return fmt.Errorf("failed to create spot node pool %s : %w", spotPool.Name, createErr)
			}

			// Verify the pool was created successfully
			if _, verifyErr := client.GetAPI().GetSpotNodePool(context.Background(), params.Org, spotPool.Name); verifyErr != nil {
				err = fmt.Errorf("failed to verify creation of spot node pool %s: %w", spotPool.Name, verifyErr)
				return err
			}
		}

		// Create on-demand node pools if any
		for _, pool := range params.OnDemandNodePools {
			// Check if context was cancelled before each pool creation
			select {
			case <-ctx.Done():
				// Clean up the cloudspace if we're cancelled mid-creation
				if err := client.GetAPI().DeleteCloudspace(ctx, params.Org, params.Name); err != nil {
					klog.Warningf("Failed to clean up cloudspace after cancellation: %v", err)
				}
				return fmt.Errorf("operation cancelled during on-demand pool creation")
			default:
				// Continue with pool creation
			}

			if pool.Name == "" {
				pool.Name = uuid.NewString()
			}
			onDemandPool := rxtspot.OnDemandNodePool{
				Name:        pool.Name,
				Org:         params.Org,
				Cloudspace:  params.Name,
				ServerClass: pool.ServerClass,
				Desired:     pool.Desired,
			}

			// Create the on-demand node pool with context
			createErr := client.GetAPI().CreateOnDemandNodePool(ctx, params.Org, onDemandPool)
			if createErr != nil {
				err = client.GetAPI().DeleteCloudspace(context.Background(), params.Org, params.Name)
				if err != nil {
					return fmt.Errorf("failed to delete cloudspace %s: %w", params.Name, err)
				}
				return fmt.Errorf("failed to create on-demand node pool %s: %w", onDemandPool.Name, createErr)
			}

			// Verify the pool was created successfully
			if _, verifyErr := client.GetAPI().GetOnDemandNodePool(context.Background(), params.Org, onDemandPool.Name); verifyErr != nil {
				return fmt.Errorf("failed to verify creation of on-demand node pool %s: %w", onDemandPool.Name, verifyErr)
			}
		}

		cloudspaceGetResponse, err := client.GetAPI().GetCloudspace(context.Background(), params.Org, params.Name)
		if err != nil {
			return fmt.Errorf("failed to get cloudspace: %w", err)
		}
		// If we got here, everything was successful
		fmt.Printf("\n%s Successfully created cloudspace '%s' in region '%s'\n",
			color.GreenString("✓"),
			color.CyanString(cloudspaceGetResponse.Name),
			color.CyanString(cloudspaceGetResponse.Region),
		)

		// Check if context was cancelled before final output
		select {
		case <-ctx.Done():
			// Clean up the cloudspace if we're cancelled at the last moment
			if err := client.GetAPI().DeleteCloudspace(ctx, params.Org, params.Name); err != nil {
				klog.Warningf("Failed to clean up cloudspace after cancellation: %v", err)
			}
			return fmt.Errorf("operation cancelled during finalization")
		default:
			// Output the created cloudspace details
			return internal.OutputData(cloudspaceGetResponse, outputFormat)
		}
	},
}

// cloudspacesGetCmd represents the cloudspaces get command
var cloudspacesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get cloudspace details",
	Long:  `Get details about a specific cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("name is required")
		}

		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		org, _ := cmd.Flags().GetString("org")
		if org == "" && cfg != nil && cfg.Org != "" {
			org = cfg.Org
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotcli configure')")
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		cloudspace, err := client.GetAPI().GetCloudspace(context.Background(), org, name)
		if err != nil {
			if rxtspot.IsNotFound(err) {
				return fmt.Errorf("cloudspace '%s' not found", name)
			}
			return fmt.Errorf("failed to get cloudspace: %w", err)
		}

		// Get output format from flags, default to "json"
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "" {
			outputFormat = "json"
		}

		// Use the OutputData function for all output formats
		return internal.OutputData(cloudspace, outputFormat)
	},
}

// cloudspacesGetConfigCmd represents the cloudspaces get-config command
var cloudspacesGetConfigCmd = &cobra.Command{
	Use:   "get-config",
	Short: "Get cloudspace/kubernetes config",
	Long:  `Get config for a specific cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.GetCLIEssentials(cmd)

		org, _ := cmd.Flags().GetString("org")
		if org == "" {
			if err == nil && cfg.Org != "" {
				org = cfg.Org
			}
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotcli configure')")
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("name is required")
		}

		var filePath string
		fileName, _ := cmd.Flags().GetString("file")
		if fileName == "" {
			filePath = filepath.Join(os.Getenv("HOME"), ".kube", name+".yaml")
		} else {
			filePath = fileName + "/" + name + ".yaml"
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		k8sConfig, err := client.GetAPI().GetCloudspaceConfig(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = os.WriteFile(filePath, []byte(k8sConfig), 0644)
		if err != nil {
			return fmt.Errorf("failed to write config to file: %w", err)
		}
		fmt.Fprintf(os.Stdout, "Config has been saved to %s successfully\n", filePath)
		return nil
	},
}

// getBidPrice parses and validates the minimum bid price
func getBidPrice(priceStr string) (string, error) {
	if priceStr == "" {
		return "", fmt.Errorf("empty price")
	}

	// Remove all whitespace and dollar signs
	trimmed := strings.TrimSpace(strings.ReplaceAll(priceStr, "$", ""))
	if trimmed == "" {
		return "", fmt.Errorf("no valid price found in: %q", priceStr)
	}

	// Parse the price as a float64
	price, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		// Try to clean up the string and parse again
		var cleanNum strings.Builder
		decimalFound := false
		for _, c := range trimmed {
			if c >= '0' && c <= '9' {
				cleanNum.WriteRune(c)
			} else if c == '.' && !decimalFound {
				cleanNum.WriteRune(c)
				decimalFound = true
			}
		}

		if cleanNum.Len() == 0 {
			return "", fmt.Errorf("invalid price format: %q (no valid numbers found)", priceStr)
		}

		price, err = strconv.ParseFloat(cleanNum.String(), 64)
		if err != nil {
			return "", fmt.Errorf("invalid price format: %q: %v", priceStr, err)
		}
	}

	if price <= 0 {
		return "", fmt.Errorf("price must be greater than 0")
	}

	// Format with up to 3 decimal places
	return fmt.Sprintf("%g", price), nil
}

// collectInteractiveInput gathers all required parameters interactively using BubbleTea
func collectInteractiveInput(client *internal.Client, cfg *config.SpotConfig) (*createCloudspaceParams, error) {
	fmt.Println("\nStarting interactive cloudspace creation...")
	// Initialize the interactive model (holds params and step functions)
	model := initInteractiveModel(client, cfg)

	// Execute each interactive step sequentially. Each step handles its own prompt.
	for _, step := range model.steps {
		if err := step(); err != nil {
			return nil, err
		}
		if model.cancelled {
			return nil, fmt.Errorf("interactive prompt cancelled")
		}
		if model.err != nil {
			return nil, model.err
		}
	}

	// Validate the collected parameters
	if err := validateCreateParams(&model.params, true); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	// Return a copy to avoid any unintended aliasing of the model's internal field
	cp := model.params
	return &cp, nil
}

// validateCreateParams validates the provided parameters
func validateCreateParams(params *createCloudspaceParams, interactive bool) error {
	// Skip validation in interactive mode as we'll collect all required parameters
	if interactive {
		return nil
	}

	// Non-interactive mode validations
	if params.Name == "" {
		return fmt.Errorf("name is required")
	}

	if params.Region == "" {
		return fmt.Errorf("region is required")

	}

	if !isValidRegion(params.Region) {
		return fmt.Errorf("region %s is not valid. Available regions: %s, %s, %s, %s, %s, %s, %s, %s", params.Region, US_CENTRAL_ORD_1, HKG_HKG_1, AUS_SYD_1, UK_LON_1, US_EAST_IAD_1, US_CENTRAL_DFW_1, US_CENTRAL_DFW_2, US_WEST_SJC_1)
	}

	// Require at least one node pool in non-interactive mode
	if len(params.SpotNodePools) == 0 && len(params.OnDemandNodePools) == 0 {
		return fmt.Errorf("at least one node pool is required when using flags (use --spot-nodepool or --ondemand-nodepool)")
	}

	// Validate spot node pools' bid prices
	for i, pool := range params.SpotNodePools {
		if pool.BidPrice == "" {
			return fmt.Errorf("bid price is required for spot node pool %s", pool.Name)
		}
		_, err := validateBidPrice(pool.BidPrice)
		if err != nil {
			return fmt.Errorf("invalid bid price for pool %s: %w", pool.Name, err)
		}
		params.SpotNodePools[i].BidPrice, _ = validateBidPrice(pool.BidPrice)
	}

	for i, pool := range params.OnDemandNodePools {
		if pool.Desired <= 0 {
			return fmt.Errorf("desired number of nodes must be greater than 0 for on-demand node pool %s", pool.Name)
		}
		params.OnDemandNodePools[i].Desired = pool.Desired
	}
	return nil
}

// loadParamsFromFlags loads parameters from command line flags and config file if provided
func loadParamsFromFlags(cmd *cobra.Command) (*createCloudspaceParams, error) {
	params := &createCloudspaceParams{}

	// First check if config file is provided
	configPath, _ := cmd.Flags().GetString("config")
	if configPath != "" {
		// Read the entire file content
		content, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// Parse based on file extension
		var fullConfig struct {
			CloudSpace        rxtspot.CloudSpace         `json:"cloudspace" yaml:"cloudspace"`
			SpotNodePools     []rxtspot.SpotNodePool     `json:"spotnodepools" yaml:"spotnodepools"`
			OnDemandNodePools []rxtspot.OnDemandNodePool `json:"ondemandnodepools" yaml:"ondemandnodepools"`
		}

		ext := strings.ToLower(filepath.Ext(configPath))
		switch ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(content, &fullConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
			}
		case ".json":
			if err := json.Unmarshal(content, &fullConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON config: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported config file format: %s (must be .yaml, .yml, or .json)", ext)
		}
		// Map the config to our params and return
		params.Name = fullConfig.CloudSpace.Name
		params.Org = fullConfig.CloudSpace.Org
		params.Region = fullConfig.CloudSpace.Region
		params.KubernetesVersion = fullConfig.CloudSpace.KubernetesVersion
		params.CNI = fullConfig.CloudSpace.CNI
		params.SpotNodePools = fullConfig.SpotNodePools
		params.OnDemandNodePools = fullConfig.OnDemandNodePools
		return params, nil
	}

	// If no config file, load all parameters from flags
	params.Name, _ = cmd.Flags().GetString("name")
	params.Org, _ = cmd.Flags().GetString("org")
	params.Region, _ = cmd.Flags().GetString("region")
	params.KubernetesVersion, _ = cmd.Flags().GetString("kubernetes-version")
	params.PreemptionWebhookURL, _ = cmd.Flags().GetString("preemption-webhook-url")
	params.CNI, _ = cmd.Flags().GetString("cni")

	// Handle node pools - these will be parsed later
	spotPools, _ := cmd.Flags().GetStringArray("spot-nodepool")
	onDemandPools, _ := cmd.Flags().GetStringArray("ondemand-nodepool")

	// Convert string pools to actual node pool objects
	for _, poolStr := range spotPools {
		poolParams, err := parseNodepoolParams(poolStr)
		if err != nil {
			klog.Warningf("Failed to parse spot nodepool params '%s': %v", poolStr, err)
			continue
		}

		desired, _ := strconv.Atoi(poolParams["desired"])
		if desired <= 0 {
			desired = 1 // Default to 1 if not specified or invalid
		}

		spotPool := rxtspot.SpotNodePool{
			Name:        uuid.New().String(),
			Org:         poolParams["org"],
			Cloudspace:  poolParams["cloudspace"],
			ServerClass: poolParams["serverclass"],
			BidPrice:    poolParams["bidprice"],
			Desired:     desired,
		}
		params.SpotNodePools = append(params.SpotNodePools, spotPool)
	}

	for _, poolStr := range onDemandPools {
		poolParams, err := parseNodepoolParams(poolStr)
		if err != nil {
			klog.Warningf("Failed to parse on-demand nodepool params '%s': %v", poolStr, err)
			continue
		}

		desired, _ := strconv.Atoi(poolParams["desired"])
		if desired <= 0 {
			desired = 1 // Default to 1 if not specified or invalid
		}

		onDemandPool := rxtspot.OnDemandNodePool{
			Name:        uuid.New().String(),
			Org:         poolParams["org"],
			Cloudspace:  poolParams["cloudspace"],
			ServerClass: poolParams["serverclass"],
			Desired:     desired,
		}
		params.OnDemandNodePools = append(params.OnDemandNodePools, onDemandPool)
	}

	// If we got here with no node pools and no config file, that's an error
	if len(params.SpotNodePools) == 0 && len(params.OnDemandNodePools) == 0 && params.ConfigPath == "" {
		return nil, fmt.Errorf("no node pools specified and no config file provided")
	}

	return params, nil
}

// isInteractiveMode checks if we should run in interactive mode
// Interactive mode should only be used when no flags are provided at all
func isInteractiveMode(cmd *cobra.Command) bool {
	// If --config flag is provided, never use interactive mode
	if configPath, _ := cmd.Flags().GetString("config"); configPath != "" {
		return false
	}

	// Check if any flags were provided
	flagSet := make(map[string]bool)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagSet[f.Name] = true
	})

	// If any flags were provided, don't use interactive mode
	return len(flagSet) == 0
}

// validateBidPrice validates and formats a bid price string to ensure it has up to 3 decimal places
func validateBidPrice(bidPrice string) (string, error) {
	// Check if it's a valid number
	val, err := strconv.ParseFloat(bidPrice, 64)
	if err != nil {
		return "", fmt.Errorf("bid price must be a valid number")
	}

	// Ensure it's a positive number
	if val <= 0 {
		return "", fmt.Errorf("bid price must be greater than 0")
	}

	// Format to exactly 3 decimal places
	formatted := fmt.Sprintf("%.3f", val)

	// Remove trailing zeros after decimal point for cleaner output
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimSuffix(formatted, ".")

	// Ensure we have at least one decimal place if it was a whole number
	if !strings.Contains(formatted, ".") && val == float64(int64(val)) {
		formatted = fmt.Sprintf("%s.000", formatted)
	} else if strings.Count(formatted, ".") > 0 {
		// Ensure exactly 3 decimal places
		parts := strings.Split(formatted, ".")
		if len(parts) == 2 && len(parts[1]) < 3 {
			formatted = fmt.Sprintf("%s%s", formatted, strings.Repeat("0", 3-len(parts[1])))
		}
	}

	return formatted, nil
}

// parseNodepoolParams parses nodepool parameters in format key1=value1,key2=value2
func parseNodepoolParams(params string) (map[string]string, error) {
	if params == "" {
		return nil, nil
	}
	result := make(map[string]string)
	pairs := strings.Split(params, ",")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid parameter format: %s, expected key=value", pair)
		}
		result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return result, nil
}

func isValidRegion(region string) bool {

	switch region {
	case HKG_HKG_1, US_CENTRAL_ORD_1, AUS_SYD_1, UK_LON_1, US_EAST_IAD_1, US_CENTRAL_DFW_1, US_CENTRAL_DFW_2, US_WEST_SJC_1:
		return true
	default:
		return false
	}
}

func initInteractiveModel(client *internal.Client, cfg *config.SpotConfig) *interactiveModel {
	m := &interactiveModel{
		client: client,
		cfg:    cfg,
		params: createCloudspaceParams{
			KubernetesVersion: "1.31.1",
			CNI:               "calico",
		},
	}

	// Define the steps of our interactive flow
	m.steps = []func() error{
		m.stepSelectRegion,
		m.stepEnterName,
		m.stepSelectKubernetesVersion,
		m.stepSelectCNI,
		m.stepAddNodePools,
		m.stepSummaryAndConfirm,
	}

	return m
}

// Init initializes the interactive model
func (m *interactiveModel) Init() tea.Cmd {
	return nil
}

func (m *interactiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *interactiveModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit\n", m.err)
	}

	if m.cancelled {
		return "Operation cancelled\n"
	}

	if m.currentStep >= len(m.steps) {
		return "Configuration complete!\n"
	}

	return ""
}

func (m *interactiveModel) stepSelectRegion() error {
	fmt.Println("Fetching available regions...")

	// Try to get available regions
	regions, err := m.client.GetAPI().ListRegions(context.Background())
	if err != nil || len(regions) == 0 {
		// Fallback to manual input if listing regions is not permitted or empty
		region, ierr := internal.PromptForString("Enter region (e.g., ord, iad, dfw)", m.params.Region)
		if ierr != nil {
			return fmt.Errorf("region input failed: %w", ierr)
		}
		region = strings.TrimSpace(region)
		if region == "" {
			fmt.Println("Region cannot be empty. Please enter a valid region.")
			return m.stepSelectRegion()
		}
		m.params.Region = region
		return nil
	}

	// Convert regions to string slice for selection
	var regionOptions []string
	for _, r := range regions {
		regionOptions = append(regionOptions, r.Name)
	}

	// Create and run the selection prompt
	p := tea.NewProgram(ui.NewSelectModel(regionOptions))
	m2, err := p.Run()
	if err != nil {
		return fmt.Errorf("region selection failed: %w", err)
	}

	if sm, ok := m2.(ui.SelectModel); ok {
		if sm.Cancelled() {
			m.cancelled = true
			return nil
		}
		if sm.Selected() != "" {
			m.params.Region = sm.Selected()
			fmt.Printf("Selected region: %s\n", color.CyanString(m.params.Region))
			return nil
		}
	}

	// If we get here, treat as cancelled
	m.cancelled = true
	return nil
}

func (m *interactiveModel) stepEnterName() error {
	fmt.Printf("\n%s Enter a name for your cloudspace:\n", color.GreenString("?"))
	for {
		// Create and run the input prompt
		p := tea.NewProgram(ui.NewInputModel("Enter cloudspace name", "", false))
		m2, err := p.Run()
		if err != nil {
			return fmt.Errorf("name input failed: %w", err)
		}

		if im, ok := m2.(ui.InputModel); ok {
			if im.Cancelled() {
				m.cancelled = true
				return nil
			}
			val := strings.TrimSpace(im.Value())
			if val != "" {
				m.params.Name = val
				fmt.Printf("%s Enter a name for your cloudspace: %s\n", color.GreenString("?"), color.CyanString(m.params.Name))
				return nil
			}
		}

		// Inform and retry unless the overall session was cancelled elsewhere
		fmt.Println("Name cannot be empty. Please enter a valid name.")
		// continue loop to re-prompt
	}
}

func (m *interactiveModel) stepSelectKubernetesVersion() error {
	fmt.Printf("\n%s Select Kubernetes version:\n", color.GreenString("?"))
	// Define available versions
	versions := []string{"1.31.1", "1.30.10", "1.29.6"}

	// Create and run the selection prompt
	p := tea.NewProgram(ui.NewSelectModel(versions))
	m2, err := p.Run()
	if err != nil {
		return fmt.Errorf("kubernetes version selection failed: %w", err)
	}

	if sm, ok := m2.(ui.SelectModel); ok {
		if sm.Cancelled() {
			m.cancelled = true
			return nil
		}
		if sm.Selected() != "" {
			m.params.KubernetesVersion = sm.Selected()
			fmt.Printf("%s Select Kubernetes version: %s\n", color.GreenString("?"), color.CyanString(m.params.KubernetesVersion))
			return nil
		}
	}

	// If we get here, treat as cancelled
	m.cancelled = true
	return nil
}

func (m *interactiveModel) stepSelectCNI() error {
	fmt.Printf("\n%s Select CNI plugin:\n", color.GreenString("?"))
	// Define available CNIs
	cniOptions := []string{"calico", "cilium", "bring your own CNI"}

	// Create and run the selection prompt
	p := tea.NewProgram(ui.NewSelectModel(cniOptions))
	m2, err := p.Run()
	if err != nil {
		return fmt.Errorf("cni selection failed: %w", err)
	}

	if sm, ok := m2.(ui.SelectModel); ok {
		if sm.Cancelled() {
			m.cancelled = true
			return nil
		}
		if sm.Selected() != "" {
			m.params.CNI = sm.Selected()
			fmt.Printf("%s Select CNI plugin: %s\n", color.GreenString("?"), color.CyanString(m.params.CNI))
			return nil
		}
	}

	// If we get here, treat as cancelled
	m.cancelled = true
	return nil
}

func (m *interactiveModel) stepAddNodePools() error {
	for {
		// Ask pool type
		poolType, err := m.client.PromptForPoolType()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				m.cancelled = true
				return nil
			}
			return fmt.Errorf("failed to select node pool type: %w", err)
		}
		fmt.Printf("%s Add a node pool: %s\n", color.GreenString("?"), color.CyanString(poolType))

		// Get server class and pricing depending on pool type
		var (
			serverClass   string
			minBidPrice   string
			onDemandPrice string
		)
		if strings.EqualFold(poolType, "Spot") {
			sc, minBid, _, err := m.client.PromptForServerClassWithBidPrice(context.Background(), m.params.Region, "spot")
			if err != nil {
				if errors.Is(err, context.Canceled) {
					m.cancelled = true
					return nil
				}
				return fmt.Errorf("failed to select server class: %w", err)
			}
			serverClass = sc
			minBidPrice = minBid

			// Get desired nodes
			desiredStr, err := m.client.PromptForNodeCount("spot")
			if err != nil {
				if errors.Is(err, context.Canceled) {
					m.cancelled = true
					return nil
				}
				return fmt.Errorf("failed to get node count: %w", err)
			}
			desired, err := strconv.Atoi(strings.TrimSpace(desiredStr))
			if err != nil || desired < 1 {
				fmt.Println("Please enter a valid number >= 1.")
				continue
			}

			// Get bid price
			bidMsg := fmt.Sprintf("Enter your maximum bid price (minimum: $%s)", minBidPrice)
			bidPrice, err := m.client.PromptForBidPrice(bidMsg, minBidPrice)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					m.cancelled = true
					return nil
				}
				return fmt.Errorf("failed to get bid price: %w", err)
			}
			bidPrice, err = validateBidPrice(bidPrice)
			if err != nil {
				fmt.Printf("Invalid bid price: %v\n", err)
				continue
			}
			fmt.Printf("%s Enter your maximum bid price (minimum: $%s) %s\n", color.GreenString("?"), minBidPrice, color.CyanString(bidPrice))

			// Add spot pool
			m.params.SpotNodePools = append(m.params.SpotNodePools, rxtspot.SpotNodePool{
				Name:        uuid.New().String(),
				ServerClass: serverClass,
				BidPrice:    bidPrice,
				Desired:     desired,
			})
		} else { // On-Demand
			sc, _, odPrice, err := m.client.PromptForServerClassWithBidPrice(context.Background(), m.params.Region, "ondemand")
			if err != nil {
				if errors.Is(err, context.Canceled) {
					m.cancelled = true
					return nil
				}
				return fmt.Errorf("failed to select server class: %w", err)
			}
			serverClass = sc
			onDemandPrice = odPrice

			// Get desired nodes
			desiredStr, err := m.client.PromptForNodeCount("on-demand")
			if err != nil {
				if errors.Is(err, context.Canceled) {
					m.cancelled = true
					return nil
				}
				return fmt.Errorf("failed to get node count: %w", err)
			}
			desired, err := strconv.Atoi(strings.TrimSpace(desiredStr))
			if err != nil || desired < 1 {
				fmt.Println("Please enter a valid number >= 1.")
				continue
			}

			// Add on-demand pool
			m.params.OnDemandNodePools = append(m.params.OnDemandNodePools, rxtspot.OnDemandNodePool{
				Name:                 uuid.New().String(),
				ServerClass:          serverClass,
				Desired:              desired,
				OnDemandPricePerHour: onDemandPrice,
			})
		}

		// Ask to add another node pool
		more, err := internal.Confirm("Add another node pool?", false)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				m.cancelled = true
				return nil
			}
			return fmt.Errorf("confirmation failed: %w", err)
		}
		if !more {
			break
		}
	}
	return nil
}

func (m *interactiveModel) stepSummaryAndConfirm() error {
	// Summary header
	fmt.Println("\nCloudspace Configuration:")
	fmt.Printf(`
Cloudspace Configuration:
• %-20s %s
• %-20s %s
• %-20s %s
• %-20s %s
`,
		"Name:", color.CyanString(m.params.Name),
		"Region:", color.CyanString(m.params.Region),
		"Kubernetes Version:", color.CyanString(m.params.KubernetesVersion),
		"CNI:", color.CyanString(m.params.CNI),
	)

	if len(m.params.SpotNodePools) > 0 {
		fmt.Println("\nSpot Node Pools:")
		for _, pool := range m.params.SpotNodePools {
			fmt.Printf(`  • %s
    %-15s %s
    %-15s %d
    %-15s $%s
`,
				color.CyanString(pool.Name),
				"Instance Type:", pool.ServerClass,
				"Desired Nodes:", pool.Desired,
				"Bid Price:", pool.BidPrice,
			)
		}
	}

	if len(m.params.OnDemandNodePools) > 0 {
		fmt.Println("\nOn-Demand Node Pools:")
		for _, pool := range m.params.OnDemandNodePools {
			fmt.Printf(`  • %s
    %-15s %s
    %-15s %d
    %-15s %s\n\n`,
				color.CyanString(pool.Name),
				"Instance Type:", pool.ServerClass,
				"Desired Nodes:", pool.Desired,
				"Price:", pool.OnDemandPricePerHour,
			)
		}
	}

	ok, err := internal.Confirm("\nCreate cloudspace with the above configuration?", true)
	if err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("cloudspace creation cancelled")
	}
	return nil
}
