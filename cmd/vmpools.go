package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/google/uuid"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
)

// vmPoolsCmd represents the vmpools command
var vmPoolsCmd = &cobra.Command{
	Use:     "vmpools",
	Short:   "Manage VM pools",
	Long:    `Manage Rackspace Spot VM pools (both spot and on-demand).`,
	Aliases: []string{"vmp", "vmpool"},
}

func init() {
	rootCmd.AddCommand(vmPoolsCmd)
	vmPoolsCmd.AddCommand(vmPoolListCmd)
	vmPoolsCmd.AddCommand(vmPoolCreateCmd)
	vmPoolsCmd.AddCommand(vmPoolGetCmd)
	vmPoolsCmd.AddCommand(vmPoolUpdateCmd)
	vmPoolsCmd.AddCommand(vmPoolDeleteCmd)

	// Flags for vmpools list
	vmPoolListCmd.Flags().String("org", "", "Organization name")
	vmPoolListCmd.Flags().String("vmcloudspace", "", "VM cloudspace name (required)")
	vmPoolListCmd.MarkFlagRequired("vmcloudspace")

	// Flags for vmpools create
	vmPoolCreateCmd.Flags().String("org", "", "Organization name")
	vmPoolCreateCmd.Flags().String("vmcloudspace", "", "VM cloudspace name (required)")
	vmPoolCreateCmd.Flags().String("serverclass", "", "Server class (required)")
	vmPoolCreateCmd.Flags().String("bidprice", "", "Bid price (required)")
	vmPoolCreateCmd.Flags().Int("desired", 1, "Desired number of VMs")
	vmPoolCreateCmd.Flags().String("pooltype", "spot", "Pool type (spot or ondemand)")
	vmPoolCreateCmd.Flags().String("vmimage", "ubuntu24.04", "VM image")
	vmPoolCreateCmd.Flags().String("vm-userdata", "", "Cloud-init user data (raw text or base64-encoded, optional)")
	vmPoolCreateCmd.Flags().String("vm-userdata-from-script", "", "Path to cloud-init script file (optional)")
	vmPoolCreateCmd.MarkFlagRequired("vmcloudspace")
	vmPoolCreateCmd.MarkFlagRequired("serverclass")
	vmPoolCreateCmd.MarkFlagRequired("bidprice")

	// Flags for vmpools get
	vmPoolGetCmd.Flags().String("org", "", "Organization name")
	vmPoolGetCmd.Flags().String("name", "", "VM pool name (required)")
	vmPoolGetCmd.MarkFlagRequired("name")

	// Flags for vmpools update
	vmPoolUpdateCmd.Flags().String("org", "", "Organization name")
	vmPoolUpdateCmd.Flags().String("name", "", "VM pool name (required)")
	vmPoolUpdateCmd.Flags().Int("desired", -1, "Desired number of VMs")
	vmPoolUpdateCmd.Flags().String("bidprice", "", "Bid price")
	vmPoolUpdateCmd.MarkFlagRequired("name")

	// Flags for vmpools delete
	vmPoolDeleteCmd.Flags().String("org", "", "Organization name")
	vmPoolDeleteCmd.Flags().String("name", "", "VM pool name (required)")
	vmPoolDeleteCmd.MarkFlagRequired("name")
	vmPoolDeleteCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts")
}

var vmPoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List VM pools",
	Long:  `List all VM pools for a VM cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return fmt.Errorf("failed to get CLI configuration: %w", err)
		}
		org, _ := cmd.Flags().GetString("org")
		if org == "" && cfg != nil && cfg.Org != "" {
			org = cfg.Org
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotctl configure')")
		}

		vmCloudSpace, _ := cmd.Flags().GetString("vmcloudspace")

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		pools, err := client.GetAPI().ListVMPools(context.Background(), org, vmCloudSpace)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(pools, outputFormat)
	},
}

var vmPoolCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a VM pool",
	Long:  `Create a new VM pool in a VM cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return fmt.Errorf("failed to get CLI configuration: %w", err)
		}

		org, _ := cmd.Flags().GetString("org")
		if org == "" && cfg != nil && cfg.Org != "" {
			org = cfg.Org
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotctl configure')")
		}

		vmPoolName, _ := cmd.Flags().GetString("name")
		if vmPoolName == "" {
			vmPoolName = uuid.NewString()
		}
		vmCloudSpace, _ := cmd.Flags().GetString("vmcloudspace")
		serverClass, _ := cmd.Flags().GetString("serverclass")
		bidPrice, _ := cmd.Flags().GetString("bidprice")
		desired, _ := cmd.Flags().GetInt("desired")
		poolType, _ := cmd.Flags().GetString("pooltype")
		vmImage, _ := cmd.Flags().GetString("vmimage")

		validatedPrice, err := validateBidPrice(bidPrice)
		if err != nil {
			return fmt.Errorf("invalid bid price: %w", err)
		}

		validateDesired, err := validateDesiredCount(desired)
		if err != nil {
			return fmt.Errorf("invalid desired count: %w", err)
		}

		// Handle cloud-init user data
		vmUserData, _ := cmd.Flags().GetString("vm-userdata")
		vmUserDataFromScript, _ := cmd.Flags().GetString("vm-userdata-from-script")

		if vmUserData != "" && vmUserDataFromScript != "" {
			return fmt.Errorf("cannot specify both --vm-userdata and --vm-userdata-from-script")
		}

		var finalUserData string
		if vmUserDataFromScript != "" {
			finalUserData, err = rxtspot.PrepareUserDataFromScript(vmUserDataFromScript)
			if err != nil {
				return fmt.Errorf("failed to read user data script: %w", err)
			}
		} else if vmUserData != "" {
			finalUserData = rxtspot.PrepareUserData(vmUserData)
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		pool := rxtspot.VMPool{
			Name:         vmPoolName,
			VMCloudSpace: vmCloudSpace,
			ServerClass:  serverClass,
			BidPrice:     validatedPrice,
			Desired:      validateDesired,
			PoolType:     poolType,
			VMImage:      vmImage,
			VMUserData:   finalUserData,
		}

		if err := client.GetAPI().CreateVMPool(context.Background(), org, pool); err != nil {
			return fmt.Errorf("failed to create VM pool: %w", err)
		}

		fmt.Printf("\n%s Successfully created VM pool in VM cloudspace %s\n",
			color.GreenString("✓"),
			color.CyanString(vmCloudSpace),
		)
		fmt.Printf(" VMPool Name: %s\n", color.CyanString(vmPoolName))
		fmt.Printf("  Server Class: %s\n", color.CyanString(serverClass))
		fmt.Printf("  Bid Price: %s\n", color.CyanString(validatedPrice))
		fmt.Printf("  Desired: %s\n", color.CyanString(fmt.Sprintf("%d", validateDesired)))
		fmt.Printf("  Pool Type: %s\n", color.CyanString(poolType))
		fmt.Printf("  VM Image: %s\n", color.CyanString(vmImage))
		if finalUserData != "" {
			fmt.Printf("  VM UserData: %s (base64-encoded)\n", color.CyanString("provided"))
		}

		return nil
	},
}

var vmPoolGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get VM pool details",
	Long:  `Get details about a specific VM pool.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		org, _ := cmd.Flags().GetString("org")
		if org == "" && cfg != nil && cfg.Org != "" {
			org = cfg.Org
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotctl configure')")
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		pool, err := client.GetAPI().GetVMPool(context.Background(), org, name)
		if err != nil {
			if rxtspot.IsNotFound(err) {
				return fmt.Errorf("VM pool '%s' not found", name)
			}
			return fmt.Errorf("failed to get VM pool: %w", err)
		}

		return internal.OutputData(pool, outputFormat)
	},
}

var vmPoolUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a VM pool",
	Long:  `Update an existing VM pool's desired count or bid price.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return fmt.Errorf("failed to get config: %w", err)
		}

		org, _ := cmd.Flags().GetString("org")
		if org == "" && cfg != nil && cfg.Org != "" {
			org = cfg.Org
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotctl configure')")
		}

		desired, _ := cmd.Flags().GetInt("desired")
		bidPrice, _ := cmd.Flags().GetString("bidprice")

		if desired < 0 && bidPrice == "" {
			return fmt.Errorf("at least one of --desired or --bidprice must be provided")
		}

		pool := rxtspot.VMPool{
			Name: name,
		}

		if desired >= 0 {
			pool.Desired = desired
		}
		if bidPrice != "" {
			validatedPrice, err := validateBidPrice(bidPrice)
			if err != nil {
				return fmt.Errorf("invalid bid price: %w", err)
			}
			pool.BidPrice = validatedPrice
		}

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		if err := client.GetAPI().UpdateVMPool(context.Background(), org, pool); err != nil {
			return fmt.Errorf("failed to update VM pool: %w", err)
		}

		fmt.Printf("%s VM pool '%s' updated successfully\n", color.GreenString("✓"), name)
		return nil
	},
}

var vmPoolDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a VM pool",
	Long:  `Delete a VM pool.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}

		org, _ := cmd.Flags().GetString("org")
		if org == "" && cfg != nil && cfg.Org != "" {
			org = cfg.Org
		}
		if org == "" {
			return fmt.Errorf("organization not specified (use --org or run 'spotctl configure')")
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			prompt := color.New(color.FgYellow).PrintfFunc()
			prompt("Are you sure you want to delete VM pool '%s'? (y/N): ", name)

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

		if err := client.GetAPI().DeleteVMPool(context.Background(), org, name); err != nil {
			if rxtspot.IsNotFound(err) {
				return fmt.Errorf("VM pool '%s' not found", name)
			}
			return fmt.Errorf("failed to delete VM pool: %w", err)
		}

		fmt.Printf("VM pool '%s' deleted successfully\n", name)
		return nil
	},
}
