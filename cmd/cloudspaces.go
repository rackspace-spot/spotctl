package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"

	"sigs.k8s.io/yaml"
)

// cloudspacesCmd represents the cloudspaces command
var cloudspacesCmd = &cobra.Command{
	Use:   "cloudspaces",
	Short: "Manage cloudspaces",
	Long:  `Manage Rackspace Spot cloudspaces (Kubernetes clusters).`,
}

const defaultServerclass = "gp.vs1.medium-iad"

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

		fmt.Printf("Config has been saved to %s successfully\n", filePath)
		return nil
	},
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

// cloudspacesCreateCmd represents the cloudspaces create command
var cloudspacesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cloudspace",
	Long:  `Create a new cloudspace (Kubernetes cluster).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)

		name, _ := cmd.Flags().GetString("name")
		org, _ := cmd.Flags().GetString("org")
		region, _ := cmd.Flags().GetString("region")

		if org == "" {
			if err == nil && cfg.Org != "" {
				org = cfg.Org
			}
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotcli configure')")
		}

		configPath, _ := cmd.Flags().GetString("config")
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes_version")
		cni, _ := cmd.Flags().GetString("cni")

		spotNodePoolJSON, _ := cmd.Flags().GetStringArray("spot-nodepool")
		onDemandNodePoolJSON, _ := cmd.Flags().GetStringArray("ondemand-nodepool")

		// Debug: Print all flag values
		//	fmt.Printf("DEBUG - Flag values: name='%s', region='%s', configPath='%s', kubernetesVersion='%s', cni='%s', spotNodePools=%d, onDemandNodePools=%d\n",
		//		name, region, configPath, kubernetesVersion, cni, len(spotNodePoolJSON), len(onDemandNodePoolJSON))

		// Check if we're in interactive mode (no flags provided)
		// Note: kubernetesVersion and cni have default values, so we only check if they're explicitly set
		hasAnyFlag := name != "" || region != "" || configPath != "" ||
			(kubernetesVersion != "" && kubernetesVersion != "1.31.1") ||
			(cni != "" && cni != "calico") ||
			len(spotNodePoolJSON) > 0 || len(onDemandNodePoolJSON) > 0

		//	fmt.Printf("DEBUG - hasAnyFlag: %v\n", hasAnyFlag)

		// 1. Create client early since we need it for interactive prompts
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Check if we're using a config file
		if configPath != "" {
			// When --config is provided, skip all other validations and prompts
			fmt.Println("Using configuration from file:", configPath)
		} else {
			// hasAnyFlag is already set above

			// If any flag is provided, ensure all required flags are provided for non-interactive mode
			if hasAnyFlag {
				// Check for required flags
				if name == "" {
					return fmt.Errorf("name is required when using flags (use --name)")
				}
				if region == "" {
					return fmt.Errorf("region is required when using flags (use --region)")
				}
				if len(spotNodePoolJSON) == 0 && len(onDemandNodePoolJSON) == 0 {
					return fmt.Errorf("at least one node pool is required when using flags (use --spot-nodepool or --ondemand-nodepool)")
				}

				// In non-interactive mode, we don't prompt for anything
				fmt.Printf("Using specified region: %s\n", color.GreenString(region))
			} else {
				// Interactive mode - show all prompts
				fmt.Println("\nStarting interactive cloudspace creation...")

				// Interactive prompt for region selection with dropdown
				fmt.Println("\nFetching available regions...")
				
				// First try with default region from config if available
				defaultRegion := ""
				if cfg != nil && cfg.Region != "" {
					defaultRegion = cfg.Region
				}

				// Use the client's interactive prompt
				selectedRegion, err := client.PromptForRegionWithDefault(context.Background(), defaultRegion)
				if err != nil {
					return fmt.Errorf("failed to select region: %w", err)
				}
				region = selectedRegion
				fmt.Printf("\nSelected region: %s\n", color.GreenString(region))

				// Interactive prompt for cloudspace name
				namePrompt := &survey.Input{
					Message: "Enter a name for your cloudspace:",
				}
				if err := survey.AskOne(namePrompt, &name); err != nil {
					return fmt.Errorf("failed to get cloudspace name: %w", err)
				}
				if name == "" {
					return fmt.Errorf("name is required")
				}

				// Set default values for other required fields
				if kubernetesVersion == "" {
					kubernetesVersion = "1.31.1"
				}
				if cni == "" {
					cni = "calico"
				}
			}
		}

		// Additional check for config and name conflict
		if configPath != "" && name != "" {
			return fmt.Errorf("cannot specify both --config and --name")
		}

		var cloudspace *rxtspot.CloudSpace
		var spotnodepool *rxtspot.SpotNodePool

		var spotnodepools []rxtspot.SpotNodePool
		var ondemandnodepools []rxtspot.OnDemandNodePool

		if configPath != "" {
			// Read config file (YAML or JSON)
			b, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			// Try YAML -> JSON conversion
			if strings.HasSuffix(configPath, ".yaml") || strings.HasSuffix(configPath, ".yml") {
				jsonBytes, err := yaml.YAMLToJSON(b)
				if err != nil {
					return fmt.Errorf("failed to convert YAML to JSON: %w", err)
				}
				b = jsonBytes
			}

			// 🔹 Expect file to contain cloudspace + multiple pools
			var fullConfig struct {
				CloudSpace        rxtspot.CloudSpace         `json:"cloudspace"`
				SpotNodePools     []rxtspot.SpotNodePool     `json:"spotnodepools"`
				OnDemandNodePools []rxtspot.OnDemandNodePool `json:"ondemandnodepools"`
			}
			if err := json.Unmarshal(b, &fullConfig); err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}

			cloudspace = &fullConfig.CloudSpace
			spotnodepools = fullConfig.SpotNodePools
			ondemandnodepools = fullConfig.OnDemandNodePools

			// 🔹 Fill missing org/cloudspace for each spot node pool
			for i := range spotnodepools {
				if spotnodepools[i].Org == "" {
					spotnodepools[i].Org = cloudspace.Org
				}
				if spotnodepools[i].Cloudspace == "" {
					spotnodepools[i].Cloudspace = cloudspace.Name
				}
			}

			// 🔹 Fill missing org/cloudspace for each on-demand node pool
			for i := range ondemandnodepools {
				if ondemandnodepools[i].Org == "" {
					ondemandnodepools[i].Org = cloudspace.Org
				}
				if ondemandnodepools[i].Cloudspace == "" {
					ondemandnodepools[i].Cloudspace = cloudspace.Name
				}
			}
		} else {
			cloudspace = &rxtspot.CloudSpace{
				Name:              name,
				Org:               org,
				Region:            region,
				KubernetesVersion: kubernetesVersion,
				Cni:               cni,
			}

			for _, poolStr := range spotNodePoolJSON {
				// Check if it's a JSON string (backward compatibility)
				if strings.TrimSpace(poolStr)[0] == '{' {
					// First, unmarshal into a map to handle type conversion
					var rawPool map[string]interface{}
					if err := json.Unmarshal([]byte(poolStr), &rawPool); err != nil {
						return fmt.Errorf("failed to parse spotnodepool JSON: %w", err)
					}

					// If server class is not provided, prompt for it
					serverClass, ok := rawPool["serverclass"].(string)
					if !ok || serverClass == "" {
						sc, err := client.PromptForServerClass(context.Background(), region)
						if err != nil {
							return fmt.Errorf("failed to select server class: %w", err)
						}
						serverClass = sc
						fmt.Printf("Selected server class: %s\n", color.GreenString(serverClass))
					}

					// Create a new SpotNodePool and populate it
					pool := rxtspot.SpotNodePool{
						Org:         org,
						Cloudspace:  cloudspace.Name,
						ServerClass: serverClass,
					}

					// Handle each field with proper type conversion
					for key, value := range rawPool {
						switch strings.ToLower(key) {
						case "name":
							if strVal, ok := value.(string); ok {
								pool.Name = strVal
							}
						case "serverclass":
							if strVal, ok := value.(string); ok {
								pool.ServerClass = strVal
							}
						case "desired":
							switch v := value.(type) {
							case float64:
								pool.Desired = int(v)
							case string:
								if intVal, err := strconv.Atoi(v); err == nil {
									pool.Desired = intVal
								}
							}
						case "bidprice":
							switch v := value.(type) {
							case string:
								if formatted, err := validateBidPrice(v); err == nil {
									pool.BidPrice = formatted
								}
							case float64:
								pool.BidPrice = fmt.Sprintf("%.3f", v)
								// Ensure it has exactly 3 decimal places
								if !strings.Contains(pool.BidPrice, ".") {
									pool.BidPrice += ".000"
								} else {
									parts := strings.Split(pool.BidPrice, ".")
									if len(parts) == 2 && len(parts[1]) < 3 {
										pool.BidPrice += strings.Repeat("0", 3-len(parts[1]))
									}
								}
							}
						}
					}

					// Set defaults if not provided
					if pool.ServerClass == "" {
						pool.ServerClass = defaultServerclass
					}
					if pool.Desired == 0 {
						pool.Desired = 1
					}

					spotnodepools = append(spotnodepools, pool)
					continue
				}

				// Parse key=value parameters
				pool := rxtspot.SpotNodePool{
					Org:        org,
					Cloudspace: cloudspace.Name,
				}

				// Parse key=value parameters
				params, err := parseNodepoolParams(poolStr)
				if err != nil {
					return fmt.Errorf("failed to parse spotnodepool parameters: %w", err)
				}

				// Apply parameters to the pool
				for key, value := range params {
					switch strings.ToLower(key) {
					case "name":
						pool.Name = value
					case "serverclass":
						pool.ServerClass = value
					case "desired":
						desired, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("invalid desired value '%s': %w", value, err)
						}
						pool.Desired = desired
					case "bidprice":
						formattedBidPrice, err := validateBidPrice(value)
						if err != nil {
							return fmt.Errorf("invalid bid price '%s': %w", value, err)
						}
						pool.BidPrice = formattedBidPrice
					default:
						return fmt.Errorf("unknown nodepool parameter: %s", key)
					}
				}

				// Set defaults if not provided
				if pool.ServerClass == "" {
					pool.ServerClass = defaultServerclass
				}
				if pool.Desired == 0 {
					pool.Desired = 1
				}
				if pool.Name == "" {
					pool.Name = fmt.Sprintf("spot-pool-%d", len(spotnodepools)+1)
				}

				spotnodepools = append(spotnodepools, pool)
			}

			for _, poolStr := range onDemandNodePoolJSON {
				// Check if it's a JSON string (backward compatibility)
				if strings.TrimSpace(poolStr)[0] == '{' {
					// First, unmarshal into a map to handle type conversion
					var rawPool map[string]interface{}
					if err := json.Unmarshal([]byte(poolStr), &rawPool); err != nil {
						return fmt.Errorf("failed to parse ondemandnodepool JSON: %w", err)
					}

					// Create a new OnDemandNodePool and populate it
					pool := rxtspot.OnDemandNodePool{
						Org:        org,
						Cloudspace: cloudspace.Name,
					}

					// Handle each field with proper type conversion
					for key, value := range rawPool {
						switch strings.ToLower(key) {
						case "name":
							if strVal, ok := value.(string); ok {
								pool.Name = strVal
							}
						case "serverclass":
							if strVal, ok := value.(string); ok {
								pool.ServerClass = strVal
							}
						case "desired":
							switch v := value.(type) {
							case float64:
								pool.Desired = int(v)
							case string:
								if intVal, err := strconv.Atoi(v); err == nil {
									pool.Desired = intVal
								}
							}
						}
					}

					// Set defaults if not provided
					if pool.ServerClass == "" {
						pool.ServerClass = defaultServerclass
					}
					if pool.Desired == 0 {
						pool.Desired = 1
					}
					if pool.Name == "" {
						pool.Name = fmt.Sprintf("ondemand-pool-%d", len(ondemandnodepools)+1)
					}

					ondemandnodepools = append(ondemandnodepools, pool)
					continue
				}

				// Parse key=value parameters
				pool := rxtspot.OnDemandNodePool{
					Org:        org,
					Cloudspace: cloudspace.Name,
				}

				// Parse key=value parameters
				params, err := parseNodepoolParams(poolStr)
				if err != nil {
					return fmt.Errorf("failed to parse ondemandnodepool parameters: %w", err)
				}

				// Apply parameters to the pool
				for key, value := range params {
					switch strings.ToLower(key) {
					case "name":
						pool.Name = value
					case "serverclass":
						pool.ServerClass = value
					case "desired":
						desired, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("invalid desired value '%s': %w", value, err)
						}
						pool.Desired = desired
					}
				}

				// Set defaults if not provided
				if pool.ServerClass == "" {
					pool.ServerClass = defaultServerclass
				}
				if pool.Desired == 0 {
					pool.Desired = 1
				}
				if pool.Name == "" {
					pool.Name = fmt.Sprintf("ondemand-pool-%d", len(ondemandnodepools)+1)
				}

				ondemandnodepools = append(ondemandnodepools, pool)
			}

			// Only show interactive node pool configuration in interactive mode
			if !hasAnyFlag && len(spotNodePoolJSON) == 0 && len(onDemandNodePoolJSON) == 0 {
				color.Yellow("⚠️  No node pool configurations were specified. Let's configure a spot node pool.")

				// Prompt for server class
				serverClass, err := client.PromptForServerClass(context.Background(), region)
				if err != nil {
					return fmt.Errorf("failed to select server class: %w", err)
				}

				// Get minimum bid price for the selected server class
				minBidPrice, err := client.GetAPI().GetMinimumBidPriceForServerClass(context.Background(), serverClass)
				if err != nil {
					minBidPrice = "0.05"
				}
				minBidPrice = strings.TrimPrefix(minBidPrice, "$")

				// Prompt for bid price with minimum value validation
				bidPricePrompt := &survey.Input{
					Message: fmt.Sprintf("Enter bid price (min: $%s per hour):", minBidPrice),
					Default: minBidPrice,
				}
				var bidPriceStr string
				if err := survey.AskOne(bidPricePrompt, &bidPriceStr); err != nil {
					return fmt.Errorf("failed to get bid price: %w", err)
				}

				// Validate bid price
				bidPrice, err := strconv.ParseFloat(bidPriceStr, 64)
				if err != nil {
					return fmt.Errorf("invalid bid price: %w", err)
				}

				minPrice, _ := strconv.ParseFloat(minBidPrice, 64)
				if bidPrice < minPrice {
					color.Yellow("⚠️  Bid price ($%.3f) is below the minimum recommended price ($%s). This may result in node termination.", bidPrice, minBidPrice)
					confirm := false
					confirmPrompt := &survey.Confirm{
						Message: "Do you want to proceed with this bid price?",
						Default: false,
					}
					if err := survey.AskOne(confirmPrompt, &confirm); err != nil || !confirm {
						return fmt.Errorf("aborted by user")
					}
				}

				// Prompt for number of nodes
				nodesPrompt := &survey.Input{
					Message: "Number of nodes:",
					Default: "1",
				}
				nodesStr := "1"
				if err := survey.AskOne(nodesPrompt, &nodesStr); err != nil {
					return fmt.Errorf("failed to get number of nodes: %w", err)
				}

				nodes, err := strconv.Atoi(nodesStr)
				if err != nil || nodes < 1 {
					nodes = 1
				}

				// Generate a UUID for the node pool name
				nodePoolName, err := uuid.NewRandom()
				if err != nil {
					return fmt.Errorf("failed to generate node pool name: %w", err)
				}

				// Create the spot node pool with a generated UUID name
				spotnodepool = &rxtspot.SpotNodePool{
					Name:        nodePoolName.String(),
					Org:         cloudspace.Org,
					Cloudspace:  cloudspace.Name,
					ServerClass: serverClass,
					Desired:     nodes,
					BidPrice:    fmt.Sprintf("%.3f", bidPrice),
				}

				// Show spot node pool summary
				color.Green("\n📋 Spot Node Pool Configuration:")
				color.Green("• Server Class  : %s", spotnodepool.ServerClass)
				color.Green("• Nodes         : %d", spotnodepool.Desired)
				color.Green("• Bid Price     : $%s/hour", spotnodepool.BidPrice)

				spotnodepools = append(spotnodepools, *spotnodepool)

				// Ask if user wants to create an on-demand node pool
				addOndemand := false
				ondemandPrompt := &survey.Confirm{
					Message: "Would you like to add an on-demand node pool?",
					Default: false,
				}
				if err := survey.AskOne(ondemandPrompt, &addOndemand); err != nil {
					return fmt.Errorf("failed to get on-demand node pool preference: %w", err)
				}

				if addOndemand {
					// Prompt for on-demand node pool configuration
					onDemandServerClass, err := client.PromptForServerClass(context.Background(), region)
					if err != nil {
						return fmt.Errorf("failed to get on-demand server class: %w", err)
					}

					nodesStr, err := client.PromptForNodeCount("on-demand")
					if err != nil {
						return fmt.Errorf("failed to get number of on-demand nodes: %w", err)
					}

					nodes, err := strconv.Atoi(nodesStr)
					if err != nil || nodes < 1 {
						nodes = 1
					}

					// Generate a UUID for the on-demand node pool name
					onDemandPoolName, err := uuid.NewRandom()
					if err != nil {
						return fmt.Errorf("failed to generate on-demand node pool name: %w", err)
					}

					onDemandPool := &rxtspot.OnDemandNodePool{
						Name:        onDemandPoolName.String(),
						Org:         cloudspace.Org,
						Cloudspace:  cloudspace.Name,
						ServerClass: onDemandServerClass,
						Desired:     nodes,
					}

					// Show on-demand pool summary
					color.Green("\n📋 On-Demand Node Pool Configuration:")
					color.Green("• Server Class  : %s", onDemandPool.ServerClass)
					color.Green("• Nodes         : %d", onDemandPool.Desired)

					ondemandnodepools = append(ondemandnodepools, *onDemandPool)
				}
			}

			// Show final configuration
			color.Green("\n📋 Final Cloudspace Configuration:")
			color.Green("• Name          : %s", cloudspace.Name)
			color.Green("• Region        : %s", region)
			color.Green("• Spot Pools    : %d", len(spotnodepools))
			color.Green("• On-Demand Pools: %d", len(ondemandnodepools))

			// Skip confirmation in non-interactive mode
			if !hasAnyFlag {
				confirm := false
				confirmPrompt := &survey.Confirm{
					Message: "Create cloudspace with this configuration?",
					Default: true,
				}
				if err := survey.AskOne(confirmPrompt, &confirm); err != nil || !confirm {
					return fmt.Errorf("aborted by user")
				}
			}
		}

		// Track created resources for cleanup in case of failure
		type cleanupFunc func() error
		var cleanupStack []cleanupFunc
		var createdCloudspace bool

		// Defer cleanup function that will run if we return with an error
		defer func() {
			if err == nil || !createdCloudspace {
				// No error or cloudspace wasn't created, nothing to clean up
				return
			}

			// Run cleanup in reverse order (LIFO)
			for i := len(cleanupStack) - 1; i >= 0; i-- {
				if cleanupErr := cleanupStack[i](); cleanupErr != nil {
					// Log but don't fail the cleanup process
					fmt.Fprintf(os.Stderr, "Warning: cleanup failed: %v\n", cleanupErr)
				}
			}
			// Finally, delete the cloudspace if we created it
			if createdCloudspace {
				if delErr := client.GetAPI().DeleteCloudspace(context.Background(), cloudspace.Name, cloudspace.Org); delErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to clean up cloudspace: %v\n", delErr)
				}
			}
		}()

		// 1. First, validate all resources
		for i, pool := range spotnodepools {
			if pool.ServerClass == "" {
				return fmt.Errorf("spot nodepool %d: server class is required", i)
			}
			if pool.Desired <= 0 {
				return fmt.Errorf("spot nodepool %d: desired must be greater than 0", i)
			}
			if pool.BidPrice == "" {
				return fmt.Errorf("spot nodepool %d: bid price is required", i)
			}
		}

		// Show creating message right before starting resource creation
		if hasAnyFlag {
			fmt.Printf("\nCreating cloudspace - %s with the provided configuration...\n", cloudspace.Name)
		}

		for i, pool := range ondemandnodepools {
			if pool.ServerClass == "" {
				return fmt.Errorf("on-demand nodepool %d: server class is required", i)
			}
			if pool.Desired <= 0 {
				return fmt.Errorf("on-demand nodepool %d: desired must be greater than 0", i)
			}
		}

		// 2. Create cloudspace
		err = client.GetAPI().CreateCloudspace(context.Background(), *cloudspace)
		if err != nil {
			if rxtspot.IsForbidden(err) {
				return fmt.Errorf("forbidden: %w", err)
			}
			if rxtspot.IsConflict(err) {
				return fmt.Errorf("a cloudspace with this name already exists")
			}
			return fmt.Errorf("%w", err)
		}
		createdCloudspace = true

		// 3. Create spot node pools
		for i, pool := range spotnodepools {
			err = client.GetAPI().CreateSpotNodePool(context.Background(), org, pool)
			if err != nil {
				return fmt.Errorf("failed to create spot nodepool %s: %w", pool.Name, err)
			}
			// Add cleanup function for this pool
			i := i // Capture loop variable
			cleanupStack = append(cleanupStack, func() error {
				return client.GetAPI().DeleteSpotNodePool(context.Background(), org, spotnodepools[i].Name)
			})
		}

		// 4. Create on-demand node pools
		for i, pool := range ondemandnodepools {
			err = client.GetAPI().CreateOnDemandNodePool(context.Background(), org, pool)
			if err != nil {
				return fmt.Errorf("failed to create on-demand nodepool %s: %w", pool.Name, err)
			}
			// Add cleanup function for this pool
			i := i // Capture loop variable
			cleanupStack = append(cleanupStack, func() error {
				return client.GetAPI().DeleteOnDemandNodePool(context.Background(), org, ondemandnodepools[i].Name)
			})
		}

		// If we got here, everything was created successfully
		// Clear the cleanup stack to prevent cleanup on success
		cleanupStack = nil

		// for range spotnodepools {
		// 	fmt.Printf("Spot node pool created successfully\n")
		// }
		// for range ondemandnodepools {
		// 	fmt.Printf("On-demand node pool created successfully\n")
		// }

		cloudspace, err = client.GetAPI().GetCloudspace(context.Background(), org, cloudspace.Name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		//fmt.Printf("Cloudspace '%s' created successfully\n", cloudspace.Name)

		return internal.OutputData(cloudspace, outputFormat)
	},
}

// cloudspacesGetCmd represents the cloudspaces get command
var cloudspacesGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get cloudspace details",
	Long:  `Get details for a specific cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("name is required")
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

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return err
		}

		cloudspace, err := client.GetAPI().GetCloudspace(context.Background(), org, name)
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

		return internal.OutputData(cloudspace, outputFormat)
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
	cloudspacesCreateCmd.Flags().StringP("kubernetes_version", "", "1.31.1", "Kubernetes version (default: 1.31.1)")

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
