package cmd

import (
	"context"
	"fmt"

	"github.com/rackerlabs/spot-cli/internal"
	"github.com/spf13/cobra"
)

// organizationsCmd represents the organizations command
var organizationsCmd = &cobra.Command{
	Use:   "organizations",
	Short: "Manage organizations",
	Long:  `Manage Rackspace Spot organizations (namespaces).`,
}

// organizationsListCmd represents the organizations list command
var organizationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations",
	Long:  `List all organizations accessible by the authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		orgs, err := client.GetAPI().ListOrganizations(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list organizations: %w", err)
		}

		return internal.OutputData(orgs, outputFormat)
	},
}

// organizationsGetCmd represents the organizations get command
var organizationsGetCmd = &cobra.Command{
	Use:   "get <org>",
	Short: "Get organization details",
	Long:  `Get details for a specific organization by org.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := internal.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		org := args[0]
		orgs, err := client.GetAPI().ListOrganizations(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list organizations: %w", err)
		}

		// Find the organization with the matching org
		for _, organization := range orgs {
			if organization.Namespace == org {
				return internal.OutputData(organization, outputFormat)
			}
		}

		return fmt.Errorf("organization with org '%s' not found", org)
	},
}

func init() {
	rootCmd.AddCommand(organizationsCmd)
	organizationsCmd.AddCommand(organizationsListCmd)
	organizationsCmd.AddCommand(organizationsGetCmd)
}
