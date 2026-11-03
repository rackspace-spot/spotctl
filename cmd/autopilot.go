package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/google/uuid"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
)

// autopilotCmd represents the autopilot nodepools command
var autopilotCmd = &cobra.Command{
	Use:     "autopilot",
	Short:   "Manage autopilot node pools",
	Long:    `Manage autopilot node pools in cloudspaces. An autopilot pool takes a total vCPU target and an hourly budget ceiling, and Spot bids across eligible server classes to fill it.`,
	Aliases: []string{"ap"},
}

func init() {
	nodepoolsCmd.AddCommand(autopilotCmd)

	autopilotCmd.AddCommand(autopilotListCmd)
	autopilotCmd.AddCommand(autopilotGetCmd)
	autopilotCmd.AddCommand(autopilotCreateCmd)
	autopilotCmd.AddCommand(autopilotUpdateCmd)
	autopilotCmd.AddCommand(autopilotDeleteCmd)

	// Flags for autopilot list
	autopilotListCmd.Flags().String("org", "", "Organization name")
	autopilotListCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	autopilotListCmd.MarkFlagRequired("cloudspace")

	// Flags for autopilot get
	autopilotGetCmd.Flags().String("org", "", "Organization name")
	autopilotGetCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	autopilotGetCmd.MarkFlagRequired("name")

	// Flags for autopilot create
	autopilotCreateCmd.Flags().String("org", "", "Organization name")
	autopilotCreateCmd.Flags().String("cloudspace", "", "Cloudspace name (required)")
	autopilotCreateCmd.Flags().String("region", "", "Region (defaults to the cloudspace's region)")
	autopilotCreateCmd.Flags().String("vcpu-total", "", "Total vCPUs to fill. Must be a multiple of the smallest eligible node size (required)")
	autopilotCreateCmd.Flags().String("budget-per-hour", "", "Hourly budget ceiling in USD, e.g. 1.50 (required)")
	autopilotCreateCmd.Flags().String("vcpu-per-node-min", "", "Smallest node size to use, in vCPUs (optional)")
	autopilotCreateCmd.Flags().String("vcpu-per-node-max", "", "Largest node size to use, in vCPUs (optional)")
	autopilotCreateCmd.Flags().String("memory-per-vcpu", "", "Memory to vCPU ratio in GB: any, 2, 4 or 8 (default any)")
	autopilotCreateCmd.Flags().String("allocation-strategy", "", "How eligible server classes are ranked. Only LowestPrice is accepted (default LowestPrice)")
	autopilotCreateCmd.Flags().String("custom-labels", "", "Custom labels to be added on the autopilot nodepool. eg: --custom-labels key1=value1,key2=value2")
	autopilotCreateCmd.Flags().String("custom-annotations", "", "Custom annotations to be added to the autopilot nodepool. eg: --custom-annotations key1=value1,key2=value2")
	autopilotCreateCmd.MarkFlagRequired("cloudspace")
	autopilotCreateCmd.MarkFlagRequired("vcpu-total")
	autopilotCreateCmd.MarkFlagRequired("budget-per-hour")

	// Flags for autopilot update. Region and cloudspace are immutable, so they are not offered.
	autopilotUpdateCmd.Flags().String("org", "", "Organization name")
	autopilotUpdateCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	autopilotUpdateCmd.Flags().String("vcpu-total", "", "Total vCPUs to fill (optional)")
	autopilotUpdateCmd.Flags().String("budget-per-hour", "", "Hourly budget ceiling in USD (optional)")
	autopilotUpdateCmd.Flags().String("vcpu-per-node-min", "", "Smallest node size to use, in vCPUs (optional)")
	autopilotUpdateCmd.Flags().String("vcpu-per-node-max", "", "Largest node size to use, in vCPUs (optional)")
	autopilotUpdateCmd.Flags().String("memory-per-vcpu", "", "Memory to vCPU ratio in GB: any, 2, 4 or 8 (optional)")
	autopilotUpdateCmd.Flags().String("allocation-strategy", "", "How eligible server classes are ranked (optional)")
	autopilotUpdateCmd.Flags().String("custom-labels", "", "Custom labels to be added on the autopilot nodepool. eg: --custom-labels key1=value1,key2=value2")
	autopilotUpdateCmd.Flags().String("custom-annotations", "", "Custom annotations to be added to the autopilot nodepool. eg: --custom-annotations key1=value1,key2=value2")
	autopilotUpdateCmd.MarkFlagRequired("name")

	// Flags for autopilot delete
	autopilotDeleteCmd.Flags().String("org", "", "Organization name")
	autopilotDeleteCmd.Flags().String("name", "", "Node pool name (Note: It should be a valid lower case UUID) (required)")
	autopilotDeleteCmd.MarkFlagRequired("name")
	autopilotDeleteCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts; assume \"yes\" as answer")
}

