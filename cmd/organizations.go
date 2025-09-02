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

// organizationsCmd represents the organizations command
var organizationsCmd = &cobra.Command{
	Use:     "organizations",
	Short:   "Manage organizations",
	Long:    `Manage Rackspace Spot organizations (namespaces).`,
	Aliases: []string{"org", "organization"},
}

// organizationsListCmd represents the organizations list command
var organizationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations",
	Long:  `List all organizations accessible by the authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		orgs, err := client.GetAPI().ListOrganizations(context.Background())
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return internal.OutputData(orgs, outputFormat)
	},
}

// organizationsGetCmd represents the organizations get command
var organizationsGetCmd = &cobra.Command{
	Use:   "get <org>",
	Short: "Get organization details",
	Long:  `Get details for a specific organization by org.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		orgName, _ := cmd.Flags().GetString("name")
		if orgName == "" {
			return fmt.Errorf("organization not specified")
		}
		orgs, err := client.GetAPI().ListOrganizations(context.Background())
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		// Find the organization with the matching org
		for _, organization := range orgs {
			if organization.Name == orgName {
				return internal.OutputData(organization, outputFormat)
			}
		}

		return fmt.Errorf("organization with org '%s' not found", orgName)
	},
}

var organizationsDeleteCmd = &cobra.Command{
	Use:   "delete <org>",
	Short: "Delete organization",
	Long:  `Delete a specific organization by org.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetCLIEssentials(cmd)
		if err != nil {
			return err
		}
		client, err := internal.NewClientWithTokens(cfg.RefreshToken, cfg.AccessToken)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		orgName, _ := cmd.Flags().GetString("name")
		if orgName == "" {
			return fmt.Errorf("organization not specified")
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			// Interactive prompt
			prompt := color.New(color.FgYellow).PrintfFunc()
			prompt("Are you sure you want to delete organization '%s'? (y/N): ", orgName)

			var response string
			_, err := fmt.Scanln(&response)
			if err != nil || (response != "y" && response != "Y") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		err = client.GetAPI().DeleteOrganization(context.Background(), orgName)
		if err != nil {
			if rxtspot.IsNotFound(err) {
				return fmt.Errorf("organization '%s' not found", orgName)
			}
			if rxtspot.IsForbidden(err) {
				return fmt.Errorf("forbidden: %w", err)
			}
			if rxtspot.IsConflict(err) {
				return fmt.Errorf("conflict: %w", err)
			}
			return fmt.Errorf("%w", err)
		}

		fmt.Printf("organization - %s deleted successfully \n", orgName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(organizationsCmd)
	organizationsCmd.AddCommand(organizationsListCmd)
	organizationsCmd.AddCommand(organizationsGetCmd)
	organizationsCmd.AddCommand(organizationsDeleteCmd)
	organizationsGetCmd.Flags().String("name", "", "Organization name (required)")
	organizationsDeleteCmd.Flags().String("name", "", "Organization name (required)")
	organizationsDeleteCmd.Flags().BoolP("yes", "y", false, "Automatic yes to prompts; assume \"yes\" as answer")

	organizationsGetCmd.MarkFlagRequired("name")
	organizationsDeleteCmd.MarkFlagRequired("name")
}
