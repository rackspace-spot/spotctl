package cmd

import (
	"context"
	"fmt"

	"github.com/rackerlabs/spot-cli/internal"
	rxtspot "github.com/rackerlabs/spot-sdk/rxtspot/api/v1"
	"github.com/spf13/cobra"
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
		region, _ := cmd.Flags().GetString("region")
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes-version")

		if name == "" || org == "" || region == "" {
			return fmt.Errorf("name, org, and region are required")
		}

		if kubernetesVersion == "" {
			kubernetesVersion = "1.28" // Default version
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		cloudspace := rxtspot.Cloudspace{
			Name:              name,
			Namespace:         org,
			Region:            region,
			KubernetesVersion: kubernetesVersion,
		}

		created, err := client.GetAPI().CreateCloudspace(context.Background(), cloudspace)
		if err != nil {
			return fmt.Errorf("failed to create cloudspace: %w", err)
		}

		fmt.Printf("Cloudspace '%s' created successfully\n", created.Name)
		return internal.OutputData(created, outputFormat)
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
			return fmt.Errorf("failed to create client: %w", err)
		}

		cloudspace, err := client.GetAPI().GetCloudspace(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("failed to get cloudspace: %w", err)
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
	cloudspacesCreateCmd.Flags().StringP("region", "r", "", "Region (required)")
	cloudspacesCreateCmd.Flags().String("kubernetes-version", "", "Kubernetes version (default: 1.28)")
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
