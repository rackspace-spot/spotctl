package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
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

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pools, err := client.GetAPI().ListSpotNodePools(context.Background(), org, cloudspace)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(pools, outputFormat)
	},
}

// spotGetCmd represents the spot get command
var spotGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get spot node pool",
	Long:  `Get a spot node pool in a org.`,
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

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pool, err := client.GetAPI().GetSpotNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(pool, outputFormat)
	},
}

// spotDeleteCmd represents the spot delete command
var spotDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete spot node pools",
	Long:  `Delete spot node pools in a org.`,
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
			prompt("Are you sure you want to delete spot nodepool '%s'? (y/N): ", name)

			var response string
			_, err := fmt.Scanln(&response)
			if err != nil || (response != "y" && response != "Y") {
				fmt.Println("Aborted.")
				return nil
			}
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = client.GetAPI().DeleteSpotNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		fmt.Printf("spot node pool - %s deleted successfully \n", name)

		return nil
	},
}

// spotCreateCmd represents the spot create command
var spotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a spot node pool",
	Long:  `Create a new spot node pool in a cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
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
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		serverClass, _ := cmd.Flags().GetString("serverclass")
		desiredStr, _ := cmd.Flags().GetString("desired")
		bidPrice, _ := cmd.Flags().GetString("bidprice")

		if name == "" || cloudspace == "" || serverClass == "" || desiredStr == "" || bidPrice == "" {
			return fmt.Errorf("name, cloudspace, serverclass, desired, and bidprice are required")
		}

		desired, err := strconv.Atoi(desiredStr)
		if err != nil {
			return fmt.Errorf("desired must be a valid integer: %w", err)
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pool := &rxtspot.SpotNodePool{
			Name:        name,
			Org:         org,
			Cloudspace:  cloudspace,
			ServerClass: serverClass,
			Desired:     desired,
			BidPrice:    bidPrice,
		}

		err = client.GetAPI().CreateSpotNodePool(context.Background(), org, *pool)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		pool, err = client.GetAPI().GetSpotNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("spot nodepool - %s created successfully \n", pool.Name)

		return internal.OutputData(pool, outputFormat)
	},
}

// spotUpdateCmd represents the spot update command
var spotUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a spot node pool",
	Long:  `Update a spot node pool in a cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
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
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		desiredStr, _ := cmd.Flags().GetString("desired")
		bidPrice, _ := cmd.Flags().GetString("bidprice")

		if name == "" || cloudspace == "" || desiredStr == "" || bidPrice == "" {
			return fmt.Errorf("name, cloudspace, desired, and bidprice are required")
		}

		desired, err := strconv.Atoi(desiredStr)
		if err != nil {
			return fmt.Errorf("desired must be a valid integer: %w", err)
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pool := &rxtspot.SpotNodePool{
			Name:       name,
			Org:        org,
			Cloudspace: cloudspace,
			Desired:    desired,
			BidPrice:   bidPrice,
		}

		err = client.GetAPI().UpdateSpotNodePool(context.Background(), org, *pool)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("spot nodepool - %s updated successfully \n", pool.Name)

		return internal.OutputData(pool, outputFormat)
	},
}

// ondemandListCmd represents the ondemand list command
var ondemandListCmd = &cobra.Command{
	Use:   "list",
	Short: "List on-demand node pools",
	Long:  `List all on-demand node pools in a org.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		cloudspace, _ := cmd.Flags().GetString("cloudspace")

		if org == "" || cloudspace == "" {
			return fmt.Errorf("org and cloudspace are required")
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pools, err := client.GetAPI().ListOnDemandNodePools(context.Background(), org, cloudspace)
		if err != nil {
			return fmt.Errorf("%w", err)
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
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		serverClass, _ := cmd.Flags().GetString("serverclass")
		desiredStr, _ := cmd.Flags().GetString("desired")

		if name == "" || cloudspace == "" || serverClass == "" || desiredStr == "" {
			return fmt.Errorf("name, org, cloudspace, serverclass, and desired are required")
		}

		desired, err := strconv.Atoi(desiredStr)
		if err != nil {
			return fmt.Errorf("desired must be a valid integer: %w", err)
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pool := &rxtspot.OnDemandNodePool{
			Name:        name,
			Org:         org,
			Cloudspace:  cloudspace,
			ServerClass: serverClass,
			Desired:     desired,
		}

		err = client.GetAPI().CreateOnDemandNodePool(context.Background(), org, *pool)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pool, err = client.GetAPI().GetOnDemandNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("on-demand nodepool - %s created successfully \n", pool.Name)

		return internal.OutputData(pool, outputFormat)
	},
}

var ondemandGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get on-demand node pool",
	Long:  `Get a on-demand node pool in a org.`,
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

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pool, err := client.GetAPI().GetOnDemandNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(pool, outputFormat)
	},
}

// ondemandUpdateCmd represents the ondemand update command
var ondemandUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a on-demand node pool",
	Long:  `Update a on-demand node pool in a cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
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
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		desiredStr, _ := cmd.Flags().GetString("desired")

		if name == "" || cloudspace == "" || desiredStr == "" {
			return fmt.Errorf("name, cloudspace, and desired are required")
		}

		desired, err := strconv.Atoi(desiredStr)
		if err != nil {
			return fmt.Errorf("desired must be a valid integer: %w", err)
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pool := &rxtspot.OnDemandNodePool{
			Name:       name,
			Org:        org,
			Cloudspace: cloudspace,
			Desired:    desired,
		}

		err = client.GetAPI().UpdateOnDemandNodePool(context.Background(), org, *pool)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("on-demand nodepool - %s updated successfully \n", pool.Name)

		return internal.OutputData(pool, outputFormat)
	},
}

