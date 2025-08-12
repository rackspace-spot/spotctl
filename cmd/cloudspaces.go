package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotcli/internal"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// cloudspacesCmd represents the cloudspaces command
var cloudspacesCmd = &cobra.Command{
	Use:   "cloudspaces",
	Short: "Manage cloudspaces",
	Long:  `Manage Rackspace Spot cloudspaces (Kubernetes clusters).`,
}

// cloudspacesListCmd represents the cloudspaces list command
var cloudspacesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloudspaces",
	Long:  `List all cloudspaces in an organization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, _ := cmd.Flags().GetString("org")
		if org == "" {
			return fmt.Errorf("org is required")
		}

		client, err := internal.NewClient()
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

// cloudspacesCreateCmd represents the cloudspaces create command
var cloudspacesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cloudspace",
	Long:  `Create a new cloudspace (Kubernetes cluster).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		org, _ := cmd.Flags().GetString("org")
		spotNodePoolJSON, _ := cmd.Flags().GetStringArray("spot_nodepool")
		onDemandNodePoolJSON, _ := cmd.Flags().GetStringArray("ondemand_nodepool")

		region, _ := cmd.Flags().GetString("region")
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes_version")
		cni, _ := cmd.Flags().GetString("cni")
		configPath, _ := cmd.Flags().GetString("config")

		// serverClass, _ := cmd.Flags().GetString("server_class")
		// bidPrice, _ := cmd.Flags().GetString("bid_price")
		// desiredNodes, _ := cmd.Flags().GetInt("desired_nodes")

		var cloudspace *rxtspot.CloudSpace
		var spotnodepool *rxtspot.SpotNodePool
		//var ondemandnodepool *rxtspot.OnDemandNodePool

		if name == "" || org == "" || region == "" {
			return fmt.Errorf("name, org, and region are required")
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		var spotnodepools []rxtspot.SpotNodePool
		var ondemandnodepools []rxtspot.OnDemandNodePool
		// cloudspace := rxtspot.CloudSpace{
		// 	Name:              name,
		// 	OrgID:             org,
		// 	Region:            region,
		// 	KubernetesVersion: kubernetesVersion,
		// 	Cni:               cni,
		// }

		// spotnodepool := rxtspot.SpotNodePool{
		// 	Org:         org,
		// 	Cloudspace:  cloudspace.Name,
		// 	ServerClass: serverClass,
		// 	BidPrice:    bidPrice,
		// 	Desired:     desiredNodes,
		// }
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
					spotnodepools[i].Org = cloudspace.OrgID
				}
				if spotnodepools[i].Cloudspace == "" {
					spotnodepools[i].Cloudspace = cloudspace.Name
				}
			}

			// 🔹 Fill missing org/cloudspace for each on-demand node pool
			for i := range ondemandnodepools {
				if ondemandnodepools[i].Org == "" {
					ondemandnodepools[i].Org = cloudspace.OrgID
				}
				if ondemandnodepools[i].Cloudspace == "" {
					ondemandnodepools[i].Cloudspace = cloudspace.Name
				}
			}
		} else {
			// Use CLI flags
			if name == "" || org == "" || region == "" {
				return fmt.Errorf("name, org, and region are required (or provide --config)")
			}

			cloudspace = &rxtspot.CloudSpace{
				Name:              name,
				OrgID:             org,
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
			if spotNodePoolJSON == nil && onDemandNodePoolJSON == nil && configPath == "" {
				spotnodepool = &rxtspot.SpotNodePool{
					Org:         cloudspace.OrgID,
					Cloudspace:  cloudspace.Name,
					ServerClass: "gp.vs1.medium-iad", // default choice
					Desired:     1,
					BidPrice:    "0.08", // match struct type
				}

				fmt.Println("No spotnodepool/ondemandpool configurations are specified.")
				fmt.Println("Default Spot Node Pool will be created with:")
				fmt.Printf("Server Class  : %s\n", spotnodepool.ServerClass)
				fmt.Printf("Desired Nodes : %d\n", spotnodepool.Desired)
				fmt.Printf("Bid Price     : %s$ per hr\n", spotnodepool.BidPrice)
				fmt.Print("Proceed? (y/N): ")

				var response string
				fmt.Scanln(&response)
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Println("Aborting as per user choice.")
					return nil
				}
			}

		}

		if cloudspace != nil {
			err = client.GetAPI().CreateCloudspace(context.Background(), *cloudspace)
			if err != nil {
				return fmt.Errorf("failed to create cloudspace: %w", err)
			}
		}

		for _, spotnodepool := range spotnodepools {
			err = client.GetAPI().CreateSpotNodePool(context.Background(), spotnodepool)
			if err != nil {
				return fmt.Errorf("failed to create spot node pool: %w", err)
			}
		}
		for _, ondemandnodepool := range ondemandnodepools {
			err = client.GetAPI().CreateOnDemandNodePool(context.Background(), ondemandnodepool)
			if err != nil {
				return fmt.Errorf("failed to create on-demand node pool: %w", err)
			}
		}

		fmt.Printf("Spot node pool created successfully\n")
		fmt.Printf("On-demand node pool created successfully\n")
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
		name, _ := cmd.Flags().GetString("name")
		org, _ := cmd.Flags().GetString("org")

		if name == "" || org == "" {
			return fmt.Errorf("name and org are required")
		}

		client, err := internal.NewClient()
		if err != nil {
			return err
		}

		cloudspace, err := client.GetAPI().GetCloudspace(context.Background(), org, name)
		if err != nil {
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
		org, _ := cmd.Flags().GetString("org")

		if name == "" || org == "" {
			return fmt.Errorf("name and org are required")
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		err = client.GetAPI().DeleteCloudspace(context.Background(), org, name)
		if err != nil {
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

	// Add flags for cloudspaces list
	cloudspacesListCmd.Flags().String("org", "", "Organization (required)") // Removed shorthand for org flag
	cloudspacesListCmd.MarkFlagRequired("org")
	cloudspacesListCmd.Flags().StringP("output", "o", "json", "Output format (json, table, yaml)")

	// Add flags for cloudspaces create
	cloudspacesCreateCmd.Flags().String("name", "", "Cloudspace name (required)")
	cloudspacesCreateCmd.Flags().String("org", "", "Organization (required)") // Removed shorthand for org flag
	cloudspacesCreateCmd.Flags().String("region", "", "Region (required)")
	cloudspacesCreateCmd.Flags().StringP("kubernetes_version", "", "1.31.1", "Kubernetes version (default: 1.31.1)")
	//cloudspacesCreateCmd.Flags().StringP("server_class", "", "gp.vs1.medium-iad", "Server Class (default: gp.vs1.medium-iad)")
	//cloudspacesCreateCmd.Flags().StringP("bid_price", "", "0.01", "Bid Price (default: 0.01$)")
	//cloudspacesCreateCmd.Flags().IntP("desired_nodes", "", 1, "Desired number of nodes (default: 1)")
	cloudspacesCreateCmd.Flags().StringArray("spot_nodepool", []string{}, "Spot nodepool details as JSON string")
	cloudspacesCreateCmd.Flags().StringArray("ondemand_nodepool", []string{}, "Ondemand nodepool details as JSON string")
	cloudspacesCreateCmd.Flags().String("config", "", "Path to config file (YAML or JSON)")
	cloudspacesCreateCmd.Flags().StringP("cni", "", "calico", "CNI (default: calico)")
	cloudspacesCreateCmd.MarkFlagRequired("name")
	cloudspacesCreateCmd.MarkFlagRequired("org")
	cloudspacesCreateCmd.MarkFlagRequired("region")

	// Add flags for cloudspaces get
	cloudspacesGetCmd.Flags().String("name", "", "Cloudspace name (required)")
	cloudspacesGetCmd.Flags().String("org", "", "Organization (required)") // Removed shorthand for org flag
	cloudspacesGetCmd.MarkFlagRequired("name")
	cloudspacesGetCmd.MarkFlagRequired("org")

	// Add flags for cloudspaces delete
	cloudspacesDeleteCmd.Flags().String("name", "", "Cloudspace name (required)")
	cloudspacesDeleteCmd.Flags().String("org", "", "Organization (required)") // Removed shorthand for org flag
	cloudspacesDeleteCmd.MarkFlagRequired("name")
	cloudspacesDeleteCmd.MarkFlagRequired("org")
}
