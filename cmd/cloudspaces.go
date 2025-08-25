package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/google/uuid"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
	// "sigs.k8s.io/yaml"
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

		// Track created resources for cleanup in case of failure
		createdResources := struct {
			cloudspaceCreated bool
			nodePoolsCreated  []struct {
				name   string
				isSpot bool
			}
		}{
			nodePoolsCreated: make([]struct {
				name   string
				isSpot bool
			}, 0),
		}

		// Validate parameters
		if err := validateCreateParams(params, interactive); err != nil {
			return fmt.Errorf("validation failed: %w", err)
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

		if err := client.GetAPI().CreateCloudspace(context.Background(), cloudspace); err != nil {
			return fmt.Errorf("failed to create cloudspace: %w", err)
		}
		// Only mark cloudspace as created after successful creation
		createdResources.cloudspaceCreated = true

		// cleanupResources handles the actual cleanup of resources
		cleanupResources := func() {

			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Delete all created node pools in reverse order
			if len(createdResources.nodePoolsCreated) > 0 {
				for i := len(createdResources.nodePoolsCreated) - 1; i >= 0; i-- {
					np := createdResources.nodePoolsCreated[i]

					var cleanupErr error
					if np.isSpot {
						cleanupErr = client.GetAPI().DeleteSpotNodePool(cleanupCtx, params.Org, np.name)
					} else {
						cleanupErr = client.GetAPI().DeleteOnDemandNodePool(cleanupCtx, params.Org, np.name)
					}

					if cleanupErr != nil {
						klog.Warningf("Failed to clean up node pool %s: %v", np.name, cleanupErr)
					}
				}
			}

			// Delete the cloudspace
			if createdResources.cloudspaceCreated {
				if delErr := client.GetAPI().DeleteCloudspace(cleanupCtx, params.Org, params.Name); delErr != nil {
					klog.Warningf("Failed to clean up cloudspace %s: %v", params.Name, delErr)
				}
			}
		}

		// Defer cleanup function in case of error
		defer func() {
			if r := recover(); r != nil {
				cleanupResources()
				panic(r) // Re-throw the panic after cleanup
			} else if err != nil {
				if createdResources.cloudspaceCreated {
					cleanupResources()
				}
			}
		}()

		// Create spot node pools if any
		for _, pool := range params.SpotNodePools {
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

			// Track the pool before creation so we can clean it up if needed
			createdResources.nodePoolsCreated = append(createdResources.nodePoolsCreated,
				struct {
					name   string
					isSpot bool
				}{name: pool.Name, isSpot: true})

			// Create the spot node pool
			createErr := client.GetAPI().CreateSpotNodePool(context.Background(), params.Org, spotPool)
			if createErr != nil {
				// Store the original error
				err = fmt.Errorf("failed to create spot node pool %s: %w", spotPool.Name, createErr)
				// Explicitly clean up resources before returning
				cleanupResources()
				return err
			}

			// Verify the pool was created successfully
			if _, verifyErr := client.GetAPI().GetSpotNodePool(context.Background(), params.Org, spotPool.Name); verifyErr != nil {
				err = fmt.Errorf("failed to verify creation of spot node pool %s: %w", spotPool.Name, verifyErr)
				// Explicitly clean up resources before returning
				cleanupResources()
				return err
			}
		}

		// Create on-demand node pools if any
		for _, pool := range params.OnDemandNodePools {
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
			// Track the pool before creation so we can clean it up if needed
			createdResources.nodePoolsCreated = append(createdResources.nodePoolsCreated,
				struct {
					name   string
					isSpot bool
				}{name: pool.Name, isSpot: false})

			// Create the on-demand node pool
			createErr := client.GetAPI().CreateOnDemandNodePool(context.Background(), params.Org, onDemandPool)
			if createErr != nil {
				// Store the original error
				err = fmt.Errorf("failed to create on-demand node pool %s: %w", onDemandPool.Name, createErr)
				// Explicitly clean up resources before returning
				cleanupResources()
				return err
			}

			// Verify the pool was created successfully
			if _, verifyErr := client.GetAPI().GetOnDemandNodePool(context.Background(), params.Org, onDemandPool.Name); verifyErr != nil {
				err = fmt.Errorf("failed to verify creation of on-demand node pool %s: %w", onDemandPool.Name, verifyErr)
				// Explicitly clean up resources before returning
				cleanupResources()
				return err
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

		// Output the created cloudspace details
		return internal.OutputData(cloudspaceGetResponse, outputFormat)
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

// getBidPrice parses and validates the minimum bid price
// It returns the price as a string with proper formatting
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

// collectInteractiveInput gathers all required parameters interactively
func collectInteractiveInput(client *internal.Client, cfg *config.SpotConfig) (*createCloudspaceParams, error) {
	params := &createCloudspaceParams{}
	var err error

	fmt.Println("\nStarting interactive cloudspace creation...")

	// Region selection
	fmt.Println("\nFetching available regions...")
	defaultRegion := ""
	if cfg != nil && cfg.Region != "" {
		defaultRegion = cfg.Region
	}

	params.Region, err = client.PromptForRegionWithDefault(context.Background(), defaultRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to select region: %w", err)
	}
	fmt.Printf("\nSelected region: %s\n", color.GreenString(params.Region))

	// Cloudspace name
	namePrompt := &survey.Input{
		Message: "Enter a name for your cloudspace:",
	}
	if err := survey.AskOne(namePrompt, &params.Name); err != nil || params.Name == "" {
		return nil, fmt.Errorf("failed to get cloudspace name: %w", err)
	}

	// Kubernetes version selection
	fmt.Println("\nSelect Kubernetes version:")
	params.KubernetesVersion, err = client.PromptForKubernetesVersion("1.31.1")
	if err != nil {
		return nil, fmt.Errorf("failed to select Kubernetes version: %w", err)
	}

	// CNI selection
	fmt.Println("\nSelect CNI plugin:")
	params.CNI, err = client.PromptForCNI("calico")
	if err != nil {
		return nil, fmt.Errorf("failed to select CNI: %w", err)
	}

	// Node pool configuration
	// var nodePools []map[string]interface{}

	for {
		// Ask for node pool type
		poolType := ""
		poolPrompt := &survey.Select{
			Message: "Add a node pool:",
			Options: []string{"Spot", "On-Demand"},
		}
		if err := survey.AskOne(poolPrompt, &poolType); err != nil {
			return nil, fmt.Errorf("failed to select node pool type: %w", err)
		}

		// Get server class
		serverClass, minBidPrice, onDemandPrice, err := client.PromptForServerClassWithBidPrice(context.Background(), params.Region, strings.ToLower(poolType))
		if err != nil {
			return nil, fmt.Errorf("failed to select server class: %w", err)
		}

		// Generate node pool name
		nodePoolUUID := uuid.New().String()
		// namePrompt := &survey.Input{
		// 	Message: fmt.Sprintf("Enter a name for your %s node pool:", strings.ToLower(poolType)),
		// 	Default: nodePoolUUID,
		// }
		// if err := survey.AskOne(namePrompt, &nodePoolUUID); err != nil || nodePoolUUID == "" {
		// 	return nil, fmt.Errorf("node pool name is required")
		// }

		// Get node count
		nodeCount, err := client.PromptForNodeCount("")
		if err != nil {
			return nil, fmt.Errorf("failed to get node count: %w", err)
		}
		desired, _ := strconv.Atoi(nodeCount)

		if poolType == "Spot" {
			// Get the minimum bid price
			minBidPrice, err := getBidPrice(minBidPrice)
			if err != nil {
				return nil, fmt.Errorf("invalid minimum bid price: %w", err)
			}

			// Get bid price for spot
			bidPrice := ""
			bidPrompt := &survey.Input{
				Message: fmt.Sprintf("Enter your maximum bid price (minimum: $%s):", minBidPrice),
				Default: minBidPrice,
			}
			if err := survey.AskOne(bidPrompt, &bidPrice); err != nil || bidPrice == "" {
				return nil, fmt.Errorf("bid price is required")
			}

			// Validate the bid price
			bidPrice, err = getBidPrice(bidPrice)
			if err != nil {
				return nil, fmt.Errorf("invalid bid price: %w", err)
			}

			// Format bid price
			if bidPriceFloat, err := strconv.ParseFloat(bidPrice, 64); err == nil {
				bidPrice = fmt.Sprintf("%.3f", bidPriceFloat)
			}

			// Add to spot node pools
			pool := rxtspot.SpotNodePool{
				Name:        nodePoolUUID,
				ServerClass: serverClass,
				BidPrice:    bidPrice,
				Desired:     desired,
			}
			params.SpotNodePools = append(params.SpotNodePools, pool)
		} else {
			// Add to on-demand node pools
			pool := rxtspot.OnDemandNodePool{
				Name:                 nodePoolUUID,
				ServerClass:          serverClass,
				Desired:              desired,
				OnDemandPricePerHour: onDemandPrice,
			}
			params.OnDemandNodePools = append(params.OnDemandNodePools, pool)
		}

		// Ask to add another node pool
		if ok, err := internal.Confirm("\nAdd another node pool?", false); err != nil || !ok {
			break
		}
	}

	// Show configuration summary
	fmt.Println("\nCloudspace Configuration:")
	fmt.Printf("• %-20s %s\n", "Name:", color.CyanString(params.Name))
	fmt.Printf("• %-20s %s\n", "Region:", color.CyanString(params.Region))
	fmt.Printf("• %-20s %s\n", "Kubernetes Version:", color.CyanString(params.KubernetesVersion))
	fmt.Printf("• %-20s %s\n", "CNI:", color.CyanString(params.CNI))

	if len(params.SpotNodePools) > 0 {
		fmt.Println("\nSpot Node Pools:")
		for _, pool := range params.SpotNodePools {
			fmt.Printf("  • %s\n", color.CyanString(pool.Name))
			fmt.Printf("    %-15s %s\n", "Instance Type:", pool.ServerClass)
			fmt.Printf("    %-15s %d\n", "Desired Nodes:", pool.Desired)
			fmt.Printf("    %-15s $%s\n\n", "Bid Price:", pool.BidPrice)
		}
	}

	if len(params.OnDemandNodePools) > 0 {
		fmt.Println("\nOn-Demand Node Pools:")
		for _, pool := range params.OnDemandNodePools {
			fmt.Printf("  • %s\n", color.CyanString(pool.Name))
			fmt.Printf("    %-15s %s\n", "Instance Type:", pool.ServerClass)
			fmt.Printf("    %-15s %d\n", "Desired Nodes:", pool.Desired)
			fmt.Printf("    %-15s $%s\n\n", "Price:", pool.OnDemandPricePerHour)
		}
	}

	// Final confirmation
	confirm, err := internal.Confirm("\nCreate cloudspace with the above configuration?", true)
	if err != nil || !confirm {
		return nil, fmt.Errorf("cloudspace creation cancelled")
	}

	// Ensure at least one node pool is configured
	if len(params.SpotNodePools) == 0 && len(params.OnDemandNodePools) == 0 {
		return nil, fmt.Errorf("at least one node pool (spot or on-demand) is required")
	}

	return params, nil
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

	// Require at least one node pool in non-interactive mode
	if len(params.SpotNodePools) == 0 && len(params.OnDemandNodePools) == 0 {
		return fmt.Errorf("at least one node pool is required when using flags (use --spot-nodepool or --ondemand-nodepool)")
	}

	// Validate spot node pools' bid prices
	for i, pool := range params.SpotNodePools {
		// fmt.Printf("pool spot - %+v \n", pool)
		if pool.BidPrice == "" {
			return fmt.Errorf("bid price is required for spot node pool %s", pool.Name)
		}
		_, err := validateBidPrice(pool.BidPrice)
		if err != nil {
			return fmt.Errorf("invalid bid price for pool %s: %w", pool.Name, err)
		}
		// Update the bid price with the validated and formatted version
		params.SpotNodePools[i].BidPrice, _ = validateBidPrice(pool.BidPrice)
		// if pool.Name == "" {
		// 	pool.Name = uuid.New().String()
		// }
		// fmt.Printf("pool spot - %+v \n", pool)
	}

	for i, pool := range params.OnDemandNodePools {
		// fmt.Printf("pool node - %+v \n", pool)
		if pool.Desired <= 0 {
			return fmt.Errorf("desired number of nodes must be greater than 0 for on-demand node pool %s", pool.Name)
		}
		// Update the desired count with the validated value
		params.OnDemandNodePools[i].Desired = pool.Desired
		// if pool.Name == "" {
		// 	pool.Name = uuid.New().String()
		// }
		// fmt.Printf("pool node - %+v \n", pool)
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