// autopilotAPI resolves the organization (--org, then the configured default) and returns an
// authenticated API client for it.
func autopilotAPI(cmd *cobra.Command) (rxtspot.SpotAPI, string, error) {
	cfg, err := config.GetCLIEssentials(cmd)
	if err != nil {
		return nil, "", err
	}
	org, _ := cmd.Flags().GetString("org")
	if org == "" {
		org = cfg.Org
	}
	if org == "" {
		return nil, "", fmt.Errorf("organization not specified (use --org or run 'spotctl configure')")
	}
	client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("%w", err)
	}
	return client.GetAPI(), org, nil
}

// flagInt reads an integer flag, returning 0 when it was not supplied.
func flagInt(cmd *cobra.Command, name string) (int, error) {
	raw, _ := cmd.Flags().GetString(name)
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", name, err)
	}
	return v, nil
}

// autopilotListCmd represents the autopilot list command
var autopilotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List autopilot node pools",
	Long:  `List all autopilot node pools in a cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		api, org, err := autopilotAPI(cmd)
		if err != nil {
			return err
		}

		pools, err := api.ListAutopilotNodePools(context.Background(), org, cloudspace)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(pools, outputFormat)
	},
}

// autopilotGetCmd represents the autopilot get command
var autopilotGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an autopilot node pool",
	Long:  `Get a single autopilot node pool, including its per-server-class allocations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		api, org, err := autopilotAPI(cmd)
		if err != nil {
			return err
		}

		pool, err := api.GetAutopilotNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(pool, outputFormat)
	},
}

