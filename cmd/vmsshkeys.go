package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	rxtspot "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rackspace-spot/spotctl/internal"
	config "github.com/rackspace-spot/spotctl/pkg"
	"github.com/spf13/cobra"
)

// vmSSHKeysCmd represents the vmsshkeys command
var vmSSHKeysCmd = &cobra.Command{
	Use:     "vmsshkeys",
	Short:   "Manage VM SSH keys",
	Long:    `Manage Rackspace Spot VM SSH keys.`,
	Aliases: []string{"vmsshkey", "vmsk"},
}

func init() {
	rootCmd.AddCommand(vmSSHKeysCmd)
	vmSSHKeysCmd.AddCommand(vmSSHKeyListCmd)
	vmSSHKeysCmd.AddCommand(vmSSHKeyCreateCmd)
	vmSSHKeysCmd.AddCommand(vmSSHKeyGetCmd)
	vmSSHKeysCmd.AddCommand(vmSSHKeyDeleteCmd)

	// Flags for vmsshkeys list
	vmSSHKeyListCmd.Flags().String("org", "", "Organization name")

	// Flags for vmsshkeys create
	vmSSHKeyCreateCmd.Flags().String("name", "", "SSH key name (required)")
	vmSSHKeyCreateCmd.Flags().String("org", "", "Organization name")
	vmSSHKeyCreateCmd.Flags().String("public-key", "", "SSH public key (required)")
	vmSSHKeyCreateCmd.Flags().String("description", "", "Description of the SSH key")
	vmSSHKeyCreateCmd.MarkFlagRequired("name")
	vmSSHKeyCreateCmd.MarkFlagRequired("public-key")

	// Flags for vmsshkeys get
	vmSSHKeyGetCmd.Flags().String("name", "", "SSH key name (required)")
	vmSSHKeyGetCmd.Flags().String("org", "", "Organization name")
	vmSSHKeyGetCmd.MarkFlagRequired("name")

	// Flags for vmsshkeys delete
	vmSSHKeyDeleteCmd.Flags().String("name", "", "SSH key name (required)")
	vmSSHKeyDeleteCmd.Flags().String("org", "", "Organization name")
	vmSSHKeyDeleteCmd.MarkFlagRequired("name")
	vmSSHKeyDeleteCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts")
}

var vmSSHKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List VM SSH keys",
	Long:  `List all VM SSH keys in an organization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("%w", err)
		}

		keys, err := client.GetAPI().ListVMSSHKeys(context.Background(), org)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(keys, outputFormat)
	},
}

var vmSSHKeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a VM SSH key",
	Long:  `Create a new VM SSH key.`,
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

		name, _ := cmd.Flags().GetString("name")
		publicKey, _ := cmd.Flags().GetString("public-key")
		description, _ := cmd.Flags().GetString("description")

		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}

		key := rxtspot.VMSSHKey{
			Name:        name,
			Org:         org,
			PublicKey:   publicKey,
			Description: description,
		}

		if err := client.GetAPI().CreateVMSSHKey(context.Background(), key); err != nil {
			return fmt.Errorf("failed to create VM SSH key: %w", err)
		}

		fmt.Printf("\n%s Successfully created VM SSH key %s\n",
			color.GreenString("✓"),
			color.CyanString(name),
		)

		return nil
	},
}

var vmSSHKeyGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get VM SSH key details",
	Long:  `Get details about a specific VM SSH key.`,
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

		key, err := client.GetAPI().GetVMSSHKey(context.Background(), org, name)
		if err != nil {
			if rxtspot.IsNotFound(err) {
				return fmt.Errorf("VM SSH key '%s' not found", name)
			}
			return fmt.Errorf("failed to get VM SSH key: %w", err)
		}

		return internal.OutputData(key, outputFormat)
	},
}

var vmSSHKeyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a VM SSH key",
	Long:  `Delete a VM SSH key.`,
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
			prompt("Are you sure you want to delete VM SSH key '%s'? (y/N): ", name)

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

		if err := client.GetAPI().DeleteVMSSHKey(context.Background(), org, name); err != nil {
			if rxtspot.IsNotFound(err) {
				return fmt.Errorf("VM SSH key '%s' not found", name)
			}
			return fmt.Errorf("failed to delete VM SSH key: %w", err)
		}

		fmt.Printf("VM SSH key '%s' deleted successfully\n", name)
		return nil
	},
}
