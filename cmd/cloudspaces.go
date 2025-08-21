package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
			return fmt.Errorf("failed to create client: %w", err)
		}

		cloudspaces, err := client.GetAPI().ListCloudspaces(context.Background(), org)
		if err != nil {
			return fmt.Errorf("failed to list cloudspaces: %w", err)
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
			return fmt.Errorf("failed to create client: %w", err)
		}
		k8sConfig, err := client.GetAPI().GetCloudspaceConfig(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("failed to get kubernetes config: %w", err)
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

		if region == "" {
			if err == nil && cfg.Region != "" {
				region = cfg.Region
			}
		}
		if region == "" {
			return fmt.Errorf("region not specified (use --region or run 'spotcli configure')")
		}

		configPath, _ := cmd.Flags().GetString("config")

		if configPath != "" && name != "" {
			return fmt.Errorf("either --config must be provided OR --name must be set")
		}

		spotNodePoolJSON, _ := cmd.Flags().GetStringArray("spot-nodepool")
		onDemandNodePoolJSON, _ := cmd.Flags().GetStringArray("ondemand-nodepool")
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes_version")
		cni, _ := cmd.Flags().GetString("cni")

		var cloudspace *rxtspot.CloudSpace
		var spotnodepool *rxtspot.SpotNodePool

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

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
			// Use CLI flags
			if name == "" {
				return fmt.Errorf("name is required")
			}

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

					// Create a new SpotNodePool and populate it
					pool := rxtspot.SpotNodePool{
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
			if len(spotNodePoolJSON) == 0 && len(onDemandNodePoolJSON) == 0 && configPath == "" {
				price, err := client.GetAPI().GetMinimumBidPriceForServerClass(context.Background(), defaultServerclass)
				if err != nil {
					price = "0.05"
				}
				price = strings.TrimPrefix(price, "$")
				spotnodepool = &rxtspot.SpotNodePool{
					Org:         cloudspace.Org,
					Cloudspace:  cloudspace.Name,
					ServerClass: defaultServerclass, // default choice
					Desired:     1,
					BidPrice:    price, // match struct type
				}

				color.Yellow("⚠️  NOTE: No spotnodepool/ondemandpool configurations are specified.")
				color.Yellow("⚙️  Default Spot Node Pool will be created with:")
				color.Yellow("• Server Class  : %s\n", spotnodepool.ServerClass)
				color.Yellow("• Desired Nodes : %d\n", spotnodepool.Desired)
				color.Yellow("• Bid Price     : %s$ per hr\n", spotnodepool.BidPrice)
				fmt.Print("Proceed? (y/N): ")

				var response string
				fmt.Scanln(&response)
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Println("Aborting as per user choice.")
					return nil
				}
				spotnodepools = append(spotnodepools, *spotnodepool)
			}

		}

		err = client.GetAPI().CreateCloudspace(context.Background(), *cloudspace)
		if err != nil {
			if rxtspot.IsForbidden(err) {
				return fmt.Errorf("forbidden: %w", err)
			}
			if rxtspot.IsConflict(err) {
				return fmt.Errorf("conflict: %w", err)
			}
			return fmt.Errorf("failed to create cloudspace: %w", err)
		}

		for _, spotnodepool := range spotnodepools {
			err = client.GetAPI().CreateSpotNodePool(context.Background(), org, spotnodepool)
			if err != nil {
				if rxtspot.IsForbidden(err) {
					return fmt.Errorf("forbidden: %w", err)
				}
				if rxtspot.IsConflict(err) {
					return fmt.Errorf("conflict: %w", err)
				}
				return fmt.Errorf("failed to create spot node pool: %w", err)
			}
			fmt.Printf("Spot node pool created successfully\n")
		}
		for _, ondemandnodepool := range ondemandnodepools {
			err = client.GetAPI().CreateOnDemandNodePool(context.Background(), org, ondemandnodepool)
			if err != nil {
				if rxtspot.IsForbidden(err) {
					return fmt.Errorf("forbidden: %w", err)
				}
				if rxtspot.IsConflict(err) {
					return fmt.Errorf("conflict: %w", err)
				}
				return fmt.Errorf("failed to create on-demand node pool: %w", err)
			}
			fmt.Printf("On-demand node pool created successfully\n")
		}

		cloudspace, err = client.GetAPI().GetCloudspace(context.Background(), org, cloudspace.Name)
		if err != nil {
			return fmt.Errorf("failed to get cloudspace: %w", err)
		}
		fmt.Printf("Cloudspace '%s' created successfully\n", cloudspace.Name)

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
			return err
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
			return fmt.Errorf("failed to delete cloudspace: %w", err)
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