// ondemandDeleteCmd represents the ondemand delete command
var ondemandDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete ondemand node pools",
	Long:  `Delete ondemand node pools in a org.`,
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
			prompt("Are you sure you want to delete ondemand nodepool '%s'? (y/N): ", name)

			var response string
			_, err := fmt.Scanln(&response)
			if err != nil || (response != "y" && response != "Y") {
				fmt.Println("Aborted.")
				return nil
			}
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = client.GetAPI().DeleteOnDemandNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		fmt.Printf("ondemand node pool - %s deleted successfully \n", name)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodepoolsCmd)
	nodepoolsCmd.AddCommand(spotCmd)
	nodepoolsCmd.AddCommand(ondemandCmd)

	// Add spot subcommands
	spotCmd.AddCommand(spotListCmd)
	spotCmd.AddCommand(spotDeleteCmd)
	spotCmd.AddCommand(spotGetCmd)
	spotCmd.AddCommand(spotUpdateCmd)
	spotCmd.AddCommand(spotCreateCmd)

	// Add ondemand subcommands
	ondemandCmd.AddCommand(ondemandListCmd)
	ondemandCmd.AddCommand(ondemandCreateCmd)
	ondemandCmd.AddCommand(ondemandGetCmd)
	ondemandCmd.AddCommand(ondemandUpdateCmd)
	ondemandCmd.AddCommand(ondemandDeleteCmd)

	spotGetCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	spotGetCmd.MarkFlagRequired("name")

	// Flags for spot list
	spotListCmd.Flags().String("org", "", "Organization (required)")
	//spotListCmd.MarkFlagRequired("org")
	spotListCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	spotListCmd.MarkFlagRequired("cloudspace")

	// Flags for spot create
	spotCreateCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	spotCreateCmd.Flags().String("org", "", "Organization ID")
	spotCreateCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	spotCreateCmd.Flags().String("serverclass", "", "Server class (required)")
	spotCreateCmd.Flags().String("desired", "", "Desired number of nodes (required)")
	spotCreateCmd.Flags().String("bidprice", "", "Maximum bid price (required)")
	spotCreateCmd.MarkFlagRequired("name")
	spotCreateCmd.MarkFlagRequired("cloudspace")
	spotCreateCmd.MarkFlagRequired("serverclass")
	spotCreateCmd.MarkFlagRequired("desired")
	spotCreateCmd.MarkFlagRequired("bidprice")

	spotUpdateCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	spotUpdateCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	spotUpdateCmd.Flags().String("desired", "", "Desired number of nodes (optional)")
	spotUpdateCmd.Flags().String("bidprice", "", "Maximum bid price (optional)")
	spotUpdateCmd.Flags().String("org", "", "Organization ID")
	spotUpdateCmd.MarkFlagRequired("name")
	spotUpdateCmd.MarkFlagRequired("cloudspace")

	spotDeleteCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	spotDeleteCmd.MarkFlagRequired("name")
	spotDeleteCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts; assume \"yes\" as answer")

	// Flags for ondemand list
	ondemandListCmd.Flags().String("org", "", "Organization ID")
	ondemandListCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	ondemandListCmd.MarkFlagRequired("cloudspace")

	ondemandGetCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	ondemandGetCmd.MarkFlagRequired("name")

	// Flags for ondemand create
	ondemandCreateCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	ondemandCreateCmd.Flags().String("org", "", "Organization ID")
	ondemandCreateCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	ondemandCreateCmd.Flags().String("serverclass", "", "Server class (required)")
	ondemandCreateCmd.Flags().String("desired", "", "Desired number of nodes (required)")
	ondemandCreateCmd.MarkFlagRequired("name")
	ondemandCreateCmd.MarkFlagRequired("cloudspace")
	ondemandCreateCmd.MarkFlagRequired("serverclass")
	ondemandCreateCmd.MarkFlagRequired("desired")

	ondemandUpdateCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	ondemandUpdateCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	ondemandUpdateCmd.Flags().String("desired", "", "Desired number of nodes (optional)")
	ondemandUpdateCmd.Flags().String("org", "", "Organization ID")
	ondemandUpdateCmd.MarkFlagRequired("name")
	ondemandUpdateCmd.MarkFlagRequired("cloudspace")

	ondemandDeleteCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	ondemandDeleteCmd.MarkFlagRequired("name")
	ondemandDeleteCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts; assume \"yes\" as answer")

}