// autopilotCreateCmd represents the autopilot create command
var autopilotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an autopilot node pool",
	Long:  `Create a new autopilot node pool in a cloudspace.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := uuid.New().String()
		cloudspace, _ := cmd.Flags().GetString("cloudspace")
		budget, _ := cmd.Flags().GetString("budget-per-hour")
		vcpuTotal, err := flagInt(cmd, "vcpu-total")
		if err != nil {
			return err
		}
		if vcpuTotal <= 0 {
			return fmt.Errorf("vcpu-total must be greater than 0")
		}
		minPerNode, err := flagInt(cmd, "vcpu-per-node-min")
		if err != nil {
			return err
		}
		maxPerNode, err := flagInt(cmd, "vcpu-per-node-max")
		if err != nil {
			return err
		}
		customLabelsStr, _ := cmd.Flags().GetString("custom-labels")
		customLabels, err := parseCustomLabels(customLabelsStr)
		if err != nil {
			return fmt.Errorf("invalid custom-labels format: %w", err)
		}
		customAnnotationsStr, _ := cmd.Flags().GetString("custom-annotations")
		customAnnotations, err := parseCustomAnnotations(customAnnotationsStr)
		if err != nil {
			return fmt.Errorf("invalid custom-annotations format: %w", err)
		}

		api, org, err := autopilotAPI(cmd)
		if err != nil {
			return err
		}

		// The pool's region must match the cloudspace's, so default it from the cloudspace
		// rather than making the caller repeat it.
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			cs, err := api.GetCloudspace(context.Background(), org, cloudspace)
			if err != nil {
				return fmt.Errorf("could not read region from cloudspace %q (pass --region to set it explicitly): %w", cloudspace, err)
			}
			region = cs.Region
		}

		memoryPerVCPU, _ := cmd.Flags().GetString("memory-per-vcpu")
		allocationStrategy, _ := cmd.Flags().GetString("allocation-strategy")

		pool := rxtspot.AutopilotNodePool{
			Name:               name,
			Org:                org,
			Cloudspace:         cloudspace,
			Region:             region,
			VCPUTotal:          vcpuTotal,
			VCPUPerNodeMin:     minPerNode,
			VCPUPerNodeMax:     maxPerNode,
			MemoryPerVCPU:      memoryPerVCPU,
			BudgetPerHour:      budget,
			AllocationStrategy: allocationStrategy,
			CustomLabels:       customLabels,
			CustomAnnotations:  customAnnotations,
		}

		if err := api.CreateAutopilotNodePool(context.Background(), org, pool); err != nil {
			return fmt.Errorf("%w", err)
		}
		created, err := api.GetAutopilotNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("autopilot nodepool - %s created successfully \n", created.Name)

		return internal.OutputData(created, outputFormat)
	},
}

// autopilotUpdateCmd represents the autopilot update command
var autopilotUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an autopilot node pool",
	Long:  `Update an autopilot node pool. Region and cloudspace are immutable; everything else may change.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		api, org, err := autopilotAPI(cmd)
		if err != nil {
			return err
		}

		// Start from the live pool so a partial update does not blank the fields the caller
		// left out. The API server rejects an update that carries vcpu.total 0.
		pool, err := api.GetAutopilotNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if cmd.Flags().Changed("vcpu-total") {
			vcpuTotal, err := flagInt(cmd, "vcpu-total")
			if err != nil {
				return err
			}
			if vcpuTotal <= 0 {
				return fmt.Errorf("vcpu-total must be greater than 0")
			}
			pool.VCPUTotal = vcpuTotal
		}
		if cmd.Flags().Changed("vcpu-per-node-min") {
			if pool.VCPUPerNodeMin, err = flagInt(cmd, "vcpu-per-node-min"); err != nil {
				return err
			}
		}
		if cmd.Flags().Changed("vcpu-per-node-max") {
			if pool.VCPUPerNodeMax, err = flagInt(cmd, "vcpu-per-node-max"); err != nil {
				return err
			}
		}
		if cmd.Flags().Changed("budget-per-hour") {
			pool.BudgetPerHour, _ = cmd.Flags().GetString("budget-per-hour")
		}
		if cmd.Flags().Changed("memory-per-vcpu") {
			pool.MemoryPerVCPU, _ = cmd.Flags().GetString("memory-per-vcpu")
		}
		if cmd.Flags().Changed("allocation-strategy") {
			pool.AllocationStrategy, _ = cmd.Flags().GetString("allocation-strategy")
		}
		if cmd.Flags().Changed("custom-labels") {
			labelsStr, _ := cmd.Flags().GetString("custom-labels")
			if pool.CustomLabels, err = parseCustomLabels(labelsStr); err != nil {
				return fmt.Errorf("invalid custom-labels format: %w", err)
			}
		}
		if cmd.Flags().Changed("custom-annotations") {
			annotationsStr, _ := cmd.Flags().GetString("custom-annotations")
			if pool.CustomAnnotations, err = parseCustomAnnotations(annotationsStr); err != nil {
				return fmt.Errorf("invalid custom-annotations format: %w", err)
			}
		}

		if err := api.UpdateAutopilotNodePool(context.Background(), org, *pool); err != nil {
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("autopilot nodepool - %s updated successfully \n", pool.Name)

		updated, err := api.GetAutopilotNodePool(context.Background(), org, name)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(updated, outputFormat)
	},
}

// autopilotDeleteCmd represents the autopilot delete command
var autopilotDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an autopilot node pool",
	Long:  `Delete an autopilot node pool. Its child spot node pools are removed with it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			prompt := color.New(color.FgYellow).PrintfFunc()
			prompt("Are you sure you want to delete autopilot nodepool '%s'? (y/N): ", name)

			var response string
			if _, err := fmt.Scanln(&response); err != nil || (response != "y" && response != "Y") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		api, org, err := autopilotAPI(cmd)
		if err != nil {
			return err
		}

		if err := api.DeleteAutopilotNodePool(context.Background(), org, name); err != nil {
			return fmt.Errorf("%w", err)
		}
		fmt.Printf("autopilot node pool - %s deleted successfully \n", name)

		return nil
	},
}
