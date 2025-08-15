package cmd

import (
	"context"
	"fmt"
	"strconv"

	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotcli/internal"
	config "github.com/rackspace-spot/spotcli/pkg"
	"github.com/spf13/cobra"
)

// nodepoolsCmd represents the nodepools command
var nodepoolsCmd = &cobra.Command{
	Use:   "nodepools",
	Short: "Manage node pools",
	Long:  `Manage Rackspace Spot node pools (both spot and on-demand).`,
}

// spotCmd represents the spot nodepools command
var spotCmd = &cobra.Command{
	Use:   "spot",
	Short: "Manage spot node pools",
	Long:  `Manage spot node pools in cloudspaces.`,
}

// ondemandCmd represents the ondemand nodepools command
var ondemandCmd = &cobra.Command{
	Use:   "ondemand",
	Short: "Manage on-demand node pools",
	Long:  `Manage on-demand node pools in cloudspaces.`,
}

// spotListCmd represents the spot list command
var spotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List spot node pools",
	Long:  `List all spot node pools in a org.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		if cloudspace == "" {
			return fmt.Errorf("cloudspace is required")
		}
		org, err := config.GetOrg(cmd)
		if err != nil {
			return err
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		pools, err := client.GetAPI().ListSpotNodePools(context.Background(), org, cloudspace)
		if err != nil {
			return fmt.Errorf("failed to list spot node pools: %w", err)
		}

		return internal.OutputData(pools, outputFormat)
	},
}

// spotCreateCmd represents the spot create command
var spotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a spot node pool",
	Long:  `Create a new spot node pool in a cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		org, err := config.GetOrg(cmd)
		if err != nil {
			return err
		}
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		serverClass, _ := cmd.Flags().GetString("server-class")
		desiredStr, _ := cmd.Flags().GetString("desired")
		bidPrice, _ := cmd.Flags().GetString("bid-price")

		if name == "" || cloudspace == "" || serverClass == "" || desiredStr == "" || bidPrice == "" {
			return fmt.Errorf("name, cloudspace, server-class, desired, and bid-price are required")
		}

		desired, err := strconv.Atoi(desiredStr)
		if err != nil {
			return fmt.Errorf("desired must be a valid integer: %w", err)
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		pool := rxtspot.SpotNodePool{
			Name:        name,
			Org:         org,
			Cloudspace:  cloudspace,
			ServerClass: serverClass,
			Desired:     desired,
			BidPrice:    bidPrice,
		}

		err = client.GetAPI().CreateSpotNodePool(context.Background(), pool)
		if err != nil {
			return fmt.Errorf("failed to create spot node pool: %w", err)
		}
		return internal.OutputData(pool, outputFormat)
	},
}

// ondemandListCmd represents the ondemand list command
var ondemandListCmd = &cobra.Command{
	Use:   "list",
	Short: "List on-demand node pools",
	Long:  `List all on-demand node pools in a org.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		org, err := config.GetOrg(cmd)
		if err != nil {
			return err
		}
		cloudspace, _ := cmd.Flags().GetString("cloudspace")

		if org == "" || cloudspace == "" {
			return fmt.Errorf("org and cloudspace are required")
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		pools, err := client.GetAPI().ListOnDemandNodePools(context.Background(), org, cloudspace)
		if err != nil {
			return fmt.Errorf("failed to list on-demand node pools: %w", err)
		}

		return internal.OutputData(pools, outputFormat)
	},
}

// ondemandCreateCmd represents the ondemand create command
var ondemandCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an on-demand node pool",
	Long:  `Create a new on-demand node pool in a cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		org, err := config.GetOrg(cmd)
		if err != nil {
			return err
		}
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		serverClass, _ := cmd.Flags().GetString("server-class")
		desiredStr, _ := cmd.Flags().GetString("desired")

		if name == "" || cloudspace == "" || serverClass == "" || desiredStr == "" {
			return fmt.Errorf("name, org, cloudspace, server-class, and desired are required")
		}

		desired, err := strconv.Atoi(desiredStr)
		if err != nil {
			return fmt.Errorf("desired must be a valid integer: %w", err)
		}

		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		pool := rxtspot.OnDemandNodePool{
			Name:        name,
			Org:         org,
			Cloudspace:  cloudspace,
			ServerClass: serverClass,
			Desired:     desired,
		}

		err = client.GetAPI().CreateOnDemandNodePool(context.Background(), pool)
		if err != nil {
			return fmt.Errorf("failed to create on-demand node pool: %w", err)
		}

		return internal.OutputData(pool, outputFormat)
	},
}

func init() {
	rootCmd.AddCommand(nodepoolsCmd)
	nodepoolsCmd.AddCommand(spotCmd)
	nodepoolsCmd.AddCommand(ondemandCmd)

	// Add spot subcommands
	spotCmd.AddCommand(spotListCmd)
	spotCmd.AddCommand(spotCreateCmd)

	// Add ondemand subcommands
	ondemandCmd.AddCommand(ondemandListCmd)
	ondemandCmd.AddCommand(ondemandCreateCmd)

	// Flags for spot list
	spotListCmd.Flags().StringP("org", "o", "", "Organization (required)")
	spotListCmd.MarkFlagRequired("org")

	// Flags for spot create
	spotCreateCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid UUID) (required)")
	spotCreateCmd.Flags().StringP("org", "o", "", "Organization ID")
	spotCreateCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	spotCreateCmd.Flags().String("server-class", "", "Server class (required)")
	spotCreateCmd.Flags().String("desired", "", "Desired number of nodes (required)")
	spotCreateCmd.Flags().String("bid-price", "", "Maximum bid price (required)")
	spotCreateCmd.MarkFlagRequired("name")
	spotCreateCmd.MarkFlagRequired("cloudspace")
	spotCreateCmd.MarkFlagRequired("server-class")
	spotCreateCmd.MarkFlagRequired("desired")
	spotCreateCmd.MarkFlagRequired("bid-price")

	// Flags for ondemand list
	ondemandListCmd.Flags().String("org", "", "Organization ID")
	ondemandListCmd.MarkFlagRequired("org")

	// Flags for ondemand create
	ondemandCreateCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid UUID) (required)")
	ondemandCreateCmd.Flags().String("org", "", "Organization ID")
	ondemandCreateCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	ondemandCreateCmd.Flags().String("server-class", "", "Server class (required)")
	ondemandCreateCmd.Flags().String("desired", "", "Desired number of nodes (required)")
	ondemandCreateCmd.MarkFlagRequired("name")
	ondemandCreateCmd.MarkFlagRequired("cloudspace")
	ondemandCreateCmd.MarkFlagRequired("server-class")
	ondemandCreateCmd.MarkFlagRequired("desired")
}
