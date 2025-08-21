package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotcli/internal"
	config "github.com/rackspace-spot/spotcli/pkg"
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

		spotNodePoolJSON, _ := cmd.Flags().GetStringArray("spot_nodepool")
		onDemandNodePoolJSON, _ := cmd.Flags().GetStringArray("ondemand_nodepool")
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

			for _, poolJSON := range spotNodePoolJSON {
				var pool rxtspot.SpotNodePool

				// Unmarshal the JSON string into the struct
				if err := json.Unmarshal([]byte(poolJSON), &pool); err != nil {
					return fmt.Errorf("failed to parse spotnodepool JSON: %w", err)
				}

				// 🔹 Set defaults if missing
				if pool.Org == "" {
					pool.Org = org
				}
				if pool.Cloudspace == "" {
					pool.Cloudspace = cloudspace.Name
				}

				spotnodepools = append(spotnodepools, pool)
			}

			for _, poolJSON := range onDemandNodePoolJSON {
				var pool rxtspot.OnDemandNodePool
				if err := json.Unmarshal([]byte(poolJSON), &pool); err != nil {
					return fmt.Errorf("failed to parse ondemandnodepool JSON: %w", err)
				}
				// 🔹 Set defaults if missing
				if pool.Org == "" {
					pool.Org = org
				}
				if pool.Cloudspace == "" {
					pool.Cloudspace = cloudspace.Name
				}

				ondemandnodepools = append(ondemandnodepools, pool)
			}
			if len(spotNodePoolJSON) == 0 && len(onDemandNodePoolJSON) == 0 && configPath == "" {
				price, err := client.GetAPI().GetMinimumBidPriceForServerClass(context.Background(), "gp.vs1.medium-iad")
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

	cloudspacesCreateCmd.Flags().StringArray("spot_nodepool", []string{}, "Spot nodepool details as JSON string")
	cloudspacesCreateCmd.Flags().StringArray("ondemand_nodepool", []string{}, "Ondemand nodepool details as JSON string")
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
